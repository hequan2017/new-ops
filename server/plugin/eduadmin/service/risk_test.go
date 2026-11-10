// Package service eduadmin：财务风险预警单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestCashRiskLevel 现金流预警三档
func TestCashRiskLevel(t *testing.T) {
	cases := []struct {
		balance, floor float64
		want           string
	}{
		{50, 100, "高"},   // 0.5 < 1
		{110, 100, "中"},  // 1.1 < 1.2
		{120, 100, "低"},  // ≥1.2
		{100, 0, "低"},   // 下限未配置
	}
	for _, c := range cases {
		if got := cashRiskLevel(c.balance, c.floor); got != c.want {
			t.Fatalf("余额 %g/下限 %g 应 %s，得 %s", c.balance, c.floor, c.want, got)
		}
	}
}
