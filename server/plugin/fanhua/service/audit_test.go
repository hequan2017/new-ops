// Package service fanhua：价值观审查单测（纯函数，无 DB）
package service

import "testing"

// TestRiskLevel 三档判定
func TestRiskLevel(t *testing.T) {
	cases := []struct {
		name string
		res  healthResult
		want string
	}{
		{"拦截即高风险", healthResult{Blocked: true, Health: 70}, "高"},
		{"低分即高风险", healthResult{Health: 30}, "高"},
		{"有命中为中", healthResult{Health: 95, Hits: []healthHit{{Word: "傻子", Weight: 5}}}, "中"},
		{"低于 70 为中", healthResult{Health: 65}, "中"},
		{"无命中为低", healthResult{Health: 100, Hits: []healthHit{}}, "低"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := riskLevel(c.res); got != c.want {
				t.Fatalf("期望 %s 实际 %s", c.want, got)
			}
		})
	}
}
