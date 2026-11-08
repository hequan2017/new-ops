// Package service artsci：Allen-Cahn 一维求解器（需求 3.3-1，显式有限差分，周期边界）
// 方程 u_t = ε·u_xx + u - u³，初值 u0=cos(2πx)，x∈[0,1)。
// 稳定性口径：显式欧拉扩散项要求 dt ≤ dx²/(4ε)，本实现收敛到 40%（dt ≤ 0.4·dx²/ε），
// 配合反应项 |f'(u)|≤2（u∈[-1,1]）保证数值有界；能量 E=∫(ε/2·u_x²+¼(u²-1)²)dx 离散单调衰减。
package service

import (
	"fmt"
	"math"
)

// errCodeACParam 参数越界/类型错（artsci 段 2500-2599）
const errCodeACParam = 2508

// acFrame 演化帧
type acFrame struct {
	Step   int       `json:"step"`
	T      float64   `json:"t"`
	Values []float64 `json:"values"`
}

func init() {
	RegisterModule("allen-cahn", RunAllenCahn)
}

// acNum 取数字参数（缺省给默认；类型错报 2508）
func acNum(args map[string]any, key string, def float64) (float64, error) {
	v, ok := args[key]
	if !ok || v == nil {
		return def, nil
	}
	n, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("[%d]参数 %s 须为数字", errCodeACParam, key)
	}
	return n, nil
}

// acEnergy 离散能量（周期边界）
func acEnergy(u []float64, dx, eps float64) float64 {
	n := len(u)
	e := 0.0
	for i := 0; i < n; i++ {
		up := u[(i+1)%n]
		grad := (up - u[i]) / dx
		w := u[i]*u[i] - 1
		e += dx * (eps/2*grad*grad + 0.25*w*w)
	}
	return e
}

// acRound 幅值六位小数取整（压缩 JSON 体积）
func acRound(u []float64) []float64 {
	out := make([]float64, len(u))
	for i, v := range u {
		out[i] = math.Round(v*1e6) / 1e6
	}
	return out
}

// RunAllenCahn 求解入口（参数白名单校验→显式差分推进→帧序列与能量返回）
func RunAllenCahn(args map[string]any) (any, error) {
	eps, err := acNum(args, "epsilon", 0.02)
	if err != nil {
		return nil, err
	}
	if eps <= 0 || eps > 1 {
		return nil, fmt.Errorf("[%d]epsilon 须在 (0,1] 内", errCodeACParam)
	}
	gridNf, err := acNum(args, "gridN", 64)
	if err != nil {
		return nil, err
	}
	gridN := int(math.Round(gridNf))
	if gridN < 16 || gridN > 512 {
		return nil, fmt.Errorf("[%d]gridN 须在 [16,512] 内", errCodeACParam)
	}
	stepsF, err := acNum(args, "steps", 800)
	if err != nil {
		return nil, err
	}
	steps := int(math.Round(stepsF))
	if steps < 1 || steps > 2000 {
		return nil, fmt.Errorf("[%d]steps 须在 [1,2000] 内", errCodeACParam)
	}
	dx := 1.0 / float64(gridN)
	dtLimit := 0.4 * dx * dx / eps
	dt, err := acNum(args, "dt", dtLimit)
	if err != nil {
		return nil, err
	}
	if dt <= 0 || dt > dtLimit {
		return nil, fmt.Errorf("[%d]dt 须在 (0,%.6f] 内（扩散稳定性 40%% 收敛）", errCodeACParam, dtLimit)
	}
	snapF, err := acNum(args, "snapshotEvery", math.Max(1, float64(steps)/40))
	if err != nil {
		return nil, err
	}
	snapshotEvery := int(math.Round(snapF))
	if snapshotEvery < 1 || snapshotEvery > steps {
		return nil, fmt.Errorf("[%d]snapshotEvery 须在 [1,%d] 内", errCodeACParam, steps)
	}

	// 初值与推进（双缓冲，周期边界）
	u := make([]float64, gridN)
	for i := range u {
		u[i] = math.Cos(2 * math.Pi * float64(i) / float64(gridN))
	}
	next := make([]float64, gridN)
	frames := []acFrame{{Step: 0, T: 0, Values: acRound(u)}}
	energy0 := acEnergy(u, dx, eps)
	for n := 1; n <= steps; n++ {
		for i := 0; i < gridN; i++ {
			up := u[(i+1)%gridN]
			um := u[(i-1+gridN)%gridN]
			next[i] = u[i] + dt*(eps*(up-2*u[i]+um)/(dx*dx)+u[i]-u[i]*u[i]*u[i])
		}
		u, next = next, u
		if n%snapshotEvery == 0 || n == steps {
			frames = append(frames, acFrame{Step: n, T: math.Round(float64(n)*dt*1e6) / 1e6, Values: acRound(u)})
		}
	}
	energy1 := acEnergy(u, dx, eps)
	// 帧数瘦身（≤60 帧，等距抽稀）
	if len(frames) > 60 {
		thin := make([]acFrame, 0, 60)
		stride := float64(len(frames)-1) / 59
		for k := 0; k < 60; k++ {
			thin = append(thin, frames[int(math.Round(float64(k)*stride))])
		}
		frames = thin
	}
	return map[string]any{
		"epsilon":       eps,
		"gridN":         gridN,
		"dt":            dt,
		"steps":         steps,
		"initialEnergy": math.Round(energy0*1e6) / 1e6,
		"finalEnergy":   math.Round(energy1*1e6) / 1e6,
		"frames":        frames,
	}, nil
}
