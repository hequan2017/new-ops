// Package service 白泽容器管理：应用模板化一键部署（M8 C3）
// 模板 = compose 内容 + {{参数}} 占位；部署 = 渲染 → 复用 Compose CreateComposeProject（校验+up）→ 实例留档。
package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

// 错误码接续 1411（portforward）
const (
	ErrCodeTplInvalid  = 1412
	ErrCodeTplNotFound = 1413
)

// 模板占位符：{{key}}（key 字母数字下划线，容忍两侧空白）
var tplPlaceholderReg = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// ParamDef 参数定义（ParamsSchema JSON 数组元素）
type ParamDef struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

// ParseParamsSchema 解析参数定义（纯函数；空 schema 合法）
func ParseParamsSchema(raw string) ([]ParamDef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []ParamDef{}, nil
	}
	var defs []ParamDef
	if err := json.Unmarshal([]byte(raw), &defs); err != nil {
		return nil, fmt.Errorf("参数定义非合法 JSON 数组: %w", err)
	}
	seen := map[string]bool{}
	for _, d := range defs {
		if d.Key == "" || seen[d.Key] {
			return nil, fmt.Errorf("参数定义存在空 key 或重复 key: %q", d.Key)
		}
		if !tplPlaceholderReg.MatchString("{{" + d.Key + "}}") {
			return nil, fmt.Errorf("参数 key 非法（仅字母数字下划线，字母开头）: %q", d.Key)
		}
		seen[d.Key] = true
	}
	return defs, nil
}

// renderComposeTpl 参数渲染（纯函数）：全部占位符必须有对应参数值（default 兜底），多余参数报错防错配
func renderComposeTpl(tpl string, schema []ParamDef, params map[string]string) (string, error) {
	merged := map[string]string{}
	for _, d := range schema {
		v, ok := params[d.Key]
		if !ok || v == "" {
			v = d.Default
		}
		if v == "" {
			if d.Required {
				return "", fmt.Errorf("必填参数缺失: %s", d.Key)
			}
			continue
		}
		merged[d.Key] = v
	}
	// 参数集外键（模板里有占位但 schema 未定义/未传值）
	var unknown []string
	for _, m := range tplPlaceholderReg.FindAllStringSubmatch(tpl, -1) {
		key := m[1]
		if _, ok := merged[key]; !ok {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return "", fmt.Errorf("模板占位符缺少参数定义或取值: %s", strings.Join(unknown, ", "))
	}
	return tplPlaceholderReg.ReplaceAllStringFunc(tpl, func(m string) string {
		sub := tplPlaceholderReg.FindStringSubmatch(m)
		return merged[sub[1]]
	}), nil
}

// GetAppTemplates 模板列表
func (s *ComposeService) GetAppTemplates() ([]model.DockerAppTemplate, error) {
	var list []model.DockerAppTemplate
	err := global.GVA_DB.Order("category, id").Find(&list).Error
	return list, err
}

// SaveAppTemplate 保存模板（schema 校验 + compose 模板非空）
func (s *ComposeService) SaveAppTemplate(t *model.DockerAppTemplate) error {
	if t == nil || strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.ComposeTpl) == "" {
		return newCtErr(ErrCodeTplInvalid, "模板名与 compose 内容不能为空")
	}
	if _, err := ParseParamsSchema(t.ParamsSchema); err != nil {
		return newCtErr(ErrCodeTplInvalid, err.Error())
	}
	if t.ID > 0 {
		return global.GVA_DB.Model(&model.DockerAppTemplate{}).Where("id = ?", t.ID).Updates(map[string]any{
			"name": t.Name, "category": t.Category, "description": t.Description,
			"compose_tpl": t.ComposeTpl, "params_schema": t.ParamsSchema, "notes": t.Notes,
		}).Error
	}
	var count int64
	global.GVA_DB.Model(&model.DockerAppTemplate{}).Where("name = ?", t.Name).Count(&count)
	if count > 0 {
		return newCtErr(ErrCodeEpDuplicate, "模板名已存在: "+t.Name)
	}
	return global.GVA_DB.Create(t).Error
}

// DeleteAppTemplate 删除模板（实例留快照不受影响）
func (s *ComposeService) DeleteAppTemplate(id uint) error {
	return global.GVA_DB.Delete(&model.DockerAppTemplate{}, id).Error
}

// DeployFromTemplate 一键部署：渲染 → 建 compose 项目（校验+up）→ 实例留档
func (s *ComposeService) DeployFromTemplate(templateID, endpointID uint, projectName string,
	params map[string]string, operator string) (*model.DockerComposeProject, error) {
	var tpl model.DockerAppTemplate
	if err := global.GVA_DB.First(&tpl, templateID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newCtErr(ErrCodeTplNotFound, "应用模板不存在")
		}
		return nil, err
	}
	schema, err := ParseParamsSchema(tpl.ParamsSchema)
	if err != nil {
		return nil, newCtErr(ErrCodeTplInvalid, err.Error())
	}
	content, rerr := renderComposeTpl(tpl.ComposeTpl, schema, params)
	if rerr != nil {
		return nil, newCtErr(ErrCodeTplInvalid, "模板渲染失败: "+rerr.Error())
	}
	proj, err := s.CreateComposeProject(projectName, endpointID, content, operator, "来自模板: "+tpl.Name)
	if err != nil {
		return nil, err
	}
	paramsJSON, _ := json.Marshal(params)
	now := time.Now()
	inst := &model.DockerAppInstance{
		ProjectName: projectName, TemplateID: tpl.ID, TemplateName: tpl.Name,
		EndpointID: endpointID, Params: string(paramsJSON), CreatedBy: operator, DeployedAt: &now,
	}
	var exist model.DockerAppInstance
	if err := global.GVA_DB.Where("project_name = ?", projectName).First(&exist).Error; err == nil {
		global.GVA_DB.Model(&exist).Updates(map[string]any{
			"template_id": inst.TemplateID, "template_name": inst.TemplateName,
			"endpoint_id": inst.EndpointID, "params": inst.Params, "deployed_at": now,
		})
	} else {
		global.GVA_DB.Create(inst)
	}
	return proj, nil
}

// GetAppInstances 模板部署实例列表（联 compose 项目状态）
func (s *ComposeService) GetAppInstances() ([]map[string]any, error) {
	var list []model.DockerAppInstance
	if err := global.GVA_DB.Order("id desc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(list))
	for _, it := range list {
		var proj model.DockerComposeProject
		row := map[string]any{
			"ID": it.ID, "projectName": it.ProjectName, "templateId": it.TemplateID,
			"templateName": it.TemplateName, "endpointId": it.EndpointID,
			"params": it.Params, "createdBy": it.CreatedBy, "deployedAt": it.DeployedAt,
			"composeStatus": "已删除",
		}
		if err := global.GVA_DB.Where("name = ?", it.ProjectName).First(&proj).Error; err == nil {
			row["composeStatus"] = proj.Status
			row["composeProjectId"] = proj.ID
		}
		out = append(out, row)
	}
	return out, nil
}
