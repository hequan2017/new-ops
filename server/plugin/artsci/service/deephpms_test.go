// Package service artsci：DeepHPMs 层次预测单测（合成数据，无 DB）
package service

import (
	"math"
	"strings"
	"testing"
)

// hpmsSynth 合成数据：y=sin(2πx)+0.05·伪噪声（确定性偏移，非随机）
func hpmsSynth(n int) ([]float64, []float64) {
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n-1)
		noise := 0.05 * math.Sin(7.0*x+1.3) // 高频小扰动模拟噪声
		xs[i] = x
		ys[i] = math.Sin(2*math.Pi*x) + noise
	}
	return xs, ys
}

// TestRunDeepHPMs 拟合质量：两级 MSE 显著小于基线且不劣于一级
func TestRunDeepHPMs(t *testing.T) {
	xs, ys := hpmsSynth(60)
	res, err := RunDeepHPMs(map[string]any{"xs": toAnySlice(xs), "ys": toAnySlice(ys)})
	if err != nil {
		t.Fatalf("拟合应成功: %v", err)
	}
	m := res.(map[string]any)
	mse := m["mse"].(float64)
	mseGlobal := m["mseGlobal"].(float64)
	mseBase := m["mseBaseline"].(float64)
	if mse >= mseGlobal+1e-9 {
		t.Fatalf("残差修正不应劣化一级：mse=%v mseGlobal=%v", mse, mseGlobal)
	}
	if mse >= mseBase {
		t.Fatalf("拟合 MSE 应显著小于均值基线：mse=%v baseline=%v", mse, mseBase)
	}
	if mse > 0.05*mseBase {
		t.Fatalf("两级模型应解释绝大部分方差：mse=%v baseline=%v", mse, mseBase)
	}
	// 外推预测走通
	res2, err := RunDeepHPMs(map[string]any{
		"xs": toAnySlice(xs), "ys": toAnySlice(ys),
		"predictXs": toAnySlice([]float64{0.25, 0.5, 0.75}),
	})
	if err != nil {
		t.Fatalf("外推预测应成功: %v", err)
	}
	preds := res2.(map[string]any)["predictions"].([]float64)
	if len(preds) != 3 {
		t.Fatalf("预测数量不符: %d", len(preds))
	}
}

// TestRunDeepHPMsGuard 参数校验
func TestRunDeepHPMsGuard(t *testing.T) {
	xs, ys := hpmsSynth(20)
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"缺xs", map[string]any{"ys": toAnySlice(ys)}, "缺少必填数组参数: xs"},
		{"长度不等", map[string]any{"xs": toAnySlice(xs[:5]), "ys": toAnySlice(ys)}, "等长"},
		{"degree越界", map[string]any{"xs": toAnySlice(xs), "ys": toAnySlice(ys), "degree": 9.0}, "[2509]degree"},
		{"样本不足", map[string]any{"xs": toAnySlice(xs[:3]), "ys": toAnySlice(ys[:3]), "degree": 5.0}, "[2509]样本不足"},
		{"类型错", map[string]any{"xs": "bad", "ys": toAnySlice(ys)}, "须为数组"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RunDeepHPMs(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("期望含 %q，实际: %v", tc.want, err)
			}
		})
	}
}

// TestSolveLinear 消元器：奇异矩阵报错
func TestSolveLinear(t *testing.T) {
	_, err := solveLinear([][]float64{{1, 2}, {2, 4}}, []float64{1, 2})
	if err == nil {
		t.Fatal("奇异矩阵应报错")
	}
}

func toAnySlice(vs []float64) []any {
	out := make([]any, len(vs))
	for i, v := range vs {
		out[i] = v
	}
	return out
}
