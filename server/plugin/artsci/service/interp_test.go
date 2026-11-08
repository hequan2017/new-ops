// Package service artsci：图像解读单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// TestParseInterpJSON 解析：干净/带前后缀/缺字段/非 JSON
func TestParseInterpJSON(t *testing.T) {
	s, c, err := parseInterpJSON(`{"surface":"构图对称","cultural":"体现礼制"}`)
	if err != nil || s != "构图对称" || c != "体现礼制" {
		t.Fatalf("干净 JSON 解析不符: %v %v %v", s, c, err)
	}
	s, _, err = parseInterpJSON(`说明{"surface":"a","cultural":"b"}尾注`)
	if err != nil || s != "a" {
		t.Fatalf("带前后缀应可解析: %v %v", s, err)
	}
	if _, _, err = parseInterpJSON(`{"surface":"only"}`); err == nil {
		t.Fatal("缺 cultural 应报错")
	}
	if _, _, err = parseInterpJSON(`不是 JSON`); err == nil {
		t.Fatal("非 JSON 应报错")
	}
}

// TestRuleInterpretation 元数据口径：两段结构化、含名称与视角
func TestRuleInterpretation(t *testing.T) {
	asset := &model.ArtsciImageAsset{
		Name: "清明上河图局部", MIME: "image/jpeg", SizeByte: 10240, Source: "url", MetaJSON: `{"年代":"北宋"}`,
	}
	surface, cultural := ruleInterpretation(asset, "历史照片")
	if !strings.Contains(surface, "清明上河图局部") || !strings.Contains(surface, "image/jpeg") {
		t.Fatalf("表面段应含名称与 MIME: %s", surface)
	}
	if !strings.Contains(cultural, "历史照片") || !strings.Contains(cultural, "北宋") {
		t.Fatalf("含义段应含视角与元数据: %s", cultural)
	}
}
