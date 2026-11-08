// Package service ziqiang：实体图谱单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// graphFixture 公司(1) -申请-> 专利号(2) -涉及技术-> 架构(3) -采用-> 训练方法(4)；基金(5) -投资-> 公司(1)
func graphFixture() (map[uint]model.ZqEntity, []model.ZqEdge) {
	mk := func(id uint, name, cat, kind string) model.ZqEntity {
		e := model.ZqEntity{Name: name, Category: cat, Kind: kind}
		e.ID = id
		return e
	}
	nodes := map[uint]model.ZqEntity{
		1: mk(1, "上海智能科技公司", "权利主体", "开发者"),
		2: mk(2, "CN202610123456.7", "法律", "专利号"),
		3: mk(3, "Transformer 架构", "技术", "模型架构"),
		4: mk(4, "LoRA 微调", "技术", "训练方法"),
		5: mk(5, "某创投基金", "权利主体", "投资方"),
	}
	edges := []model.ZqEdge{
		{FromId: 1, ToId: 2, Relation: "申请"},
		{FromId: 2, ToId: 3, Relation: "涉及技术"},
		{FromId: 3, ToId: 4, Relation: "采用"},
		{FromId: 5, ToId: 1, Relation: "投资"},
	}
	return nodes, edges
}

// TestBfsPaths 多跳最短路径
func TestBfsPaths(t *testing.T) {
	nodes, edges := graphFixture()
	paths := bfsPaths(nodes, edges, 1, 3)
	if len(paths) != 3 {
		t.Fatalf("从公司出发应达 3 个实体: %d", len(paths))
	}
	// 两跳：公司→专利号→架构
	found := false
	for _, p := range paths {
		if len(p.NodeIds) == 3 && p.NodeIds[2] == 3 {
			found = true
			if len(p.Edges) != 2 || p.Edges[0].Relation != "申请" {
				t.Fatalf("路径边不符: %+v", p.Edges)
			}
		}
	}
	if !found {
		t.Fatalf("应存在两跳路径: %+v", paths)
	}
	// 跳数限制：hop=1 只达专利号
	paths1 := bfsPaths(nodes, edges, 1, 1)
	if len(paths1) != 1 || paths1[0].NodeIds[1] != 2 {
		t.Fatalf("hop=1 应只达专利号: %+v", paths1)
	}
}

// TestValidEntityCategory2 分类白名单
func TestValidEntityCategory2(t *testing.T) {
	for _, c := range EntityCategories {
		if !ValidEntityCategory2(c) {
			t.Fatalf("%s 应在白名单", c)
		}
	}
	if ValidEntityCategory2("其他") {
		t.Fatal("白名单外应拒绝")
	}
}
