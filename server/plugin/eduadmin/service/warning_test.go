// Package service eduadmin：办学预警单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestRunLevel 双向阈值判定
func TestRunLevel(t *testing.T) {
	cases := []struct {
		value, threshold float64
		higherBetter     bool
		want             string
	}{
		{0.7, 0.85, true, "高"},   // 完成率 82% < 90% 阈值
		{0.8, 0.85, true, "中"},   // 94%
		{0.9, 0.85, true, ""},    // 达标
		{0.95, 0.9, false, "中"},  // 越低越好：超阈 5% → 中
		{1.2, 0.9, false, "高"},   // 越低越好：超阈 33% → 高
		{0.5, 0, true, ""},       // 阈值未配置
	}
	for _, c := range cases {
		if got := runLevel(c.value, c.threshold, c.higherBetter); got != c.want {
			t.Fatalf("value=%v threshold=%v higher=%v 应 %s 得 %s", c.value, c.threshold, c.higherBetter, c.want, got)
		}
	}
}

// TestAdmissionStrategy 招生策略报告模板
func TestAdmissionStrategy(t *testing.T) {
	if s := admissionStrategy(nil); !strings.Contains(s, "达标") {
		t.Fatalf("无预警应达标口径: %s", s)
	}
	alerts := []model.EduRunAlert{
		{IndexName: "就业率", Value: 0.75, Threshold: 0.9},
		{IndexName: "招生完成率", Value: 0.8, Threshold: 0.95},
	}
	s := admissionStrategy(alerts)
	if !strings.Contains(s, "2 项") || !strings.Contains(s, "订单班") || !strings.Contains(s, "生源基地") {
		t.Fatalf("报告模板不符: %s", s)
	}
}
