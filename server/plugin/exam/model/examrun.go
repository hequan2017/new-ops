// Package model exam：考试场次
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// ExamExam 考试场次（采购件③ 分布式考试：时间窗 + 试卷绑定 + 四态状态机）
type ExamExam struct {
	global.GVA_MODEL
	Name        string    `json:"name" gorm:"comment:考试名;unique"`
	PaperId     uint      `json:"paperId" gorm:"comment:试卷ID"`
	StartTime   time.Time `json:"startTime" gorm:"comment:开考时间"`
	EndTime     time.Time `json:"endTime" gorm:"comment:截止时间"`
	DurationMin int       `json:"durationMin" gorm:"default:60;comment:单场作答时长(分钟)"`
	Status      string    `json:"status" gorm:"type:varchar(16);default:draft;index;comment:draft/published/ongoing/finished"`
}

func (ExamExam) TableName() string { return "exam_exam" }
