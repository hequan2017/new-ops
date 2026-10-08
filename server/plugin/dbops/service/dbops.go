// Package service 白泽数据库工单：实例纳管与 SQL 工单（M6 dbops）
// dbops 插件错误码段：1700-1799（DEV_PLAN 3.6）
package service

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/dbops/model"
	"gorm.io/gorm"
)

// 错误码
const (
	ErrCodeInstNameRequired  = 1701
	ErrCodeInstDuplicate     = 1702
	ErrCodeInstAddrInvalid   = 1703
	ErrCodeInstNotFound      = 1704
	ErrCodeInstOffline       = 1705
	ErrCodeOrderInvalid      = 1706
	ErrCodeOrderNotFound     = 1707
	ErrCodeAuditNotConfigured = 1708
)

func newDbErr(code int, msg string) error { return &dbopsError{Code: code, Msg: msg} }

type dbopsError struct {
	Code int
	Msg  string
}

func (e *dbopsError) Error() string { return e.Msg }

// DbopsService 服务出口
type DbopsService struct{}

var DbopsServiceI = new(DbopsService)

// validateInstance 实例校验（纯逻辑，可单测）
func validateInstance(inst *model.DbopsInstance) error {
	if inst == nil || strings.TrimSpace(inst.Name) == "" {
		return newDbErr(ErrCodeInstNameRequired, "实例名称不能为空")
	}
	if net.ParseIP(strings.TrimSpace(inst.Host)) == nil && strings.ContainsAny(inst.Host, "/\\") {
		return newDbErr(ErrCodeInstAddrInvalid, "主机格式非法: "+inst.Host)
	}
	if inst.Port <= 0 || inst.Port > 65535 {
		inst.Port = 3306
	}
	return nil
}

// CreateInstance 创建实例（密码 AES-GCM 加密落库）
func (s *DbopsService) CreateInstance(inst *model.DbopsInstance) error {
	if err := validateInstance(inst); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.DbopsInstance{}).Where("name = ?", inst.Name).Count(&count)
	if count > 0 {
		return newDbErr(ErrCodeInstDuplicate, "实例已存在: "+inst.Name)
	}
	if inst.PasswordEnc != "" {
		enc, err := crypto.Encrypt(inst.PasswordEnc)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		inst.PasswordEnc = enc
	}
	inst.Status = model.InstanceUnknown
	return global.GVA_DB.Create(inst).Error
}

// UpdateInstance 更新实例（密码留空不改）
func (s *DbopsService) UpdateInstance(inst *model.DbopsInstance) error {
	if inst == nil || inst.ID == 0 {
		return newDbErr(ErrCodeInstNameRequired, "实例ID不能为空")
	}
	if err := validateInstance(inst); err != nil {
		return err
	}
	updates := map[string]any{
		"name": inst.Name, "host": inst.Host, "port": inst.Port,
		"username": inst.Username, "database": inst.Database, "notes": inst.Notes,
	}
	if inst.PasswordEnc != "" {
		enc, err := crypto.Encrypt(inst.PasswordEnc)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		updates["password_enc"] = enc
	}
	return global.GVA_DB.Model(&model.DbopsInstance{}).Where("id = ?", inst.ID).Updates(updates).Error
}

// DeleteInstance 删除实例（有未结束工单拒绝）
func (s *DbopsService) DeleteInstance(id uint) error {
	var running int64
	global.GVA_DB.Model(&model.DbopsOrder{}).
		Where("instance_id = ? AND status IN ?", id,
			[]string{model.OrderPending, model.OrderApproved, model.OrderExecuting}).Count(&running)
	if running > 0 {
		return newDbErr(ErrCodeOrderInvalid, fmt.Sprintf("仍有 %d 个未结束工单，不可删除", running))
	}
	return global.GVA_DB.Delete(&model.DbopsInstance{}, id).Error
}

// GetInstanceList 实例列表
func (s *DbopsService) GetInstanceList(keyword, status string) (list []*model.DbopsInstance, err error) {
	db := global.GVA_DB.Model(&model.DbopsInstance{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR host LIKE ? OR notes LIKE ?", kw, kw, kw)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// GetInstance 详情（密码永不回显）
func (s *DbopsService) GetInstance(id uint) (*model.DbopsInstance, error) {
	var inst model.DbopsInstance
	if err := global.GVA_DB.First(&inst, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newDbErr(ErrCodeInstNotFound, "实例不存在")
		}
		return nil, err
	}
	inst.PasswordEnc = ""
	return &inst, nil
}

// GetInstanceCredential 取实例连接凭据（内部使用，密码解密即用即毁）
func (s *DbopsService) GetInstanceCredential(id uint) (host string, port int, user, password string, err error) {
	var inst model.DbopsInstance
	if e := global.GVA_DB.First(&inst, id).Error; e != nil {
		return "", 0, "", "", newDbErr(ErrCodeInstNotFound, "实例不存在")
	}
	password = ""
	if inst.PasswordEnc != "" {
		p, e := crypto.Decrypt(inst.PasswordEnc)
		if e != nil {
			return "", 0, "", "", fmt.Errorf("密码解密失败: %w", e)
		}
		password = p
	}
	return inst.Host, inst.Port, inst.Username, password, nil
}

// TestConnection 实例连通检测（TCP 探活级；SQL 层检测待 mysql 驱动引入）
func (s *DbopsService) TestConnection(id uint) (bool, string, error) {
	inst, err := s.GetInstance(id)
	if err != nil {
		return false, "", err
	}
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(inst.Host, strconv.Itoa(inst.Port)), 2*time.Second)
	now := time.Now()
	if err != nil {
		global.GVA_DB.Model(&model.DbopsInstance{}).Where("id = ?", id).
			Updates(map[string]any{"status": model.InstanceOffline, "last_check_at": now})
		return false, "", err
	}
	_ = conn.Close()
	global.GVA_DB.Model(&model.DbopsInstance{}).Where("id = ?", id).
		Updates(map[string]any{"status": model.InstanceOnline, "last_check_at": now})
	return true, "TCP 可达", nil
}

// ---------- SQL 工单 ----------

// CreateOrder 创建 SQL 工单（待审核）
func (s *DbopsService) CreateOrder(instID uint, title, sqlText, creator string, userID uint) (*model.DbopsOrder, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(sqlText) == "" {
		return nil, newDbErr(ErrCodeOrderInvalid, "标题与 SQL 不能为空")
	}
	var inst model.DbopsInstance
	if err := global.GVA_DB.First(&inst, instID).Error; err != nil {
		return nil, newDbErr(ErrCodeInstNotFound, "实例不存在")
	}
	order := &model.DbopsOrder{
		InstanceID: instID, InstanceNm: inst.Name,
		Title: title, AuditPayload: sqlText,
		Status: model.OrderPending, Creator: creator, UserID: userID,
	}
	return order, global.GVA_DB.Create(order).Error
}

// AuditOrder 已迁移至 inception.go（goInception 接入，M6 收官）

// CancelOrder 取消工单（待审核可取消）
func (s *DbopsService) CancelOrder(orderID uint, operator string) error {
	var order model.DbopsOrder
	if err := global.GVA_DB.First(&order, orderID).Error; err != nil {
		return newDbErr(ErrCodeOrderNotFound, "工单不存在")
	}
	if order.Status != model.OrderPending {
		return newDbErr(ErrCodeOrderInvalid, "仅待审核工单可取消")
	}
	now := time.Now()
	return global.GVA_DB.Model(&model.DbopsOrder{}).Where("id = ?", orderID).Updates(map[string]any{
		"status": model.OrderCanceled, "finished_at": &now,
	}).Error
}

// GetOrderList 工单分页（普通用户仅本人）
func (s *DbopsService) GetOrderList(page, pageSize int, creator string, userID uint, isSuperAdmin bool) (list []*model.DbopsOrder, total int64, err error) {
	db := global.GVA_DB.Model(&model.DbopsOrder{})
	if !isSuperAdmin {
		db = db.Where("user_id = ?", userID)
	}
	if creator != "" {
		db = db.Where("creator LIKE ?", "%"+creator+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	err = db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
