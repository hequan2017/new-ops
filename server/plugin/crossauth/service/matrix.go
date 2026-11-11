// Package service crossauth：四级权限矩阵（需求 3.7-7）
// 口径：校-部门-用户-场景 四级 grant（* 通配）；校验=命中授权里"非通配维度数"
// 最具体者优先（同级并列默认拒绝）；纯函数可测。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaMatrixParam = 2911

// specificity 非通配维度数（纯函数辅助）
func specificity(g model.CaPermGrant) int {
	n := 0
	for _, v := range []string{g.School, g.Dept, g.UserName, g.Scene} {
		if v != "*" && v != "" {
			n++
		}
	}
	return n
}

// matrixMatches 授权是否命中请求（纯函数辅助）
func matrixMatches(g model.CaPermGrant, school, dept, user, scene string) bool {
	match := func(pattern, val string) bool {
		return pattern == "*" || pattern == "" || pattern == val
	}
	return match(g.School, school) && match(g.Dept, dept) && match(g.UserName, user) && match(g.Scene, scene)
}

// matrixCheck 权限校验（纯函数，测试锚点）：最具体优先，无命中默认拒绝
func matrixCheck(grants []model.CaPermGrant, school, dept, user, scene string) (bool, int) {
	best := -1
	allowed := false
	for _, g := range grants {
		if !matrixMatches(g, school, dept, user, scene) {
			continue
		}
		sp := specificity(g)
		if sp > best {
			best = sp
			allowed = g.Allowed
		}
	}
	return allowed, best
}

// CheckCaMatrix 权限校验
func (s *CaService) CheckCaMatrix(school, dept, user, scene string) (map[string]any, error) {
	if strings.TrimSpace(scene) == "" {
		return nil, fmt.Errorf("[%d]场景必填", ErrCodeCaMatrixParam)
	}
	var grants []model.CaPermGrant
	if err := global.GVA_DB.Limit(2000).Find(&grants).Error; err != nil {
		return nil, err
	}
	allowed, level := matrixCheck(grants, school, dept, user, scene)
	return map[string]any{
		"allowed": allowed, "specificity": level,
		"reason": map[bool]string{true: "命中授权放行", false: "无命中或最具体授权拒绝"}[allowed],
	}, nil
}

// SaveCaPermGrant 保存授权
func (s *CaService) SaveCaPermGrant(in *model.CaPermGrant) (*model.CaPermGrant, error) {
	if strings.TrimSpace(in.Scene) == "" {
		return nil, fmt.Errorf("[%d]场景必填", ErrCodeCaMatrixParam)
	}
	for _, v := range []string{in.School, in.Dept, in.UserName, in.Scene} {
		if strings.TrimSpace(v) == "" {
			inSchoolDefaults(in)
			break
		}
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// inSchoolDefaults 空维度归一为通配（就地修正）
func inSchoolDefaults(in *model.CaPermGrant) {
	if strings.TrimSpace(in.School) == "" {
		in.School = "*"
	}
	if strings.TrimSpace(in.Dept) == "" {
		in.Dept = "*"
	}
	if strings.TrimSpace(in.UserName) == "" {
		in.UserName = "*"
	}
}

// GetCaPermGrants 授权列表（院校过滤）
func (s *CaService) GetCaPermGrants(school string) ([]model.CaPermGrant, error) {
	db := global.GVA_DB.Model(&model.CaPermGrant{})
	if school != "" {
		db = db.Where("school = ?", school)
	}
	var list []model.CaPermGrant
	err := db.Order("id ASC").Limit(500).Find(&list).Error
	return list, err
}

// DeleteCaPermGrant 删除授权
func (s *CaService) DeleteCaPermGrant(id uint) error {
	res := global.GVA_DB.Delete(&model.CaPermGrant{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]授权不存在", ErrCodeCaMatrixParam)
	}
	return nil
}
