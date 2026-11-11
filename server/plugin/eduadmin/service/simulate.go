// Package service eduadmin：政策模拟（需求 3.6-26）
// 口径：决策场景参数（预算增减率/费率调整幅度）→ 财务影响模拟——弹性系数模型
// （费率收入弹性 ΔRev=基期收入×Δ费率×弹性系数），风险三档（净影响<-5% 基期为高/
// <0 为中/≥0 为低）+ 对策建议；双方案对比输出。纯函数可测。
package service

import (
	"fmt"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduSimParam = 2827

// SimInput 决策场景参数
type SimInput struct {
	Name             string  `json:"name"`
	BaseBudget       float64 `json:"baseBudget"`       // 基期预算(万元)
	BaseRevenue      float64 `json:"baseRevenue"`      // 基期费率性收入(万元)
	BudgetChangeRate float64 `json:"budgetChangeRate"` // 预算增减率(-1~1)
	FeeRateChange    float64 `json:"feeRateChange"`    // 费率调整幅度(-0.5~0.5)
	Elasticity       float64 `json:"elasticity"`       // 收入弹性系数(默认 0.6)
}

// SimResult 模拟结果
type SimResult struct {
	Name         string  `json:"name"`
	NewBudget    float64 `json:"newBudget"`
	DeltaBudget  float64 `json:"deltaBudget"`
	DeltaRevenue float64 `json:"deltaRevenue"`
	Net          float64 `json:"net"`
	NetRate      float64 `json:"netRate"` // 净影响/基期预算
	RiskLevel    string  `json:"riskLevel"`
	Advice       string  `json:"advice"`
}

// simulateFinance 财务影响模拟（纯函数，测试锚点）
func simulateFinance(in SimInput) (*SimResult, error) {
	if in.BaseBudget <= 0 {
		return nil, fmt.Errorf("[%d]基期预算须为正数", ErrCodeEduSimParam)
	}
	if in.BudgetChangeRate < -1 || in.BudgetChangeRate > 1 {
		return nil, fmt.Errorf("[%d]预算增减率须在 [-1,1]", ErrCodeEduSimParam)
	}
	if in.FeeRateChange < -0.5 || in.FeeRateChange > 0.5 {
		return nil, fmt.Errorf("[%d]费率调整幅度须在 [-0.5,0.5]", ErrCodeEduSimParam)
	}
	elasticity := in.Elasticity
	if elasticity <= 0 {
		elasticity = 0.6
	}
	deltaBudget := round2(in.BaseBudget * in.BudgetChangeRate)
	deltaRevenue := round2(in.BaseRevenue * in.FeeRateChange * elasticity)
	net := round2(deltaBudget + deltaRevenue)
	netRate := round2(net / in.BaseBudget)
	level := "低"
	switch {
	case netRate < -0.05:
		level = "高"
	case net < 0:
		level = "中"
	}
	res := &SimResult{
		Name: in.Name, NewBudget: round2(in.BaseBudget + deltaBudget),
		DeltaBudget: deltaBudget, DeltaRevenue: deltaRevenue,
		Net: net, NetRate: netRate, RiskLevel: level,
	}
	switch level {
	case "高":
		res.Advice = "净影响超基期预算 5%：建议分期实施并设置年度复核点"
	case "中":
		res.Advice = "净影响为负：建议压缩非必要支出对冲缺口"
	default:
		res.Advice = "净影响可控：按计划执行并跟踪执行率"
	}
	return res, nil
}

// compareSims 方案对比（纯函数，测试锚点）
func compareSims(a, b *SimResult) string {
	d := round2(b.Net - a.Net)
	if d > 0 {
		return fmt.Sprintf("方案 B 净影响优 %+g 万元；A 风险 %s / B 风险 %s。", d, a.RiskLevel, b.RiskLevel)
	}
	if d < 0 {
		return fmt.Sprintf("方案 A 净影响优 %+g 万元；A 风险 %s / B 风险 %s。", -d, a.RiskLevel, b.RiskLevel)
	}
	return fmt.Sprintf("两方案净影响相同；A 风险 %s / B 风险 %s。", a.RiskLevel, b.RiskLevel)
}

// SimulateEduPolicy 单方案模拟
func (s *EduadminService) SimulateEduPolicy(in SimInput) (*SimResult, error) {
	return simulateFinance(in)
}

// CompareEduPolicySims 双方案模拟对比
func (s *EduadminService) CompareEduPolicySims(a, b SimInput) (*SimResult, *SimResult, string, error) {
	ra, err := simulateFinance(a)
	if err != nil {
		return nil, nil, "", err
	}
	rb, err := simulateFinance(b)
	if err != nil {
		return nil, nil, "", err
	}
	return ra, rb, compareSims(ra, rb), nil
}
