// Package service crossauth：SSO 单点登录（需求 3.7-4）
// 口径：一次性票据=SHA256(随机数|appKey|用户|时间戳) 64 位 hex，60s 有效；
// 换取防重放（已用/过期/应用不符/票据不存在 四路拒绝，成功即置 used）；
// 跳转 URL=回调地址拼接 ticket+appKey（纯函数）。
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaSsoParam    = 2906
	ErrCodeCaSsoNotFound = 2907
)

// CaTicketTTL 票据有效期（秒）
const CaTicketTTL = 60

// ticketExpired 票据过期判定（纯函数，测试锚点）
func ticketExpired(now, expiredAt int64) bool {
	return now > expiredAt
}

// buildSsoRedirect 跳转 URL 组装（纯函数，测试锚点）
func buildSsoRedirect(callbackUrl, ticket, appKey string) (string, error) {
	u, err := url.Parse(callbackUrl)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("[%d]回调地址须为合法 http(s) URL", ErrCodeCaSsoParam)
	}
	q := u.Query()
	q.Set("ticket", ticket)
	q.Set("appKey", appKey)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// IssueCaSsoTicket 签发一次性票据
func (s *CaService) IssueCaSsoTicket(appKey, userName string) (*model.CaSsoTicket, string, error) {
	if strings.TrimSpace(appKey) == "" || strings.TrimSpace(userName) == "" {
		return nil, "", fmt.Errorf("[%d]应用标识与用户必填", ErrCodeCaSsoParam)
	}
	var app model.CaSsoApp
	if err := global.GVA_DB.Where("app_key = ?", appKey).First(&app).Error; err != nil {
		return nil, "", fmt.Errorf("[%d]应用未注册: %s", ErrCodeCaSsoNotFound, appKey)
	}
	if app.Status != "启用" {
		return nil, "", fmt.Errorf("[%d]应用已停用: %s", ErrCodeCaSsoParam, appKey)
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d", hex.EncodeToString(nonce), appKey, userName, time.Now().UnixNano())))
	ticket := &model.CaSsoTicket{
		Ticket: hex.EncodeToString(sum[:]), AppKey: appKey, UserName: userName,
		ExpiredAt: time.Now().Unix() + CaTicketTTL,
	}
	if err := global.GVA_DB.Create(ticket).Error; err != nil {
		return nil, "", err
	}
	redirect, err := buildSsoRedirect(app.CallbackUrl, ticket.Ticket, appKey)
	if err != nil {
		return nil, "", err
	}
	return ticket, redirect, nil
}

// RedeemCaSsoTicket 换取票据（防重放）
func (s *CaService) RedeemCaSsoTicket(ticketStr, appKey string) (*model.CaSsoTicket, error) {
	if strings.TrimSpace(ticketStr) == "" {
		return nil, fmt.Errorf("[%d]票据必填", ErrCodeCaSsoParam)
	}
	var ticket model.CaSsoTicket
	if err := global.GVA_DB.Where("ticket = ?", ticketStr).First(&ticket).Error; err != nil {
		return nil, fmt.Errorf("[%d]票据不存在", ErrCodeCaSsoNotFound)
	}
	if ticket.Used {
		return nil, fmt.Errorf("[%d]票据已使用（防重放拒绝）", ErrCodeCaSsoParam)
	}
	if ticketExpired(time.Now().Unix(), ticket.ExpiredAt) {
		return nil, fmt.Errorf("[%d]票据已过期", ErrCodeCaSsoParam)
	}
	if appKey != "" && ticket.AppKey != appKey {
		return nil, fmt.Errorf("[%d]票据与应用不符", ErrCodeCaSsoParam)
	}
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05")
	if err := global.GVA_DB.Model(&ticket).Updates(map[string]any{"used": true, "used_at": now}).Error; err != nil {
		return nil, err
	}
	ticket.Used = true
	ticket.UsedAt = now
	return &ticket, nil
}

// SaveCaSsoApp 注册/更新应用
func (s *CaService) SaveCaSsoApp(in *model.CaSsoApp) (*model.CaSsoApp, error) {
	if strings.TrimSpace(in.AppName) == "" || strings.TrimSpace(in.AppKey) == "" {
		return nil, fmt.Errorf("[%d]应用名与标识必填", ErrCodeCaSsoParam)
	}
	if _, err := buildSsoRedirect(in.CallbackUrl, "t", in.AppKey); err != nil {
		return nil, err
	}
	if in.Status == "" {
		in.Status = "启用"
	}
	var exist model.CaSsoApp
	if err := global.GVA_DB.Where("app_key = ?", in.AppKey).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"app_name": in.AppName, "callback_url": in.CallbackUrl, "status": in.Status,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaSsoApps 应用列表
func (s *CaService) GetCaSsoApps() ([]model.CaSsoApp, error) {
	var list []model.CaSsoApp
	err := global.GVA_DB.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// CleanupCaSsoTickets 过期票据清理（纯决策复用 ticketExpired，删除 已用或过期 票据）
func (s *CaService) CleanupCaSsoTickets(keepHours int) (map[string]any, error) {
	if keepHours < 1 {
		keepHours = 24 // 已用票据默认保留 24h 供审计
	}
	cutoff := time.Now().Unix() - int64(keepHours)*3600
	var usedCnt, expiredCnt int64
	global.GVA_DB.Model(&model.CaSsoTicket{}).Where("used = ?", true).Count(&usedCnt)
	global.GVA_DB.Model(&model.CaSsoTicket{}).Where("used = ? AND expired_at < ?", false, cutoff).Count(&expiredCnt)
	res := global.GVA_DB.Where("used = ? OR expired_at < ?", true, cutoff).Delete(&model.CaSsoTicket{})
	if res.Error != nil {
		return nil, res.Error
	}
	return map[string]any{
		"deleted": res.RowsAffected, "usedRemoved": usedCnt, "expiredRemoved": expiredCnt,
		"keepHours": keepHours,
	}, nil
}

// CaSsoTicketStats 票据存量统计
type CaSsoTicketStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"` // 未用未过期
	Used   int64 `json:"used"`
	Expired int64 `json:"expired"`
}

// CaSsoTicketInventory 票据存量（纯函数，测试锚点）
func CaSsoTicketInventory(total, used, expired int64) CaSsoTicketStats {
	active := total - used - expired
	if active < 0 {
		active = 0
	}
	return CaSsoTicketStats{Total: total, Active: active, Used: used, Expired: expired}
}

// GetCaSsoTicketStats 票据存量统计
func (s *CaService) GetCaSsoTicketStats() (*CaSsoTicketStats, error) {
	now := time.Now().Unix()
	var total, used, expired int64
	if err := global.GVA_DB.Model(&model.CaSsoTicket{}).Count(&total).Error; err != nil {
		return nil, err
	}
	global.GVA_DB.Model(&model.CaSsoTicket{}).Where("used = ?", true).Count(&used)
	global.GVA_DB.Model(&model.CaSsoTicket{}).Where("used = ? AND expired_at < ?", false, now).Count(&expired)
	out := CaSsoTicketInventory(total, used, expired)
	return &out, nil
}
