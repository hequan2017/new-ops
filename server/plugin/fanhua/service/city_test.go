// Package service fanhua：城市科学单测（纯函数，无 DB）
package service

import (
	"math"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

func govList() []model.FanhuaGovIndicator {
	list := []model.FanhuaGovIndicator{
		{Name: "环境品质", ParentId: 0, Weight: 30},
		{Name: "视觉评价", ParentId: 0, Weight: 30, Visual: true},
		{Name: "空气质量", ParentId: 1, Weight: 50},
		{Name: "噪声水平", ParentId: 1, Weight: 50},
		{Name: "街景美观度", ParentId: 2, Weight: 100, Visual: true},
	}
	for i := range list {
		list[i].ID = uint(i + 1)
	}
	return list
}

func govPoints() []model.FanhuaGovDataPoint {
	return []model.FanhuaGovDataPoint{
		{Region: "田林新村", IndicatorId: 3, Value: 80, Period: "2026-10"},
		{Region: "田林新村", IndicatorId: 4, Value: 60, Period: "2026-10"},
		{Region: "田林新村", IndicatorId: 5, Value: 70, Period: "2026-10"},
	}
}

// TestBuildIndicatorTree 平铺转树
func TestBuildIndicatorTree(t *testing.T) {
	roots := buildIndicatorTree(govList())
	if len(roots) != 2 {
		t.Fatalf("应有 2 个根: %d", len(roots))
	}
	if len(roots[0].Children) != 2 || len(roots[1].Children) != 1 {
		t.Fatalf("子叶挂载不符: %+v", roots)
	}
}

// TestScoreTree 加权得分：叶均值、父归一、根加权
func TestScoreTree(t *testing.T) {
	roots := buildIndicatorTree(govList())
	// 环境=(80×50+60×50)/100=70；视觉=70
	total := 0.0
	for _, r := range roots {
		scoreTree(r, govPoints())
	}
	env, visual := roots[0].Avg, roots[1].Avg
	if math.Abs(env-70) > 1e-9 {
		t.Fatalf("环境应 70: %v", env)
	}
	if math.Abs(visual-70) > 1e-9 {
		t.Fatalf("视觉应 70: %v", visual)
	}
	total = (env*30 + visual*30) / 60
	if math.Abs(total-70) > 1e-9 {
		t.Fatalf("总分应 70: %v", total)
	}
	// 无数据点叶子得 0
	roots = buildIndicatorTree(govList())
	scoreTree(roots[0], nil)
	if roots[0].Avg != 0 {
		t.Fatalf("无数据应得 0: %v", roots[0].Avg)
	}
}

// TestRoundScore 取整
func TestRoundScore(t *testing.T) {
	if roundScore(70.456) != 70.46 {
		t.Fatalf("两位小数不符: %v", roundScore(70.456))
	}
}
