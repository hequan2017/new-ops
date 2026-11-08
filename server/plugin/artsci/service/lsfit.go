// Package service artsci：多项式最小二乘底座（正规方程+高斯消元偏主元，M04/M05 共用）
// 口径：x 先 min-max 归一化到 [0,1] 再取幂基，控制正规方程条件数；奇异（pivot<1e-12）报错。
package service

import (
	"fmt"
	"math"
)

// lsErrCode 拟合类参数/矩阵错误码（artsci 段）
const lsErrCode = 2509

// polyModel 归一化多项式模型 y=Σ c_j·t^j，t=(x-xmin)/span
type polyModel struct {
	coeffs []float64
	xmin   float64
	span   float64
}

// predict 模型预测
func (m *polyModel) predict(x float64) float64 {
	t := (x - m.xmin) / m.span
	s := 0.0
	for j := len(m.coeffs) - 1; j >= 0; j-- {
		s = s*t + m.coeffs[j]
	}
	return s
}

// fitPolyLeastSquares 多项式最小二乘（正规方程；n ≥ degree+2 且 span>0）
func fitPolyLeastSquares(xs, ys []float64, degree int) (*polyModel, error) {
	n := len(xs)
	if n < degree+2 {
		return nil, fmt.Errorf("[%d]样本不足：n=%d 至少 degree+2=%d", lsErrCode, n, degree+2)
	}
	xmin, xmax := xs[0], xs[0]
	for _, x := range xs {
		if x < xmin {
			xmin = x
		}
		if x > xmax {
			xmax = x
		}
	}
	span := xmax - xmin
	if span <= 0 {
		return nil, fmt.Errorf("[%d]x 范围为零，无法归一化", lsErrCode)
	}
	m := degree + 1
	ata := make([][]float64, m)
	for i := range ata {
		ata[i] = make([]float64, m)
	}
	aty := make([]float64, m)
	for i := 0; i < n; i++ {
		t := (xs[i] - xmin) / span
		base := make([]float64, m)
		p := 1.0
		for j := 0; j < m; j++ {
			base[j] = p
			p *= t
		}
		for j := 0; j < m; j++ {
			aty[j] += base[j] * ys[i]
			for k := 0; k < m; k++ {
				ata[j][k] += base[j] * base[k]
			}
		}
	}
	coeffs, err := solveLinear(ata, aty)
	if err != nil {
		return nil, err
	}
	return &polyModel{coeffs: coeffs, xmin: xmin, span: span}, nil
}

// solveLinear 高斯消元（偏主元）解方阵线性组；奇异报错
func solveLinear(a [][]float64, b []float64) ([]float64, error) {
	m := len(b)
	ma := make([][]float64, m)
	for i := range ma {
		row := make([]float64, m+1)
		copy(row, a[i])
		row[m] = b[i]
		ma[i] = row
	}
	for col := 0; col < m; col++ {
		piv := col
		for r := col + 1; r < m; r++ {
			if math.Abs(ma[r][col]) > math.Abs(ma[piv][col]) {
				piv = r
			}
		}
		if math.Abs(ma[piv][col]) < 1e-12 {
			return nil, fmt.Errorf("[%d]正规方程奇异（degree 过高或数据退化）", lsErrCode)
		}
		ma[col], ma[piv] = ma[piv], ma[col]
		for r := col + 1; r < m; r++ {
			f := ma[r][col] / ma[col][col]
			for c := col; c <= m; c++ {
				ma[r][c] -= f * ma[col][c]
			}
		}
	}
	x := make([]float64, m)
	for i := m - 1; i >= 0; i-- {
		s := ma[i][m]
		for j := i + 1; j < m; j++ {
			s -= ma[i][j] * x[j]
		}
		x[i] = s / ma[i][i]
	}
	return x, nil
}

// acFloatSlice 取浮点数组参数（JSON 数组→[]float64，类型错报错）
func acFloatSlice(args map[string]any, key string, required bool) ([]float64, error) {
	v, ok := args[key]
	if !ok || v == nil {
		if required {
			return nil, fmt.Errorf("[%d]缺少必填数组参数: %s", lsErrCode, key)
		}
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("[%d]参数 %s 须为数组", lsErrCode, key)
	}
	out := make([]float64, 0, len(arr))
	for _, e := range arr {
		n, ok := e.(float64)
		if !ok {
			return nil, fmt.Errorf("[%d]参数 %s 须为数字数组", lsErrCode, key)
		}
		out = append(out, n)
	}
	return out, nil
}

// round6 六位小数取整
func round6(v float64) float64 { return math.Round(v*1e6) / 1e6 }

// roundSlice6 数组六位小数取整
func roundSlice6(vs []float64) []float64 {
	out := make([]float64, len(vs))
	for i, v := range vs {
		out[i] = round6(v)
	}
	return out
}
