// Package service artsci：勒让德基与任意基最小二乘（M05 DeepONet branch/trunk 共用）
package service

// LegendreP 第 n 阶勒让德多项式（三项递推，P0=1, P1=x）
func LegendreP(n int, x float64) float64 {
	if n == 0 {
		return 1
	}
	p0, p1 := 1.0, x
	for k := 1; k < n; k++ {
		p0, p1 = p1, ((2*float64(k)+1)*x*p1-float64(k)*p0)/float64(k+1)
	}
	return p1
}

// legendreBasisRow 前 p 阶基在 t 处取值行
func legendreBasisRow(p int, t float64) []float64 {
	row := make([]float64, p)
	for k := 0; k < p; k++ {
		row[k] = LegendreP(k, t)
	}
	return row
}

// fitBasisCoeffs 任意基函数行最小二乘（AᵀA 正规方程，复用 solveLinear）
func fitBasisCoeffs(rows [][]float64, ys []float64) ([]float64, error) {
	if len(rows) == 0 || len(rows) != len(ys) {
		return nil, artErr(lsErrCode, "基行与目标等长非空")
	}
	m := len(rows[0])
	ata := make([][]float64, m)
	for i := range ata {
		ata[i] = make([]float64, m)
	}
	aty := make([]float64, m)
	for i, row := range rows {
		for j := 0; j < m; j++ {
			aty[j] += row[j] * ys[i]
			for k := 0; k < m; k++ {
				ata[j][k] += row[j] * row[k]
			}
		}
	}
	return solveLinear(ata, aty)
}

// fitLegendreCoeffs 传感器值→前 p 阶勒让德系数（均匀传感器、归一化域 [0,1]，branch 编码器）
func fitLegendreCoeffs(values []float64, p int) ([]float64, error) {
	s := len(values)
	if s < p {
		return nil, artErr(lsErrCode, "传感器数 %d 不得少于基阶数 %d", s, p)
	}
	rows := make([][]float64, s)
	for i := 0; i < s; i++ {
		rows[i] = legendreBasisRow(p, float64(i)/float64(s-1))
	}
	return fitBasisCoeffs(rows, values)
}
