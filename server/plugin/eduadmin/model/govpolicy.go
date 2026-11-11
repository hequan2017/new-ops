// Package model eduadmin 教育管理一模通管：政策数据更新（需求 3.6-35）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduGovPolicy 上级政策条目（时效性：发布/抓取时间；内容哈希幂等）
type EduGovPolicy struct {
	global.GVA_MODEL
	Title       string `json:"title" gorm:"type:varchar(300);comment:政策标题"`
	Source      string `json:"source" gorm:"type:varchar(32);index;comment:教育部/上海市教委"`
	Url         string `json:"url" gorm:"type:varchar(500);comment:原文链接"`
	PublishedAt string `json:"publishedAt" gorm:"type:varchar(32);comment:发布时间"`
	FetchedAt   string `json:"fetchedAt" gorm:"type:varchar(32);comment:抓取时间"`
	ContentHash string `json:"contentHash" gorm:"type:char(64);uniqueIndex;comment:SHA256(title+url) 幂等键"`
}

func (EduGovPolicy) TableName() string { return "edu_gov_policy" }
