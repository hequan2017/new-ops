// Package service crossauth：四级权限矩阵单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// TestMatrixCheck 最具体优先
func TestMatrixCheck(t *testing.T) {
	grants := []model.CaPermGrant{
		{School: "上海大学", Dept: "*", UserName: "*", Scene: "算力平台", Allowed: true},          // 2 维
		{School: "上海大学", Dept: "计算机学院", UserName: "*", Scene: "算力平台", Allowed: false},   // 3 维
		{School: "上海大学", Dept: "计算机学院", UserName: "zhangsan", Scene: "算力平台", Allowed: true}, // 4 维
	}
	// 具名用户命中 4 维授权 → 放行
	allowed, sp := matrixCheck(grants, "上海大学", "计算机学院", "zhangsan", "算力平台")
	if !allowed || sp != 4 {
		t.Fatalf("具名用户应放行(4 维): %v %d", allowed, sp)
	}
	// 同院其他用户命中 3 维拒绝
	allowed, sp = matrixCheck(grants, "上海大学", "计算机学院", "lisi", "算力平台")
	if allowed || sp != 3 {
		t.Fatalf("同院他人应拒绝(3 维): %v %d", allowed, sp)
	}
	// 其他院命中 2 维放行
	allowed, sp = matrixCheck(grants, "上海大学", "外国语学院", "wangwu", "算力平台")
	if !allowed || sp != 2 {
		t.Fatalf("他院应 2 维放行: %v %d", allowed, sp)
	}
	// 无命中默认拒绝
	allowed, _ = matrixCheck(grants, "复旦大学", "计算机学院", "zhangsan", "算力平台")
	if allowed {
		t.Fatal("他校应默认拒绝")
	}
	// 场景不匹配
	allowed, _ = matrixCheck(grants, "上海大学", "计算机学院", "zhangsan", "图书馆")
	if allowed {
		t.Fatal("场景不匹配应拒绝")
	}
}
