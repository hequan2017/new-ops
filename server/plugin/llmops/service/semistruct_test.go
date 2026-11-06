// Package service llmops 半结构化解析单测（格式探测/JSON/YAML/XML 顶层字段抽取）
package service

import (
	"strings"
	"testing"
)

func TestDetectSemiFormat(t *testing.T) {
	cases := map[string]string{
		`{"a":1}`:               "json",
		` [1,2]`:                "json",
		"\xef\xbb\xbf{\"a\":1}": "json", // BOM 前缀不影响
		"<root><a>1</a></root>": "xml",
		"a: 1\nb: 2":            "yaml",
		"name: x":               "yaml",
		"":                      "",
		"   ":                   "",
	}
	for in, want := range cases {
		if got := DetectSemiFormat([]byte(in)); got != want {
			t.Fatalf("DetectSemiFormat(%q)=%q want %q", in, got, want)
		}
	}
}

func TestParseJSONSchema(t *testing.T) {
	fields, err := ParseJSONSchema([]byte(`{"name":"白泽","port":8080,"ok":true,"tags":["a","b"],"cfg":{"k":"v"},"empty":null}`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	byName := map[string]SemiField{}
	for _, f := range fields {
		byName[f.Name] = f
	}
	want := map[string]SemiField{
		"name":  {Name: "name", Type: "string", Sample: "白泽"},
		"port":  {Name: "port", Type: "number", Sample: "8080"},
		"ok":    {Name: "ok", Type: "bool", Sample: "true"},
		"tags":  {Name: "tags", Type: "array", Sample: "[a b]"},
		"cfg":   {Name: "cfg", Type: "object", Sample: "map[k:v]"},
		"empty": {Name: "empty", Type: "null", Sample: ""},
	}
	for n, w := range want {
		got := byName[n]
		if got.Type != w.Type || got.Sample != w.Sample {
			t.Fatalf("字段 %s = %+v want %+v", n, got, w)
		}
	}
}

// TestParseJSONSchema_ArrayOfRecords 数组口径：取首元素字段并标注 array
func TestParseJSONSchema_ArrayOfRecords(t *testing.T) {
	fields, err := ParseJSONSchema([]byte(`[{"host":"h1","port":22},{"host":"h2"}]`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(fields) != 2 || fields[0].Name != "host" || fields[0].Type != "string[]" || fields[1].Type != "number[]" {
		t.Fatalf("数组口径异常: %+v", fields)
	}
}

func TestParseYAMLSchema(t *testing.T) {
	fields, err := ParseYAMLSchema([]byte("name: 白泽\nreplicas: 3\nenabled: true\n"))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(fields) != 3 ||
		fields[0].Name != "enabled" || fields[0].Type != "bool" ||
		fields[1].Name != "name" || fields[1].Type != "string" ||
		fields[2].Name != "replicas" || fields[2].Type != "number" {
		t.Fatalf("YAML 字段异常(应按名排序): %+v", fields)
	}
	// 坏 YAML
	if _, err := ParseYAMLSchema([]byte("a:\n - b\n  bad indent: [")); err == nil || !strings.Contains(err.Error(), "2216") {
		t.Fatalf("坏 YAML 应报 2216: %v", err)
	}
}

func TestParseXMLSchema(t *testing.T) {
	xmlDoc := `<catalog>
  <book id="1">Go 入门<price>45.5</price></book>
  <book id="2">深入 K8s</book>
  <magazine>运维周刊</magazine>
</catalog>`
	fields, err := ParseXMLSchema([]byte(xmlDoc))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(fields) != 3 {
		t.Fatalf("顶层元素应为 3: %+v", fields)
	}
	if fields[0].Name != "book" || fields[0].Type != "element" {
		t.Fatalf("book 字段异常: %+v", fields[0])
	}
	if fields[0].Sample != "Go 入门" { // 嵌套 <price> 文本不混入直属文本
		t.Fatalf("样本应仅含直属文本: %q", fields[0].Sample)
	}
	if fields[2].Sample != "运维周刊" {
		t.Fatalf("magazine 样本异常: %+v", fields[2])
	}
	if _, err := ParseXMLSchema([]byte("<bad>unclosed")); err == nil || !strings.Contains(err.Error(), "2216") {
		t.Fatalf("坏 XML 应报 2216: %v", err)
	}
}

// TestTruncateSample 样本 64 rune 截断
func TestTruncateSample(t *testing.T) {
	long := strings.Repeat("字", 100)
	got := truncateSample(long)
	if len([]rune(got)) != 65 || !strings.HasSuffix(got, "…") {
		t.Fatalf("截断口径异常: %d", len([]rune(got)))
	}
}
