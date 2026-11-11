// Package service policyreg：制度分词与属性单测（纯函数，无 DB）
package service

import (
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"strings"
	"testing"
)

// TestPolicySegment 词典最大正向匹配
func TestPolicySegment(t *testing.T) {
	tokens := PolicySegment("根据学位评定委员会审议，自2026年9月1日起施行学分互认办法。")
	joined := strings.Join(tokens, "|")
	if !strings.Contains(joined, "学位评定委员会") || !strings.Contains(joined, "学分互认") {
		t.Fatalf("词典词应成词: %s", joined)
	}
	if !strings.Contains(joined, "施") || !strings.Contains(joined, "行") {
		t.Fatalf("OOV 汉字应单字输出: %s", joined)
	}
	// 拉丁数字连串
	tokens2 := PolicySegment("参照 GB/T 32907 标准 2026 版执行")
	joined2 := strings.Join(tokens2, "|")
	if !strings.Contains(joined2, "GB/T") || !strings.Contains(joined2, "32907") {
		t.Fatalf("拉丁/数字应连串: %s", joined2)
	}
}

// TestExtractEffectiveDate 生效日期抽取
func TestExtractEffectiveDate(t *testing.T) {
	if d := ExtractEffectiveDate("本办法自2026年9月1日起施行"); d != "2026-09-01" {
		t.Fatalf("应抽取 2026-09-01: %s", d)
	}
	if d := ExtractEffectiveDate("自发布之日起 2027年12月31日 废止"); d != "2027-12-31" {
		t.Fatalf("应抽取 2027-12-31: %s", d)
	}
	if d := ExtractEffectiveDate("未写日期"); d != "" {
		t.Fatalf("无日期应空: %s", d)
	}
}

// TestClassifyPolicyLabel 分类标签
func TestClassifyPolicyLabel(t *testing.T) {
	if l := ClassifyPolicyLabel("研究生招生与培养方案管理办法"); l != "教学" {
		t.Fatalf("招生培养应教学: %s", l)
	}
	if l := ClassifyPolicyLabel("横向课题经费及成果转化管理办法"); l != "科研" {
		t.Fatalf("课题转化应科研: %s", l)
	}
	if l := ClassifyPolicyLabel("管理办法"); l != "综合" {
		t.Fatalf("无关键词应综合: %s", l)
	}
}

// TestTermLabelStats 标签聚合固定序
func TestTermLabelStats(t *testing.T) {
	tags := []model.PolicyTermTag{
		{CategoryLabel: "教学"}, {CategoryLabel: "教学"}, {CategoryLabel: "科研"}, {CategoryLabel: "综合"},
	}
	stats := termLabelStats(tags)
	if len(stats) != 6 {
		t.Fatalf("应六类固定序: %d", len(stats))
	}
	if stats[0].Label != "教学" || stats[0].Count != 2 || stats[1].Count != 1 {
		t.Fatalf("教学 2 科研 1: %+v", stats)
	}
	if stats[5].Label != "综合" || stats[5].Count != 1 {
		t.Fatalf("综合应单列: %+v", stats[5])
	}
}
