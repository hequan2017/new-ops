// Package service 白泽工单引擎（M7）：定义校验 + 实例流转
package service

import (
	pipelineBuild "github.com/hequan2017/new-ops/server/plugin/pipeline/service"
	pipelineModel "github.com/hequan2017/new-ops/server/plugin/pipeline/model"
)

// PipelineTrigger 发版钩子的流水线依赖缝：解 workflow→pipeline 具体类型耦合，
// 单测可注入假触发器验证钩子行为。
type PipelineTrigger interface {
	CreateBuild(pipelineID uint, params map[string]string, operator string, userID uint, triggerIP string) (*pipelineModel.PipelineBuild, error)
}

// defaultTrigger 生产默认触发器（具体实现由 pipeline 插件提供）
func defaultTrigger() PipelineTrigger {
	return new(pipelineBuild.PipelineBuildService)
}
