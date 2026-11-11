// Package service crossauth：身份互认规则库（需求 3.7-2）
// 口径：校验=身份精确匹配 + 资源包含匹配（规则 TargetResource 含查询资源即命中）
// + allowed 判定；多条命中时任一 allowed=true 即放行（白名单并集）；种子 8 条幂等。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaTrustParam   = 2902
	ErrCodeCaTrustNotFound = 2903
)

// trustVerdict 互认判定（纯函数，测试锚点）
func trustVerdict(rules []model.CaTrustRule, identity, resource string) (bool, string) {
	if strings.TrimSpace(identity) == "" || strings.TrimSpace(resource) == "" {
		return false, "身份与资源必填"
	}
	allowed := false
	matched := 0
	for _, r := range rules {
		if r.HomeIdentity != identity {
			continue
		}
		if !strings.Contains(r.TargetResource, resource) && !strings.Contains(resource, r.TargetResource) {
			continue
		}
		matched++
		if r.Allowed {
			allowed = true
		}
	}
	if matched == 0 {
		return false, fmt.Sprintf("无匹配规则（身份 %s × 资源 %s）", identity, resource)
	}
	if !allowed {
		return false, "命中规则均为拒绝"
	}
	return true, fmt.Sprintf("命中 %d 条规则，互认放行", matched)
}

// CheckCaTrust 互认校验
func (s *CaService) CheckCaTrust(identity, resource string) (map[string]any, error) {
	var rules []model.CaTrustRule
	if err := global.GVA_DB.Limit(1000).Find(&rules).Error; err != nil {
		return nil, err
	}
	allow, reason := trustVerdict(rules, identity, resource)
	return map[string]any{"allowed": allow, "reason": reason}, nil
}

// SaveCaTrustRule 新建/更新规则
func (s *CaService) SaveCaTrustRule(in *model.CaTrustRule) (*model.CaTrustRule, error) {
	if strings.TrimSpace(in.HomeIdentity) == "" || strings.TrimSpace(in.TargetResource) == "" {
		return nil, fmt.Errorf("[%d]本校身份与目标资源必填", ErrCodeCaTrustParam)
	}
	if in.ID > 0 {
		var exist model.CaTrustRule
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]规则不存在", ErrCodeCaTrustNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"home_identity": in.HomeIdentity, "target_resource": in.TargetResource,
			"scope": in.Scope, "allowed": in.Allowed, "remark": in.Remark,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaTrustRules 规则列表（身份过滤）
func (s *CaService) GetCaTrustRules(identity string) ([]model.CaTrustRule, error) {
	db := global.GVA_DB.Model(&model.CaTrustRule{})
	if identity != "" {
		db = db.Where("home_identity = ?", identity)
	}
	var list []model.CaTrustRule
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// DeleteCaTrustRule 删除规则
func (s *CaService) DeleteCaTrustRule(id uint) error {
	res := global.GVA_DB.Delete(&model.CaTrustRule{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]规则不存在", ErrCodeCaTrustNotFound)
	}
	return nil
}
