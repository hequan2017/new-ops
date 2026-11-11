// Package service eduadmin：工单进度与督办（需求 3.6-36）
// 口径：时限=受理（创建）时间+时限类别小时数（复用 P25 TicketDeadlineHours）；
// 超时扫描三档——超时限=超时自动督办、剩余<6h=临近提醒、其余正常；
// 督办记录按 工单+等级 幂等（同一等级只督办一次）。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduProgressParam = 2840

// urgeLevel 督办等级（纯函数，测试锚点）：超时/临近/空
func urgeLevel(now, deadline time.Time) string {
	switch {
	case now.After(deadline):
		return "超时"
	case deadline.Sub(now) < 6*time.Hour:
		return "临近"
	default:
		return ""
	}
}

// ticketDeadlineTime 工单时限时点（纯函数，测试锚点）：创建时间+时限小时
func ticketDeadlineTime(createdAt time.Time, deadlineClass string) (time.Time, error) {
	hours, err := TicketDeadlineHours(deadlineClass)
	if err != nil {
		return time.Time{}, err
	}
	return createdAt.Add(time.Duration(hours) * time.Hour), nil
}

// AddEduTicketTimeline 新增时间线
func (s *EduadminService) AddEduTicketTimeline(in *model.EduTicketTimeline) (*model.EduTicketTimeline, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, in.TicketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduProgressParam)
	}
	if strings.TrimSpace(in.Content) == "" || strings.TrimSpace(in.Dept) == "" {
		return nil, fmt.Errorf("[%d]办理单位与进展说明必填", ErrCodeEduProgressParam)
	}
	if in.OccurredAt == "" {
		in.OccurredAt = time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04")
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduTicketTimelines 时间线列表（工单过滤，倒序）
func (s *EduadminService) GetEduTicketTimelines(ticketId uint) ([]model.EduTicketTimeline, error) {
	var list []model.EduTicketTimeline
	err := global.GVA_DB.Where("ticket_id = ?", ticketId).Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// ScanEduTicketOverdue 超时扫描（分派/办理态工单，临近/超时幂等督办落库）
func (s *EduadminService) ScanEduTicketOverdue() ([]model.EduTicketUrge, error) {
	var tickets []model.EduTicket
	if err := global.GVA_DB.Where("status IN ?", []string{"分派", "办理"}).Limit(1000).Find(&tickets).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	cst := time.FixedZone("CST", 8*3600)
	created := make([]model.EduTicketUrge, 0)
	for _, t := range tickets {
		deadline, err := ticketDeadlineTime(t.CreatedAt, t.DeadlineClass)
		if err != nil {
			continue
		}
		level := urgeLevel(now, deadline)
		if level == "" {
			continue
		}
		var count int64
		global.GVA_DB.Model(&model.EduTicketUrge{}).
			Where("ticket_id = ? AND level = ?", t.ID, level).Count(&count)
		if count > 0 {
			continue
		}
		urge := model.EduTicketUrge{
			TicketId: t.ID, Level: level,
			DeadlineAt: deadline.In(cst).Format("2006-01-02 15:04"),
			UrgedAt:    now.In(cst).Format("2006-01-02 15:04"),
		}
		if err := global.GVA_DB.Create(&urge).Error; err != nil {
			return nil, err
		}
		created = append(created, urge)
	}
	return created, nil
}

// GetEduTicketUrges 督办记录列表
func (s *EduadminService) GetEduTicketUrges(level string) ([]model.EduTicketUrge, error) {
	db := global.GVA_DB.Model(&model.EduTicketUrge{})
	if level != "" && level != "临近" && level != "超时" {
		return nil, fmt.Errorf("[%d]等级须为 临近/超时", ErrCodeEduProgressParam)
	}
	if level != "" {
		db = db.Where("level = ?", level)
	}
	var list []model.EduTicketUrge
	err := db.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}
