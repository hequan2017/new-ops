// Package service crossauth：五维监控看板（需求 3.7-13）
// 口径：身份认证/权限变更/数据传输/资源使用/安全事件 五维聚合（各维计数+关键
// 汇总），五维结构纯函数可测。
package service

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// CaFiveDim 五维看板
type CaFiveDim struct {
	IdentityAuth struct {
		SyncUsers   int64 `json:"syncUsers"`
		SyncTasks24 int64 `json:"syncTasks24h"`
	} `json:"identityAuth"`
	PermChange struct {
		Assignments int64 `json:"assignments"`
		MatrixGrants int64 `json:"matrixGrants"`
	} `json:"permChange"`
	DataTransfer struct {
		Federations  int64 `json:"federations"`
		PulledTotal  int64 `json:"pulledTotal"`
	} `json:"dataTransfer"`
	ResourceUsage struct {
		UsageWindows int64 `json:"usageWindows"`
		CoreHours    float64 `json:"coreHours"`
	} `json:"resourceUsage"`
	SecurityEvents struct {
		AccessAlerts    int64 `json:"accessAlerts"`
		UnhandledAlerts int64 `json:"unhandledAlerts"`
		ComputeAlerts   int64 `json:"computeAlerts"`
	} `json:"securityEvents"`
	Healthy bool `json:"healthy"`
}

// fiveDimHealthy 健康判定（纯函数，测试锚点）：未处置告警=0 即健康
func fiveDimHealthy(unhandledAccess, unhandledCompute int64) bool {
	return unhandledAccess == 0 && unhandledCompute == 0
}

// GetCaFiveDim 五维聚合
func (s *CaService) GetCaFiveDim() (*CaFiveDim, error) {
	out := &CaFiveDim{}
	db := global.GVA_DB
	if err := db.Model(&model.CaSyncUser{}).Count(&out.IdentityAuth.SyncUsers).Error; err != nil {
		return nil, err
	}
	// 24h 窗口语义：仅统计近 24 小时内执行的任务（修正此前名不符实的全量计数）
	if err := db.Model(&model.CaSyncTask{}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&out.IdentityAuth.SyncTasks24).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.CaGrantAssignment{}).Count(&out.PermChange.Assignments).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.CaPermGrant{}).Count(&out.PermChange.MatrixGrants).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.CaFederation{}).Count(&out.DataTransfer.Federations).Error; err != nil {
		return nil, err
	}
	type pulledRow struct{ Total int64 }
	var pulled pulledRow
	if err := db.Model(&model.CaSyncTask{}).Select("COALESCE(SUM(pulled),0) as total").Scan(&pulled).Error; err != nil {
		return nil, err
	}
	out.DataTransfer.PulledTotal = pulled.Total
	if err := db.Model(&model.CaComputeUsage{}).Count(&out.ResourceUsage.UsageWindows).Error; err != nil {
		return nil, err
	}
	var cores struct{ Total float64 }
	if err := db.Model(&model.CaComputeUsage{}).Select("COALESCE(SUM(core_hours),0) as total").Scan(&cores).Error; err != nil {
		return nil, err
	}
	out.ResourceUsage.CoreHours = cores.Total
	if err := db.Model(&model.CaAlert{}).Count(&out.SecurityEvents.AccessAlerts).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.CaAlert{}).Where("handled = ?", false).Count(&out.SecurityEvents.UnhandledAlerts).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&model.CaComputeAlert{}).Count(&out.SecurityEvents.ComputeAlerts).Error; err != nil {
		return nil, err
	}
	out.Healthy = fiveDimHealthy(out.SecurityEvents.UnhandledAlerts, 0)
	return out, nil
}
