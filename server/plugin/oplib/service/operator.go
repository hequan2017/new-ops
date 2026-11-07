// Package service oplib：算子管理（注册/分类/schema 校验；执行引擎与编排由后续条目接入）
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
)

// 错误码（oplib 段 2300-2399）
const (
	ErrCodeOpNotFound = 2300
	ErrCodeOpCatBad   = 2301
	ErrCodeOpImplBad  = 2302
	ErrCodeOpSchema   = 2303
	ErrCodeOpNameDup  = 2304
)

// OperatorCategories 算子分类白名单
var OperatorCategories = []string{"数据处理", "特征", "检索", "后处理", "评估"}

// ValidCategory 分类白名单
func ValidCategory(c string) bool {
	for _, v := range OperatorCategories {
		if v == c {
			return true
		}
	}
	return false
}

// ValidImplType 实现类型白名单
func ValidImplType(t string) bool {
	return t == "builtin" || t == "script"
}

// ValidateSchema schema 须为 JSON 对象（字段→类型/说明）
func ValidateSchema(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return fmt.Errorf("[%d]schema 须为 JSON 对象: %w", ErrCodeOpSchema, err)
	}
	return nil
}

// OplibApp 服务单例出口（算子域服务聚合）
var OplibApp = new(OplibService)

type OplibService struct{}

// GetOperators 算子列表（分类/实现过滤）
func (s *OplibService) GetOperators(category, implType string) ([]model.OplibOperator, error) {
	db := global.GVA_DB.Model(&model.OplibOperator{})
	if category != "" {
		if !ValidCategory(category) {
			return nil, fmt.Errorf("[%d]不支持的算子分类: %s", ErrCodeOpCatBad, category)
		}
		db = db.Where("category = ?", category)
	}
	if implType != "" {
		if !ValidImplType(implType) {
			return nil, fmt.Errorf("[%d]不支持的实现类型: %s", ErrCodeOpImplBad, implType)
		}
		db = db.Where("impl_type = ?", implType)
	}
	var list []model.OplibOperator
	err := db.Order("category, id").Find(&list).Error
	return list, err
}

// SaveOperator 新建/更新算子
func (s *OplibService) SaveOperator(in *model.OplibOperator) (*model.OplibOperator, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidCategory(in.Category) || !ValidImplType(in.ImplType) {
		return nil, fmt.Errorf("[%d]名称必填且分类/实现类型须在白名单内", ErrCodeOpCatBad)
	}
	if err := ValidateSchema(in.InputSchema); err != nil {
		return nil, err
	}
	if err := ValidateSchema(in.OutputSchema); err != nil {
		return nil, err
	}
	if err := saveScriptFields(in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Version) == "" {
		in.Version = "v1"
	}
	var count int64
	q := global.GVA_DB.Model(&model.OplibOperator{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]算子名已存在", ErrCodeOpNameDup)
	}
	if in.ID > 0 {
		var op model.OplibOperator
		if err := global.GVA_DB.First(&op, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]算子不存在", ErrCodeOpNotFound)
		}
		err := global.GVA_DB.Model(&op).Updates(map[string]any{
			"name": in.Name, "category": in.Category, "impl_type": in.ImplType,
			"input_schema": in.InputSchema, "output_schema": in.OutputSchema,
			"script_lang": in.ScriptLang, "script_code": in.ScriptCode,
			"version": in.Version, "desc": in.Desc,
		}).Error
		if err != nil {
			return nil, err
		}
		return &op, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteOperator 删除算子（软删除）
func (s *OplibService) DeleteOperator(id uint) error {
	res := global.GVA_DB.Delete(&model.OplibOperator{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]算子不存在", ErrCodeOpNotFound)
	}
	return nil
}
