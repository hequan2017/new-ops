// Package service eduadmin：审计知识语料（需求 3.6-37）
// 口径：六域白名单（财务/采购/人事/资产/工程/数据安全）；检索=查询 bigram 对
// 标题+风险要点 命中计数降序（复用 P05 eduBigramSet），纯函数可测。
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
	ErrCodeEduAuditParam   = 2844
	ErrCodeEduAuditNotFound = 2845
)

// EduAuditDomains 六域白名单
var EduAuditDomains = []string{"财务", "采购", "人事", "资产", "工程", "数据安全"}

// ValidEduAuditDomain 六域白名单
func ValidEduAuditDomain(d string) bool {
	for _, v := range EduAuditDomains {
		if v == d {
			return true
		}
	}
	return false
}

// auditHit 检索命中项
type auditHit struct {
	model.EduAuditItem
	Score int `json:"score"`
}

// rankAuditItems bigram 命中降序（纯函数，测试锚点）
func rankAuditItems(list []model.EduAuditItem, query string) []auditHit {
	qSet := eduBigramSet(query)
	hits := make([]auditHit, 0, len(list))
	for _, a := range list {
		aSet := eduBigramSet(a.Title + a.RiskPoints)
		score := 0
		for bg := range qSet {
			if aSet[bg] {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, auditHit{EduAuditItem: a, Score: score})
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

// SearchEduAudit 审计知识检索（域过滤 + bigram 打分，辅助审计与内控信息化）
func (s *EduadminService) SearchEduAudit(query, domain string, topK int) ([]auditHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("[%d]检索词必填", ErrCodeEduAuditParam)
	}
	if domain != "" && !ValidEduAuditDomain(domain) {
		return nil, fmt.Errorf("[%d]域须为 财务/采购/人事/资产/工程/数据安全", ErrCodeEduAuditParam)
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	db := global.GVA_DB.Model(&model.EduAuditItem{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	var list []model.EduAuditItem
	if err := db.Limit(2000).Find(&list).Error; err != nil {
		return nil, err
	}
	hits := rankAuditItems(list, query)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

// SaveEduAuditItem 新建/更新条目
func (s *EduadminService) SaveEduAuditItem(in *model.EduAuditItem) (*model.EduAuditItem, error) {
	if strings.TrimSpace(in.Title) == "" || !ValidEduAuditDomain(in.Domain) {
		return nil, fmt.Errorf("[%d]标题必填且域须在白名单内", ErrCodeEduAuditParam)
	}
	if in.ID > 0 {
		var exist model.EduAuditItem
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]条目不存在", ErrCodeEduAuditNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"domain": in.Domain, "title": in.Title,
			"risk_points": in.RiskPoints, "control_advice": in.ControlAdvice,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduAuditItems 条目列表（域过滤，倒序分页）
func (s *EduadminService) GetEduAuditItems(domain string, page, pageSize int) ([]model.EduAuditItem, int64, error) {
	if domain != "" && !ValidEduAuditDomain(domain) {
		return nil, 0, fmt.Errorf("[%d]域须为 财务/采购/人事/资产/工程/数据安全", ErrCodeEduAuditParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduAuditItem{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduAuditItem
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
