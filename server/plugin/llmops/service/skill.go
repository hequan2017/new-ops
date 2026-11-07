// Package service llmops：技能生态（要求(6)）——技能 CRUD + 定义协议 + AgentTool 适配
// 口径：定义协议 {kind:"prompt", system, template, params}（{input}/{param} 占位渲染）；
// 启用技能经适配器注册进 E03 工具注册表，chat 调用走 B05 能力调度。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeSkillNotFound = 2280
	ErrCodeSkillTypeBad  = 2281
	ErrCodeSkillNameDup  = 2282
	ErrCodeSkillDef      = 2283
)

// SkillTypes 技能类型白名单
var SkillTypes = []string{"builtin", "extended", "custom"}

// ValidSkillType 类型白名单
func ValidSkillType(t string) bool {
	for _, v := range SkillTypes {
		if v == t {
			return true
		}
	}
	return false
}

// SkillDef 技能定义（prompt 型）
type SkillDef struct {
	Kind     string                       `json:"kind"` // prompt（其余形态后续扩展）
	System   string                       `json:"system"`
	Template string                       `json:"template"`
	Params   map[string]map[string]string `json:"params"` // 参数名→{desc,default}
}

// ParseSkillDef 定义解析（纯函数，测试锚点）
func ParseSkillDef(raw string) (*SkillDef, error) {
	var def SkillDef
	if err := json.Unmarshal([]byte(raw), &def); err != nil {
		return nil, fmt.Errorf("[%d]技能定义解析失败: %w", ErrCodeSkillDef, err)
	}
	if def.Kind != "prompt" {
		return nil, fmt.Errorf("[%d]暂仅支持 prompt 型技能（当前 %s）", ErrCodeSkillDef, def.Kind)
	}
	if strings.TrimSpace(def.Template) == "" {
		return nil, fmt.Errorf("[%d]技能模板不能为空", ErrCodeSkillDef)
	}
	return &def, nil
}

// renderSkillTemplate 模板渲染：{input} + {param}（未提供参数保留原占位，纯函数）
func renderSkillTemplate(tpl, input string, params map[string]string) string {
	out := strings.ReplaceAll(tpl, "{input}", input)
	for k, v := range params {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return out
}

// GetSkills 技能列表（类型过滤）
func (s *LlmopsServiceStruct) GetSkills(skillType, status string) ([]model.LlmopsSkill, error) {
	db := global.GVA_DB.Model(&model.LlmopsSkill{})
	if skillType != "" {
		if !ValidSkillType(skillType) {
			return nil, fmt.Errorf("[%d]不支持的技能类型: %s", ErrCodeSkillTypeBad, skillType)
		}
		db = db.Where("type = ?", skillType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []model.LlmopsSkill
	err := db.Order("id").Find(&list).Error
	return list, err
}

// SaveSkill 新建/更新技能（名称唯一/类型白名单/定义即检）
func (s *LlmopsServiceStruct) SaveSkill(in *model.LlmopsSkill) (*model.LlmopsSkill, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidSkillType(in.Type) {
		return nil, fmt.Errorf("[%d]名称必填且类型须为 builtin/extended/custom", ErrCodeSkillTypeBad)
	}
	if _, err := ParseSkillDef(in.DefJSON); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsSkill{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]技能名已存在", ErrCodeSkillNameDup)
	}
	if in.ID > 0 {
		var sk model.LlmopsSkill
		if err := global.GVA_DB.First(&sk, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]技能不存在", ErrCodeSkillNotFound)
		}
		err := global.GVA_DB.Model(&sk).Updates(map[string]any{
			"name": in.Name, "type": in.Type, "desc": in.Desc, "def_json": in.DefJSON, "status": in.Status,
		}).Error
		if err != nil {
			return nil, err
		}
		return &sk, nil
	}
	if in.Status == "" {
		in.Status = "enabled"
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteSkill 删除技能（软删除）
func (s *LlmopsServiceStruct) DeleteSkill(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsSkill{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]技能不存在", ErrCodeSkillNotFound)
	}
	return nil
}

// skillToolAdapter 技能→AgentTool 适配器（E03 回路面）
type skillToolAdapter struct {
	skill model.LlmopsSkill
	fn    turnChatFunc
}

func (a *skillToolAdapter) Name() string        { return "skill:" + a.skill.Name }
func (a *skillToolAdapter) Description() string { return a.skill.Desc }

// Execute 技能执行：模板渲染 → chat（system + 渲染后模板）
func (a *skillToolAdapter) Execute(_ context.Context, input string) (string, error) {
	def, err := ParseSkillDef(a.skill.DefJSON)
	if err != nil {
		return "", err
	}
	// 输入兼容两种形态：纯文本（占 {input}）或 {"params":{...},"input":"..."}
	inputText := input
	params := map[string]string{}
	var wrapper struct {
		Input  string            `json:"input"`
		Params map[string]string `json:"params"`
	}
	if err := json.Unmarshal([]byte(input), &wrapper); err == nil && wrapper.Input != "" {
		inputText = wrapper.Input
		params = wrapper.Params
	}
	// 缺省参数补齐
	for name, p := range def.Params {
		if _, ok := params[name]; !ok {
			params[name] = p["default"]
		}
	}
	prompt := renderSkillTemplate(def.Template, inputText, params)
	msgs := make([]ChatMessage, 0, 2)
	if def.System != "" {
		msgs = append(msgs, ChatMessage{Role: "system", Content: def.System})
	}
	msgs = append(msgs, ChatMessage{Role: "user", Content: prompt})
	return a.fn(prompt, msgs)
}

// RegisterSkillAdapters 启用技能批量注册进工具注册表（chat 经能力调度；启动期 best-effort）
func (s *LlmopsServiceStruct) RegisterSkillAdapters() {
	var skills []model.LlmopsSkill
	if err := global.GVA_DB.Where("status = 'enabled'").Find(&skills).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 技能适配器装载失败: %v", err))
		return
	}
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
		defer cancel()
		m, err := s.SelectModel("chat")
		if err != nil {
			return "", err
		}
		content, _, err := s.ChatCompletion(ctx, m.ID, msgs, nil)
		return content, err
	}
	for i := range skills {
		RegisterAgentTool(&skillToolAdapter{skill: skills[i], fn: fn})
	}
}
