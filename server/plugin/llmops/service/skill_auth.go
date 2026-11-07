// Package service llmops：技能鉴权与限流（要求(6) 认证鉴权 + 流量控制 + 调用日志审计）
// 口径：令牌 SHA256 落库不可逆（明文仅创建时返回一次，prefix 识别）；限流为内存令牌桶
// （时钟注入可测）；开放调用入口 verify→限流→沙箱执行→审计留痕全链。
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeTokenBad    = 2292
	ErrCodeRateLimited = 2293
)

// GenerateTokenPlain 生成随机令牌（32 hex 字符）
func GenerateTokenPlain() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken 令牌哈希（SHA256 hex，校验口径）
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// VerifyTokenHash 常量时间比较（纯函数）
func VerifyTokenHash(plain, stored string) bool {
	return subtle.ConstantTimeCompare([]byte(HashToken(plain)), []byte(stored)) == 1
}

// CreateSkillToken 创建令牌（明文仅本次返回）
func (s *LlmopsServiceStruct) CreateSkillToken(name string, rateLimit int) (*model.LlmopsSkillToken, string, error) {
	if strings.TrimSpace(name) == "" {
		return nil, "", fmt.Errorf("[%d]令牌名不能为空", ErrCodeTokenBad)
	}
	if rateLimit <= 0 || rateLimit > 1000 {
		rateLimit = 10
	}
	var count int64
	global.GVA_DB.Model(&model.LlmopsSkillToken{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return nil, "", fmt.Errorf("[%d]令牌名已存在", ErrCodeTokenBad)
	}
	plain, err := GenerateTokenPlain()
	if err != nil {
		return nil, "", err
	}
	t := &model.LlmopsSkillToken{
		Name: name, TokenHash: HashToken(plain), Prefix: plain[:4], RateLimit: rateLimit, Status: "enabled",
	}
	if err := global.GVA_DB.Create(t).Error; err != nil {
		return nil, "", err
	}
	return t, plain, nil
}

// GetSkillTokens 令牌列表
func (s *LlmopsServiceStruct) GetSkillTokens() ([]model.LlmopsSkillToken, error) {
	var list []model.LlmopsSkillToken
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// RevokeSkillToken 吊销令牌
func (s *LlmopsServiceStruct) RevokeSkillToken(id uint) error {
	res := global.GVA_DB.Model(&model.LlmopsSkillToken{}).Where("id = ?", id).Update("status", "revoked")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]令牌不存在", ErrCodeTokenBad)
	}
	return nil
}

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens float64
	last   time.Time
}

// rateLimiter 内存令牌桶限流器（now 注入可测）
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	now     func() time.Time
}

var skillLimiter = &rateLimiter{buckets: map[string]*tokenBucket{}, now: time.Now}

// Allow 限流判定：桶初始满额 rate，按流逝时间补充
func (l *rateLimiter) Allow(key string, rate int) bool {
	if rate <= 0 {
		rate = 10
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &tokenBucket{tokens: float64(rate) - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * float64(rate) // 按速率补充
	if b.tokens > float64(rate) {
		b.tokens = float64(rate)
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// VerifySkillToken 令牌校验（存在+启用+哈希匹配）
func (s *LlmopsServiceStruct) VerifySkillToken(plain string) (*model.LlmopsSkillToken, error) {
	if len(plain) != 32 {
		return nil, fmt.Errorf("[%d]令牌格式非法", ErrCodeTokenBad)
	}
	var candidates []model.LlmopsSkillToken
	if err := global.GVA_DB.Where("status = 'enabled'").Find(&candidates).Error; err != nil {
		return nil, err
	}
	for i := range candidates {
		if VerifyTokenHash(plain, candidates[i].TokenHash) {
			return &candidates[i], nil
		}
	}
	return nil, fmt.Errorf("[%d]令牌无效或已吊销", ErrCodeTokenBad)
}

// SkillCallWithAuth 开放调用入口：校验→限流→沙箱执行→审计留痕
func (s *LlmopsServiceStruct) SkillCallWithAuth(tokenPlain, toolName, argsJSON string) (string, error) {
	token, err := s.VerifySkillToken(tokenPlain)
	if err != nil {
		return "", err
	}
	if !skillLimiter.Allow(fmt.Sprintf("token-%d", token.ID), token.RateLimit) {
		return "", fmt.Errorf("[%d]触发限流(%d/s)，请稍后重试", ErrCodeRateLimited, token.RateLimit)
	}
	tool, ok := GetAgentTool(toolName)
	if !ok {
		s.recordSkillCall(token.ID, toolName, SanitizeArgs(argsJSON, nil), false, 0, "工具不存在")
		return "", fmt.Errorf("[%d]工具不存在: %s", ErrCodeMcpNotFound, toolName)
	}
	started := time.Now()
	_, result, callErr := RunToolSandboxed(tool, argsJSON, SandboxConfig{})
	s.recordSkillCall(token.ID, toolName, SanitizeArgs(argsJSON, nil), callErr == nil,
		time.Since(started).Milliseconds(), errSummary(callErr))
	return result, callErr
}

func errSummary(err error) string {
	if err == nil {
		return ""
	}
	e := truncateSample(err.Error())
	if len([]rune(e)) > 500 {
		e = string([]rune(e)[:500])
	}
	return e
}

// recordSkillCall 审计留痕（失败仅记日志）
func (s *LlmopsServiceStruct) recordSkillCall(tokenId uint, skill, argsMasked string, ok bool, latencyMs int64, errMsg string) {
	if len([]rune(argsMasked)) > 2000 {
		argsMasked = string([]rune(argsMasked)[:2000])
	}
	if err := global.GVA_DB.Create(&model.LlmopsSkillCallLog{
		TokenId: tokenId, Skill: skill, ArgsMasked: argsMasked, Ok: ok, LatencyMs: latencyMs, ErrMsg: errMsg,
	}).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 技能审计日志写入失败: %v", err))
	}
}

// GetSkillCallLogs 审计日志列表
func (s *LlmopsServiceStruct) GetSkillCallLogs(page, pageSize int, tokenId uint) ([]model.LlmopsSkillCallLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsSkillCallLog{})
	if tokenId > 0 {
		db = db.Where("token_id = ?", tokenId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsSkillCallLog
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
