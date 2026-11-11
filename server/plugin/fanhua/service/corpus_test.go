// Package service fanhua：语料导入解析单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// TestParseCorpusJSONL JSONL：合法/坏 JSON/来源白名单外/缺正文
func TestParseCorpusJSONL(t *testing.T) {
	content := `{"source":"新闻","title":"沪语新闻","content":"今朝天气交关好","lang":"沪语"}
{"source":"社交","title":"弄堂闲话","content":"侬饭吃过了伐","lang":"沪语"}
{bad json}
{"source":"体育","title":"x","content":"y"}
{"source":"文学","title":"","content":"z"}`
	list, errs, err := parseCorpusImport("jsonl", content)
	if err != nil {
		t.Fatalf("解析不应整体失败: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("合法行应 2 条: %d", len(list))
	}
	if len(errs) != 3 {
		t.Fatalf("错误行应 3 条: %v", errs)
	}
	if errs[0].Line != 3 || !strings.Contains(errs[0].Reason, "JSON 解析失败") {
		t.Fatalf("第 3 行应为 JSON 错误: %+v", errs[0])
	}
	if !strings.Contains(errs[1].Reason, "白名单") || !strings.Contains(errs[2].Reason, "必填") {
		t.Fatalf("错误原因不符: %+v %+v", errs[1], errs[2])
	}
}

// TestParseCorpusCSV CSV：表头跳过/列数不符整体报错
func TestParseCorpusCSV(t *testing.T) {
	csv := "source,title,lang,content\n新闻,沪语新闻,沪语,今朝天气好\n社交,弄堂闲话,沪语,侬好"
	list, errs, err := parseCorpusImport("csv", csv)
	if err != nil {
		t.Fatalf("CSV 应成功: %v", err)
	}
	if len(list) != 2 || len(errs) != 0 {
		t.Fatalf("应 2 条合法 0 错误: %d %v", len(list), errs)
	}
	if _, _, err = parseCorpusImport("csv", "source,title,lang\n新闻,a,沪语"); err == nil {
		t.Fatal("列数不符应报错")
	}
	if _, _, err = parseCorpusImport("xlsx", "x"); err == nil || !strings.Contains(err.Error(), "jsonl/csv") {
		t.Fatalf("不支持格式应报错: %v", err)
	}
}

// TestValidSourceLang 白名单
func TestValidSourceLang(t *testing.T) {
	for _, s := range CorpusSources {
		if !ValidSource(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if ValidSource("体育") || ValidLang("粤语") {
		t.Fatal("白名单外应拒绝")
	}
	if !ValidLang("") {
		t.Fatal("空语种应放行（默认沪语）")
	}
}

// TestCorporaToCSV CSV 导出（I07 口径）：表头+逗号替换
func TestCorporaToCSV(t *testing.T) {
	list := []model.FanhuaCorpus{
		{Source: "新闻", Title: "侬好,上海", Lang: "沪语", Content: "第一行\n第二行,含逗号"},
	}
	csv := corporaToCSV(list)
	if !strings.HasPrefix(csv, "来源,标题,语种,内容\n") {
		t.Fatalf("应有表头: %s", csv)
	}
	if !strings.Contains(csv, "侬好，上海") {
		t.Fatalf("标题逗号应转全角: %s", csv)
	}
	if strings.Contains(csv, "\n第二行,") {
		t.Fatalf("内容换行应替换且逗号转全角: %s", csv)
	}
}
