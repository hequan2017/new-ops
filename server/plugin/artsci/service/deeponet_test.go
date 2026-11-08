// Package service artsci：DeepONet 算子逼近单测（以积分算子为已知真值，无 DB）
package service

import (
	"math"
	"strings"
	"testing"
)

// donetSensors 函数族 f_c(s)=c1·sin2πs+c2·cos2πs+c3·sin4πs+c4·cos4πs+c5·s+c6·s² 传感器采样
func donetSensors(c [6]float64, s int) []float64 {
	f := make([]float64, s)
	for i := range f {
		x := float64(i) / float64(s-1)
		f[i] = c[0]*math.Sin(2*math.Pi*x) + c[1]*math.Cos(2*math.Pi*x) +
			c[2]*math.Sin(4*math.Pi*x) + c[3]*math.Cos(4*math.Pi*x) + c[4]*x + c[5]*x*x
	}
	return f
}

// donetTrueIntegral 真值算子：G(f_c)(y)=∫₀ʸ f_c(s)ds（解析）
func donetTrueIntegral(c [6]float64, y float64) float64 {
	return c[0]*(1-math.Cos(2*math.Pi*y))/(2*math.Pi) + c[1]*math.Sin(2*math.Pi*y)/(2*math.Pi) +
		c[2]*(1-math.Cos(4*math.Pi*y))/(4*math.Pi) + c[3]*math.Sin(4*math.Pi*y)/(4*math.Pi) +
		c[4]*y*y/2 + c[5]*y*y*y/3
}

// donetCoeff(i) 确定性系数扫描（12 条训练函数，张成 6 个独立方向）
func donetCoeff(i int) [6]float64 {
	return [6]float64{
		0.6 * math.Sin(float64(i)+0.4),
		0.5*math.Cos(1.7*float64(i)) + 0.2,
		0.4 * math.Sin(2.3*float64(i)),
		0.3*math.Cos(0.9*float64(i)) + 0.1,
		0.4 - 0.05*float64(i),
		0.2*math.Sin(1.1*float64(i)) + 0.1,
	}
}

// TestRunDeepONet 积分算子拟合：训练/外推误差小
func TestRunDeepONet(t *testing.T) {
	const s, p, m = 32, 12, 12
	ys := make([]float64, p)
	for j := range ys {
		ys[j] = float64(j) / float64(p-1)
	}
	var funcs, targets [][]any
	for i := 0; i < m; i++ {
		c := donetCoeff(i)
		funcs = append(funcs, toAnySlice(donetSensors(c, s)))
		row := make([]any, p)
		for j, y := range ys {
			row[j] = donetTrueIntegral(c, y)
		}
		targets = append(targets, row)
	}
	res, err := RunDeepONet(map[string]any{
		"trainFunctions": funcs, "trainYs": toAnySlice(ys), "trainTargets": targets,
		"basisCount": 10.0,
	})
	if err != nil {
		t.Fatalf("训练应成功: %v", err)
	}
	mm := res.(map[string]any)
	if mm["trainMSE"].(float64) > 1e-5 { // 截断极限口径（p=10 时 ~2e-6）
		t.Fatalf("训练 MSE 应极小（模型族内可表示），实际 %v", mm["trainMSE"])
	}
	// 外推：两条未见过函数 + 8 个查询点，对比解析真值
	pys := make([]float64, 8)
	for j := range pys {
		pys[j] = 0.08 + 0.12*float64(j)
	}
	var pf, pt [][]any
	for _, c := range [][6]float64{{0.9, -0.4, 0.3, 0.2, 0.5, -0.2}, {-0.5, 0.7, -0.2, 0.4, -0.3, 0.3}} {
		pf = append(pf, toAnySlice(donetSensors(c, s)))
		row := make([]any, len(pys))
		for j, y := range pys {
			row[j] = donetTrueIntegral(c, y)
		}
		pt = append(pt, row)
	}
	res2, err := RunDeepONet(map[string]any{
		"trainFunctions": funcs, "trainYs": toAnySlice(ys), "trainTargets": targets,
		"basisCount": 10.0, "predictFunction": pf, "predictYs": toAnySlice(pys), "predictTargets": pt,
	})
	if err != nil {
		t.Fatalf("预测应成功: %v", err)
	}
	m2 := res2.(map[string]any)
	if m2["predictMSE"].(float64) > 1e-4 { // 截断极限口径
		t.Fatalf("外推 MSE 应小，实际 %v", m2["predictMSE"])
	}
}

// TestRunDeepONetGuard 参数校验
func TestRunDeepONetGuard(t *testing.T) {
	f := toAnySlice(donetSensors([6]float64{1, 0, 0, 0, 0, 0}, 16))
	good := map[string]any{
		"trainFunctions": [][]any{f, f},
		"trainYs":        toAnySlice([]float64{0.2, 0.6}),
		"trainTargets":   [][]any{{0.1, 0.3}, {0.1, 0.3}},
	}
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"缺训练函数", map[string]any{"trainYs": toAnySlice([]float64{0.2})}, "缺少必填二维数组参数"},
		{"basisCount越界", mergeArgs(good, "basisCount", 20.0), "[2509]basisCount"},
		{"目标行数不符", func() map[string]any {
			bad := map[string]any{}
			for k, v := range good {
				bad[k] = v
			}
			bad["trainTargets"] = [][]any{{0.1, 0.3}}
			return bad
		}(), "行数一致"},
		{"传感器数不足", func() map[string]any {
			bad := map[string]any{}
			for k, v := range good {
				bad[k] = v
			}
			bad["trainFunctions"] = [][]any{toAnySlice(donetSensors([6]float64{1, 0, 0, 0, 0, 0}, 4))}
			bad["trainTargets"] = [][]any{{0.1, 0.3}}
			return bad
		}(), "不得少于基阶数"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RunDeepONet(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("期望含 %q，实际: %v", tc.want, err)
			}
		})
	}
}

func mergeArgs(base map[string]any, k string, v any) map[string]any {
	out := map[string]any{}
	for key, val := range base {
		out[key] = val
	}
	out[k] = v
	return out
}

// TestLegendreP 递推正确性锚点：P0=1, P1=x, P2=(3x²-1)/2
func TestLegendreP(t *testing.T) {
	if LegendreP(0, 0.7) != 1 {
		t.Fatal("P0 应恒为 1")
	}
	if math.Abs(LegendreP(1, 0.7)-0.7) > 1e-12 {
		t.Fatal("P1 应为 x")
	}
	if math.Abs(LegendreP(2, 0.7)-(3*0.7*0.7-1)/2) > 1e-12 {
		t.Fatal("P2 应为 (3x²-1)/2")
	}
}
