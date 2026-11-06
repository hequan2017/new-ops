// Package service llmops 混合检索单测（RRF 融合/路由白名单/图谱路径文本化）
package service

import "testing"

func TestRrfFuse(t *testing.T) {
	// 两路：A 路顺序 x,y,z；B 路顺序 y,x,w → x(1/61+1/62) > y(1/62+1/61) 相同？rank 不同则不同
	listA := []RetrievalHit{
		{Route: "vector", Content: "x"}, {Route: "vector", Content: "y"}, {Route: "vector", Content: "z"},
	}
	listB := []RetrievalHit{
		{Route: "fulltext", Content: "w"}, {Route: "fulltext", Content: "x"},
	}
	fused := rrfFuse([][]RetrievalHit{listA, listB}, 60)
	// x: 1/61+1/62 ≈ 0.03252；y: 1/62；z: 1/63；w: 1/61 → w 与 x 接近，x>w
	if len(fused) != 4 {
		t.Fatalf("应 4 个去重后条目: %d", len(fused))
	}
	if fused[0].Content != "x" {
		t.Fatalf("双路命中 x 应第一: %+v", fused[0])
	}
	if fused[0].Rrf <= fused[1].Rrf {
		t.Fatalf("融合分应降序: %+v", fused)
	}
	// 跨路同内容去重（vector 与 fulltext 都含 y → 合并）
	listC := []RetrievalHit{{Route: "fulltext", Content: "y"}}
	fused2 := rrfFuse([][]RetrievalHit{listA, listC}, 60)
	count := 0
	for _, h := range fused2 {
		if h.Content == "y" {
			count++
			if h.Rrf <= 1.0/61 {
				t.Fatalf("跨路命中 y 应累加两路分: %v", h.Rrf)
			}
		}
	}
	if count != 1 {
		t.Fatalf("y 应去重为 1 条: %d", count)
	}
	// 空输入
	if got := rrfFuse(nil, 60); len(got) != 0 {
		t.Fatalf("空输入应空: %d", len(got))
	}
}

func TestValidRoute(t *testing.T) {
	for _, ok := range HybridRoutes {
		if !ValidRoute(ok) {
			t.Fatalf("合法路由 %q 被拒绝", ok)
		}
	}
	if ValidRoute("kg") || ValidRoute("") {
		t.Fatal("非法路由应拒绝")
	}
}

func TestFormatGraphPath(t *testing.T) {
	path := []PathStep{
		{FromName: "白泽", Relation: "属于", ToName: "平台"},
		{Relation: "位于", ToName: "中国"},
	}
	if got := formatGraphPath(path); got != "白泽 -属于-> 平台 -位于-> 中国" {
		t.Fatalf("路径文本化异常: %q", got)
	}
	if got := formatGraphPath(nil); got != "" {
		t.Fatalf("空路径应空串: %q", got)
	}
}

// TestHybridSearch_RouteGuard 非法路由入口即拒（不触 DB）
func TestHybridSearch_RouteGuard(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	if _, _, err := svc.HybridSearch(1, "q", 5, []string{"bad-route"}, 0); err == nil {
		t.Fatal("非法路由应拒绝")
	}
}
