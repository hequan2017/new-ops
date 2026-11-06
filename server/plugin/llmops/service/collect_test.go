// Package service llmops 文档采集纯函数单测（解析模式/命名归一/元数据口径）
package service

import (
	"encoding/json"
	"mime/multipart"
	"strings"
	"testing"
)

// TestExtParseMode 文本类直读、二进制仅编目的解析口径契约
func TestExtParseMode(t *testing.T) {
	for _, ext := range []string{".txt", ".md", ".csv", ".json", ".yaml", ".xml", ".log", ".html", ".SQL", ".Py"} {
		if got := ExtParseMode(ext); got != "text" {
			t.Fatalf("文本扩展名 %q 判为 %q", ext, got)
		}
	}
	for _, ext := range []string{".pdf", ".docx", ".xlsx", ".png", ".mp4", "", ".weird"} {
		if got := ExtParseMode(ext); got != "binary" {
			t.Fatalf("二进制/未知扩展名 %q 判为 %q", ext, got)
		}
	}
}

func TestBuildDocName(t *testing.T) {
	cases := []struct{ provided, filename, want string }{
		{"", "需求说明.pdf", "需求说明"},
		{"  自定义名  ", "a.txt", "自定义名"},
		{"", "/path/to/report.2024.md", "report.2024"}, // 只去最后一个扩展
		{"", "   ", "未命名文档"},
		{strings.Repeat("长", 200), "x.txt", strings.Repeat("长", 128)},
	}
	for _, c := range cases {
		if got := BuildDocName(c.provided, c.filename); got != c.want {
			t.Fatalf("BuildDocName(%q,%q)=%q want %q", c.provided, c.filename, got, c.want)
		}
	}
}

// TestImportDocRejects 空文件与超限文件的守卫分支（不入库、不触盘）
func TestImportDocRejects(t *testing.T) {
	if _, err := LlmopsService.ImportDoc(nil, "", ""); err == nil || !strings.Contains(err.Error(), "2202") {
		t.Fatalf("nil 文件应报 2202: %v", err)
	}
	big := &multipart.FileHeader{}
	big.Size = ImportMaxBytes + 1
	if _, err := LlmopsService.ImportDoc(big, "", ""); err == nil || !strings.Contains(err.Error(), "2203") {
		t.Fatalf("超限应报 2203: %v", err)
	}
}

// TestDocMetaShape 编目元数据 JSON 字段契约（前端/编目检索依赖键名稳定）
func TestDocMetaShape(t *testing.T) {
	m := docMeta{Filename: "a.pdf", Ext: ".pdf", MIME: mimeExts[".pdf"], ParseMode: "binary", Note: "x"}
	b, _ := json.Marshal(m)
	for _, key := range []string{`"filename":"a.pdf"`, `"ext":".pdf"`, `"mime":"application/pdf"`, `"parseMode":"binary"`, `"note":"x"`} {
		if !strings.Contains(string(b), key) {
			t.Fatalf("metaJSON 缺键值 %s: %s", key, b)
		}
	}
	// note 空时应省略（omitempty），文本类不携带占位说明
	m2 := docMeta{Filename: "a.txt", ParseMode: "text"}
	b2, _ := json.Marshal(m2)
	if strings.Contains(string(b2), "note") {
		t.Fatalf("空 note 应省略: %s", b2)
	}
}
