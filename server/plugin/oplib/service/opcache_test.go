// Package service oplib 结果缓存单测（键规范化/命中与 TTL/LRU 驱逐/执行链）
package service

import (
	"strings"
	"testing"
	"time"
)

func TestCacheKey(t *testing.T) {
	// 键序无关
	a := cacheKey("text_clean", map[string]any{"text": "x", "mode": "on"})
	b := cacheKey("text_clean", map[string]any{"mode": "on", "text": "x"})
	if a != b {
		t.Fatal("键序不应影响指纹")
	}
	// 参数不同 → 键不同
	c := cacheKey("text_clean", map[string]any{"text": "y"})
	if a == c {
		t.Fatal("不同参数应不同键")
	}
	// 算子不同 → 键不同
	d := cacheKey("count_stats", map[string]any{"text": "x", "mode": "on"})
	if a == d {
		t.Fatal("不同算子应不同键")
	}
	if len(a) != 64 {
		t.Fatalf("应 SHA256 hex: %d", len(a))
	}
}

func TestOpCache_RW(t *testing.T) {
	c := &opCache{data: map[string]opCacheEntry{}, lastAccess: map[string]time.Time{}}
	if _, ok := c.get("k"); ok {
		t.Fatal("空缓存不应命中")
	}
	c.put("k", "v1")
	if v, ok := c.get("k"); !ok || v != "v1" {
		t.Fatalf("写入应命中: %v %v", v, ok)
	}
	// TTL 过期（直接改 expiresAt 模拟）
	c.data["k"] = opCacheEntry{result: "v1", expiresAt: time.Now().Add(-time.Second)}
	if _, ok := c.get("k"); ok {
		t.Fatal("过期不应命中")
	}
}

func TestOpCache_LRU(t *testing.T) {
	c := &opCache{data: map[string]opCacheEntry{}, lastAccess: map[string]time.Time{}}
	for i := 0; i < opCacheCapacity; i++ {
		c.put(string(rune('a'+i%26))+string(rune('a'+i/26))+string(rune(i)), i)
	}
	if len(c.data) != opCacheCapacity {
		t.Fatalf("容量内不应驱逐: %d", len(c.data))
	}
	// 再写一条 → 驱逐最久未访问
	c.put("newest", "n")
	if len(c.data) > opCacheCapacity {
		t.Fatalf("超容应驱逐: %d", len(c.data))
	}
	if _, ok := c.get("newest"); !ok {
		t.Fatal("最新写入应存活")
	}
}

// TestExecuteWithCache 执行链：首算未命中、再算命中、参数不同不命中
func TestExecuteWithCache(t *testing.T) {
	args := map[string]any{"text": "  hello  world  "}
	r1, err := executeWithCache("text_clean", args)
	if err != nil || r1.FromCache {
		t.Fatalf("首算应未命中: %+v %v", r1, err)
	}
	if r1.Result != "hello world" {
		t.Fatalf("结果异常: %v", r1.Result)
	}
	r2, err := executeWithCache("text_clean", args)
	if err != nil || !r2.FromCache {
		t.Fatalf("再算应命中: %+v %v", r2, err)
	}
	if r2.Result != "hello world" {
		t.Fatalf("缓存结果应一致: %v", r2.Result)
	}
	// 不同参数
	r3, err := executeWithCache("text_clean", map[string]any{"text": "other"})
	if err != nil || r3.FromCache {
		t.Fatalf("不同参数应未命中: %+v %v", r3, err)
	}
	// 失败不缓存
	if _, err := executeWithCache("regex_extract", map[string]any{"pattern": "([bad", "text": "x"}); err == nil {
		t.Fatal("失败应报错")
	}
	if _, ok := operatorCache.get(cacheKey("regex_extract", map[string]any{"pattern": "([bad", "text": "x"})); ok {
		t.Fatal("失败结果不应入缓存")
	}
	_ = strings.Contains
}
