// Package service llmops 图谱多跳推理单测（BFS 路径枚举：链/分支/环防护/hops 上限/路径数截断）
package service

import "testing"

// edgeOf 辅助构造邻接表
func edgeOf(adj map[uint][]GraphEdgeRef, from, to uint, rel string) {
	adj[from] = append(adj[from], GraphEdgeRef{To: to, Relation: rel})
}

func TestBfsPaths_Chain(t *testing.T) {
	adj := map[uint][]GraphEdgeRef{}
	edgeOf(adj, 1, 2, "属于")
	edgeOf(adj, 2, 3, "位于")
	paths := BfsPaths(adj, 1, 3)
	if len(paths) != 2 {
		t.Fatalf("链 1→2→3 应 2 条路径(1跳+2跳): %d", len(paths))
	}
	longest := paths[len(paths)-1]
	if len(longest) != 2 || longest[0].Relation != "属于" || longest[1].Relation != "位于" || longest[1].ToId != 3 {
		t.Fatalf("最长路径异常: %+v", longest)
	}
}

func TestBfsPaths_Branch(t *testing.T) {
	adj := map[uint][]GraphEdgeRef{}
	edgeOf(adj, 1, 2, "a")
	edgeOf(adj, 1, 3, "b")
	paths := BfsPaths(adj, 1, 2)
	if len(paths) != 2 {
		t.Fatalf("两分支应 2 条路径: %d", len(paths))
	}
}

// TestBfsPaths_Cycle 环引用防护：simple path 不重复节点、不死循环
func TestBfsPaths_Cycle(t *testing.T) {
	adj := map[uint][]GraphEdgeRef{}
	edgeOf(adj, 1, 2, "r")
	edgeOf(adj, 2, 1, "r")
	paths := BfsPaths(adj, 1, 5)
	// 无环路径：1→2；含环回 1 被禁 → 仅 1 条
	if len(paths) != 1 {
		t.Fatalf("环应被防护: %d", len(paths))
	}
	if paths[0][0].ToId != 2 {
		t.Fatalf("路径异常: %+v", paths[0])
	}
}

// TestBfsPaths_HopLimit hops 上限截断与夹取
func TestBfsPaths_HopLimit(t *testing.T) {
	adj := map[uint][]GraphEdgeRef{}
	for i := uint(1); i < 10; i++ {
		edgeOf(adj, i, i+1, "r")
	}
	paths := BfsPaths(adj, 1, 3)
	maxLen := 0
	for _, p := range paths {
		if len(p) > maxLen {
			maxLen = len(p)
		}
	}
	if maxLen != 3 {
		t.Fatalf("最长路径应 3 跳: %d", maxLen)
	}
	// hops=0 按契约收敛为 1 跳；超大夹取为 5
	if got := BfsPaths(adj, 1, 0); len(got) != 1 {
		t.Fatalf("hops=0 应收敛 1 跳(链上仅 1 条): %d", len(got))
	}
	if got := BfsPaths(adj, 1, 999); len(got) == 0 || len(got[0]) > 5 {
		t.Fatal("超大 hops 应夹取为 5")
	}
}

// TestBfsPaths_PathCap 路径数截断（组合爆炸防护）
func TestBfsPaths_PathCap(t *testing.T) {
	adj := map[uint][]GraphEdgeRef{}
	// 三层扇出 10×10×10 → 路径 10+100+1000 远超 100 条
	for i := uint(2); i <= 11; i++ {
		edgeOf(adj, 1, i, "r")
		for j := uint(0); j < 10; j++ {
			mid := i*10 + j
			leaf := i*100 + j
			edgeOf(adj, i, mid, "r")
			edgeOf(adj, mid, leaf, "r")
		}
	}
	paths := BfsPaths(adj, 1, 3)
	if len(paths) != graphMaxPaths {
		t.Fatalf("路径数应截断为 %d: %d", graphMaxPaths, len(paths))
	}
}
