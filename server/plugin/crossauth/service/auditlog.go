// Package service crossauth：三级审计（需求 3.7-14）
// 口径：统一事件表+层级过滤——校级=全量（可按校筛选）、院系级=校+院系必填、
// 用户级=用户必填；层级白名单校验；事件登记 CST 时间。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaAuditEventParam = 2922

// CaAuditLevels 层级白名单
var CaAuditLevels = []string{"校级", "院系级", "用户级"}

// ValidCaAuditLevel 层级白名单
func ValidCaAuditLevel(l string) bool {
	for _, v := range CaAuditLevels {
		if v == l {
			return true
		}
	}
	return false
}

// RecordCaAuditEvent 登记事件
func (s *CaService) RecordCaAuditEvent(in *model.CaAuditEvent) (*model.CaAuditEvent, error) {
	if !ValidCaAuditLevel(in.Level) {
		return nil, fmt.Errorf("[%d]层级须为 校级/院系级/用户级", ErrCodeCaAuditEventParam)
	}
	if strings.TrimSpace(in.Action) == "" {
		return nil, fmt.Errorf("[%d]动作必填", ErrCodeCaAuditEventParam)
	}
	if in.OccurredAt == "" {
		in.OccurredAt = time.Now().In(caCst).Format("2006-01-02 15:04:05")
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaAuditEvents 分层查询
func (s *CaService) GetCaAuditEvents(level, school, dept, userName string, page, pageSize int) ([]model.CaAuditEvent, int64, error) {
	if !ValidCaAuditLevel(level) {
		return nil, 0, fmt.Errorf("[%d]层级须为 校级/院系级/用户级", ErrCodeCaAuditEventParam)
	}
	// 层级口径约束
	switch level {
	case "院系级":
		if school == "" || dept == "" {
			return nil, 0, fmt.Errorf("[%d]院系级查询须指定院校与院系", ErrCodeCaAuditEventParam)
		}
	case "用户级":
		if strings.TrimSpace(userName) == "" {
			return nil, 0, fmt.Errorf("[%d]用户级查询须指定用户", ErrCodeCaAuditEventParam)
		}
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	db := global.GVA_DB.Model(&model.CaAuditEvent{}).Where("level = ?", level)
	if school != "" {
		db = db.Where("school = ?", school)
	}
	if dept != "" {
		db = db.Where("dept = ?", dept)
	}
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CaAuditEvent
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
