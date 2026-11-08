// Package service artsci：DeepONet 简化算子逼近（需求 3.3-6）
// 口径：branch=输入函数传感器值的勒让德投影系数（固定线性编码器），
// trunk=查询点处前 p 阶勒让德基，输出 G(f)(y)≈Σ_k b_k·a_k(f)·P_k(y)；
// 组合系数 b 对训练三元组线性，最小二乘闭式拟合（复用 fitBasisCoeffs）。
package service

import (
	"fmt"
	"math"
)

func init() {
	RegisterModule("deep-onet", RunDeepONet)
}

// acMatrix 取二维数字数组参数（JSON [][]number → [][]float64）
func acMatrix(args map[string]any, key string, required bool) ([][]float64, error) {
	v, ok := args[key]
	if !ok || v == nil {
		if required {
			return nil, fmt.Errorf("[%d]缺少必填二维数组参数: %s", lsErrCode, key)
		}
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		// 直传 [][]any（JSON 解码产物为 []any，元素 []any；测试可直传强类型）
		if m2, ok2 := v.([][]any); ok2 {
			out := make([][]float64, 0, len(m2))
			for _, row := range m2 {
				r, err := acFloatSlice(map[string]any{"k": row}, "k", true)
				if err != nil {
					return nil, fmt.Errorf("[%d]参数 %s 须为数字二维数组", lsErrCode, key)
				}
				out = append(out, r)
			}
			return out, nil
		}
		return nil, fmt.Errorf("[%d]参数 %s 须为二维数组", lsErrCode, key)
	}
	out := make([][]float64, 0, len(arr))
	for _, e := range arr {
		row, err := acFloatSlice(map[string]any{"k": e}, "k", true)
		if err != nil {
			return nil, fmt.Errorf("[%d]参数 %s 须为数字二维数组", lsErrCode, key)
		}
		out = append(out, row)
	}
	return out, nil
}

// RunDeepONet 算子逼近入口：训练三元组拟合 + 可选预测
// 参数：trainFunctions(M×s 传感器值)/trainYs(P 查询点)/trainTargets(M×P 算子真值) 必填；
// basisCount∈[2,12] 默认 6；predictFunction/predictYs/predictTargets 可选
func RunDeepONet(args map[string]any) (any, error) {
	trainFunctions, err := acMatrix(args, "trainFunctions", true)
	if err != nil {
		return nil, err
	}
	trainYs, err := acFloatSlice(args, "trainYs", true)
	if err != nil {
		return nil, err
	}
	trainTargets, err := acMatrix(args, "trainTargets", true)
	if err != nil {
		return nil, err
	}
	basisF, err := acNum(args, "basisCount", 6)
	if err != nil {
		return nil, err
	}
	basisCount := int(math.Round(basisF))
	if basisCount < 2 || basisCount > 12 {
		return nil, fmt.Errorf("[%d]basisCount 须在 [2,12] 内", lsErrCode)
	}
	m := len(trainFunctions)
	if m == 0 || len(trainTargets) != m || len(trainYs) == 0 {
		return nil, fmt.Errorf("[%d]训练函数/目标行数一致且查询点非空", lsErrCode)
	}
	sensorN := len(trainFunctions[0])
	for i := range trainFunctions {
		if len(trainFunctions[i]) != sensorN || len(trainTargets[i]) != len(trainYs) {
			return nil, fmt.Errorf("[%d]传感器列数一致且目标与查询点等长", lsErrCode)
		}
	}

	// branch 编码：每条训练函数投影到勒让德系数
	branch := make([][]float64, m)
	for i := 0; i < m; i++ {
		a, err := fitLegendreCoeffs(trainFunctions[i], basisCount)
		if err != nil {
			return nil, err
		}
		branch[i] = a
	}
	// 组合映射 B(p×p)：行=(i,j) 特征 z_{j'p+k}=P_{j'}(y)·a_k(f_i)，目标对 B 线性
	rows := make([][]float64, 0, m*len(trainYs))
	targets := make([]float64, 0, m*len(trainYs))
	for i := 0; i < m; i++ {
		for j, y := range trainYs {
			feat := make([]float64, basisCount*basisCount)
			for jp := 0; jp < basisCount; jp++ {
				Py := LegendreP(jp, y)
				for k := 0; k < basisCount; k++ {
					feat[jp*basisCount+k] = Py * branch[i][k]
				}
			}
			rows = append(rows, feat)
			targets = append(targets, trainTargets[i][j])
		}
	}
	z, err := fitBasisCoeffs(rows, targets)
	if err != nil {
		return nil, err
	}
	bm := make([][]float64, basisCount)
	for jp := range bm {
		bm[jp] = z[jp*basisCount : (jp+1)*basisCount]
	}

	// 训练集指标
	predict := func(a []float64, y float64) float64 {
		s := 0.0
		for jp := 0; jp < basisCount; jp++ {
			inner := 0.0
			for k := 0; k < basisCount; k++ {
				inner += bm[jp][k] * a[k]
			}
			s += LegendreP(jp, y) * inner
		}
		return s
	}
	trainMSE := 0.0
	for i := 0; i < m; i++ {
		for j, y := range trainYs {
			d := predict(branch[i], y) - trainTargets[i][j]
			trainMSE += d * d
		}
	}
	trainMSE /= float64(m * len(trainYs))

	out := map[string]any{
		"basisCount": basisCount,
		"sensorN":    sensorN,
		"trainMSE":   round6(trainMSE),
	}
	// 可选预测
	pf, err := acMatrix(args, "predictFunction", false)
	if err != nil {
		return nil, err
	}
	if len(pf) > 0 {
		pys, err := acFloatSlice(args, "predictYs", false)
		if err != nil {
			return nil, err
		}
		if len(pys) == 0 {
			pys = append([]float64{}, trainYs...)
		}
		all := make([][]float64, 0, len(pf))
		for _, f := range pf {
			a, err := fitLegendreCoeffs(f, basisCount)
			if err != nil {
				return nil, err
			}
			row := make([]float64, len(pys))
			for j, y := range pys {
				row[j] = predict(a, y)
			}
			all = append(all, row)
		}
		out["predictYs"] = pys
		out["predictions"] = all
		if pt, err := acMatrix(args, "predictTargets", false); err != nil {
			return nil, err
		} else if len(pt) == len(pf) {
			mse := 0.0
			cnt := 0
			for i := range pf {
				if len(pt[i]) != len(pys) {
					return nil, fmt.Errorf("[%d]predictTargets 与查询点等长", lsErrCode)
				}
				for j := range pys {
					d := all[i][j] - pt[i][j]
					mse += d * d
					cnt++
				}
			}
			out["predictMSE"] = round6(mse / float64(cnt))
		}
	}
	return out, nil
}
