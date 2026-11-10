// Package model eduadmin 教育管理一模通管：智能招聘（需求 3.6-12）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduResume 简历
type EduResume struct {
	global.GVA_MODEL
	Name       string `json:"name" gorm:"type:varchar(100);comment:姓名"`
	Education  string `json:"education" gorm:"type:varchar(32);comment:学历(博士/硕士/本科)"`
	Years      int    `json:"years" gorm:"comment:工作年限"`
	Skills     string `json:"skills" gorm:"type:varchar(500);comment:技能标签(逗号分隔)"`
	ResumeText string `json:"resumeText" gorm:"type:text;comment:简历正文"`
}

func (EduResume) TableName() string { return "edu_resume" }

// EduPosition 招聘岗位
type EduPosition struct {
	global.GVA_MODEL
	Title         string `json:"title" gorm:"type:varchar(200);comment:岗位名"`
	Dept          string `json:"dept" gorm:"type:varchar(100);comment:用人部门"`
	RequiredSkill string `json:"requiredSkill" gorm:"type:varchar(500);comment:必备技能(逗号分隔)"`
	MinYears      int    `json:"minYears" gorm:"comment:最低工作年限"`
	MinEducation  string `json:"minEducation" gorm:"type:varchar(32);comment:最低学历"`
	Headcount     int    `json:"headcount" gorm:"comment:招聘人数"`
}

func (EduPosition) TableName() string { return "edu_position" }
