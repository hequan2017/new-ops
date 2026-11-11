// Package service crossauth：权限生命周期（需求 3.7-8）
// 口径：有效期=到期日(yyyy-MM-dd)；扫描两档——已到期=回收（Revoked 置位+CST 时间）
// /剩余 ≤7 天=提醒（用户×角色幂等）；决策纯函数可测。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaLifecycleParam = 2913

// expiryAction 生命周期决策（纯函数，测试锚点）："回收"/"提醒"/"无"
func expiryAction(now, expiredAt time.Time) string {
	switch {
	case !expiredAt.Before(now):
		return "无" // 未到期（含当日）
	case now.Sub(expiredAt) <= 7*24*time.Hour && now.Sub(expiredAt) >= 0:
		return "提醒" // 到期前 7 天窗口（含当日到期后顺延语义：剩余 ≤7 天）
	default:
		return "回收"
	}
}

// daysLeft 剩余天数（纯函数辅助，向下取整）
func daysLeft(now, expiredAt time.Time) int {
	d := int(expiredAt.Sub(now).Hours() / 24)
	if d < 0 {
		d = 0
	}
	return d
}

// SaveCaGrantExpiry 登记有效期
func (s *CaService) SaveCaGrantExpiry(in *model.CaGrantExpiry) (*model.CaGrantExpiry, error) {
	if strings.TrimSpace(in.UserName) == "" {
		return nil, fmt.Errorf("[%d]用户必填", ErrCodeCaLifecycleParam)
	}
	if _, err := time.ParseInLocation("2006-01-02", in.ExpiredAt, time.Local); err != nil {
		return nil, fmt.Errorf("[%d]到期日格式须为 yyyy-MM-dd", ErrCodeCaLifecycleParam)
	}
	if in.Revoked {
		return nil, fmt.Errorf("[%d]新登记不可直接置回收", ErrCodeCaLifecycleParam)
	}
	var exist model.CaGrantExpiry
	if err := global.GVA_DB.Where("user_name = ? AND role_name = ? AND revoked = ?", in.UserName, in.RoleName, false).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"school": in.School, "expired_at": in.ExpiredAt,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// ScanCaGrantExpiry 到期扫描（回收+7 天提醒，幂等）
func (s *CaService) ScanCaGrantExpiry() ([]model.CaExpiryReminder, []model.CaGrantExpiry, error) {
	var grants []model.CaGrantExpiry
	if err := global.GVA_DB.Where("revoked = ?", false).Limit(1000).Find(&grants).Error; err != nil {
		return nil, nil, err
	}
	now := time.Now()
	cst := time.FixedZone("CST", 8*3600)
	reminders := make([]model.CaExpiryReminder, 0)
	revoked := make([]model.CaGrantExpiry, 0)
	for _, g := range grants {
		expiredAt, err := time.ParseInLocation("2006-01-02", g.ExpiredAt, time.Local)
		if err != nil {
			continue
		}
		switch expiryAction(now, expiredAt) {
		case "回收":
			if err := global.GVA_DB.Model(&g).Updates(map[string]any{
				"revoked": true, "revoked_at": now.In(cst).Format("2006-01-02 15:04"),
			}).Error; err != nil {
				return nil, nil, err
			}
			g.Revoked = true
			revoked = append(revoked, g)
		case "提醒":
			var count int64
			global.GVA_DB.Model(&model.CaExpiryReminder{}).
				Where("user_name = ? AND role_name = ?", g.UserName, g.RoleName).Count(&count)
			if count > 0 {
				continue
			}
			reminder := model.CaExpiryReminder{
				UserName: g.UserName, RoleName: g.RoleName,
				DaysLeft: daysLeft(now, expiredAt),
				RemindedAt: now.In(cst).Format("2006-01-02 15:04"),
			}
			if err := global.GVA_DB.Create(&reminder).Error; err != nil {
				return nil, nil, err
			}
			reminders = append(reminders, reminder)
		}
	}
	return reminders, revoked, nil
}

// GetCaGrantExpiries 有效期列表（用户过滤）
func (s *CaService) GetCaGrantExpiries(userName string) ([]model.CaGrantExpiry, error) {
	db := global.GVA_DB.Model(&model.CaGrantExpiry{})
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	var list []model.CaGrantExpiry
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// GetCaExpiryReminders 提醒列表
func (s *CaService) GetCaExpiryReminders() ([]model.CaExpiryReminder, error) {
	var list []model.CaExpiryReminder
	err := global.GVA_DB.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}
