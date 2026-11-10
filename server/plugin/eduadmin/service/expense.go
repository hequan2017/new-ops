// Package service eduadmin：自动报销（需求 3.6-16）
// 口径：类别限额（差旅 5000/办公 2000/设备 20000/招待 1000）；自动审核规则=
// 类别白名单+限额+重复报销（同发票号或同人同金额同日期）；提交即自动审核，
// 审核中可人工支付/驳回；状态机提交→审核中→支付/驳回（纯函数可测）。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduExpenseParam    = 2813
	ErrCodeEduExpenseNotFound = 2814
)

// ExpenseCategoryLimits 类别限额（元）
var ExpenseCategoryLimits = map[string]float64{
	"差旅": 5000,
	"办公": 2000,
	"设备": 20000,
	"招待": 1000,
}

// ExpenseStatuses 状态机节点
var ExpenseStatuses = []string{"提交", "审核中", "支付", "驳回"}

// auditResult 审核结论
type auditResult struct {
	Pass    bool     `json:"pass"`
	Reasons []string `json:"reasons"`
}

// auditInvoice 规则审核（纯函数，测试锚点）：全部规则通过才放行
func auditInvoice(in *model.EduInvoice) auditResult {
	reasons := make([]string, 0)
	limit, ok := ExpenseCategoryLimits[in.Category]
	if !ok {
		reasons = append(reasons, fmt.Sprintf("类别不在白名单: %s", in.Category))
	} else if in.Amount > limit {
		reasons = append(reasons, fmt.Sprintf("超类别限额: %g > %g 元", in.Amount, limit))
	}
	if in.Amount <= 0 {
		reasons = append(reasons, "金额须为正数")
	}
	if strings.TrimSpace(in.InvoiceNo) == "" {
		reasons = append(reasons, "发票号必填")
	}
	if strings.TrimSpace(in.InvoiceDate) == "" {
		reasons = append(reasons, "发票日期必填")
	}
	return auditResult{Pass: len(reasons) == 0, Reasons: reasons}
}

// dupCheck 重复报销检测（库内比对）
func dupCheck(in *model.EduInvoice) (string, bool) {
	var count int64
	global.GVA_DB.Model(&model.EduInvoice{}).Where("invoice_no = ?", in.InvoiceNo).Count(&count)
	if count > 0 {
		return "发票号已存在（重复报销）", true
	}
	global.GVA_DB.Model(&model.EduInvoice{}).Where(
		"applicant = ? AND amount = ? AND invoice_date = ? AND status <> ?",
		in.Applicant, in.Amount, in.InvoiceDate, "驳回").Count(&count)
	if count > 0 {
		return "同人同金额同日期已报销", true
	}
	return "", false
}

// SubmitEduExpense 提交报销（登记+自动规则审核）
func (s *EduadminService) SubmitEduExpense(in *model.EduInvoice) (*model.EduInvoice, error) {
	in.Status = "提交"
	in.RejectReason = ""
	if res := auditInvoice(in); !res.Pass {
		return nil, fmt.Errorf("[%d]%s", ErrCodeEduExpenseParam, strings.Join(res.Reasons, "；"))
	}
	if reason, dup := dupCheck(in); dup {
		in.Status = "驳回"
		in.RejectReason = reason
		if err := global.GVA_DB.Create(in).Error; err != nil {
			return nil, err
		}
		return in, nil
	}
	in.Status = "审核中"
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// PayEduExpense 支付（审核中→支付）
func (s *EduadminService) PayEduExpense(id uint) (*model.EduInvoice, error) {
	var inv model.EduInvoice
	if err := global.GVA_DB.First(&inv, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]报销单不存在", ErrCodeEduExpenseNotFound)
	}
	if inv.Status != "审核中" {
		return nil, fmt.Errorf("[%d]仅审核中单据可支付，当前 %s", ErrCodeEduExpenseParam, inv.Status)
	}
	err := global.GVA_DB.Model(&inv).Update("status", "支付").Error
	return &inv, err
}

// RejectEduExpense 人工驳回（审核中→驳回）
func (s *EduadminService) RejectEduExpense(id uint, reason string) (*model.EduInvoice, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("[%d]驳回原因必填", ErrCodeEduExpenseParam)
	}
	var inv model.EduInvoice
	if err := global.GVA_DB.First(&inv, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]报销单不存在", ErrCodeEduExpenseNotFound)
	}
	if inv.Status != "审核中" {
		return nil, fmt.Errorf("[%d]仅审核中单据可驳回，当前 %s", ErrCodeEduExpenseParam, inv.Status)
	}
	err := global.GVA_DB.Model(&inv).Updates(map[string]any{"status": "驳回", "reject_reason": reason}).Error
	return &inv, err
}

// GetEduInvoices 报销列表（状态/部门过滤，倒序分页）
func (s *EduadminService) GetEduInvoices(status, dept string, page, pageSize int) ([]model.EduInvoice, int64, error) {
	if status != "" {
		valid := false
		for _, st := range ExpenseStatuses {
			if st == status {
				valid = true
			}
		}
		if !valid {
			return nil, 0, fmt.Errorf("[%d]状态须为 提交/审核中/支付/驳回", ErrCodeEduExpenseParam)
		}
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduInvoice{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if dept != "" {
		db = db.Where("dept LIKE ?", "%"+dept+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduInvoice
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
