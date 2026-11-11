// Package service eduadmin：审计知识语料单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestRankAuditItems bigram 命中降序
func TestRankAuditItems(t *testing.T) {
	list := []model.EduAuditItem{
		{Title: "费用报销审计要点", RiskPoints: "超标准报销、连号发票、重复报销"},
		{Title: "科研经费专项审计", RiskPoints: "横向经费私账、劳务费虚领"},
	}
	hits := rankAuditItems(list, "重复报销风险")
	if len(hits) != 1 || hits[0].Title != "费用报销审计要点" {
		t.Fatalf("应仅费用报销命中: %+v", hits)
	}
	if hits2 := rankAuditItems(list, "云计算底层"); len(hits2) != 0 {
		t.Fatalf("无关查询应零命中: %+v", hits2)
	}
}

// TestValidEduAuditDomain 六域白名单
func TestValidEduAuditDomain(t *testing.T) {
	for _, d := range EduAuditDomains {
		if !ValidEduAuditDomain(d) {
			t.Fatalf("%s 应在白名单", d)
		}
	}
	if ValidEduAuditDomain("后勤") {
		t.Fatal("白名单外应拒绝")
	}
}
