// Package service llmops：智能体记忆（要求(5) 长期与短期记忆）
// 口径：短期=近 N 轮消息窗口截断（E02 已用固定 10 轮，此处统一纯函数口径）；
// 长期=记忆条目表，检索复用 bigram+BM25 打分注入上下文；自动抽取经 LLM JSON 协议可选。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeMemoryMiss = 2272
	memoryMaxRunes    = 1024
)

// TruncateHistoryWindow 短期记忆窗口截断（保留最近 maxTurns*2 条，纯函数）
func TruncateHistoryWindow(msgs []ChatMessage, maxTurns int) []ChatMessage {
	if maxTurns <= 0 {
		maxTurns = 5
	}
	keep := maxTurns * 2
	if len(msgs) <= keep {
		return msgs
	}
	return msgs[len(msgs)-keep:]
}

// SaveMemory 保存长期记忆（内容限 1024 rune）
func (s *LlmopsServiceStruct) SaveMemory(agentId uint, userKey, content, source string) (*model.LlmopsAgentMemory, error) {
	c := strings.TrimSpace(content)
	if c == "" {
		return nil, fmt.Errorf("[%d]记忆内容不能为空", ErrCodeMemoryMiss)
	}
	if r := []rune(c); len(r) > memoryMaxRunes {
		c = string(r[:memoryMaxRunes])
	}
	if source != "manual" && source != "auto" {
		source = "manual"
	}
	m := &model.LlmopsAgentMemory{AgentId: agentId, UserKey: strings.TrimSpace(userKey), Content: c, Source: source, Weight: 1}
	if err := global.GVA_DB.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// ListMemories 记忆列表（倒序限条）
func (s *LlmopsServiceStruct) ListMemories(agentId uint, userKey string, limit int) ([]model.LlmopsAgentMemory, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var list []model.LlmopsAgentMemory
	err := global.GVA_DB.Where("agent_id = ?", agentId).
		Where("user_key = ? OR user_key = ''", strings.TrimSpace(userKey)). // 用户专属 + 全局记忆
		Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}

// DeleteMemory 删除记忆
func (s *LlmopsServiceStruct) DeleteMemory(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsAgentMemory{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]记忆不存在", ErrCodeMemoryMiss)
	}
	return nil
}

// buildMemoryContext 记忆注入上下文块（纯函数，测试锚点）
func buildMemoryContext(hits []FtHit) string {
	if len(hits) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[长期记忆]\n")
	for i, h := range hits {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, h.Content)
	}
	return sb.String()
}

// InjectMemories 记忆检索注入：query 与记忆内容 bigram+BM25 打分 top-k
func (s *LlmopsServiceStruct) InjectMemories(agentId uint, userKey, query string, topK int) (string, error) {
	if topK <= 0 || topK > 10 {
		topK = 3
	}
	memories, err := s.ListMemories(agentId, userKey, 100)
	if err != nil {
		return "", err
	}
	if len(memories) == 0 {
		return "", nil
	}
	queryTerms := Bigrams(query)
	if len(queryTerms) == 0 {
		return "", nil
	}
	freqs := make([]map[string]int, len(memories))
	df := make(map[string]int, 32)
	var totalLen int
	for i := range memories {
		terms := Bigrams(memories[i].Content)
		freqs[i] = TermFreqs(terms)
		totalLen += len(terms)
		for t := range freqs[i] {
			df[t]++
		}
	}
	avgLen := float64(totalLen) / float64(len(memories))
	hits := make([]FtHit, 0, topK)
	for i := range memories {
		score := BM25Score(queryTerms, freqs[i], len(freqs[i]), avgLen, df, len(memories))
		if score <= 0 {
			continue
		}
		hit := FtHit{ChunkId: memories[i].ID, Content: memories[i].Content, Score: score}
		if len(hits) < topK {
			hits = append(hits, hit)
			if len(hits) == topK {
				sortFtHits(hits)
			}
			continue
		}
		if score > hits[topK-1].Score {
			hits[topK-1] = hit
			sortFtHits(hits)
		}
	}
	return buildMemoryContext(hits), nil
}

// ParseMemoryExtract 解析自动记忆抽取结果（纯函数）：{"memory":"..."} 或 {"memory":null}
func ParseMemoryExtract(text string) (string, bool) {
	obj, ok := extractJSONObject(text)
	if !ok {
		return "", false
	}
	var parsed struct {
		Memory *string `json:"memory"`
	}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil {
		return "", false
	}
	if parsed.Memory == nil {
		return "", false
	}
	c := strings.TrimSpace(*parsed.Memory)
	if c == "" {
		return "", false
	}
	return c, true
}

// ExtractAutoMemory 自动记忆抽取（LLM 判断对话中值得长期记住的事实）
func (s *LlmopsServiceStruct) ExtractAutoMemory(agentId uint, userKey string, userMsg, reply string, modelId uint) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	prompt := "从以下对话中提取值得长期记住的用户事实或偏好（如姓名/单位/习惯）。若无，输出 {\"memory\":null}；有则输出 {\"memory\":\"事实一句\"}。只输出 JSON。\n用户：" +
		userMsg + "\n助手：" + reply
	content, _, err := s.ChatCompletion(ctx, modelId, []ChatMessage{{Role: "user", Content: prompt}}, nil)
	if err != nil {
		return "", err
	}
	fact, ok := ParseMemoryExtract(content)
	if !ok {
		return "", nil
	}
	m, err := s.SaveMemory(agentId, userKey, fact, "auto")
	if err != nil {
		return "", err
	}
	return m.Content, nil
}
