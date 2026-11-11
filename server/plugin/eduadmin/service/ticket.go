// Package service eduadmin：AI-12345 工单（需求 3.6-36）
// 口径：唯一编号 T-yyyyMMdd-序号；渠道白名单（网页/企微/小程序）；时限类别
// （简单24h=24h/一般3工作日=72h/复杂7工作日=168h，纯函数）；状态机线性流转
// 受理→分派→办理→办结→评价→归档（非法跳转拒绝）。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduTicketParam    = 2836
	ErrCodeEduTicketNotFound = 2837
)

// EduTicketChannels 渠道白名单
var EduTicketChannels = []string{"网页", "企微", "小程序"}

// EduTicketStatuses 状态机节点（线性）
var EduTicketStatuses = []string{"受理", "分派", "办理", "办结", "评价", "归档"}

// EduDeadlineClasses 时限类别白名单
var EduDeadlineClasses = []string{"简单24h", "一般3工作日", "复杂7工作日"}

// TicketDeadlineHours 时限类别→时限小时（纯函数，测试锚点）
func TicketDeadlineHours(class string) (int, error) {
	switch class {
	case "简单24h":
		return 24, nil
	case "一般3工作日":
		return 72, nil
	case "复杂7工作日":
		return 168, nil
	default:
		return 0, fmt.Errorf("[%d]时限类别须为 简单24h/一般3工作日/复杂7工作日", ErrCodeEduTicketParam)
	}
}

// nextTicketStatus 状态机下一节点（纯函数，测试锚点）
func nextTicketStatus(from string) (string, error) {
	for i, st := range EduTicketStatuses {
		if st == from {
			if i+1 >= len(EduTicketStatuses) {
				return "", fmt.Errorf("[%d]已处于终态 %s", ErrCodeEduTicketParam, from)
			}
			return EduTicketStatuses[i+1], nil
		}
	}
	return "", fmt.Errorf("[%d]未知状态: %s", ErrCodeEduTicketParam, from)
}

func validTicketStatus(st string) bool {
	for _, s := range EduTicketStatuses {
		if s == st {
			return true
		}
	}
	return false
}

// genEduTicketNo 工单号：T-yyyyMMdd-序号
func (s *EduadminService) genEduTicketNo() string {
	today := time.Now().Format("20060102")
	var count int64
	global.GVA_DB.Model(&model.EduTicket{}).Where("ticket_no LIKE ?", "T-"+today+"-%").Count(&count)
	return fmt.Sprintf("T-%s-%04d", today, count+1)
}

// CreateEduTicket 创建工单（受理）
func (s *EduadminService) CreateEduTicket(in *model.EduTicket) (*model.EduTicket, error) {
	validChannel := false
	for _, c := range EduTicketChannels {
		if c == in.Channel {
			validChannel = true
		}
	}
	if !validChannel {
		return nil, fmt.Errorf("[%d]渠道须为 网页/企微/小程序", ErrCodeEduTicketParam)
	}
	if strings.TrimSpace(in.Content) == "" || strings.TrimSpace(in.DemandType) == "" {
		return nil, fmt.Errorf("[%d]诉求类型与内容必填", ErrCodeEduTicketParam)
	}
	if _, err := TicketDeadlineHours(in.DeadlineClass); err != nil {
		return nil, err
	}
	in.Status = "受理"
	in.TicketNo = s.genEduTicketNo()
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// AdvanceEduTicket 状态推进（线性下一节点）
func (s *EduadminService) AdvanceEduTicket(id uint) (*model.EduTicket, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduTicketNotFound)
	}
	next, err := nextTicketStatus(ticket.Status)
	if err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&ticket).Update("status", next).Error; err != nil {
		return nil, err
	}
	ticket.Status = next
	return &ticket, nil
}

// GetEduTickets 工单列表（渠道/状态过滤，倒序分页）
func (s *EduadminService) GetEduTickets(channel, status string, page, pageSize int) ([]model.EduTicket, int64, error) {
	if channel != "" {
		valid := false
		for _, c := range EduTicketChannels {
			if c == channel {
				valid = true
			}
		}
		if !valid {
			return nil, 0, fmt.Errorf("[%d]渠道须为 网页/企微/小程序", ErrCodeEduTicketParam)
		}
	}
	if status != "" && !validTicketStatus(status) {
		return nil, 0, fmt.Errorf("[%d]状态须在状态机内", ErrCodeEduTicketParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduTicket{})
	if channel != "" {
		db = db.Where("channel = ?", channel)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduTicket
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetEduTicket 工单详情
func (s *EduadminService) GetEduTicket(id uint) (*model.EduTicket, error) {
	var ticket model.EduTicket
	if err := global.GVA_DB.First(&ticket, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduTicketNotFound)
	}
	return &ticket, nil
}
