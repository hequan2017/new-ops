// Package service eduadmin：跨校协同（需求 3.6-27/28）
// 口径：资源调度规则（状态=可共享 且 容量>0 且 非归属校自用 → 通过并置占用，
// 容量-1）；联合预算分析=多校经费聚合+统筹分配建议（按基数占比，单校下限 10%
// 重归一，纯函数）；培训材料/最佳实践条目种子幂等。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduShareParam    = 2828
	ErrCodeEduShareNotFound = 2829
)

// scheduleCheck 调度规则判定（纯函数，测试锚点）
func scheduleCheck(res *model.EduShareResource, applySchool string) (bool, string) {
	if strings.TrimSpace(applySchool) == "" {
		return false, "申请校必填"
	}
	if applySchool == res.OwnerSchool {
		return false, "归属校资源自用，无须走共享调度"
	}
	if res.Status != "可共享" {
		return false, "资源当前不可共享（占用中）"
	}
	if res.Capacity <= 0 {
		return false, "可共享容量已用尽"
	}
	return true, "调度通过"
}

// jointAlloc 统筹分配建议（纯函数，测试锚点）：按基数占比缩放，单校下限 10% 重归一
func jointAlloc(bases map[string]float64, total float64) map[string]float64 {
	if total <= 0 || len(bases) == 0 {
		return nil
	}
	var sum float64
	for _, v := range bases {
		if v > 0 {
			sum += v
		}
	}
	if sum <= 0 {
		return nil
	}
	const floor = 0.10
	out := make(map[string]float64, len(bases))
	floorTotal := 0.0
	var aboveSum float64
	for school, v := range bases {
		if v <= 0 {
			out[school] = 0
			continue
		}
		ratio := v / sum
		if ratio < floor {
			out[school] = floor
			floorTotal += floor
		} else {
			out[school] = ratio
			aboveSum += ratio
		}
	}
	remain := 1 - floorTotal
	if aboveSum > 0 && remain > 0 {
		for school := range out {
			if out[school] > floor {
				out[school] = round2(out[school] / aboveSum * remain)
			}
		}
	}
	for school := range out {
		out[school] = round2(out[school] * total)
	}
	return out
}

// SchoolBudgetReport 联合预算报告
type SchoolBudgetReport struct {
	Year        int                `json:"year"`
	SchoolTotals map[string]float64 `json:"schoolTotals"`
	Total       float64            `json:"total"`
	Proposal    map[string]float64 `json:"proposal"`
	Summary     string             `json:"summary"`
}

// AnalyzeEduJointBudget 联合预算分析（多校聚合+统筹建议）
func (s *EduadminService) AnalyzeEduJointBudget(year int) (*SchoolBudgetReport, error) {
	if year < 2000 {
		return nil, fmt.Errorf("[%d]年度必填", ErrCodeEduShareParam)
	}
	var list []model.EduJointBudget
	if err := global.GVA_DB.Where("year = ?", year).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("[%d]该年度无联合预算数据", ErrCodeEduShareNotFound)
	}
	bases := map[string]float64{}
	report := &SchoolBudgetReport{Year: year, SchoolTotals: map[string]float64{}}
	for _, b := range list {
		bases[b.School] += b.Amount
		report.SchoolTotals[b.School] = round2(report.SchoolTotals[b.School] + b.Amount)
		report.Total += b.Amount
	}
	report.Total = round2(report.Total)
	report.Proposal = jointAlloc(bases, report.Total)
	schools := make([]string, 0, len(report.Proposal))
	for k := range report.Proposal {
		schools = append(schools, k)
	}
	sort.Strings(schools)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d 年度联合经费 %.0f 万元，统筹建议：", year, report.Total))
	for _, k := range schools {
		sb.WriteString(fmt.Sprintf("%s %.0f 万；", k, report.Proposal[k]))
	}
	report.Summary = sb.String()
	return report, nil
}

// ScheduleEduShareResource 资源调度申请
func (s *EduadminService) ScheduleEduShareResource(id uint, applySchool string) (*model.EduShareResource, error) {
	var res model.EduShareResource
	if err := global.GVA_DB.First(&res, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]共享资源不存在", ErrCodeEduShareNotFound)
	}
	ok, reason := scheduleCheck(&res, applySchool)
	if !ok {
		return nil, fmt.Errorf("[%d]%s", ErrCodeEduShareParam, reason)
	}
	if err := global.GVA_DB.Model(&res).Updates(map[string]any{
		"capacity": res.Capacity - 1,
		"status":   map[bool]string{true: "可共享", false: "占用"}[res.Capacity-1 > 0],
	}).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

// SaveEduShareResource 新建/更新共享资源
func (s *EduadminService) SaveEduShareResource(in *model.EduShareResource) (*model.EduShareResource, error) {
	if strings.TrimSpace(in.ResourceName) == "" || strings.TrimSpace(in.OwnerSchool) == "" {
		return nil, fmt.Errorf("[%d]资源名与归属校必填", ErrCodeEduShareParam)
	}
	switch in.Category {
	case "课程", "设备", "师资", "场地":
	default:
		return nil, fmt.Errorf("[%d]类别须为 课程/设备/师资/场地", ErrCodeEduShareParam)
	}
	if in.Status == "" {
		in.Status = "可共享"
	}
	if in.Status != "可共享" && in.Status != "占用" {
		return nil, fmt.Errorf("[%d]状态须为 可共享/占用", ErrCodeEduShareParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduShareResources 共享资源列表
func (s *EduadminService) GetEduShareResources(category string) ([]model.EduShareResource, error) {
	db := global.GVA_DB.Model(&model.EduShareResource{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var list []model.EduShareResource
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// GetEduBestPractices 培训材料/最佳实践列表
func (s *EduadminService) GetEduBestPractices(category string) ([]model.EduBestPractice, error) {
	db := global.GVA_DB.Model(&model.EduBestPractice{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var list []model.EduBestPractice
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}
