// Package service llmops：智能体发布/回滚与调用统计（要求(5) 发布与运营阶段）
// 口径：回滚=版本快照字段回写（仅回写快照内字段，状态保持 published）；
// 调用日志在 chat/debugChat 双入口落库；统计含 total/ok/fail/avg/p95（近 N 小时窗口）。
package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeAgentRollback = 2278

// snapshotToUpdates 快照→回写字段映射（纯函数，测试锚点）
func snapshotToUpdates(snapshot string) (map[string]any, error) {
	var snap map[string]any
	if err := json.Unmarshal([]byte(snapshot), &snap); err != nil {
		return nil, fmt.Errorf("[%d]版本快照解析失败: %w", ErrCodeAgentRollback, err)
	}
	updates := map[string]any{}
	for _, key := range []string{"name", "mode", "sysPrompt", "modelId", "knowledgeIds", "maxTurns", "configJSON"} {
		if v, ok := snap[key]; ok {
			updates[snapshotColumn(key)] = v
		}
	}
	return updates, nil
}

// snapshotColumn 快照键→列名
func snapshotColumn(key string) string {
	switch key {
	case "sysPrompt":
		return "sys_prompt"
	case "modelId":
		return "model_id"
	case "knowledgeIds":
		return "knowledge_ids"
	case "maxTurns":
		return "max_turns"
	case "configJSON":
		return "config_json"
	}
	return key
}

// RollbackAgent 回滚到指定版本（快照回写 + 状态保持 published）
func (s *LlmopsServiceStruct) RollbackAgent(agentId uint, version int) (*model.LlmopsAgent, error) {
	var ver model.LlmopsAgentVersion
	if err := global.GVA_DB.Where("agent_id = ? AND version = ?", agentId, version).First(&ver).Error; err != nil {
		return nil, fmt.Errorf("[%d]版本不存在", ErrCodeAgentRollback)
	}
	updates, err := snapshotToUpdates(ver.Snapshot)
	if err != nil {
		return nil, err
	}
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, agentId).Error; err != nil {
		return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	if err := global.GVA_DB.Model(&ag).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ag).Update("status", "published").Error; err != nil {
		return nil, err
	}
	ag.Status = "published"
	return &ag, nil
}

// recordAgentCall 智能体调用日志落库（chat/debugChat 双入口；失败仅记日志）
func (s *LlmopsServiceStruct) recordAgentCall(agentId, sessionId uint, started time.Time, callErr error) {
	log := model.LlmopsAgentCallLog{
		AgentId: agentId, SessionId: sessionId,
		LatencyMs: time.Since(started).Milliseconds(), Ok: callErr == nil,
	}
	if callErr != nil {
		e := truncateSample(callErr.Error())
		if len([]rune(e)) > 500 {
			e = string([]rune(e)[:500])
		}
		log.ErrMsg = e
	}
	if err := global.GVA_DB.Create(&log).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 智能体调用日志写入失败: %v", err))
	}
}

// AgentCallStat 调用统计
type AgentCallStat struct {
	Total      int   `json:"total"`
	OkCount    int   `json:"okCount"`
	FailCount  int   `json:"failCount"`
	AvgLatency int64 `json:"avgLatencyMs"`
	P95Latency int64 `json:"p95LatencyMs"`
}

// percentile95 P95 时延（纯函数：最近秩法 ceil(0.95n)-1，空返回 0）
func percentile95(latencies []int64) int64 {
	if len(latencies) == 0 {
		return 0
	}
	sorted := make([]int64, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Ceil(0.95*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

// summarizeCalls 统计聚合（纯函数，测试锚点）
func summarizeCalls(rows []model.LlmopsAgentCallLog) AgentCallStat {
	stat := AgentCallStat{Total: len(rows)}
	var sum int64
	var latencies []int64
	for _, r := range rows {
		if r.Ok {
			stat.OkCount++
		} else {
			stat.FailCount++
		}
		sum += r.LatencyMs
		latencies = append(latencies, r.LatencyMs)
	}
	if stat.Total > 0 {
		stat.AvgLatency = sum / int64(stat.Total)
	}
	stat.P95Latency = percentile95(latencies)
	return stat
}

// GetAgentCallStats 调用统计（近 hours 小时；0/越界默认 24）
func (s *LlmopsServiceStruct) GetAgentCallStats(agentId uint, hours int) (*AgentCallStat, error) {
	if hours <= 0 || hours > 720 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var rows []model.LlmopsAgentCallLog
	if err := global.GVA_DB.Where("agent_id = ? AND created_at > ?", agentId, since).Find(&rows).Error; err != nil {
		return nil, err
	}
	stat := summarizeCalls(rows)
	return &stat, nil
}
