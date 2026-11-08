// Package service ziqiang：文本鉴权语料单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestParseTextImport JSONL/CSV 解析
func TestParseTextImport(t *testing.T) {
	jsonl := `{"type":"专利","title":"一种相变储能装置","content":"本发明涉及…","aigcFlag":false}
{"type":"论文","title":"AIGC 综述","content":"…","aigcFlag":true}
{bad}
{"type":"散文","title":"x","content":"y"}`
	list, errs, err := parseTextImport("jsonl", jsonl)
	if err != nil || len(list) != 2 || len(errs) != 2 {
		t.Fatalf("应 2 合法 2 错误: %d %v %v", len(list), errs, err)
	}
	if !list[1].AigcFlag {
		t.Fatal("第二条应为 AIGC")
	}
	csv := "type,title,aigcFlag,content\n专利,装置,false,本发明…\n合同,服务合同,false,甲方乙方…"
	list2, errs2, err2 := parseTextImport("csv", csv)
	if err2 != nil || len(list2) != 2 || len(errs2) != 0 {
		t.Fatalf("CSV 应 2 合法: %d %v %v", len(list2), errs2, err2)
	}
	if _, _, err3 := parseTextImport("xlsx", "x"); err3 == nil || !strings.Contains(err3.Error(), "jsonl/csv") {
		t.Fatalf("格式不支持应报错: %v", err3)
	}
}

// TestValidTextType 类型白名单
func TestValidTextType(t *testing.T) {
	for _, t2 := range TextTypes {
		if !ValidTextType(t2) {
			t.Fatalf("%s 应在白名单", t2)
		}
	}
	if ValidTextType("散文") {
		t.Fatal("白名单外应拒绝")
	}
}

// TestCorpusFromRow 行校验
func TestCorpusFromRow(t *testing.T) {
	c, reason := corpusFromRow("专利", "一种装置", "false", "正文")
	if reason != "" || c.AigcFlag {
		t.Fatalf("合法行应通过: %s", reason)
	}
	if _, reason2 := corpusFromRow("专利", "", "false", "正文"); !strings.Contains(reason2, "必填") {
		t.Fatalf("空标题应报必填: %s", reason2)
	}
	if _, reason3 := corpusFromRow("专利", "t", "是", "正文"); reason3 != "" {
		t.Fatalf("中文「是」应识别为 true: %s", reason3)
	}
}
