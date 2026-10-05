// Package service 白泽监控：告警规则引擎与钉钉推送（M6）
// 评估时机：CollectAll 采集完成后统一评估（指标规则）+ 端口探活规则同轮执行；
// 触发去重：静默窗口内同规则不重复触发；恢复：未触发条件且存在未恢复事件时关闭。
package service

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
)

// evaluateMetric 规则评估（纯逻辑，可单测）：
// 最近 duration 个采样点全部满足比较条件 → 触发
func evaluateMetric(values []float64, operator string, threshold float64) bool {
	if len(values) == 0 {
		return false
	}
	for _, v := range values {
		ok := false
		switch operator {
		case model.OpGT:
			ok = v > threshold
		case model.OpLT:
			ok = v < threshold
		}
		if !ok {
			return false
		}
	}
	return true
}

// EvalAllRules 全量规则评估（CollectAll 末尾调用）
func (s *MonitorService) EvalAllRules() {
	var rules []model.MonitorAlertRule
	global.GVA_DB.Where("enabled = ?", true).Find(&rules)
	for i := range rules {
		switch rules[i].Type {
		case model.RuleTypeMetric:
			s.evalMetricRule(&rules[i])
		case model.RuleTypePort:
			s.evalPortRule(&rules[i])
		}
	}
}

// evalMetricRule 指标规则评估
func (s *MonitorService) evalMetricRule(rule *model.MonitorAlertRule) {
	dur := rule.Duration
	if dur <= 0 {
		dur = 1
	}
	var pts []model.MonitorMetric
	global.GVA_DB.Where("asset_id = ? AND name = ?", rule.AssetID, rule.MetricName).
		Order("ts DESC").Limit(dur).Find(&pts)
	if len(pts) < dur {
		return // 采样不足，不评估
	}
	vals := make([]float64, len(pts))
	for i, p := range pts {
		vals[i] = p.Value
	}
	if evaluateMetric(vals, rule.Operator, rule.Threshold) {
		s.fireAlert(rule, fmt.Sprintf("%s %s %.2f（当前 %.2f，连续 %d 次）",
			rule.MetricName, rule.Operator, rule.Threshold, vals[0], dur), vals[0])
	} else {
		s.resolveAlert(rule, fmt.Sprintf("%s 已恢复（当前 %.2f）", rule.MetricName, vals[0]))
	}
}

// evalPortRule 端口探活规则评估（TCP 可达，2s 超时）
func (s *MonitorService) evalPortRule(rule *model.MonitorAlertRule) {
	ip := ruleAssetIP(rule.AssetID)
	if ip == "" {
		return
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", rule.Port)), 2*time.Second)
	reachable := err == nil
	if conn != nil {
		_ = conn.Close()
	}
	// 摘要中主机名由 fireAlert 统一拼装
	if !reachable {
		s.fireAlert(rule, fmt.Sprintf("端口 %d 不可达", rule.Port), 0)
	} else {
		s.resolveAlert(rule, fmt.Sprintf("端口 %d 恢复可达", rule.Port))
	}
}

// fireAlert 触发告警（静默窗口去重 + 钉钉推送）
func (s *MonitorService) fireAlert(rule *model.MonitorAlertRule, summary string, value float64) {
	silence := time.Duration(rule.SilenceMin) * time.Minute
	if silence <= 0 {
		silence = 30 * time.Minute
	}
	var recent model.MonitorAlertEvent
	q := global.GVA_DB.Where("rule_id = ? AND fired_at > ?", rule.ID, time.Now().Add(-silence)).
		Order("id DESC").First(&recent)
	if q.Error == nil {
		return // 静默窗口内已触发过
	}
	ev := &model.MonitorAlertEvent{
		RuleID: rule.ID, AssetID: rule.AssetID,
		Summary: "[白泽告警] " + ruleAssetName(rule.AssetID) + " " + summary,
		Value:   value, FiredAt: time.Now(),
	}
	if err := global.GVA_DB.Create(ev).Error; err != nil {
		return
	}
	if rule.WebhookURL != "" && isHTTPURL(rule.WebhookURL) {
		if err := pushDingTalk(rule.WebhookURL, ev.Summary); err == nil {
			global.GVA_DB.Model(ev).Update("notified", true)
		}
	}
}

// resolveAlert 恢复告警（关闭该规则最近的未恢复事件）
func (s *MonitorService) resolveAlert(rule *model.MonitorAlertRule, summary string) {
	now := time.Now()
	global.GVA_DB.Model(&model.MonitorAlertEvent{}).
		Where("rule_id = ? AND resolved_at IS NULL", rule.ID).
		Update("resolved_at", &now)
	_ = summary // 恢复通知暂不推送（降噪），事件表可查恢复时间
}

// pushDingTalk 钉钉机器人文本消息（webhook 由规则配置，仅 https/http）
func pushDingTalk(webhook, text string) error {
	if !isHTTPURL(webhook) {
		return fmt.Errorf("webhook 必须为 http(s) 地址")
	}
	body := fmt.Sprintf(`{"msgtype":"text","text":{"content":%q}}`, text)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook 响应 %d", resp.StatusCode)
	}
	return nil
}

// isHTTPURL 仅允许 http/https（安全约束）
func isHTTPURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

// ruleAssetName 规则关联资产名（缓存友好取一次）
func ruleAssetName(assetID uint) string {
	var h assetModel.AssetHost
	if err := global.GVA_DB.Select("hostname", "ip").First(&h, assetID).Error; err != nil {
		return fmt.Sprintf("资产#%d", assetID)
	}
	return fmt.Sprintf("%s(%s)", h.Hostname, h.IP)
}

// ruleAssetIP 资产 IP
func ruleAssetIP(assetID uint) string {
	var h assetModel.AssetHost
	if err := global.GVA_DB.Select("ip").First(&h, assetID).Error; err != nil {
		return ""
	}
	return h.IP
}
