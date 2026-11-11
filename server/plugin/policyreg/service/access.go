// Package service policyreg：模型鉴权知识库（需求 3.8-7）
// 口径：多角色精细 grant（角色×分类 唯一，MaxLevel=可访问最高密级）；密级序
// 公开(0)<内部(1)<涉密(2)；检索=关键字包含匹配×角色过滤（无 grant 或
// docLevel>MaxLevel 或 Allowed=false 均不可见），纯函数可测。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const ErrCodePrAclParam = 3008

// prLevelRank 密级序（纯函数辅助）
func prLevelRank(level string) int {
	switch level {
	case "公开":
		return 0
	case "内部":
		return 1
	case "涉密":
		return 2
	}
	return -1
}

// PolicyCategories 分类白名单
var PolicyCategories = []string{"教学", "科研", "人事", "财务", "资产", "综合"}

// canAccessRole 角色访问判定（纯函数，测试锚点）
func canAccessRole(grants []model.PolicyPermGrant, roleName, docCategory, docLevel string) bool {
	rank := prLevelRank(docLevel)
	if rank < 0 {
		return false
	}
	for _, g := range grants {
		if g.RoleName != roleName {
			continue
		}
		if g.DocCategory != "综合" && g.DocCategory != docCategory {
			continue
		}
		if !g.Allowed {
			continue
		}
		if rank <= prLevelRank(g.MaxLevel) {
			return true
		}
	}
	return false
}

// SavePolicyPermGrant 保存 grant（角色×分类幂等）
func (s *PolicyService) SavePolicyPermGrant(in *model.PolicyPermGrant) (*model.PolicyPermGrant, error) {
	if strings.TrimSpace(in.RoleName) == "" {
		return nil, fmt.Errorf("[%d]角色必填", ErrCodePrAclParam)
	}
	validCat := false
	for _, c := range PolicyCategories {
		if c == in.DocCategory {
			validCat = true
		}
	}
	if !validCat {
		return nil, fmt.Errorf("[%d]分类须为 教学/科研/人事/财务/资产/综合", ErrCodePrAclParam)
	}
	if prLevelRank(in.MaxLevel) < 0 {
		return nil, fmt.Errorf("[%d]最高密级须为 公开/内部/涉密", ErrCodePrAclParam)
	}
	var exist model.PolicyPermGrant
	if err := global.GVA_DB.Where("role_name = ? AND doc_category = ?", in.RoleName, in.DocCategory).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"max_level": in.MaxLevel, "allowed": in.Allowed,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// SearchPolicyDocsByRole 按角色过滤检索
func (s *PolicyService) SearchPolicyDocsByRole(roleName, keyword string, page, pageSize int) ([]model.PolicyDoc, int64, error) {
	if strings.TrimSpace(roleName) == "" {
		return nil, 0, fmt.Errorf("[%d]角色必填", ErrCodePrAclParam)
	}
	var grants []model.PolicyPermGrant
	if err := global.GVA_DB.Limit(500).Find(&grants).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.PolicyDoc{})
	if strings.TrimSpace(keyword) != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	var all []model.PolicyDoc
	if err := db.Order("id DESC").Limit(1000).Find(&all).Error; err != nil {
		return nil, 0, err
	}
	visible := make([]model.PolicyDoc, 0, len(all))
	for _, d := range all {
		if canAccessRole(grants, roleName, ClassifyPolicyLabel(d.Name), d.SecurityLevel) {
			visible = append(visible, d)
		}
	}
	total := int64(len(visible))
	start := (page - 1) * pageSize
	if start > len(visible) {
		start = len(visible)
	}
	end := start + pageSize
	if end > len(visible) {
		end = len(visible)
	}
	return visible[start:end], total, nil
}

// GetPolicyPermGrants grant 列表
func (s *PolicyService) GetPolicyPermGrants(roleName string) ([]model.PolicyPermGrant, error) {
	db := global.GVA_DB.Model(&model.PolicyPermGrant{})
	if roleName != "" {
		db = db.Where("role_name = ?", roleName)
	}
	var list []model.PolicyPermGrant
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}
