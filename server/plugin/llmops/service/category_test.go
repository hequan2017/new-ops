// Package service llmops 编目治理单测（分类树构建：正常/孤儿挂根/环引用不丢/动作白名单）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func mkCat(id, parent uint) model.LlmopsCategory {
	c := model.LlmopsCategory{Name: fmtName(id), ParentId: parent}
	c.ID = id
	return c
}

func fmtName(id uint) string { return "cat" + string(rune('0'+int(id%10))) }

func TestBuildCategoryTree_Normal(t *testing.T) {
	cats := []model.LlmopsCategory{mkCat(1, 0), mkCat(2, 1), mkCat(3, 1), mkCat(4, 2)}
	tree := BuildCategoryTree(cats)
	if len(tree) != 1 {
		t.Fatalf("应只有 1 个根: %d", len(tree))
	}
	if len(tree[0].Children) != 2 || len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("树结构异常: %+v", tree)
	}
}

// TestBuildCategoryTree_Orphan 父缺失挂根，数据不丢
func TestBuildCategoryTree_Orphan(t *testing.T) {
	cats := []model.LlmopsCategory{mkCat(1, 0), mkCat(2, 99)} // 2 的父 99 不存在
	tree := BuildCategoryTree(cats)
	if len(tree) != 2 {
		t.Fatalf("孤儿应挂根(共 2 根): %d", len(tree))
	}
}

// TestBuildCategoryTree_Cycle 环引用断链挂根：不死循环、节点不丢不重
func TestBuildCategoryTree_Cycle(t *testing.T) {
	cats := []model.LlmopsCategory{mkCat(1, 0), mkCat(2, 3), mkCat(3, 2)} // 2↔3 环
	tree := BuildCategoryTree(cats)
	var ids []uint
	var walk func(nodes []CategoryNode)
	walk = func(nodes []CategoryNode) {
		for _, n := range nodes {
			ids = append(ids, n.ID)
			walk(n.Children)
		}
	}
	walk(tree)
	seen := map[uint]int{}
	for _, id := range ids {
		seen[id]++
	}
	if len(seen) != 3 {
		t.Fatalf("环成员应全部保留(3 个唯一节点): %v", seen)
	}
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("节点 %d 重复出现 %d 次", id, n)
		}
	}
}

func TestBuildCategoryTree_Empty(t *testing.T) {
	if tree := BuildCategoryTree(nil); tree != nil {
		t.Fatalf("空输入应返回 nil: %+v", tree)
	}
}

func TestValidAssetAction(t *testing.T) {
	for _, ok := range []string{"collect", "parse", "catalog", "tag", "update", "offline", "restore", "delete"} {
		if !ValidAssetAction(ok) {
			t.Fatalf("合法动作 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "DROP", "collect;"} {
		if ValidAssetAction(bad) {
			t.Fatalf("非法动作 %q 被放行", bad)
		}
	}
}
