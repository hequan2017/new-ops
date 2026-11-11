// Package service crossauth：资源访问权限匹配与告警（需求 3.7-11）
// 口径：访问校验=矩阵(Q08)+互认(Q03)双门，任一拒绝即阻断并落告警（推送 归属校
// 管理员+安全团队，响应 ≤5 分钟口径字段）；放行仅留访问日志。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaAccessParam = 2919

// accessVerdict 双门判定（纯函数，测试锚点）
func accessVerdict(matrixAllowed bool, trustAllowed bool, matrixReason, trustReason string) (bool, string) {
	switch {
	case !matrixAllowed && !trustAllowed:
		return false, "矩阵拒绝(" + matrixReason + ")；互认拒绝(" + trustReason + ")"
	case !matrixAllowed:
		return false, "矩阵拒绝: " + matrixReason
	case !trustAllowed:
		return false, "互认拒绝: " + trustReason
	default:
		return true, "矩阵与互认均放行"
	}
}

// AccessCheckReq 访问校验入参（导出供 API 层绑定）
type AccessCheckReq struct {
	UserName string `json:"userName"`
	School   string `json:"school"`
	Dept     string `json:"dept"`
	Identity string `json:"identity"`
	Resource string `json:"resource"`
}

// CheckCaAccess 访问校验（双门+阻断告警）
func (s *CaService) CheckCaAccess(in AccessCheckReq) (map[string]any, error) {
	if strings.TrimSpace(in.Resource) == "" {
		return nil, fmt.Errorf("[%d]资源必填", ErrCodeCaAccessParam)
	}
	matrixOut, err := s.CheckCaMatrix(in.School, in.Dept, in.UserName, in.Resource)
	if err != nil {
		return nil, err
	}
	trustOut, err := s.CheckCaTrust(in.Identity, in.Resource)
	if err != nil {
		return nil, err
	}
	allowed, reason := accessVerdict(
		matrixOut["allowed"].(bool), trustOut["allowed"].(bool),
		fmt.Sprintf("%v", matrixOut["reason"]), fmt.Sprintf("%v", trustOut["reason"]),
	)
	log := model.CaAccessLog{
		UserName: in.UserName, School: in.School, Resource: in.Resource,
		Allowed: allowed, Reason: reason,
	}
	if err := global.GVA_DB.Create(&log).Error; err != nil {
		return nil, err
	}
	out := map[string]any{"allowed": allowed, "reason": reason, "logId": log.ID}
	if !allowed {
		alert := model.CaAlert{
			UserName: in.UserName, School: in.School, Resource: in.Resource,
			BlockReason: reason, PushedTo: "归属校管理员+安全团队", ResponseMinutes: 5,
		}
		if err := global.GVA_DB.Create(&alert).Error; err != nil {
			return nil, err
		}
		out["alertId"] = alert.ID
	}
	return out, nil
}

// GetCaAccessLogs 访问日志（用户过滤，倒序分页）
func (s *CaService) GetCaAccessLogs(userName string, page, pageSize int) ([]model.CaAccessLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.CaAccessLog{})
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CaAccessLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetCaAlerts 告警列表（处置状态过滤，倒序分页）
func (s *CaService) GetCaAlerts(handled string, page, pageSize int) ([]model.CaAlert, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.CaAlert{})
	if handled == "yes" || handled == "true" {
		db = db.Where("handled = ?", true)
	} else if handled == "no" || handled == "false" {
		db = db.Where("handled = ?", false)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CaAlert
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// HandleCaAlert 处置告警
func (s *CaService) HandleCaAlert(id uint) error {
	res := global.GVA_DB.Model(&model.CaAlert{}).Where("id = ?", id).Update("handled", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]告警不存在", ErrCodeCaAccessParam)
	}
	return nil
}
