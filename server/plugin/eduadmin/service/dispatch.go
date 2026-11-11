// Package service eduadmin：智能分派（需求 3.6-36）
// 口径：关键词命中计数×规则优先级排序（命中数>0 取最高分规则→自动分派并推进
// 受理→分派）；零命中工单留在待派池（受理态）由人工兜底；分派记录落库。
// 规则种子：教务/后勤/财务/信息化/人事五类。
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
	ErrCodeEduDispatchParam    = 2838
	ErrCodeEduDispatchNotFound = 2839
)

// dispatchCandidate 分派候选
type dispatchCandidate struct {
	Rule  model.EduDispatchRule
	Hits  int
	Score int // 命中数×1000+优先级
}

// matchDispatchRule 规则匹配（纯函数，测试锚点）：命中计数降序+优先级，零命中返回 nil
func matchDispatchRule(rules []model.EduDispatchRule, content string) *dispatchCandidate {
	lower := strings.ToLower(content)
	cands := make([]dispatchCandidate, 0, len(rules))
	for _, r := range rules {
		hits := 0
		for _, kw := range strings.Split(r.Keywords, ",") {
			k := strings.ToLower(strings.TrimSpace(kw))
			if k != "" && strings.Contains(lower, k) {
				hits++
			}
		}
		if hits > 0 {
			cands = append(cands, dispatchCandidate{Rule: r, Hits: hits, Score: hits*1000 + r.Priority})
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].Score != cands[j].Score {
			return cands[i].Score > cands[j].Score
		}
		return cands[i].Rule.ID < cands[j].Rule.ID
	})
	return &cands[0]
}

// AutoDispatchEduTicket 自动分派（零命中留待派池）
func (s *EduadminService) AutoDispatchEduTicket(ticketId uint) (*model.EduDispatchLog, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduDispatchNotFound)
	}
	if ticket.Status != "受理" {
		return nil, fmt.Errorf("[%d]仅受理态工单可分派，当前 %s", ErrCodeEduDispatchParam, ticket.Status)
	}
	var rules []model.EduDispatchRule
	if err := global.GVA_DB.Order("priority DESC").Limit(200).Find(&rules).Error; err != nil {
		return nil, err
	}
	cand := matchDispatchRule(rules, ticket.Content+" "+ticket.DemandType)
	log := &model.EduDispatchLog{TicketId: ticketId, Auto: true}
	if cand == nil {
		log.RuleId = 0
		log.Dept = ""
		log.Reason = "零命中，留待派池人工兜底"
		if err := global.GVA_DB.Create(log).Error; err != nil {
			return nil, err
		}
		return log, nil
	}
	log.RuleId = cand.Rule.ID
	log.Dept = cand.Rule.Dept
	log.Reason = fmt.Sprintf("命中 %d 个关键词（%s），优先级 %d", cand.Hits, cand.Rule.Keywords, cand.Rule.Priority)
	if err := global.GVA_DB.Create(log).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ticket).Update("status", "分派").Error; err != nil {
		return nil, err
	}
	return log, nil
}

// ManualDispatchEduTicket 人工兜底分派
func (s *EduadminService) ManualDispatchEduTicket(ticketId uint, dept string) (*model.EduDispatchLog, error) {
	if strings.TrimSpace(dept) == "" {
		return nil, fmt.Errorf("[%d]承办部门必填", ErrCodeEduDispatchParam)
	}
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduDispatchNotFound)
	}
	if ticket.Status != "受理" {
		return nil, fmt.Errorf("[%d]仅受理态工单可分派，当前 %s", ErrCodeEduDispatchParam, ticket.Status)
	}
	log := &model.EduDispatchLog{
		TicketId: ticketId, RuleId: 0, Dept: dept, Auto: false,
		Reason: "待派池人工兜底分派",
	}
	if err := global.GVA_DB.Create(log).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ticket).Update("status", "分派").Error; err != nil {
		return nil, err
	}
	return log, nil
}

// GetEduPendingTickets 待派池（受理态工单）
func (s *EduadminService) GetEduPendingTickets(page, pageSize int) ([]model.EduTicket, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduTicket{}).Where("status = ?", "受理")
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduTicket
	err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetEduDispatchLogs 分派记录（工单过滤，倒序分页）
func (s *EduadminService) GetEduDispatchLogs(ticketId uint, page, pageSize int) ([]model.EduDispatchLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduDispatchLog{})
	if ticketId > 0 {
		db = db.Where("ticket_id = ?", ticketId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduDispatchLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveEduDispatchRule 新建/更新规则
func (s *EduadminService) SaveEduDispatchRule(in *model.EduDispatchRule) (*model.EduDispatchRule, error) {
	if strings.TrimSpace(in.Keywords) == "" || strings.TrimSpace(in.Dept) == "" {
		return nil, fmt.Errorf("[%d]关键词与承办部门必填", ErrCodeEduDispatchParam)
	}
	if in.ID > 0 {
		var exist model.EduDispatchRule
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]规则不存在", ErrCodeEduDispatchNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"keywords": in.Keywords, "dept": in.Dept, "priority": in.Priority,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}
