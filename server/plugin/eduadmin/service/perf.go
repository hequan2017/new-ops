// Package service eduadmin：绩效考评（需求 3.6-13）
// 口径：教师权重 教学0.5/科研0.3/服务0.2，行政权重 教学0.3/科研0.1/服务0.6；
// 综合分=Σ权重×分项（教学分按学时达标折算系数 min(1,学时/160)）；
// 等级 优秀≥90/良好≥75/合格≥60/不合格；报告=模板+数据填充纯函数，幂等存档。
package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduPerfParam    = 2809
	ErrCodeEduPerfNotFound = 2810
)

// PerfWeights 岗位权重（教学/科研/服务）
var PerfWeights = map[string][3]float64{
	"教师": {0.5, 0.3, 0.2},
	"行政": {0.3, 0.1, 0.6},
}

// perfTotal 综合分（纯函数，测试锚点）：教学分按学时达标系数折算
func perfTotal(rec model.EduPerfRecord) (float64, error) {
	w, ok := PerfWeights[rec.StaffType]
	if !ok {
		return 0, fmt.Errorf("[%d]人员类型须为 教师/行政", ErrCodeEduPerfParam)
	}
	teachFactor := math.Min(1, rec.TeachingHours/160)
	total := w[0]*rec.TeachingScore*teachFactor + w[1]*rec.ResearchScore + w[2]*rec.ServiceScore
	return round2(total), nil
}

// perfGrade 等级（纯函数，测试锚点）
func perfGrade(total float64) string {
	switch {
	case total >= 90:
		return "优秀"
	case total >= 75:
		return "良好"
	case total >= 60:
		return "合格"
	default:
		return "不合格"
	}
}

// renderPerfReport 报告模板填充（纯函数，测试锚点）
func renderPerfReport(rec model.EduPerfRecord, total float64, grade string) string {
	w := PerfWeights[rec.StaffType]
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %d 年度绩效报告（%s）\n", rec.StaffName, rec.Year, rec.StaffType)
	fmt.Fprintf(&sb, "一、教学：全年 %g 学时（达标系数 %.2f），教学评价分 %g，权重 %.1f。\n", rec.TeachingHours, math.Min(1, rec.TeachingHours/160), rec.TeachingScore, w[0])
	fmt.Fprintf(&sb, "二、科研：科研分 %g，权重 %.1f。\n", rec.ResearchScore, w[1])
	fmt.Fprintf(&sb, "三、服务：服务/行政分 %g，权重 %.1f。\n", rec.ServiceScore, w[2])
	fmt.Fprintf(&sb, "四、综合得分 %g，等级 %s。", total, grade)
	switch grade {
	case "优秀":
		sb.WriteString("建议纳入年度表彰与绩效倾斜。")
	case "不合格":
		sb.WriteString("建议制定改进计划并于下年度复评。")
	}
	return sb.String()
}

// GeneratePerfReport 生成并归档绩效报告（同年同人幂等覆盖）
func (s *EduadminService) GeneratePerfReport(staffName string, year int) (*model.EduPerfReport, error) {
	if strings.TrimSpace(staffName) == "" || year < 2000 || year > 2100 {
		return nil, fmt.Errorf("[%d]姓名与年度必填", ErrCodeEduPerfParam)
	}
	var records []model.EduPerfRecord
	if err := global.GVA_DB.Where("staff_name = ? AND year = ?", staffName, year).Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("[%d]该人员年度无绩效数据", ErrCodeEduPerfNotFound)
	}
	// 多条记录取均值聚合
	var tHours, tScore, rScore, sScore float64
	for _, r := range records {
		tHours += r.TeachingHours
		tScore += r.TeachingScore
		rScore += r.ResearchScore
		sScore += r.ServiceScore
	}
	n := float64(len(records))
	agg := model.EduPerfRecord{
		StaffName: staffName, StaffType: records[0].StaffType, Year: year,
		TeachingHours: tHours / n, TeachingScore: tScore / n,
		ResearchScore: rScore / n, ServiceScore: sScore / n,
	}
	total, err := perfTotal(agg)
	if err != nil {
		return nil, err
	}
	grade := perfGrade(total)
	report := model.EduPerfReport{
		StaffName: staffName, Year: year, TotalScore: total, Grade: grade,
		ReportText: renderPerfReport(agg, total, grade),
	}
	var exist model.EduPerfReport
	if err := global.GVA_DB.Where("staff_name = ? AND year = ?", staffName, year).First(&exist).Error; err == nil {
		if err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"total_score": total, "grade": grade, "report_text": report.ReportText,
		}).Error; err != nil {
			return nil, err
		}
		return &exist, nil
	}
	if err := global.GVA_DB.Create(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// SaveEduPerfRecord 新建/更新绩效数据
func (s *EduadminService) SaveEduPerfRecord(in *model.EduPerfRecord) (*model.EduPerfRecord, error) {
	if strings.TrimSpace(in.StaffName) == "" {
		return nil, fmt.Errorf("[%d]姓名必填", ErrCodeEduPerfParam)
	}
	if _, ok := PerfWeights[in.StaffType]; !ok {
		return nil, fmt.Errorf("[%d]人员类型须为 教师/行政", ErrCodeEduPerfParam)
	}
	if in.ID > 0 {
		var exist model.EduPerfRecord
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]绩效数据不存在", ErrCodeEduPerfNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"staff_name": in.StaffName, "staff_type": in.StaffType, "year": in.Year,
			"teaching_hours": in.TeachingHours, "teaching_score": in.TeachingScore,
			"research_score": in.ResearchScore, "service_score": in.ServiceScore,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduPerfRecords 绩效数据列表
func (s *EduadminService) GetEduPerfRecords(staffName string, year int, page, pageSize int) ([]model.EduPerfRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduPerfRecord{})
	if staffName != "" {
		db = db.Where("staff_name LIKE ?", "%"+staffName+"%")
	}
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduPerfRecord
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetEduPerfReports 报告列表
func (s *EduadminService) GetEduPerfReports(staffName string, year int, page, pageSize int) ([]model.EduPerfReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduPerfReport{})
	if staffName != "" {
		db = db.Where("staff_name LIKE ?", "%"+staffName+"%")
	}
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduPerfReport
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
