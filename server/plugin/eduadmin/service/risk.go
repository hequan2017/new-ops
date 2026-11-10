// Package service eduadmin：财务风险预警（需求 3.6-17）
// 口径：支出 3σ 异常（已支付发票金额复用 clean.go sigmaOutliers，n≥2 防误报）；
// 现金流阈值（余额/下限比 <1 高、<1.2 中）；预警记录按 type+refKey 幂等去重。
package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduRiskParam = 2815

// RiskScanReq 扫描请求（导出供 API 层绑定）
type RiskScanReq struct {
	CashBalance float64 `json:"cashBalance"`
	CashFloor   float64 `json:"cashFloor"`
}

// riskLevel 现金流预警等级（纯函数，测试锚点）
func cashRiskLevel(balance, floor float64) string {
	if floor <= 0 {
		return "低"
	}
	ratio := balance / floor
	switch {
	case ratio < 1:
		return "高"
	case ratio < 1.2:
		return "中"
	default:
		return "低"
	}
}

// ScanEduRisk 执行财务风险扫描（支出 3σ + 现金流阈值，幂等去重落库）
func (s *EduadminService) ScanEduRisk(in RiskScanReq) ([]model.EduRiskAlert, error) {
	created := make([]model.EduRiskAlert, 0)
	// ①已支付发票金额 3σ 异常
	var invoices []model.EduInvoice
	if err := global.GVA_DB.Where("status = ?", "支付").Order("id ASC").Limit(2000).Find(&invoices).Error; err != nil {
		return nil, err
	}
	amounts := make([]float64, len(invoices))
	for i, inv := range invoices {
		amounts[i] = inv.Amount
	}
	outliers := sigmaOutliers(amounts)
	for idx := range outliers {
		inv := invoices[idx]
		key := fmt.Sprintf("invoice:%s", inv.InvoiceNo)
		if s.riskExists("支出异常", key) {
			continue
		}
		alert := model.EduRiskAlert{
			Type: "支出异常", RefKey: key, Value: inv.Amount, Threshold: 3,
			Level: "高",
			Detail: fmt.Sprintf("部门 %s 发票 %s 金额 %g 元偏离同批已支付金额 3σ", inv.Dept, inv.InvoiceNo, inv.Amount),
		}
		if err := global.GVA_DB.Create(&alert).Error; err != nil {
			return nil, err
		}
		created = append(created, alert)
	}
	// ②现金流阈值
	if in.CashFloor > 0 {
		if in.CashBalance < 0 {
			return nil, fmt.Errorf("[%d]现金余额不可为负", ErrCodeEduRiskParam)
		}
		level := cashRiskLevel(in.CashBalance, in.CashFloor)
		if level != "低" {
			key := fmt.Sprintf("cash:%.0f", math.Floor(in.CashBalance))
			if !s.riskExists("现金流不足", key) {
				alert := model.EduRiskAlert{
					Type: "现金流不足", RefKey: key, Value: in.CashBalance, Threshold: in.CashFloor,
					Level:  level,
					Detail: fmt.Sprintf("现金余额 %g 低于下限 %g（%.0f%%）", in.CashBalance, in.CashFloor, in.CashBalance/in.CashFloor*100),
				}
				if err := global.GVA_DB.Create(&alert).Error; err != nil {
					return nil, err
				}
				created = append(created, alert)
			}
		}
	}
	return created, nil
}

func (s *EduadminService) riskExists(alertType, refKey string) bool {
	var count int64
	global.GVA_DB.Model(&model.EduRiskAlert{}).Where("type = ? AND ref_key = ?", alertType, refKey).Count(&count)
	return count > 0
}

// GetEduRiskAlerts 预警列表（类型/处理状态过滤，倒序分页）
func (s *EduadminService) GetEduRiskAlerts(alertType string, handled string, page, pageSize int) ([]model.EduRiskAlert, int64, error) {
	if alertType != "" && alertType != "支出异常" && alertType != "现金流不足" {
		return nil, 0, fmt.Errorf("[%d]类型须为 支出异常/现金流不足", ErrCodeEduRiskParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduRiskAlert{})
	if alertType != "" {
		db = db.Where("type = ?", alertType)
	}
	if strings.TrimSpace(handled) != "" {
		if handled == "yes" || handled == "true" {
			db = db.Where("handled = ?", true)
		} else if handled == "no" || handled == "false" {
			db = db.Where("handled = ?", false)
		}
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduRiskAlert
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// HandleEduRiskAlert 标记处理
func (s *EduadminService) HandleEduRiskAlert(id uint) error {
	res := global.GVA_DB.Model(&model.EduRiskAlert{}).Where("id = ?", id).Update("handled", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]预警不存在", ErrCodeEduRiskParam)
	}
	return nil
}
