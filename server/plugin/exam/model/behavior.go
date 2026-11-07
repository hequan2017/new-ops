// Package model exam：防作弊行为日志
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamBehavior 考试行为日志（切屏/粘贴/复制/退出全屏上报留痕）
type ExamBehavior struct {
	global.GVA_MODEL
	SheetId uint   `json:"sheetId" gorm:"index;comment:答卷ID"`
	Event   string `json:"event" gorm:"type:varchar(16);index;comment:blur/paste/copy/fullscreen_exit"`
	Detail  string `json:"detail" gorm:"type:varchar(500);comment:详情"`
}

func (ExamBehavior) TableName() string { return "exam_behavior" }
