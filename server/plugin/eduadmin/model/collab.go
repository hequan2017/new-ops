// Package model eduadmin 教育管理一模通管：跨校协同（需求 3.6-27/28）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduShareResource 共享资源（课程/设备/师资/场地）
type EduShareResource struct {
	global.GVA_MODEL
	ResourceName string `json:"resourceName" gorm:"type:varchar(200);comment:资源名"`
	Category     string `json:"category" gorm:"type:varchar(32);index;comment:课程/设备/师资/场地"`
	OwnerSchool  string `json:"ownerSchool" gorm:"type:varchar(100);index;comment:归属高校"`
	Capacity     int    `json:"capacity" gorm:"comment:可共享容量(名额/台数)"`
	Status       string `json:"status" gorm:"type:varchar(16);index;comment:可共享/占用"`
}

func (EduShareResource) TableName() string { return "edu_share_resource" }

// EduJointBudget 联合预算（多校年度经费）
type EduJointBudget struct {
	global.GVA_MODEL
	School string  `json:"school" gorm:"type:varchar(100);index;comment:高校"`
	Year   int     `json:"year" gorm:"index;comment:年度"`
	Amount float64 `json:"amount" gorm:"comment:经费(万元)"`
}

func (EduJointBudget) TableName() string { return "edu_joint_budget" }

// EduBestPractice 培训材料/最佳实践条目
type EduBestPractice struct {
	global.GVA_MODEL
	Title    string `json:"title" gorm:"type:varchar(200);uniqueIndex;comment:标题"`
	Category string `json:"category" gorm:"type:varchar(32);index;comment:培训材料/最佳实践"`
	Content  string `json:"content" gorm:"type:text;comment:内容"`
}

func (EduBestPractice) TableName() string { return "edu_best_practice" }
