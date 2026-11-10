// Package service eduadmin：人才预测单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestLinearRegression 正规方程：已知直线 y=2x+1
func TestLinearRegression(t *testing.T) {
	x := []float64{1, 2, 3, 4}
	y := []float64{3, 5, 7, 9} // y = 2x + 1
	slope, intercept, err := linearRegression(x, y)
	if err != nil {
		t.Fatalf("回归失败: %v", err)
	}
	if slope != 2 || intercept != 1 {
		t.Fatalf("应为 slope=2, intercept=1: %v %v", slope, intercept)
	}
	if _, _, err := linearRegression([]float64{1}, []float64{1}); err == nil {
		t.Fatal("单点应报错")
	}
	// 恒定 x 退化均值
	slope, intercept, _ = linearRegression([]float64{5, 5, 5}, []float64{1, 3, 5})
	if slope != 0 || intercept != 3 {
		t.Fatalf("恒定 x 应退化均值 3: %v %v", slope, intercept)
	}
}

// TestMovingAverage 移动平均
func TestMovingAverage(t *testing.T) {
	ma := movingAverage([]float64{0.1, 0.2, 0.3, 0.4}, 2)
	if len(ma) != 3 || ma[0] != 0.15 || ma[2] != 0.35 {
		t.Fatalf("窗口 2 应得 3 项: %+v", ma)
	}
	if movingAverage([]float64{1}, 3) != nil {
		t.Fatal("样本不足应返回空")
	}
}

// TestTalentSuggestions 培训建议规则
func TestTalentSuggestions(t *testing.T) {
	out := talentSuggestions(0.18, 0.2, 420, 400)
	joined := strings.Join(out, "|")
	if !strings.Contains(joined, "偏高") || !strings.Contains(joined, "净增 20") || !strings.Contains(joined, "增长超 10%") {
		t.Fatalf("高流失+高增长应三条建议: %s", joined)
	}
	out2 := talentSuggestions(0.08, 0.02, 390, 400)
	if len(out2) != 2 {
		t.Fatalf("正常流失+净减应两条(正常区间+净减): %+v", out2)
	}
	if !strings.Contains(out2[1], "净减 10") {
		t.Fatalf("净减建议不符: %+v", out2)
	}
	out3 := talentSuggestions(0, 0.05, 400, 400)
	if len(out3) != 1 || !strings.Contains(out3[0], "平稳") {
		t.Fatalf("平稳态应兜底建议: %+v", out3)
	}
}
