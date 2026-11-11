// Package service crossauth：算力使用防护（需求 3.7-12）
// 口径：单用户 1 小时窗口累计核时——>1000 核时=限流阻断、>800（80%）=降优先级
// 标记、其余放行（决策纯函数）；用量按 用户×小时窗口 幂等累计；越阈自动告警留痕。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaComputeParam = 2921

// 算力阈值口径
const (
	CaCoreHoursLimit     = 1000.0 // 单用户 1 小时核时上限
	CaCoreHoursSoftLimit = 800.0  // 降优先级阈值（80%）
)

// computeDecision 阈值决策（纯函数，测试锚点）："放行"/"降优先级"/"限流"
func computeDecision(coreHours float64) string {
	switch {
	case coreHours > CaCoreHoursLimit:
		return "限流"
	case coreHours > CaCoreHoursSoftLimit:
		return "降优先级"
	default:
		return "放行"
	}
}

// RecordCaComputeUsage 记录用量（用户×小时窗口累计+决策）
func (s *CaService) RecordCaComputeUsage(userName string, gpuHours, coreHours float64) (map[string]any, error) {
	if strings.TrimSpace(userName) == "" || coreHours < 0 || gpuHours < 0 {
		return nil, fmt.Errorf("[%d]用户必填且用量非负", ErrCodeCaComputeParam)
	}
	window := time.Now().In(caCst).Format("2006-01-02 15")
	var usage model.CaComputeUsage
	err := global.GVA_DB.Where("user_name = ? AND window_hour = ?", userName, window).First(&usage).Error
	if err == nil {
		if err := global.GVA_DB.Model(&usage).Updates(map[string]any{
			"gpu_hours": usage.GpuHours + gpuHours, "core_hours": usage.CoreHours + coreHours,
		}).Error; err != nil {
			return nil, err
		}
		usage.GpuHours += gpuHours
		usage.CoreHours += coreHours
	} else {
		usage = model.CaComputeUsage{
			UserName: userName, WindowHour: window,
			GpuHours: gpuHours, CoreHours: coreHours,
		}
		if err := global.GVA_DB.Create(&usage).Error; err != nil {
			return nil, err
		}
	}
	decision := computeDecision(usage.CoreHours)
	out := map[string]any{
		"userName": userName, "windowHour": window,
		"coreHours": usage.CoreHours, "decision": decision,
		"limit":     CaCoreHoursLimit, "softLimit": CaCoreHoursSoftLimit,
	}
	if decision != "放行" {
		alert := model.CaComputeAlert{
			UserName: userName, WindowHour: window, CoreHours: usage.CoreHours,
			Action: decision,
			Reason: fmt.Sprintf("小时窗口核时 %.0f 触发阈值（限流 >%.0f / 降优先级 >%.0f）", usage.CoreHours, CaCoreHoursLimit, CaCoreHoursSoftLimit),
		}
		if err := global.GVA_DB.Create(&alert).Error; err != nil {
			return nil, err
		}
		out["alerted"] = true
	}
	return out, nil
}

// GetCaComputeUsages 用量列表（用户过滤）
func (s *CaService) GetCaComputeUsages(userName string) ([]model.CaComputeUsage, error) {
	db := global.GVA_DB.Model(&model.CaComputeUsage{})
	if userName != "" {
		db = db.Where("user_name = ?", userName)
	}
	var list []model.CaComputeUsage
	err := db.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}

// GetCaComputeAlerts 算力告警列表
func (s *CaService) GetCaComputeAlerts(handled string) ([]model.CaComputeAlert, error) {
	db := global.GVA_DB.Model(&model.CaComputeAlert{})
	if handled == "yes" || handled == "true" {
		db = db.Where("handled = ?", true)
	} else if handled == "no" || handled == "false" {
		db = db.Where("handled = ?", false)
	}
	var list []model.CaComputeAlert
	err := db.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}

// HandleCaComputeAlert 处置
func (s *CaService) HandleCaComputeAlert(id uint) error {
	res := global.GVA_DB.Model(&model.CaComputeAlert{}).Where("id = ?", id).Update("handled", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]告警不存在", ErrCodeCaComputeParam)
	}
	return nil
}
