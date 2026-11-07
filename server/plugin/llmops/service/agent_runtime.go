// Package service llmops：智能体运行时（要求(5) 意图理解 + 任务自主拆解 + 会话记忆底座）
// 口径：拆解计划由 chat 模型输出结构化 JSON（容忍 markdown 围栏与前后缀散文，均衡截取），
// 解析失败带更严格指令重试一次；qa 模式直答，其余模式先拆解后作答；chat 调用注入可测。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeSessionMiss   = 2269
	ErrCodePlanParse     = 2270
	agentHistoryWindow   = 10
)

// AgentPlan 任务拆解计划
type AgentPlan struct {
	Intent string   `json:"intent"`
	Steps  []string `json:"steps"`
	Reply  string   `json:"reply"`
}

// buildPlanPrompt 拆解 prompt（纯函数，测试锚点）
func buildPlanPrompt(userMsg string) string {
	return "请先理解用户意图并拆解任务，只输出 JSON 对象（不要 markdown 代码块）：\n" +
		`{"intent":"用户意图一句话","steps":["步骤1","步骤2"],"reply":"可直接回复用户的内容（无后续步骤时给出答案）"}` +
		"\n用户输入：" + userMsg
}

// extractJSONObject 从文本中截取首个均衡 JSON 对象（容忍围栏/散文，纯函数）
func extractJSONObject(text string) (string, bool) {
	start := strings.Index(text, "{")
	if start < 0 {
		return "", false
	}
	depth := 0
	inStr := false
	escape := false
	for i := start; i < len(text); i++ {
		c := text[i]
		if escape {
			escape = false
			continue
		}
		switch c {
		case '\\':
			if inStr {
				escape = true
			}
		case '"':
			inStr = !inStr
		case '{':
			if !inStr {
				depth++
			}
		case '}':
			if !inStr {
				depth--
				if depth == 0 {
					return text[start : i+1], true
				}
			}
		}
	}
	return "", false
}

// ParseAgentPlan 解析拆解计划（直接解析 → 均衡截取兜底；intent 必填）
func ParseAgentPlan(text string) (*AgentPlan, error) {
	candidates := []string{strings.TrimSpace(text)}
	if obj, ok := extractJSONObject(text); ok {
		candidates = append([]string{obj}, candidates...)
	}
	for _, cand := range candidates {
		var p AgentPlan
		if err := json.Unmarshal([]byte(cand), &p); err == nil && strings.TrimSpace(p.Intent) != "" {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("[%d]任务拆解计划解析失败", ErrCodePlanParse)
}

// CreateSession 创建会话
func (s *LlmopsServiceStruct) CreateSession(agentId uint, userKey, title string) (*model.LlmopsAgentSession, error) {
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, agentId).Error; err != nil {
		return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	if strings.TrimSpace(title) == "" {
		title = "会话 " + timeNowShort()
	}
	sess := &model.LlmopsAgentSession{AgentId: agentId, Title: title, UserKey: strings.TrimSpace(userKey)}
	if err := global.GVA_DB.Create(sess).Error; err != nil {
		return nil, err
	}
	return sess, nil
}

// GetSessionMessages 会话消息（正序，近 N 条）
func (s *LlmopsServiceStruct) GetSessionMessages(sessionId uint, limit int) ([]model.LlmopsAgentMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var list []model.LlmopsAgentMessage
	err := global.GVA_DB.Where("session_id = ?", sessionId).Order("id desc").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, err
	}
	// 反转为正序
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// historyMessages 近 N 轮历史转 chat 消息
func historyMessages(msgs []model.LlmopsAgentMessage) []ChatMessage {
	out := make([]ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "user" || m.Role == "assistant" {
			out = append(out, ChatMessage{Role: m.Role, Content: m.Content})
		}
	}
	return out
}

// turnChatFunc 单轮对话调用注入点（prompt 与完整消息列表，独立于 D05 评测的单参 chatFunc）
type turnChatFunc func(prompt string, msgs []ChatMessage) (string, error)

// runAgentTurn 单轮运行核心（chatFn 注入，测试锚点）：
// qa 模式直答；其余模式先拆解（失败重试一次带更严格指令）后作答。
func runAgentTurn(fn turnChatFunc, mode, sysPrompt string, history []ChatMessage, userMsg string, maxTurns int) (string, *AgentPlan, error) {
	if mode == "qa" {
		msgs := make([]ChatMessage, 0, len(history)+2)
		if sysPrompt != "" {
			msgs = append(msgs, ChatMessage{Role: "system", Content: sysPrompt})
		}
		msgs = append(msgs, history...)
		msgs = append(msgs, ChatMessage{Role: "user", Content: userMsg})
		reply, err := fn(userMsg, msgs)
		return reply, nil, err
	}
	// 拆解阶段
	planCall := func(strict bool) (*AgentPlan, error) {
		prompt := buildPlanPrompt(userMsg)
		if strict {
			prompt += "\n注意：只输出纯 JSON 对象，禁止任何其它文字、注释或代码块围栏。"
		}
		msgs := []ChatMessage{{Role: "user", Content: prompt}}
		out, err := fn(prompt, msgs)
		if err != nil {
			return nil, err
		}
		return ParseAgentPlan(out)
	}
	plan, err := planCall(false)
	if err != nil {
		plan, err = planCall(true) // 重试一次
		if err != nil {
			return "", nil, err
		}
	}
	// 作答阶段（携带计划）
	answerPrompt := fmt.Sprintf("已理解意图：%s\n拆解步骤：%s\n请基于以上计划回答用户：%s",
		plan.Intent, strings.Join(plan.Steps, "；"), userMsg)
	msgs := make([]ChatMessage, 0, len(history)+2)
	if sysPrompt != "" {
		msgs = append(msgs, ChatMessage{Role: "system", Content: sysPrompt})
	}
	msgs = append(msgs, history...)
	msgs = append(msgs, ChatMessage{Role: "user", Content: answerPrompt})
	reply, err := fn(userMsg, msgs)
	if err != nil {
		return "", plan, err
	}
	_ = maxTurns // 轮次上限供 E03 工具回路使用
	return reply, plan, nil
}

// AgentChat 会话单轮（消息双写落库；返回回复与计划）
func (s *LlmopsServiceStruct) AgentChat(sessionId uint, userMsg string) (string, *AgentPlan, error) {
	var sess model.LlmopsAgentSession
	if err := global.GVA_DB.First(&sess, sessionId).Error; err != nil {
		return "", nil, fmt.Errorf("[%d]会话不存在", ErrCodeSessionMiss)
	}
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, sess.AgentId).Error; err != nil {
		return "", nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	history, err := s.GetSessionMessages(sessionId, agentHistoryWindow)
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		content, _, err := s.ChatCompletion(ctx, ag.ModelId, msgs, nil)
		return content, err
	}
	started := time.Now()
	reply, plan, err := runAgentTurn(fn, ag.Mode, ag.SysPrompt, historyMessages(history), userMsg, ag.MaxTurns)
	s.recordAgentCall(ag.ID, sessionId, started, err)
	if err != nil {
		return "", nil, err
	}
	// 消息双写
	planJSON := ""
	if plan != nil {
		b, _ := json.Marshal(plan)
		planJSON = string(b)
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.LlmopsAgentMessage{SessionId: sessionId, Role: "user", Content: userMsg}).Error; err != nil {
			return err
		}
		return tx.Create(&model.LlmopsAgentMessage{SessionId: sessionId, Role: "assistant", Content: reply, PlanJSON: planJSON}).Error
	})
	if err != nil {
		return "", nil, err
	}
	return reply, plan, nil
}

// timeNowShort 会话默认标题时间戳
func timeNowShort() string {
	return time.Now().Format("0102-1504")
}
