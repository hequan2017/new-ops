// Package service llmops：智能体监控（要求(5) 监控阶段）——时延分布 + 异常会话
// 口径：时延分布四桶（<0.5s/0.5-2s/2-10s/>10s）；异常会话=窗口内存在失败调用的去重会话（含失败次数）。
package service

import (
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// LatencyBuckets 时延分布（<500ms / 500-2000 / 2000-10000 / >10000）
func LatencyBuckets(latencies []int64) [4]int64 {
	var b [4]int64
	for _, ms := range latencies {
		switch {
		case ms < 500:
			b[0]++
		case ms < 2000:
			b[1]++
		case ms < 10000:
			b[2]++
		default:
			b[3]++
		}
	}
	return b
}

// AgentMonitorView 监控视图
type AgentMonitorView struct {
	Hours     int            `json:"hours"`
	Stat      AgentCallStat  `json:"stat"`
	Buckets   [4]int64       `json:"buckets"` // <0.5s/0.5-2s/2-10s/>10s
	AbnormalCount int64      `json:"abnormalCount"`
}

// GetAgentMonitor 监控聚合
func (s *LlmopsServiceStruct) GetAgentMonitor(agentId uint, hours int) (*AgentMonitorView, error) {
	if hours <= 0 || hours > 720 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var rows []model.LlmopsAgentCallLog
	if err := global.GVA_DB.Where("agent_id = ? AND created_at > ?", agentId, since).Find(&rows).Error; err != nil {
		return nil, err
	}
	latencies := make([]int64, 0, len(rows))
	for _, r := range rows {
		latencies = append(latencies, r.LatencyMs)
	}
	var abnormal int64
	global.GVA_DB.Model(&model.LlmopsAgentCallLog{}).
		Where("agent_id = ? AND created_at > ? AND ok = ?", agentId, since, false).
		Distinct("session_id").Count(&abnormal)
	return &AgentMonitorView{
		Hours: hours, Stat: summarizeCalls(rows), Buckets: LatencyBuckets(latencies), AbnormalCount: abnormal,
	}, nil
}

// AbnormalSession 异常会话项
type AbnormalSession struct {
	SessionId uint   `json:"sessionId"`
	Title     string `json:"title"`
	FailCount int64  `json:"failCount"`
	LastErr   string `json:"lastErr"`
}

// GetAgentAbnormalSessions 异常会话列表（窗口内存在失败调用的会话，按失败次数降序）
func (s *LlmopsServiceStruct) GetAgentAbnormalSessions(agentId uint, hours int, page, pageSize int) ([]AbnormalSession, int64, error) {
	if hours <= 0 || hours > 720 {
		hours = 24
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	// 失败聚合（按会话分组）
	type aggRow struct {
		SessionId uint
		FailCount int64
		LastErr   string
	}
	var aggs []aggRow
	err := global.GVA_DB.Model(&model.LlmopsAgentCallLog{}).
		Select("session_id, COUNT(*) as fail_count, MAX(err_msg) as last_err").
		Where("agent_id = ? AND created_at > ? AND ok = ?", agentId, since, false).
		Group("session_id").Order("fail_count desc").Find(&aggs).Error
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(aggs))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(aggs) {
		start = len(aggs)
	}
	if end > len(aggs) {
		end = len(aggs)
	}
	out := make([]AbnormalSession, 0, end-start)
	for _, a := range aggs[start:end] {
		var sess model.LlmopsAgentSession
		title := ""
		if err := global.GVA_DB.First(&sess, a.SessionId).Error; err == nil {
			title = sess.Title
		} else if err != gorm.ErrRecordNotFound {
			return nil, 0, err
		}
		out = append(out, AbnormalSession{SessionId: a.SessionId, Title: title, FailCount: a.FailCount, LastErr: a.LastErr})
	}
	return out, total, nil
}
