// Package service policyreg：训练数据标注单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// TestPreLabelTokens 实体预标
func TestPreLabelTokens(t *testing.T) {
	tokens := []string{"根据", "教", "〔", "2026", "〕", "学位评定委员会", "审议"}
	entities := []model.PolicyEntity{
		{Name: "教〔2026〕15号", EType: "文件号", Source: "规则"},
		{Name: "学位评定委员会", EType: "机构", Source: "词典"},
	}
	labels := PreLabelTokens(tokens, entities)
	if len(labels) != len(tokens) {
		t.Fatalf("应逐词对齐: %d", len(labels))
	}
	if labels[1].Label != "文件号" {
		t.Fatalf("「教」应标文件号: %+v", labels[1])
	}
	if labels[5].Label != "机构" {
		t.Fatalf("「学位评定委员会」应标机构: %+v", labels[5])
	}
	if labels[0].Label != "O" || labels[6].Label != "O" {
		t.Fatalf("未命中词应 O: %+v %+v", labels[0], labels[6])
	}
	for _, l := range labels {
		if l.Source != "规则预标" {
			t.Fatalf("预标来源应规则预标: %+v", l)
		}
	}
}

// TestPreLabelRelations 关系预标
func TestPreLabelRelations(t *testing.T) {
	entities := []model.PolicyEntity{
		{Name: "教育部", EType: "机构"},
		{Name: "上海市教育委员会", EType: "机构"},
		{Name: "教〔2026〕15号", EType: "文件号"},
		{Name: "2026年9月1日", EType: "日期"},
	}
	rels := PreLabelRelations(entities)
	if len(rels) != 2 { // 2 机构 × 1 文件号（日期不参与）
		t.Fatalf("应 2 条关系: %+v", rels)
	}
	for _, r := range rels {
		if r.Relation != "发文机构" || r.Head == "" || r.Tail == "" {
			t.Fatalf("关系口径不符: %+v", r)
		}
	}
}
