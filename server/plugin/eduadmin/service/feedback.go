// Package service eduadmin：反馈闭环（需求 3.6-36）
// 口径：办结态可评价（好评/差评白名单）；差评自动回流——状态回退 办理 并写
// 时间线留痕（异议工单同口径）；评价态归档终态；评价结果纯函数可测。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduReviewParam = 2842

// reviewOutcome 评价结果（纯函数，测试锚点）：好评→评价态可归档；差评→回流办理
func reviewOutcome(score string) (string, error) {
	switch score {
	case "好评":
		return "评价", nil
	case "差评":
		return "办理", nil
	default:
		return "", fmt.Errorf("[%d]评分须为 好评/差评", ErrCodeEduReviewParam)
	}
}

// ReviewEduTicket 评价工单（办结态；差评自动回流办理+时间线留痕）
func (s *EduadminService) ReviewEduTicket(ticketId uint, score, comment string) (*model.EduTicketReview, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduReviewParam)
	}
	if ticket.Status != "办结" {
		return nil, fmt.Errorf("[%d]仅办结工单可评价，当前 %s", ErrCodeEduReviewParam, ticket.Status)
	}
	next, err := reviewOutcome(score)
	if err != nil {
		return nil, err
	}
	review := &model.EduTicketReview{
		TicketId: ticketId, Score: score, Comment: comment,
		ReviewedAt: time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04"),
	}
	if err := global.GVA_DB.Create(review).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ticket).Update("status", next).Error; err != nil {
		return nil, err
	}
	if score == "差评" {
		timeline := model.EduTicketTimeline{
			TicketId: ticketId, Dept: "12345 平台", Operator: "系统自动",
			Content:    fmt.Sprintf("差评自动回流：评语「%s」，请重新办理并再次提交办结", comment),
			OccurredAt: review.ReviewedAt,
		}
		if err := global.GVA_DB.Create(&timeline).Error; err != nil {
			return nil, err
		}
	}
	return review, nil
}

// ArchiveEduTicket 归档（评价态→归档）
func (s *EduadminService) ArchiveEduTicket(ticketId uint) (*model.EduTicket, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduReviewParam)
	}
	if ticket.Status != "评价" {
		return nil, fmt.Errorf("[%d]仅评价态工单可归档，当前 %s", ErrCodeEduReviewParam, ticket.Status)
	}
	if err := global.GVA_DB.Model(&ticket).Update("status", "归档").Error; err != nil {
		return nil, err
	}
	ticket.Status = "归档"
	return &ticket, nil
}

// GetEduTicketReviews 评价列表（评分过滤，倒序分页）
func (s *EduadminService) GetEduTicketReviews(score string, page, pageSize int) ([]model.EduTicketReview, int64, error) {
	if score != "" && score != "好评" && score != "差评" {
		return nil, 0, fmt.Errorf("[%d]评分须为 好评/差评", ErrCodeEduReviewParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduTicketReview{})
	if strings.TrimSpace(score) != "" {
		db = db.Where("score = ?", score)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduTicketReview
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
