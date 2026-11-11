// Package service crossauth：数据分级引擎（需求 3.7-9）
// 口径：分类=关键词映射（教学/科研/管理，多者胜平局先命中）；分级三级取最高——
// 敏感（身份证/成绩单/薪酬/病历）>内部（项目/预算/内部/未公开）>公开（默认）；
// 批量扫描落结果表（CST 时间）。分级纯函数可测。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaClassifyParam = 2914

// CaLevels 分级白名单（低→高）
var CaLevels = []string{"公开", "内部", "敏感"}

// CaCategories 分类白名单
var CaCategories = []string{"教学", "科研", "管理"}

// classifyKeywords 分级关键词规则库（高→低优先）
var classifyKeywords = map[string][]string{
	"敏感": {"身份证", "成绩单", "薪酬", "病历", "学号+姓名", "银行卡"},
	"内部": {"项目", "预算", "内部", "未公开", "会议纪要", "工资"},
}

// classifyKeywordsCategory 分类关键词规则库
var classifyCategoryKeywords = map[string][]string{
	"教学": {"课程", "成绩", "教学", "课件", "实训", "考试"},
	"科研": {"论文", "专利", "实验", "科研", "课题", "成果"},
	"管理": {"人事", "财务", "后勤", "预算", "档案", "考核"},
}

// classifyData 数据分级（纯函数，测试锚点）：分级取最高、分类多者胜平局先命中
func classifyData(text string) (category, level, hitRule string) {
	for _, lv := range []string{"敏感", "内部"} {
		for _, kw := range classifyKeywords[lv] {
			if strings.Contains(text, kw) {
				level = lv
				hitRule = "关键词:" + kw
				break
			}
		}
		if level != "" {
			break
		}
	}
	if level == "" {
		level = "公开"
		hitRule = "默认公开"
	}
	best, bestCnt := "", 0
	order := make([]string, 0, len(classifyCategoryKeywords))
	for cat := range classifyCategoryKeywords {
		order = append(order, cat)
	}
	// 固定顺序保证平局确定性：按 CaCategories 白名单序
	for _, cat := range CaCategories {
		cnt := 0
		for _, kw := range classifyCategoryKeywords[cat] {
			if strings.Contains(text, kw) {
				cnt++
			}
		}
		if cnt > bestCnt {
			best, bestCnt = cat, cnt
		}
	}
	_ = order
	if best == "" {
		best = "管理" // 默认归管理类
	}
	return best, level, hitRule
}

// ClassifyScanItem 扫描入参（导出供 API 层绑定）
type ClassifyScanItem struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ScanCaClassify 批量扫描（分级落库）
func (s *CaService) ScanCaClassify(items []ClassifyScanItem) ([]model.CaDataClassify, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("[%d]扫描列表必填", ErrCodeCaClassifyParam)
	}
	if len(items) > 500 {
		return nil, fmt.Errorf("[%d]单次扫描 ≤500 条", ErrCodeCaClassifyParam)
	}
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04")
	out := make([]model.CaDataClassify, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.Name) == "" {
			return nil, fmt.Errorf("[%d]资源名必填", ErrCodeCaClassifyParam)
		}
		category, level, hitRule := classifyData(it.Name + " " + it.Content)
		rec := model.CaDataClassify{
			ResourceName: it.Name, Category: category, Level: level,
			HitRule: hitRule, ScannedAt: now,
		}
		if err := global.GVA_DB.Create(&rec).Error; err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

// GetCaClassifyResults 分级结果列表（分级/分类过滤）
func (s *CaService) GetCaClassifyResults(level, category string) ([]model.CaDataClassify, error) {
	db := global.GVA_DB.Model(&model.CaDataClassify{})
	if level != "" {
		valid := false
		for _, v := range CaLevels {
			if v == level {
				valid = true
			}
		}
		if !valid {
			return nil, fmt.Errorf("[%d]分级须为 公开/内部/敏感", ErrCodeCaClassifyParam)
		}
		db = db.Where("level = ?", level)
	}
	if category != "" {
		valid := false
		for _, v := range CaCategories {
			if v == category {
				valid = true
			}
		}
		if !valid {
			return nil, fmt.Errorf("[%d]分类须为 教学/科研/管理", ErrCodeCaClassifyParam)
		}
		db = db.Where("category = ?", category)
	}
	var list []model.CaDataClassify
	err := db.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}
