// Package service eduadmin：自动报销单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestAuditInvoice 规则审核：限额/白名单/必填
func TestAuditInvoice(t *testing.T) {
	pass := model.EduInvoice{InvoiceNo: "FP001", Category: "差旅", Amount: 3000, InvoiceDate: "2026-11-01"}
	if res := auditInvoice(&pass); !res.Pass {
		t.Fatalf("合规单应通过: %+v", res)
	}
	over := model.EduInvoice{InvoiceNo: "FP002", Category: "招待", Amount: 1500, InvoiceDate: "2026-11-01"}
	res := auditInvoice(&over)
	if res.Pass || !strings.Contains(strings.Join(res.Reasons, ""), "超类别限额") {
		t.Fatalf("招待 1500 应超限: %+v", res)
	}
	bad := model.EduInvoice{InvoiceNo: "", Category: "礼品", Amount: 0, InvoiceDate: ""}
	res = auditInvoice(&bad)
	if res.Pass || len(res.Reasons) < 3 {
		t.Fatalf("多规则应聚合报错: %+v", res)
	}
	// 设备高额放行
	device := model.EduInvoice{InvoiceNo: "FP003", Category: "设备", Amount: 18000, InvoiceDate: "2026-11-02"}
	if res := auditInvoice(&device); !res.Pass {
		t.Fatalf("设备 18000 应放行: %+v", res)
	}
}

// TestExpenseStatuses 状态机白名单
func TestExpenseStatuses(t *testing.T) {
	if len(ExpenseStatuses) != 4 {
		t.Fatalf("状态机应 4 节点: %+v", ExpenseStatuses)
	}
	for _, st := range ExpenseStatuses {
		if st == "" {
			t.Fatal("状态不可为空")
		}
	}
}
