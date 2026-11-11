// Package service crossauth：跨校联合运维（需求 3.7-15）
// 口径：工单状态机 提交→分配→处理→闭环（线性，非法跳转拒绝）；协同消息文字口径
// （CST 时间）；远程协助授权记录（范围 只读/操作级，可撤销）。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaOpsParam    = 2923
	ErrCodeCaOpsNotFound = 2924
)

// caOpsNext 状态机（纯函数，测试锚点）
func caOpsNext(from string) (string, error) {
	switch from {
	case "提交":
		return "分配", nil
	case "分配":
		return "处理", nil
	case "处理":
		return "闭环", nil
	default:
		return "", fmt.Errorf("[%d]状态 %s 不可推进", ErrCodeCaOpsParam, from)
	}
}

func caOpsNow() string {
	return time.Now().In(caCst).Format("2006-01-02 15:04:05")
}

// CreateCaOpsTicket 创建工单
func (s *CaService) CreateCaOpsTicket(school, title string) (*model.CaOpsTicket, error) {
	if strings.TrimSpace(school) == "" || strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("[%d]院校与标题必填", ErrCodeCaOpsParam)
	}
	var count int64
	global.GVA_DB.Model(&model.CaOpsTicket{}).Count(&count)
	ticket := &model.CaOpsTicket{
		TicketNo: fmt.Sprintf("CAOPS-%04d", count+1),
		School:   school, Title: title, Status: "提交",
	}
	if err := global.GVA_DB.Create(ticket).Error; err != nil {
		return nil, err
	}
	return ticket, nil
}

// AdvanceCaOpsTicket 推进（分配需承接方）
func (s *CaService) AdvanceCaOpsTicket(id uint, assignee string) (*model.CaOpsTicket, error) {
	var ticket model.CaOpsTicket
	if err := global.GVA_DB.First(&ticket, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeCaOpsNotFound)
	}
	next, err := caOpsNext(ticket.Status)
	if err != nil {
		return nil, err
	}
	updates := map[string]any{"status": next}
	if next == "分配" {
		if strings.TrimSpace(assignee) == "" {
			return nil, fmt.Errorf("[%d]进入分配须指定承接运维方", ErrCodeCaOpsParam)
		}
		updates["assignee"] = assignee
	}
	if err := global.GVA_DB.Model(&ticket).Updates(updates).Error; err != nil {
		return nil, err
	}
	ticket.Status = next
	return &ticket, nil
}

// PostCaOpsMessage 发协同消息
func (s *CaService) PostCaOpsMessage(ticketId uint, from, content string) (*model.CaOpsMessage, error) {
	var ticket model.CaOpsTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeCaOpsNotFound)
	}
	if strings.TrimSpace(from) == "" || strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("[%d]发送方与内容必填", ErrCodeCaOpsParam)
	}
	msg := &model.CaOpsMessage{
		TicketId: ticketId, From: from, Content: content, SentAt: caOpsNow(),
	}
	if err := global.GVA_DB.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// GetCaOpsMessages 工单消息（升序）
func (s *CaService) GetCaOpsMessages(ticketId uint) ([]model.CaOpsMessage, error) {
	var list []model.CaOpsMessage
	err := global.GVA_DB.Where("ticket_id = ?", ticketId).Order("id ASC").Limit(500).Find(&list).Error
	return list, err
}

// GrantCaRemoteAssist 远程协助授权
func (s *CaService) GrantCaRemoteAssist(ticketId uint, requester, approver, scope string) (*model.CaRemoteAssist, error) {
	var ticket model.CaOpsTicket
	if err := global.GVA_DB.First(&ticket, ticketId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeCaOpsNotFound)
	}
	if strings.TrimSpace(requester) == "" || strings.TrimSpace(approver) == "" {
		return nil, fmt.Errorf("[%d]申请方与授权方必填", ErrCodeCaOpsParam)
	}
	if scope != "只读" && scope != "操作级" {
		return nil, fmt.Errorf("[%d]授权范围须为 只读/操作级", ErrCodeCaOpsParam)
	}
	assist := &model.CaRemoteAssist{
		TicketId: ticketId, Requester: requester, Approver: approver, Scope: scope,
		AuthorizedAt: caOpsNow(),
	}
	if err := global.GVA_DB.Create(assist).Error; err != nil {
		return nil, err
	}
	return assist, nil
}

// RevokeCaRemoteAssist 撤销授权
func (s *CaService) RevokeCaRemoteAssist(id uint) error {
	res := global.GVA_DB.Model(&model.CaRemoteAssist{}).Where("id = ? AND revoked = ?", id, false).
		Updates(map[string]any{"revoked": true, "authorized_at": caOpsNow()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]授权不存在或已撤销", ErrCodeCaOpsNotFound)
	}
	return nil
}

// GetCaOpsTickets 工单列表（状态/院校过滤）
func (s *CaService) GetCaOpsTickets(status, school string) ([]model.CaOpsTicket, error) {
	db := global.GVA_DB.Model(&model.CaOpsTicket{})
	if status != "" {
		if status != "提交" && status != "分配" && status != "处理" && status != "闭环" {
			return nil, fmt.Errorf("[%d]状态须为 提交/分配/处理/闭环", ErrCodeCaOpsParam)
		}
		db = db.Where("status = ?", status)
	}
	if school != "" {
		db = db.Where("school = ?", school)
	}
	var list []model.CaOpsTicket
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// GetCaRemoteAssists 授权记录列表（工单过滤）
func (s *CaService) GetCaRemoteAssists(ticketId uint) ([]model.CaRemoteAssist, error) {
	db := global.GVA_DB.Model(&model.CaRemoteAssist{})
	if ticketId > 0 {
		db = db.Where("ticket_id = ?", ticketId)
	}
	var list []model.CaRemoteAssist
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}
