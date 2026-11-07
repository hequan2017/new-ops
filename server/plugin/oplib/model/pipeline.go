// Package model oplib：算子管道
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// OplibPipeline 算子管道（采购件② DAG 编排；Kahn 拓扑校验环）
type OplibPipeline struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"comment:管道名;unique"`
	Desc    string `json:"desc" gorm:"type:varchar(500);comment:描述"`
	DefJSON string `json:"defJSON" gorm:"type:text;comment:编排定义JSON(nodes+edges)"`
	Status  string `json:"status" gorm:"type:varchar(16);default:draft;comment:draft/ready/disabled"`
}

func (OplibPipeline) TableName() string { return "oplib_pipeline" }
