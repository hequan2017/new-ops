// Package service llmops：智能体调测（要求(5) 调测阶段）——工具回路接入的调试对话 + 会话历史
// 口径：debugChat 与普通 chat 的差别——非 qa 模式走 E03 工具回路（注册表内全部工具可用），
// trace 落消息 TraceJSON；qa 模式复用普通链路。
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// collectAgentTools 注册表快照（调试会话可用工具面）
func collectAgentTools() map[string]AgentTool {
	toolRegMu.RLock()
	defer toolRegMu.RUnlock()
	out := make(map[string]AgentTool, len(toolRegistry))
	for k, v := range toolRegistry {
		out[k] = v
	}
	return out
}

// AgentDebugChat 调试对话：非 qa 模式走工具回路，trace 留档
func (s *LlmopsServiceStruct) AgentDebugChat(sessionId uint, userMsg string) (string, *AgentPlan, []ToolTrace, error) {
	var sess model.LlmopsAgentSession
	if err := global.GVA_DB.First(&sess, sessionId).Error; err != nil {
		return "", nil, nil, fmt.Errorf("[%d]会话不存在", ErrCodeSessionMiss)
	}
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, sess.AgentId).Error; err != nil {
		return "", nil, nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	history, err := s.GetSessionMessages(sessionId, agentHistoryWindow)
	if err != nil {
		return "", nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	chatFn := func(_ string, msgs []ChatMessage) (string, error) {
		content, _, err := s.ChatCompletion(ctx, ag.ModelId, msgs, nil)
		return content, err
	}

	var reply string
	var plan *AgentPlan
	var traces []ToolTrace
	if ag.Mode == "qa" {
		reply, plan, err = runAgentTurn(chatFn, "qa", ag.SysPrompt, historyMessages(history), userMsg, ag.MaxTurns)
	} else {
		// 先拆解（容忍失败——调试模式下拆解失败继续走工具回路）
		_, plan, _ = runAgentTurn(chatFn, "workflow", ag.SysPrompt, nil, userMsg, 1)
		msgs := make([]ChatMessage, 0, len(history)+2)
		if ag.SysPrompt != "" {
			msgs = append(msgs, ChatMessage{Role: "system", Content: ag.SysPrompt})
		}
		msgs = append(msgs, historyMessages(history)...)
		msgs = append(msgs, ChatMessage{Role: "user", Content: userMsg})
		loopFn := func(_ string, loopMsgs []ChatMessage) (string, error) {
			return chatFn("", loopMsgs)
		}
		reply, traces, err = runToolLoop(loopFn, collectAgentTools(), userMsg, ag.MaxTurns)
	}
	if err != nil {
		return "", nil, nil, err
	}
	planJSON := ""
	if plan != nil {
		b, _ := json.Marshal(plan)
		planJSON = string(b)
	}
	traceJSON := ""
	if len(traces) > 0 {
		b, _ := json.Marshal(traces)
		traceJSON = string(b)
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.LlmopsAgentMessage{SessionId: sessionId, Role: "user", Content: userMsg}).Error; err != nil {
			return err
		}
		return tx.Create(&model.LlmopsAgentMessage{SessionId: sessionId, Role: "assistant", Content: reply, PlanJSON: planJSON, TraceJSON: traceJSON}).Error
	})
	if err != nil {
		return "", nil, nil, err
	}
	return reply, plan, traces, nil
}

// ListAgentSessions 会话列表（调试历史入口）
func (s *LlmopsServiceStruct) ListAgentSessions(agentId uint, userKey string, page, pageSize int) ([]model.LlmopsAgentSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsAgentSession{}).Where("agent_id = ?", agentId)
	if userKey != "" {
		db = db.Where("user_key = ?", userKey)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsAgentSession
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
