// Package service llmops：智能体管理（要求(5)）——CRUD + 六阶段状态 + 发布版本快照
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeAgentNotFound = 2264
	ErrCodeAgentModeBad  = 2265
	ErrCodeAgentNameDup  = 2266
	ErrCodeAgentConfig   = 2267
	ErrCodeAgentStage    = 2268
)

// AgentModes 构建模式白名单（要求(5)：问答助手/交互应用/工作流/多智能体/Agentic）
var AgentModes = []string{"qa", "interactive", "workflow", "multi", "agentic"}

// AgentStages 六阶段白名单
var AgentStages = []string{"draft", "debug", "eval", "published", "operating", "monitor"}

// ValidAgentMode 模式白名单
func ValidAgentMode(m string) bool {
	for _, v := range AgentModes {
		if v == m {
			return true
		}
	}
	return false
}

// ValidAgentStage 阶段白名单
func ValidAgentStage(s string) bool {
	for _, v := range AgentStages {
		if v == s {
			return true
		}
	}
	return false
}

// clampMaxTurns 轮次上限夹取 1-20
func clampMaxTurns(n int) int {
	if n <= 0 {
		return 5
	}
	if n > 20 {
		return 20
	}
	return n
}

// AgentIn 智能体保存入参
type AgentIn struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Mode         string `json:"mode"`
	Desc         string `json:"desc"`
	SysPrompt    string `json:"sysPrompt"`
	ModelId      uint   `json:"modelId"`
	KnowledgeIds string `json:"knowledgeIds"`
	MaxTurns     int    `json:"maxTurns"`
	ConfigJSON   string `json:"configJSON"`
}

// GetAgents 智能体列表
func (s *LlmopsServiceStruct) GetAgents() ([]model.LlmopsAgent, error) {
	var list []model.LlmopsAgent
	err := global.GVA_DB.Order("id desc").Find(&list).Error
	return list, err
}

// SaveAgent 新建/更新智能体（名称唯一/模式白名单/配置 JSON 对象即检）
func (s *LlmopsServiceStruct) SaveAgent(in *AgentIn) (*model.LlmopsAgent, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidAgentMode(in.Mode) {
		return nil, fmt.Errorf("[%d]名称必填且模式须为 qa/interactive/workflow/multi/agentic", ErrCodeAgentModeBad)
	}
	if strings.TrimSpace(in.ConfigJSON) != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(in.ConfigJSON), &m); err != nil {
			return nil, fmt.Errorf("[%d]配置须为 JSON 对象: %w", ErrCodeAgentConfig, err)
		}
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsAgent{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]智能体名已存在", ErrCodeAgentNameDup)
	}
	maxTurns := clampMaxTurns(in.MaxTurns)
	if in.ID > 0 {
		var ag model.LlmopsAgent
		if err := global.GVA_DB.First(&ag, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
		}
		err := global.GVA_DB.Model(&ag).Updates(map[string]any{
			"name": in.Name, "mode": in.Mode, "desc": in.Desc, "sys_prompt": in.SysPrompt,
			"model_id": in.ModelId, "knowledge_ids": in.KnowledgeIds,
			"max_turns": maxTurns, "config_json": in.ConfigJSON,
		}).Error
		if err != nil {
			return nil, err
		}
		return &ag, nil
	}
	ag := &model.LlmopsAgent{
		Name: in.Name, Mode: in.Mode, Desc: in.Desc, SysPrompt: in.SysPrompt,
		ModelId: in.ModelId, KnowledgeIds: in.KnowledgeIds,
		MaxTurns: maxTurns, ConfigJSON: in.ConfigJSON, Status: "draft",
	}
	if err := global.GVA_DB.Create(ag).Error; err != nil {
		return nil, err
	}
	return ag, nil
}

// SetAgentStage 生命周期阶段流转（白名单校验）
func (s *LlmopsServiceStruct) SetAgentStage(id uint, stage string) (*model.LlmopsAgent, error) {
	if !ValidAgentStage(stage) {
		return nil, fmt.Errorf("[%d]不支持的生命周期阶段: %s", ErrCodeAgentStage, stage)
	}
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	if err := global.GVA_DB.Model(&ag).Update("status", stage).Error; err != nil {
		return nil, err
	}
	ag.Status = stage
	return &ag, nil
}

// PublishAgent 发布：配置快照落版本表（版本号自增）+ 状态置 published
func (s *LlmopsServiceStruct) PublishAgent(id uint, remark string) (*model.LlmopsAgentVersion, error) {
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	snapshot, _ := json.Marshal(map[string]any{
		"name": ag.Name, "mode": ag.Mode, "sysPrompt": ag.SysPrompt,
		"modelId": ag.ModelId, "knowledgeIds": ag.KnowledgeIds,
		"maxTurns": ag.MaxTurns, "configJSON": ag.ConfigJSON,
	})
	var maxVer int
	global.GVA_DB.Model(&model.LlmopsAgentVersion{}).Where("agent_id = ?", id).
		Select("COALESCE(MAX(version),0)").Scan(&maxVer)
	ver := &model.LlmopsAgentVersion{
		AgentId: id, Version: maxVer + 1, Snapshot: string(snapshot), Remark: remark,
	}
	if err := global.GVA_DB.Create(ver).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ag).Update("status", "published").Error; err != nil {
		return nil, err
	}
	return ver, nil
}

// GetAgentVersions 版本列表（倒序）
func (s *LlmopsServiceStruct) GetAgentVersions(agentId uint) ([]model.LlmopsAgentVersion, error) {
	var list []model.LlmopsAgentVersion
	err := global.GVA_DB.Where("agent_id = ?", agentId).Order("version desc").Find(&list).Error
	return list, err
}

// DeleteAgent 删除智能体（软删除）
func (s *LlmopsServiceStruct) DeleteAgent(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsAgent{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	return nil
}
