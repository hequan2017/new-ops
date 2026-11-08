// Package service artsci：Allen-Cahn 求解器数值单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestRunAllenCahn 默认参数求解：帧序列有界、能量衰减、帧数收口
func TestRunAllenCahn(t *testing.T) {
	res, err := RunAllenCahn(map[string]any{})
	if err != nil {
		t.Fatalf("默认参数应成功: %v", err)
	}
	m, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("结果类型不符: %T", res)
	}
	frames := m["frames"].([]acFrame)
	if len(frames) < 2 || len(frames) > 60 {
		t.Fatalf("帧数应在 [2,60]，实际 %d", len(frames))
	}
	for _, f := range frames {
		if len(f.Values) != 64 {
			t.Fatalf("帧长应与 gridN 一致")
		}
		for _, v := range f.Values {
			if v < -1.5 || v > 1.5 {
				t.Fatalf("幅值应数值有界，step=%d 出现 %v", f.Step, v)
			}
		}
	}
	e0 := m["initialEnergy"].(float64)
	e1 := m["finalEnergy"].(float64)
	if e1 > e0+1e-9 {
		t.Fatalf("能量应单调衰减：initial=%v final=%v", e0, e1)
	}
	if e1 >= e0 {
		t.Fatalf("能量应显著衰减：initial=%v final=%v", e0, e1)
	}
}

// TestRunAllenCahnParamGuard 参数白名单校验
func TestRunAllenCahnParamGuard(t *testing.T) {
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"dt超稳定上限", map[string]any{"dt": 0.1}, "[2508]dt"},
		{"gridN过小", map[string]any{"gridN": 8.0}, "[2508]gridN"},
		{"gridN过大", map[string]any{"gridN": 600.0}, "[2508]gridN"},
		{"steps越界", map[string]any{"steps": 0.0}, "[2508]steps"},
		{"epsilon类型错", map[string]any{"epsilon": "x"}, "epsilon 须为数字"},
		{"snapshotEvery越界", map[string]any{"steps": 10.0, "snapshotEvery": 99.0}, "[2508]snapshotEvery"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RunAllenCahn(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("期望含 %q，实际: %v", tc.want, err)
			}
		})
	}
}

// TestRunAllenCahnCustom 自定义参数走通（gridN=32 小网格快速演化）
func TestRunAllenCahnCustom(t *testing.T) {
	res, err := RunAllenCahn(map[string]any{"gridN": 32.0, "steps": 50.0})
	if err != nil {
		t.Fatalf("自定义参数应成功: %v", err)
	}
	m := res.(map[string]any)
	if m["gridN"].(int) != 32 || m["steps"].(int) != 50 {
		t.Fatalf("回显参数不符: %v", m)
	}
	if len(m["frames"].([]acFrame)) < 2 {
		t.Fatal("至少含初值帧与终帧")
	}
}
