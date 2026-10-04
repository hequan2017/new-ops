// Package service 白泽流水线：定义 CRUD（三层嵌套，事务内整体替换）
// pipeline 插件错误码段：1300-1399（DEV_PLAN 3.6）
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"gorm.io/gorm"
)

// 错误码
const (
	ErrCodePlNameRequired  = 1301
	ErrCodePlDuplicate     = 1302
	ErrCodePlStructInvalid = 1303
	ErrCodePlNotFound      = 1304
)

// PipelineService 流水线服务
type PipelineService struct{}

// validatePipeline 定义结构校验（纯逻辑，可单测）：名称必填、阶段/步骤结构与类型合法性
func validatePipeline(p *model.Pipeline) error {
	if p == nil || strings.TrimSpace(p.Name) == "" {
		return newPlErr(ErrCodePlNameRequired, "流水线名称不能为空")
	}
	for si := range p.Stages {
		st := &p.Stages[si]
		if strings.TrimSpace(st.Name) == "" {
			return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("第 %d 个阶段缺少名称", si+1))
		}
		if len(st.Steps) == 0 {
			return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("阶段「%s」至少需要一个步骤", st.Name))
		}
		for ti := range st.Steps {
			sp := &st.Steps[ti]
			if strings.TrimSpace(sp.Name) == "" {
				return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("阶段「%s」第 %d 个步骤缺少名称", st.Name, ti+1))
			}
			if !model.ValidStepTypes()[sp.Type] {
				return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("步骤「%s」类型非法: %s", sp.Name, sp.Type))
			}
			switch sp.Type {
			case model.StepTypeShell:
				if strings.TrimSpace(sp.ShellContent) == "" {
					return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("shell 步骤「%s」内容不能为空", sp.Name))
				}
				if sp.HostID == 0 {
					return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("shell 步骤「%s」必须绑定目标主机（SSH 执行）", sp.Name))
				}
			case model.StepTypeHTTP:
				if !strings.HasPrefix(sp.HTTPURL, "http://") && !strings.HasPrefix(sp.HTTPURL, "https://") {
					return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("http 步骤「%s」地址必须以 http(s):// 开头", sp.Name))
				}
				if sp.HTTPMethod == "" {
					sp.HTTPMethod = "GET"
				}
			}
			if sp.TimeoutSec < 0 {
				return newPlErr(ErrCodePlStructInvalid, fmt.Sprintf("步骤「%s」超时不能为负", sp.Name))
			}
		}
	}
	return nil
}

func newPlErr(code int, msg string) error { return &pipelineError{Code: code, Msg: msg} }

type pipelineError struct {
	Code int
	Msg  string
}

func (e *pipelineError) Error() string { return e.Msg }

// CreatePipeline 创建定义（嵌套创建阶段与步骤）
func (s *PipelineService) CreatePipeline(p *model.Pipeline) error {
	if err := validatePipeline(p); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.Pipeline{}).Where("name = ?", p.Name).Count(&count)
	if count > 0 {
		return newPlErr(ErrCodePlDuplicate, fmt.Sprintf("流水线已存在: %s", p.Name))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// Omit 防止 GVA 关联自动落库丢失 Sort 归一，阶段/步骤统一走 replaceStages
		if err := tx.Omit("Stages").Create(p).Error; err != nil {
			return err
		}
		return replaceStages(tx, p.ID, p.Stages)
	})
}

// UpdatePipeline 更新定义（阶段/步骤整体替换；执行快照语义由构建侧保证，改定义不影响历史构建）
func (s *PipelineService) UpdatePipeline(p *model.Pipeline) error {
	if p == nil || p.ID == 0 {
		return newPlErr(ErrCodePlNameRequired, "流水线ID不能为空")
	}
	if err := validatePipeline(p); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.Pipeline{}).Where("name = ? AND id <> ?", p.Name, p.ID).Count(&count)
	if count > 0 {
		return newPlErr(ErrCodePlDuplicate, fmt.Sprintf("流水线已存在: %s", p.Name))
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var exist model.Pipeline
		if err := tx.First(&exist, p.ID).Error; err != nil {
			return newPlErr(ErrCodePlNotFound, "流水线不存在")
		}
		if err := tx.Model(&exist).Omit("Stages").Updates(map[string]any{
			"name": p.Name, "description": p.Description, "enabled": p.Enabled,
		}).Error; err != nil {
			return err
		}
		// 整体替换：先删旧（步骤→阶段），再插新
		var oldStages []model.PipelineStage
		if err := tx.Where("pipeline_id = ?", p.ID).Find(&oldStages).Error; err != nil {
			return err
		}
		for _, st := range oldStages {
			if err := tx.Where("stage_id = ?", st.ID).Delete(&model.PipelineStep{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("pipeline_id = ?", p.ID).Delete(&model.PipelineStage{}).Error; err != nil {
			return err
		}
		return replaceStages(tx, p.ID, p.Stages)
	})
}

// replaceStages 写入阶段与步骤（Sort 按下标归一）
func replaceStages(tx *gorm.DB, pipelineID uint, stages []model.PipelineStage) error {
	for i := range stages {
		st := &stages[i]
		st.PipelineID = pipelineID
		st.Sort = i + 1
		st.ID = 0
		steps := make([]model.PipelineStep, len(st.Steps))
		copy(steps, st.Steps)
		for j := range steps {
			steps[j].StageID = 0
			steps[j].Sort = j + 1
			steps[j].ID = 0
		}
		st.Steps = nil
		if err := tx.Create(st).Error; err != nil {
			return err
		}
		for j := range steps {
			steps[j].StageID = st.ID
		}
		if len(steps) > 0 {
			if err := tx.Create(&steps).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// DeletePipeline 删除定义（级联清理阶段与步骤）
func (s *PipelineService) DeletePipeline(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var stages []model.PipelineStage
		if err := tx.Where("pipeline_id = ?", id).Find(&stages).Error; err != nil {
			return err
		}
		for _, st := range stages {
			if err := tx.Where("stage_id = ?", st.ID).Delete(&model.PipelineStep{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("pipeline_id = ?", id).Delete(&model.PipelineStage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Pipeline{}, id).Error
	})
}

// GetPipelineList 列表（keyword 过滤，含嵌套定义）
func (s *PipelineService) GetPipelineList(keyword string) (list []*model.Pipeline, err error) {
	db := global.GVA_DB.Model(&model.Pipeline{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err = db.Preload("Stages").Preload("Stages.Steps").Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// GetPipelineDetail 详情（嵌套预载，阶段/步骤按 Sort 排序）
func (s *PipelineService) GetPipelineDetail(id uint) (*model.Pipeline, error) {
	var p model.Pipeline
	if err := global.GVA_DB.Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("Stages.Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).First(&p, id).Error; err != nil {
		return nil, newPlErr(ErrCodePlNotFound, "流水线不存在")
	}
	return &p, nil
}
