// Package service policyreg：实体构建（需求 3.8-5）
// 口径：机构=机构词典 + 后缀规则（大学/学院/委员会/办公会 收尾的前缀片段）；
// 文件号=〔yyyy〕N号 规则正则；日期=yyyy年M月d日 全量抽取；去重保序，纯函数可测。
package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const ErrCodePrEntityParam = 3005

// orgDictionary 机构词典
var orgDictionary = []string{
	"教育部", "上海市教育委员会", "上海市人民政府", "国务院",
	"学位评定委员会", "学术委员会", "校长办公会", "教职工代表大会",
	"教务处", "财务处", "人事处", "研究生院", "发展规划处",
}

// docNoPattern 文件号正则：X〔yyyy〕N号
var docNoPattern = regexp.MustCompile(`[^\s，。、；：]{1,12}〔([0-9]{4})〕([0-9]{1,6})号`)

// datePatternEntity 日期正则
var datePatternEntity = regexp.MustCompile(`[0-9]{4}年[0-9]{1,2}月[0-9]{1,2}日`)

// orgSuffixPattern 机构后缀规则：前缀 2-10 个非标点字符 + 后缀
var orgSuffixPattern = regexp.MustCompile(`[^\s，。、；：]{2,10}(大学|学院|委员会|办公会|研究院|研究所)`)

// extractPolicyEntities 实体抽取（纯函数，测试锚点）：去重保序
func extractPolicyEntities(text string) []model.PolicyEntity {
	seen := map[string]bool{}
	out := make([]model.PolicyEntity, 0, 16)
	add := func(name, eType, source string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name+eType] {
			return
		}
		seen[name+eType] = true
		out = append(out, model.PolicyEntity{Name: name, EType: eType, Source: source})
	}
	for _, org := range orgDictionary {
		if strings.Contains(text, org) {
			add(org, "机构", "词典")
		}
	}
	for _, m := range orgSuffixPattern.FindAllStringSubmatch(text, -1) {
		add(m[0], "机构", "规则")
	}
	for _, m := range docNoPattern.FindAllString(text, -1) {
		add(m, "文件号", "规则")
	}
	for _, m := range datePatternEntity.FindAllString(text, -1) {
		add(m, "日期", "规则")
	}
	return out
}

// ExtractPolicyEntities 实体抽取（按 docId 幂等覆盖）
func (s *PolicyService) ExtractPolicyEntities(docId uint) ([]model.PolicyEntity, error) {
	var doc model.PolicyDoc
	if err := global.GVA_DB.First(&doc, docId).Error; err != nil {
		return nil, fmt.Errorf("[%d]制度不存在", ErrCodePrEntityParam)
	}
	text := doc.RawText
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("[%d]二进制编目文档无可抽取文本", ErrCodePrEntityParam)
	}
	entities := extractPolicyEntities(text)
	// 幂等：先清后插
	if err := global.GVA_DB.Where("doc_id = ?", docId).Delete(&model.PolicyEntity{}).Error; err != nil {
		return nil, err
	}
	for i := range entities {
		entities[i].DocId = docId
		if err := global.GVA_DB.Create(&entities[i]).Error; err != nil {
			return nil, err
		}
	}
	return entities, nil
}

// GetPolicyEntities 实体列表（类型过滤）
func (s *PolicyService) GetPolicyEntities(docId uint, eType string) ([]model.PolicyEntity, error) {
	db := global.GVA_DB.Model(&model.PolicyEntity{}).Where("doc_id = ?", docId)
	if eType != "" {
		db = db.Where("e_type = ?", eType)
	}
	var list []model.PolicyEntity
	err := db.Order("id ASC").Limit(500).Find(&list).Error
	return list, err
}
