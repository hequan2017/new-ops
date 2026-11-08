// Package service fanhua：海派文化知识库单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

func voiceSeed() []model.FanhuaVoiceEntry {
	return []model.FanhuaVoiceEntry{
		{Title: "笃笃笃卖糖粥", Category: "弄堂俚语", Content: "笃笃笃，卖糖粥，三斤胡桃四斤壳。"},
		{Title: "再别康桥·节选", Category: "经典朗读", Content: "轻轻的我走了，正如我轻轻的来。"},
		{Title: "燕燕做媒·选段", Category: "沪剧台词", Content: "燕燕做媒选段，媒人两边说合。"},
	}
}

// TestRankVoice token 命中计数与降序
func TestRankVoice(t *testing.T) {
	hits := rankVoice(voiceSeed(), "糖粥")
	if len(hits) != 1 || hits[0].Title != "笃笃笃卖糖粥" || hits[0].Score < 1 {
		t.Fatalf("「糖粥」应命中卖糖粥条目: %+v", hits)
	}
	// 多 token 命中排序
	hits = rankVoice(voiceSeed(), "卖糖粥的媒人")
	if len(hits) < 2 || hits[0].Score < hits[1].Score {
		t.Fatalf("应降序: %+v", hits)
	}
	if hits[0].Title != "笃笃笃卖糖粥" {
		t.Fatalf("多 token 命中者应居首: %+v", hits[0])
	}
	// 无关查询零命中
	if hits := rankVoice(voiceSeed(), "量子力学"); len(hits) != 0 {
		t.Fatalf("无关查询应零命中: %+v", hits)
	}
}

// TestVoiceWhitelists 分类与声部白名单
func TestVoiceWhitelists(t *testing.T) {
	for _, c := range VoiceCategories {
		if !validIn(VoiceCategories, c) {
			t.Fatalf("%s 应在白名单", c)
		}
	}
	for _, s := range SpeakerTypes {
		if !validIn(SpeakerTypes, s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if validIn(VoiceCategories, "相声") || validIn(SpeakerTypes, "机器人") {
		t.Fatal("白名单外应拒绝")
	}
}
