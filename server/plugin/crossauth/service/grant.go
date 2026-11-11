// Package service crossauth：权限分配双模式（需求 3.7-6）
// 口径：自动匹配=院校精确+部门精确→通配 dept=* 兜底（纯函数）；人工审核流=
// 提交→本校审核→归属校审核→通过（任一环节可驳回），通过即建分配+推送留痕。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaGrantParam    = 2909
	ErrCodeCaGrantNotFound = 2910
)

// autoMatchRole 自动匹配（纯函数，测试锚点）：部门精确优先，通配 * 兜底
func autoMatchRole(rules []model.CaAutoGrantRule, school, dept string) *model.CaAutoGrantRule {
	var wildcard *model.CaAutoGrantRule
	for i := range rules {
		r := &rules[i]
		if r.School != school {
			continue
		}
		if r.Dept == dept {
			return r
		}
		if r.Dept == "*" && wildcard == nil {
			wildcard = r
		}
	}
	return wildcard
}

// grantNextStage 审核状态机（纯函数，测试锚点）
func grantNextStage(from string) (string, error) {
	switch from {
	case "提交":
		return "本校审核", nil
	case "本校审核":
		return "归属校审核", nil
	case "归属校审核":
		return "通过", nil
	default:
		return "", fmt.Errorf("[%d]状态 %s 不可推进", ErrCodeCaGrantParam, from)
	}
}

// AutoGrantCa 自动匹配分配（命中模板即建分配）
func (s *CaService) AutoGrantCa(userName, school, dept string) (*model.CaGrantAssignment, error) {
	if strings.TrimSpace(userName) == "" || strings.TrimSpace(school) == "" {
		return nil, fmt.Errorf("[%d]用户与院校必填", ErrCodeCaGrantParam)
	}
	var rules []model.CaAutoGrantRule
	if err := global.GVA_DB.Limit(500).Find(&rules).Error; err != nil {
		return nil, err
	}
	r := autoMatchRole(rules, school, dept)
	if r == nil {
		return nil, fmt.Errorf("[%d]无匹配规则，请走人工审核", ErrCodeCaGrantParam)
	}
	var tpl model.CaRoleTemplate
	if err := global.GVA_DB.Where("role_name = ?", r.RoleName).First(&tpl).Error; err != nil {
		return nil, fmt.Errorf("[%d]角色模板不存在: %s", ErrCodeCaGrantNotFound, r.RoleName)
	}
	assign := &model.CaGrantAssignment{
		UserName: userName, School: school, RoleName: r.RoleName,
		Source: "自动", PermJSON: tpl.PermJSON,
	}
	if err := global.GVA_DB.Create(assign).Error; err != nil {
		return nil, err
	}
	return assign, nil
}

// SubmitCaGrantRequest 提交人工申请
func (s *CaService) SubmitCaGrantRequest(in *model.CaGrantRequest) (*model.CaGrantRequest, error) {
	if strings.TrimSpace(in.UserName) == "" || strings.TrimSpace(in.DesiredRole) == "" {
		return nil, fmt.Errorf("[%d]申请人与角色必填", ErrCodeCaGrantParam)
	}
	in.Status = "提交"
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// ApproveCaGrantRequest 审核推进（本校/归属校两段；通过即建分配+推送留痕）
func (s *CaService) ApproveCaGrantRequest(id uint, approver string) (*model.CaGrantRequest, error) {
	if strings.TrimSpace(approver) == "" {
		return nil, fmt.Errorf("[%d]审核人必填", ErrCodeCaGrantParam)
	}
	var req model.CaGrantRequest
	if err := global.GVA_DB.First(&req, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]申请不存在", ErrCodeCaGrantNotFound)
	}
	next, err := grantNextStage(req.Status)
	if err != nil {
		return nil, err
	}
	updates := map[string]any{"status": next}
	switch next {
	case "本校审核":
		updates["home_approver"] = approver
	case "归属校审核":
		updates["host_approver"] = approver
	case "通过":
		// 建分配
		var tpl model.CaRoleTemplate
		if err := global.GVA_DB.Where("role_name = ?", req.DesiredRole).First(&tpl).Error; err == nil {
			assign := model.CaGrantAssignment{
				UserName: req.UserName, School: req.School, RoleName: req.DesiredRole,
				Source: "人工", PermJSON: tpl.PermJSON,
			}
			if err := global.GVA_DB.Create(&assign).Error; err != nil {
				return nil, err
			}
		}
		push := fmt.Sprintf("%s 审核通过已实时推送至申请人 %s 与资源归属校", approver, req.UserName)
		updates["push_log"] = time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04") + " " + push
	}
	if err := global.GVA_DB.Model(&req).Updates(updates).Error; err != nil {
		return nil, err
	}
	req.Status = next
	return &req, nil
}

// RejectCaGrantRequest 驳回（任一环节）
func (s *CaService) RejectCaGrantRequest(id uint, approver string) (*model.CaGrantRequest, error) {
	var req model.CaGrantRequest
	if err := global.GVA_DB.First(&req, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]申请不存在", ErrCodeCaGrantNotFound)
	}
	if req.Status == "通过" || req.Status == "驳回" {
		return nil, fmt.Errorf("[%d]终态不可驳回，当前 %s", ErrCodeCaGrantParam, req.Status)
	}
	if err := global.GVA_DB.Model(&req).Updates(map[string]any{
		"status": "驳回",
		"push_log": time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04") +
			fmt.Sprintf(" %s 驳回并实时推送至申请人 %s", approver, req.UserName),
	}).Error; err != nil {
		return nil, err
	}
	req.Status = "驳回"
	return &req, nil
}

// GetCaGrantRequests 申请列表（状态过滤）
func (s *CaService) GetCaGrantRequests(status string) ([]model.CaGrantRequest, error) {
	db := global.GVA_DB.Model(&model.CaGrantRequest{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []model.CaGrantRequest
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// GetCaGrantAssignments 分配列表（用户过滤）
func (s *CaService) GetCaGrantAssignments(userName string) ([]model.CaGrantAssignment, error) {
	db := global.GVA_DB.Model(&model.CaGrantAssignment{})
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	var list []model.CaGrantAssignment
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// SaveCaAutoGrantRule 保存自动匹配规则
func (s *CaService) SaveCaAutoGrantRule(in *model.CaAutoGrantRule) (*model.CaAutoGrantRule, error) {
	if strings.TrimSpace(in.School) == "" || strings.TrimSpace(in.RoleName) == "" {
		return nil, fmt.Errorf("[%d]院校与角色必填", ErrCodeCaGrantParam)
	}
	if in.Dept == "" {
		in.Dept = "*"
	}
	var exist model.CaAutoGrantRule
	if err := global.GVA_DB.Where("school = ? AND dept = ?", in.School, in.Dept).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Update("role_name", in.RoleName).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}
