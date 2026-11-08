// Package service fanhua：实体匹配单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

func seedEntities() []model.FanhuaEntity {
	return []model.FanhuaEntity{
		{Category: "文化符号", Term: "石库门", AliasJSON: `["石库门房子"]`},
		{Category: "文化符号", Term: "弄堂", AliasJSON: `["里弄"]`},
		{Category: "文化符号", Term: "生煎", AliasJSON: `["生煎馒头"]`},
		{Category: "地标", Term: "外滩"},
		{Category: "地域", Term: "浦东新区", AliasJSON: `["浦东"]`},
	}
}

// TestEntityMatch 词条+别名计数、按次数降序
func TestEntityMatch(t *testing.T) {
	text := "漫步外滩与弄堂，弄堂深处是石库门，石库门房子里飘出生煎馒头的香气，弄堂口排队买生煎，后来乘地铁去了浦东。"
	hits := entityMatch(text, seedEntities())
	if len(hits) != 5 {
		t.Fatalf("5 个实体均应命中: %+v", hits)
	}
	if hits[0].Term != "弄堂" || hits[0].Count != 3 {
		t.Fatalf("「弄堂」应 3 次居首: %+v", hits[0])
	}
	if hits[1].Term != "石库门" || hits[1].Count != 2 {
		t.Fatalf("「石库门」应 2 次（含别名）: %+v", hits[1])
	}
	if hits[2].Term != "生煎" || hits[2].Count != 2 {
		t.Fatalf("「生煎」应 2 次（含别名）: %+v", hits[2])
	}
	if hits[4].Term != "浦东新区" || hits[4].Count != 1 {
		t.Fatalf("「浦东新区」应经别名命中 1 次: %+v", hits[4])
	}
}

// TestEntityMatchEmpty 无命中
func TestEntityMatchEmpty(t *testing.T) {
	if hits := entityMatch("与上海无关的文本", seedEntities()); len(hits) != 0 {
		t.Fatalf("无关文本应零命中: %+v", hits)
	}
}

// TestValidEntityCategory 分类白名单
func TestValidEntityCategory(t *testing.T) {
	for _, c := range EntityCategories {
		if !ValidEntityCategory(c) {
			t.Fatalf("%s 应在白名单", c)
		}
	}
	if ValidEntityCategory("其他") {
		t.Fatal("白名单外应拒绝")
	}
}

// TestCountOccurrences 计数
func TestCountOccurrences(t *testing.T) {
	if n := countOccurrences("ababab", "ab"); n != 3 {
		t.Fatalf("应 3 次: %d", n)
	}
	if n := countOccurrences("abc", ""); n != 0 {
		t.Fatalf("空子串应 0: %d", n)
	}
}
