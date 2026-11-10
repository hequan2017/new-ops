// Package service eduadmin：智能预算（需求 3.6-15）
// 口径：同比/环比=逐年增长额与增长率（纯函数）；分配建议=上年占比缩放至新总额，
// 单部门下限 10% 后重归一（纯函数）；方案对比=逐部门差额+总额差额+集中度提示。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduBudgetParam = 2812

// budgetDelta 单年增长
type budgetDelta struct {
	Year   int     `json:"year"`
	Amount float64 `json:"amount"`
	Growth float64 `json:"growth"` // 同比增长率（首年 0）
}

// DeptAmount 部门金额（导出供 API 层绑定）
type DeptAmount struct {
	Dept   string  `json:"dept"`
	Amount float64 `json:"amount"`
	Ratio  float64 `json:"ratio"`
}

// deptAmount 内部别名
type deptAmount = DeptAmount

// budgetTrend 同比/环比趋势（纯函数，测试锚点）：按年升序
func budgetTrend(list []model.EduBudget) []budgetDelta {
	sorted := make([]model.EduBudget, len(list))
	copy(sorted, list)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Year < sorted[j].Year })
	out := make([]budgetDelta, 0, len(sorted))
	var prev float64
	for i, b := range sorted {
		growth := 0.0
		if i > 0 && prev > 0 {
			growth = round2((b.Amount - prev) / prev)
		}
		out = append(out, budgetDelta{Year: b.Year, Amount: b.Amount, Growth: growth})
		prev = b.Amount
	}
	return out
}

// suggestAllocation 分配建议（纯函数，测试锚点）：上年占比缩放至 total；
// 单部门占比 <10% 抬升到 10%，>0 的部门重归一。零历史均分。
func suggestAllocation(history []deptAmount, total float64) []deptAmount {
	valid := make([]deptAmount, 0, len(history))
	var sum float64
	for _, h := range history {
		if h.Amount > 0 {
			valid = append(valid, deptAmount{Dept: h.Dept, Amount: h.Amount})
			sum += h.Amount
		}
	}
	if len(valid) == 0 || total <= 0 {
		return nil
	}
	for i := range valid {
		valid[i].Ratio = round2(valid[i].Amount / sum)
	}
	const floor = 0.10
	// 下限抬升：触发下限的部门合计 floorTotal，其余部门按原占比分摊 remain=1-floorTotal
	floorTotal := 0.0
	var aboveIdx []int
	var aboveOrigSum float64
	for i := range valid {
		if valid[i].Ratio < floor {
			valid[i].Ratio = floor
			floorTotal += floor
		} else {
			aboveIdx = append(aboveIdx, i)
			aboveOrigSum += valid[i].Ratio
		}
	}
	remain := round2(1 - floorTotal)
	if len(aboveIdx) > 0 && remain > 0 {
		for _, i := range aboveIdx {
			valid[i].Ratio = round2(valid[i].Ratio / aboveOrigSum * remain)
		}
	}
	for i := range valid {
		valid[i].Amount = round2(total * valid[i].Ratio)
	}
	return valid
}

// planDiff 方案对比
type planDiff struct {
	TotalA  float64 `json:"totalA"`
	TotalB  float64 `json:"totalB"`
	Delta   float64 `json:"delta"`
	Summary string  `json:"summary"`
}

func comparePlans(planA, planB []deptAmount) planDiff {
	depts := map[string][2]float64{}
	var ta, tb float64
	for _, p := range planA {
		v := depts[p.Dept]
		v[0] += p.Amount
		depts[p.Dept] = v
		ta += p.Amount
	}
	for _, p := range planB {
		v := depts[p.Dept]
		v[1] += p.Amount
		depts[p.Dept] = v
		tb += p.Amount
	}
	names := make([]string, 0, len(depts))
	for k := range depts {
		names = append(names, k)
	}
	sort.Strings(names)
	out := planDiff{TotalA: round2(ta), TotalB: round2(tb)}
	var biggestDelta float64
	var biggestDept string
	for _, name := range names {
		v := depts[name]
		d := round2(v[1] - v[0])
		out.Summary += fmt.Sprintf("%s: %g→%g (%+g)；", name, v[0], v[1], d)
		if d > 0 && d > biggestDelta {
			biggestDelta = d
			biggestDept = name
		}
	}
	out.Delta = round2(tb - ta)
	if biggestDept != "" {
		out.Summary += fmt.Sprintf("方案 B 相对更倾斜 %s（+%g）。", biggestDept, biggestDelta)
	} else {
		out.Summary += "两方案总量持平或 B 未显著倾斜。"
	}
	return out
}

// SaveEduBudget 新建/更新预算
func (s *EduadminService) SaveEduBudget(in *model.EduBudget) (*model.EduBudget, error) {
	if strings.TrimSpace(in.Dept) == "" || in.Year < 2000 || in.Year > 2100 {
		return nil, fmt.Errorf("[%d]部门与年度必填", ErrCodeEduBudgetParam)
	}
	var exist model.EduBudget
	if err := global.GVA_DB.Where("year = ? AND dept = ?", in.Year, in.Dept).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Update("amount", in.Amount).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduBudgets 预算列表（年度过滤，部门升序）
func (s *EduadminService) GetEduBudgets(year int) ([]model.EduBudget, error) {
	db := global.GVA_DB.Model(&model.EduBudget{})
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var list []model.EduBudget
	err := db.Order("year ASC, dept ASC").Limit(500).Find(&list).Error
	return list, err
}

// AnalyzeEduBudget 部门预算同比分析
func (s *EduadminService) AnalyzeEduBudget(dept string) ([]budgetDelta, error) {
	if strings.TrimSpace(dept) == "" {
		return nil, fmt.Errorf("[%d]部门必填", ErrCodeEduBudgetParam)
	}
	var list []model.EduBudget
	if err := global.GVA_DB.Where("dept = ?", dept).Order("year ASC").Limit(50).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("[%d]该部门无预算数据", ErrCodeEduBudgetParam)
	}
	return budgetTrend(list), nil
}

// SuggestEduBudget 分配建议（基于最近年度占比）
func (s *EduadminService) SuggestEduBudget(total float64) ([]deptAmount, error) {
	if total <= 0 {
		return nil, fmt.Errorf("[%d]总预算须为正数", ErrCodeEduBudgetParam)
	}
	var latest int
	if err := global.GVA_DB.Model(&model.EduBudget{}).Select("MAX(year) as latest").Scan(&latest).Error; err != nil || latest == 0 {
		return nil, fmt.Errorf("[%d]库内无预算历史", ErrCodeEduBudgetParam)
	}
	var list []model.EduBudget
	if err := global.GVA_DB.Where("year = ?", latest).Find(&list).Error; err != nil {
		return nil, err
	}
	history := make([]deptAmount, 0, len(list))
	for _, b := range list {
		history = append(history, deptAmount{Dept: b.Dept, Amount: b.Amount})
	}
	out := suggestAllocation(history, total)
	if out == nil {
		return nil, fmt.Errorf("[%d]最近年度无可分配预算", ErrCodeEduBudgetParam)
	}
	return out, nil
}

// CompareEduBudgetPlans 方案对比
func (s *EduadminService) CompareEduBudgetPlans(planA, planB []deptAmount) (planDiff, error) {
	if len(planA) == 0 || len(planB) == 0 {
		return planDiff{}, fmt.Errorf("[%d]两方案均须非空", ErrCodeEduBudgetParam)
	}
	return comparePlans(planA, planB), nil
}
