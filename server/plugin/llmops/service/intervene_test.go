// Package service llmops 干预机制单测（词表匹配纯函数：block/replace/多词/禁用）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestMatchSensitiveWords(t *testing.T) {
	words := []model.LlmopsSensitiveWord{
		{Word: "违禁词", Action: "block", Enabled: true},
		{Word: "垃圾", Action: "replace", Replacement: "内容", Enabled: true},
		{Word: "已禁用词", Action: "block", Enabled: false},
	}
	// block 命中
	hits, replaced := matchSensitiveWords("包含违禁词的文本", words)
	if len(hits) != 1 || hits[0].Action != "block" {
		t.Fatalf("block 命中异常: %+v", hits)
	}
	if replaced != "包含违禁词的文本" { // block 不改写
		t.Fatalf("block 不应替换: %q", replaced)
	}
	// replace 命中
	hits, replaced = matchSensitiveWords("这是垃圾内容", words)
	if len(hits) != 1 || hits[0].Action != "replace" {
		t.Fatalf("replace 命中异常: %+v", hits)
	}
	if replaced != "这是内容内容" {
		t.Fatalf("替换异常: %q", replaced)
	}
	// 禁用词不命中
	hits, _ = matchSensitiveWords("含已禁用词", words)
	if len(hits) != 0 {
		t.Fatalf("禁用词不应命中: %+v", hits)
	}
	// 无命中
	hits, replaced = matchSensitiveWords("完全干净的文本", words)
	if len(hits) != 0 || replaced != "完全干净的文本" {
		t.Fatalf("无命中应原样: %+v %q", hits, replaced)
	}
	// 多词同时命中（block+replace）
	hits, replaced = matchSensitiveWords("违禁词和垃圾都在", words)
	if len(hits) != 2 {
		t.Fatalf("应双命中: %+v", hits)
	}
	if !strings.Contains(replaced, "内容") || strings.Contains(replaced, "垃圾") {
		t.Fatalf("replace 词应已替换: %q", replaced)
	}
}
