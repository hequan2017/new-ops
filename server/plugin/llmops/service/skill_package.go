// Package service llmops：技能包导入导出（要求(6) 一键安装扩展技能）
// 口径：包格式 {"version":1,"skills":[{name,desc,defJSON}]}；导入前计划规划为纯函数
// （存在冲突→默认跳过并报告，overwrite=true 覆盖更新）；导入技能统一归 extended 型。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeSkillPkg = 2285

// skillPkg 技能包
type skillPkg struct {
	Version int        `json:"version"`
	Skills  []pkgSkill `json:"skills"`
}

// pkgSkill 包内技能
type pkgSkill struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	DefJSON string `json:"defJSON"`
}

// marshalSkillPackage 打包（纯函数，测试锚点）
func marshalSkillPackage(skills []model.LlmopsSkill) string {
	pkg := skillPkg{Version: 1}
	for i := range skills {
		pkg.Skills = append(pkg.Skills, pkgSkill{
			Name: skills[i].Name, Desc: skills[i].Desc, DefJSON: skills[i].DefJSON,
		})
	}
	b, _ := json.Marshal(pkg)
	return string(b)
}

// ParseSkillPackage 包解析（纯函数：版本/非空/逐个 def 校验）
func ParseSkillPackage(raw string) ([]pkgSkill, error) {
	var pkg skillPkg
	if err := json.Unmarshal([]byte(raw), &pkg); err != nil {
		return nil, fmt.Errorf("[%d]技能包格式非法: %w", ErrCodeSkillPkg, err)
	}
	if pkg.Version != 1 {
		return nil, fmt.Errorf("[%d]不支持的包版本: %d", ErrCodeSkillPkg, pkg.Version)
	}
	if len(pkg.Skills) == 0 {
		return nil, fmt.Errorf("[%d]技能包为空", ErrCodeSkillPkg)
	}
	for i := range pkg.Skills {
		s := &pkg.Skills[i]
		s.Name = strings.TrimSpace(s.Name)
		s.DefJSON = strings.TrimSpace(s.DefJSON)
		if s.Name == "" {
			return nil, fmt.Errorf("[%d]第 %d 个技能名为空", ErrCodeSkillPkg, i+1)
		}
		if _, err := ParseSkillDef(s.DefJSON); err != nil {
			return nil, fmt.Errorf("[%d]技能 %s 定义非法: %w", ErrCodeSkillPkg, s.Name, err)
		}
	}
	return pkg.Skills, nil
}

// importPlan 导入规划（纯函数，测试锚点）：existing=已存在技能名集合
func importPlan(existing map[string]bool, skills []pkgSkill, overwrite bool) (toCreate, toUpdate []pkgSkill, conflicts []string) {
	for _, s := range skills {
		if !existing[s.Name] {
			toCreate = append(toCreate, s)
			continue
		}
		if overwrite {
			toUpdate = append(toUpdate, s)
		} else {
			conflicts = append(conflicts, s.Name)
		}
	}
	return
}

// ExportSkillPackage 按导出 ID 列表打包
func (s *LlmopsServiceStruct) ExportSkillPackage(ids []uint) (string, error) {
	if len(ids) == 0 {
		return "", fmt.Errorf("[%d]未选择导出技能", ErrCodeSkillPkg)
	}
	var skills []model.LlmopsSkill
	if err := global.GVA_DB.Where("id IN ?", ids).Find(&skills).Error; err != nil {
		return "", err
	}
	if len(skills) == 0 {
		return "", fmt.Errorf("[%d]导出技能不存在", ErrCodeSkillNotFound)
	}
	return marshalSkillPackage(skills), nil
}

// ImportSkillPackage 一键安装（默认冲突跳过并报告；overwrite 覆盖）
func (s *LlmopsServiceStruct) ImportSkillPackage(raw string, overwrite bool) (imported int, conflicts []string, err error) {
	skills, err := ParseSkillPackage(raw)
	if err != nil {
		return 0, nil, err
	}
	var existList []model.LlmopsSkill
	if err := global.GVA_DB.Find(&existList).Error; err != nil {
		return 0, nil, err
	}
	existing := make(map[string]bool, len(existList))
	for _, e := range existList {
		existing[e.Name] = true
	}
	toCreate, toUpdate, conflicts := importPlan(existing, skills, overwrite)
	for i := range toCreate {
		sk := &model.LlmopsSkill{
			Name: toCreate[i].Name, Type: "extended", Desc: toCreate[i].Desc,
			DefJSON: toCreate[i].DefJSON, Status: "enabled",
		}
		if err := global.GVA_DB.Create(sk).Error; err != nil {
			return imported, conflicts, err
		}
		imported++
	}
	for i := range toUpdate {
		if err := global.GVA_DB.Model(&model.LlmopsSkill{}).Where("name = ?", toUpdate[i].Name).
			Updates(map[string]any{"desc": toUpdate[i].Desc, "def_json": toUpdate[i].DefJSON}).Error; err != nil {
			return imported, conflicts, err
		}
		imported++
	}
	return imported, conflicts, nil
}
