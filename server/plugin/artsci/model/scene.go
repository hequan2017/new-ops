// Package model artsci 学科工具集：文生图教学场景（需求 3.3-10/11/12/14/15）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciSceneTemplate 教学场景模板（提示词模板 + 教学要点）
type ArtsciSceneTemplate struct {
	global.GVA_MODEL
	Scene          string `json:"scene" gorm:"type:varchar(32);uniqueIndex;comment:场景名"`
	PromptTemplate string `json:"promptTemplate" gorm:"type:text;comment:提示词模板({{text}}/{{scene}})"`
	TeachingPoints string `json:"teachingPoints" gorm:"type:varchar(500);comment:教学要点"`
	Status         string `json:"status" gorm:"type:varchar(16);default:enabled;comment:enabled/disabled"`
}

func (ArtsciSceneTemplate) TableName() string { return "artsci_scene_template" }

// ArtsciImageTask 文生图任务（状态机 queued/done/failed）
type ArtsciImageTask struct {
	global.GVA_MODEL
	Scene          string `json:"scene" gorm:"type:varchar(32);index;comment:场景"`
	InputText      string `json:"inputText" gorm:"type:text;comment:教学输入文本"`
	RenderedPrompt string `json:"renderedPrompt" gorm:"type:text;comment:渲染后提示词"`
	Status         string `json:"status" gorm:"type:varchar(16);default:queued;index;comment:queued/done/failed"`
	ImageUrl       string `json:"imageUrl" gorm:"type:varchar(500);comment:生成图像地址"`
	ErrMsg         string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (ArtsciImageTask) TableName() string { return "artsci_image_task" }
