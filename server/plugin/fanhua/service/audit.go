// Package service fanhua：价值观审查 API（需求 3.4-21）
// 口径：中英输入 → 敏感词命中（N10 词表）→ 风险等级自动标注（高/中/低）+ 健康度评分 +
// 情感三分类 + 异常预警信号（blocked 或 health<40）；语义解析时延落日志（800ms 阈值，需求口径）。
package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const ErrCodeAuditParam = 2629

// auditLatencyThresholdMs 语义解析时延阈值（需求口径 <800ms）
const auditLatencyThresholdMs = 800

// riskLevel 风险等级判定（纯函数，测试锚点）
func riskLevel(res healthResult) string {
	switch {
	case res.Blocked || res.Health < 40:
		return "高"
	case res.Health < 70 || len(res.Hits) > 0:
		return "中"
	default:
		return "低"
	}
}

// AuditValues 价值观审查（词表→健康度→分级→预警→日志）
func (s *FanhuaService) AuditValues(text string) (*model.FanhuaValuesAuditLog, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errx.New(ErrCodeAuditParam, "审查文本必填")
	}
	words, err := s.repo.EnabledWords(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询敏感词失败", err)
	}
	started := time.Now()
	res := computeHealth(text, words)
	latency := time.Since(started).Milliseconds()
	level := riskLevel(res)
	log := &model.FanhuaValuesAuditLog{
		Text: text, RiskLevel: level, Health: res.Health, Sentiment: res.Sentiment,
		Blocked: res.Blocked, Alarm: res.Blocked || res.Health < 40,
		LatencyMs: latency, OverThreshold: latency > auditLatencyThresholdMs,
	}
	b, _ := json.Marshal(res.Hits)
	log.HitsJSON = string(b)
	if err := s.repo.CreateValuesAuditLog(s.dbh(), log); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "落审查日志失败", err)
	}
	return log, nil
}

// GetValuesAuditLogs 审查日志列表（等级/预警过滤，倒序分页）
func (s *FanhuaService) GetValuesAuditLogs(riskLevelFilter string, alarmOnly bool, page, pageSize int) ([]model.FanhuaValuesAuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	if riskLevelFilter != "" && riskLevelFilter != "高" && riskLevelFilter != "中" && riskLevelFilter != "低" {
		return nil, 0, errx.New(ErrCodeAuditParam, "风险等级须为 高/中/低")
	}
	list, total, err := s.repo.ValuesAuditLogList(s.dbh(), riskLevelFilter, alarmOnly, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询审查日志失败", err)
	}
	return list, total, nil
}
