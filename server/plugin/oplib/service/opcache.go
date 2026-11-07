// Package service oplib：算子执行结果缓存（备用池 P04）
// 口径：进程内 LRU——键 = 算子名 + 规范化参数 SHA256；仅缓存 builtin 成功结果
// （script 有副作用不缓存，失败不缓存）；容量 256 条、TTL 10 分钟；命中返回缓存副本并标记 FromCache。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"
	"time"
)

// 缓存口径常量
const (
	opCacheCapacity = 256
	opCacheTTL      = 10 * time.Minute
)

// opCacheEntry 缓存条目
type opCacheEntry struct {
	result    any
	expiresAt time.Time
}

// opCache LRU 缓存体
type opCache struct {
	mu   sync.Mutex
	data map[string]opCacheEntry
	// 访问序（LRU 驱逐）：key → 最近访问
	lastAccess map[string]time.Time
}

var operatorCache = &opCache{data: map[string]opCacheEntry{}, lastAccess: map[string]time.Time{}}

// cacheKey 规范化参数指纹（纯函数，测试锚点）：map 键排序后序列化再哈希
func cacheKey(operator string, args map[string]any) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b, _ := json.Marshal(args)
	sum := sha256.Sum256(append([]byte(operator+"|"), b...))
	return hex.EncodeToString(sum[:])
}

// get 命中（含 TTL 与 LRU 访问序维护）
func (c *opCache) get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		delete(c.data, key)
		delete(c.lastAccess, key)
		return nil, false
	}
	c.lastAccess[key] = time.Now()
	return e.result, true
}

// put 写入（容量超限驱逐最久未访问）
func (c *opCache) put(key string, result any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = opCacheEntry{result: result, expiresAt: time.Now().Add(opCacheTTL)}
	c.lastAccess[key] = time.Now()
	if len(c.data) <= opCacheCapacity {
		return
	}
	// 找最久未访问驱逐
	var oldestKey string
	var oldest time.Time
	first := true
	for k, t := range c.lastAccess {
		if first || t.Before(oldest) {
			oldest, oldestKey, first = t, k, false
		}
	}
	delete(c.data, oldestKey)
	delete(c.lastAccess, oldestKey)
}

// CacheHit 缓存命中响应包装
type CacheHit struct {
	Result    any  `json:"result"`
	FromCache bool `json:"fromCache"`
}

// executeWithCache 带缓存的内置执行（RunOperator 的缓存包装面）
func executeWithCache(operator string, args map[string]any) (*CacheHit, error) {
	key := cacheKey(operator, args)
	if cached, ok := operatorCache.get(key); ok {
		return &CacheHit{Result: cached, FromCache: true}, nil
	}
	result, err := runBuiltin(operator, mustJSONArgs(args))
	if err != nil {
		return nil, err
	}
	operatorCache.put(key, result)
	return &CacheHit{Result: result, FromCache: false}, nil
}

// mustJSONArgs map → JSON 字符串（runBuiltin 入参口径）
func mustJSONArgs(args map[string]any) string {
	b, err := json.Marshal(args)
	if err != nil {
		return "{}"
	}
	return string(b)
}
