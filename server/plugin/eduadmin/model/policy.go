// Package model eduadmin 教育管理一模通管：六大知识库（需求 3.6-4）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduPolicy 知识库条目（人事/财务/资产/后勤/决策/跨校 六大域）
type EduPolicy struct {
	global.GVA_MODEL
	Domain  string `json:"domain" gorm:"type:varchar(16);index;comment:人事/财务/资产/后勤/决策/跨校"`
	Title   string `json:"title" gorm:"type:varchar(300);comment:条目标题"`
	Content string `json:"content" gorm:"type:text;comment:条目内容"`
	Tags    string `json:"tags" gorm:"type:varchar(500);comment:标签(逗号分隔)"`
}

func (EduPolicy) TableName() string { return "edu_policy" }
