// Package service artsci：DeepHPMs 简化层次预测模型（需求 3.3-4）
// 口径：两级层次——一级全局多项式最小二乘（正规方程）拟合主趋势，
// 二级对残差再拟合局部修正项；预测=一级+二级叠加。
// 训练 MSE 单调不增（二级最小二乘在训练集上不可能劣化一级）。
package service

import (
	"math"
)

func init() {
	RegisterModule("deep-hpms", RunDeepHPMs)
}

// hpmsMSE 均方误差
func hpmsMSE(m *polyModel, xs, ys []float64) float64 {
	s := 0.0
	for i := range xs {
		d := m.predict(xs[i]) - ys[i]
		s += d * d
	}
	return s / float64(len(xs))
}

// RunDeepHPMs 两级层次拟合与预测入口
// 参数：xs/ys 必填等长；degree∈[1,6] 默认 3；residualDegree∈[1,6] 默认 2；predictXs 可选（缺省即 xs）
func RunDeepHPMs(args map[string]any) (any, error) {
	xs, err := acFloatSlice(args, "xs", true)
	if err != nil {
		return nil, err
	}
	ys, err := acFloatSlice(args, "ys", true)
	if err != nil {
		return nil, err
	}
	if len(xs) != len(ys) || len(xs) == 0 {
		return nil, artErr(lsErrCode, "xs/ys 须等长非空")
	}
	degF, err := acNum(args, "degree", 3)
	if err != nil {
		return nil, err
	}
	degree := int(math.Round(degF))
	if degree < 1 || degree > 6 {
		return nil, artErr(lsErrCode, "degree 须在 [1,6] 内")
	}
	resDegF, err := acNum(args, "residualDegree", 2)
	if err != nil {
		return nil, err
	}
	resDegree := int(math.Round(resDegF))
	if resDegree < 1 || resDegree > 6 {
		return nil, artErr(lsErrCode, "residualDegree 须在 [1,6] 内")
	}
	predictXs, err := acFloatSlice(args, "predictXs", false)
	if err != nil {
		return nil, err
	}
	if len(predictXs) == 0 {
		predictXs = xs
	}

	// 一级：全局主趋势
	m1, err := fitPolyLeastSquares(xs, ys, degree)
	if err != nil {
		return nil, err
	}
	// 二级：残差局部修正
	residuals := make([]float64, len(xs))
	for i := range xs {
		residuals[i] = ys[i] - m1.predict(xs[i])
	}
	m2, err := fitPolyLeastSquares(xs, residuals, resDegree)
	if err != nil {
		return nil, err
	}
	// 指标
	mse1 := hpmsMSE(m1, xs, ys)
	mse2 := 0.0
	for i := range xs {
		d := m1.predict(xs[i]) + m2.predict(xs[i]) - ys[i]
		mse2 += d * d
	}
	mse2 /= float64(len(xs))
	mean := 0.0
	for _, y := range ys {
		mean += y
	}
	mean /= float64(len(ys))
	mseBase := 0.0
	for _, y := range ys {
		d := mean - y
		mseBase += d * d
	}
	mseBase /= float64(len(ys))

	preds := make([]float64, len(predictXs))
	for i, x := range predictXs {
		preds[i] = m1.predict(x) + m2.predict(x)
	}
	return map[string]any{
		"degree":         degree,
		"residualDegree": resDegree,
		"coeffs":         roundSlice6(m1.coeffs),
		"residualCoeffs": roundSlice6(m2.coeffs),
		"mseGlobal":      round6(mse1),
		"mse":            round6(mse2),
		"mseBaseline":    round6(mseBase),
		"predictXs":      predictXs,
		"predictions":    roundSlice6(preds),
	}, nil
}
