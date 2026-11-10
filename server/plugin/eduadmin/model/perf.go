// Package model eduadmin 教育管理一模通管：绩效考评（需求 3.6-13）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduPerfRecord 绩效数据记录（教师/行政人员年度教学科研服务数据）
type EduPerfRecord struct {
	global.GVA_MODEL
	StaffName    string  `json:"staffName" gorm:"type:varchar(100);index;comment:人员姓名"`
	StaffType    string  `json:"staffType" gorm:"type:varchar(16);comment:教师/行政"`
	Year         int     `json:"year" gorm:"index;comment:年度"`
	TeachingHours float64 `json:"teachingHours" gorm:"comment:教学学时"`
	TeachingScore float64 `json:"teachingScore" gorm:"comment:教学评价分(0-100)"`
	ResearchScore float64 `json:"researchScore" gorm:"comment:科研分(0-100)"`
	ServiceScore  float64 `json:"serviceScore" gorm:"comment:服务/行政分(0-100)"`
}

func (EduPerfRecord) TableName() string { return "edu_perf_record" }

// EduPerfReport 绩效报告记录
type EduPerfReport struct {
	global.GVA_MODEL
	StaffName  string  `json:"staffName" gorm:"type:varchar(100);index;comment:人员姓名"`
	Year       int     `json:"year" gorm:"index;comment:年度"`
	TotalScore float64 `json:"totalScore" gorm:"comment:综合得分"`
	Grade      string  `json:"grade" gorm:"type:varchar(16);comment:优秀/良好/合格/不合格"`
	ReportText string  `json:"reportText" gorm:"type:text;comment:个性化报告"`
}

func (EduPerfReport) TableName() string { return "edu_perf_report" }
