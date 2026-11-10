// Package service eduadmin：六大知识库（需求 3.6-4）
// 口径：六域白名单（人事/财务/资产/后勤/决策/跨校）；检索=查询 bigram 对标题+内容
// 命中计数降序（对齐 ziqiang O07 口径），纯函数可测；教育管理指标库种子幂等。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduPolicyParam   = 2805
	ErrCodeEduPolicyNotFound = 2806
)

// EduPolicyDomains 六域白名单
var EduPolicyDomains = []string{"人事", "财务", "资产", "后勤", "决策", "跨校"}

// ValidEduPolicyDomain 六域白名单
func ValidEduPolicyDomain(d string) bool {
	for _, v := range EduPolicyDomains {
		if v == d {
			return true
		}
	}
	return false
}

// eduBigramSet 文本 bigram 集合（rune 级，对齐 ziqiang 口径）
func eduBigramSet(text string) map[string]bool {
	runes := []rune(text)
	set := make(map[string]bool, len(runes))
	for i := 0; i+1 < len(runes); i++ {
		set[string(runes[i:i+2])] = true
	}
	return set
}

// policyHit 检索命中项
type policyHit struct {
	model.EduPolicy
	Score int `json:"score"`
}

// rankPolicies bigram 命中降序（纯函数，测试锚点）
func rankPolicies(list []model.EduPolicy, query string) []policyHit {
	qSet := eduBigramSet(query)
	hits := make([]policyHit, 0, len(list))
	for _, p := range list {
		pSet := eduBigramSet(p.Title + p.Content)
		score := 0
		for bg := range qSet {
			if pSet[bg] {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, policyHit{EduPolicy: p, Score: score})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].ID < hits[j].ID
	})
	return hits
}

// SearchEduPolicies 知识库检索（域过滤 + bigram 打分）
func (s *EduadminService) SearchEduPolicies(query, domain string, topK int) ([]policyHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("[%d]检索词必填", ErrCodeEduPolicyParam)
	}
	if domain != "" && !ValidEduPolicyDomain(domain) {
		return nil, fmt.Errorf("[%d]域须为人事/财务/资产/后勤/决策/跨校", ErrCodeEduPolicyParam)
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	db := global.GVA_DB.Model(&model.EduPolicy{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	var list []model.EduPolicy
	if err := db.Limit(2000).Find(&list).Error; err != nil {
		return nil, err
	}
	hits := rankPolicies(list, query)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

// SaveEduPolicy 新建/更新条目
func (s *EduadminService) SaveEduPolicy(in *model.EduPolicy) (*model.EduPolicy, error) {
	if strings.TrimSpace(in.Title) == "" || !ValidEduPolicyDomain(in.Domain) {
		return nil, fmt.Errorf("[%d]标题必填且域须在白名单内", ErrCodeEduPolicyParam)
	}
	if in.ID > 0 {
		var exist model.EduPolicy
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]条目不存在", ErrCodeEduPolicyNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"domain": in.Domain, "title": in.Title, "content": in.Content, "tags": in.Tags,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteEduPolicy 删除条目（软删除）
func (s *EduadminService) DeleteEduPolicy(id uint) error {
	res := global.GVA_DB.Delete(&model.EduPolicy{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]条目不存在", ErrCodeEduPolicyNotFound)
	}
	return nil
}

// GetEduPolicies 条目列表（域过滤，倒序分页）
func (s *EduadminService) GetEduPolicies(domain string, page, pageSize int) ([]model.EduPolicy, int64, error) {
	if domain != "" && !ValidEduPolicyDomain(domain) {
		return nil, 0, fmt.Errorf("[%d]域须为人事/财务/资产/后勤/决策/跨校", ErrCodeEduPolicyParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduPolicy{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduPolicy
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
