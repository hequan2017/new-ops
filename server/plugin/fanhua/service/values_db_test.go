// Package service fanhua：价值观知识库/审查 DB 路径单测（data.OpenMemory 内存库注入）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// TestValuesAuditServiceDB 敏感词→健康度→审查落日志→日志过滤（内存库）
func TestValuesAuditServiceDB(t *testing.T) {
	db, err := data.OpenMemory(&model.FanhuaSensitiveWord{}, &model.FanhuaValuesAuditLog{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	s := NewFanhuaService(db)

	// 落一个 block 词与一个 replace 词
	if _, err = s.SaveWord(&model.FanhuaSensitiveWord{Word: "赌博", Action: "block", Weight: 30, Enabled: true}); err != nil {
		t.Fatalf("SaveWord: %v", err)
	}
	if _, err = s.SaveWord(&model.FanhuaSensitiveWord{Word: "傻子", Action: "replace", Replacement: "反应较慢的人", Weight: 5, Enabled: true}); err != nil {
		t.Fatalf("SaveWord replace: %v", err)
	}
	// 健康度评估走库上词表
	res, err := s.GetHealth("参与赌博还骂人傻子")
	if err != nil || !res.Blocked || res.Health != 65 {
		t.Fatalf("健康度评估不符: err=%v %+v", err, res)
	}
	if !strings.Contains(res.Sanitized, "反应较慢的人") {
		t.Fatalf("替换输出不符: %s", res.Sanitized)
	}
	// 审查落日志（blocked → 高风险 + 预警）
	log, err := s.AuditValues("参与赌博是违法行为")
	if err != nil || !log.Blocked || log.RiskLevel != "高" || !log.Alarm {
		t.Fatalf("审查结果不符: err=%v %+v", err, log)
	}
	// 日志列表：预警过滤 + 等级过滤 + 非法等级 2629
	logs, total, err := s.GetValuesAuditLogs("高", true, 1, 20)
	if err != nil || total != 1 || len(logs) != 1 || logs[0].Text != "参与赌博是违法行为" {
		t.Fatalf("日志过滤不符: err=%v total=%d", err, total)
	}
	if _, _, err = s.GetValuesAuditLogs("严重", false, 1, 20); errx.CodeOf(err) != ErrCodeAuditParam {
		t.Fatalf("非法等级应 2629，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	// 删除不存在词 → 2620
	if err = s.DeleteWord(999); errx.CodeOf(err) != ErrCodeValuesNotFound {
		t.Fatalf("删除不存在词应 2620，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
}
