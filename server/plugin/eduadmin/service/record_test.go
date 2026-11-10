// Package service eduadmin：数据接入单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestParseEduJSON JSON 数组导入：白名单/唯一键/负载校验
func TestParseEduJSON(t *testing.T) {
	content := `[
		{"domain":"人事","externalKey":"HR-001","source":"人事系统","payload":{"name":"张三","dept":"计算机学院"}},
		{"domain":"后勤","externalKey":"LOG-9","source":"后勤系统","payload":{"dorm":"A栋"}},
		{"domain":"其他","externalKey":"X-1","payload":{}},
		{"domain":"财务","externalKey":"","payload":{}},
		{"domain":"财务","externalKey":"F-1","payload":[1,2]},
		{"domain":"人事","externalKey":"HR-001","payload":{}}
	]`
	list, errs, err := parseEduJSON(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("应入库 2 条: %d", len(list))
	}
	if len(errs) != 4 {
		t.Fatalf("应报 4 行错误: %+v", errs)
	}
	joined := ""
	for _, e := range errs {
		joined += e.Reason
	}
	for _, want := range []string{"白名单", "外部唯一键必填", "JSON 对象", "唯一键重复"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("错误报告缺 %s: %+v", want, errs)
		}
	}
}

// TestParseEduCSV CSV 导入：表头跳过/列数校验
func TestParseEduCSV(t *testing.T) {
	content := "domain,externalKey,source,payloadJSON\n" +
		"资产,ZC-01,资产系统,\"{\"\"tag\"\":\"\"笔记本\"\"}\"\n" +
		"教务,KC-01,教务系统,{}\n" +
		"资产,ZC-01,资产系统,{}\n"
	list, errs, err := parseEduCSV(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(list) != 1 || list[0].ExternalKey != "ZC-01" {
		t.Fatalf("应入库 1 条 ZC-01: %+v", list)
	}
	if len(errs) != 2 {
		t.Fatalf("应报 2 行错误（域白名单+唯一键重复）: %+v", errs)
	}
}

// TestValidPayloadJSON 负载 JSON 对象校验
func TestValidPayloadJSON(t *testing.T) {
	for _, ok := range []string{`{"a":1}`, ` {} `, `{"a":{"b":"c"}}`} {
		if !validPayloadJSON(ok) {
			t.Fatalf("%s 应合法", ok)
		}
	}
	for _, bad := range []string{`[1,2]`, `{"a":`, `"text"`, ``} {
		if validPayloadJSON(bad) {
			t.Fatalf("%s 应非法", bad)
		}
	}
}

// TestValidEduDomain 域白名单
func TestValidEduDomain(t *testing.T) {
	for _, d := range EduDomains {
		if !ValidEduDomain(d) {
			t.Fatalf("%s 应在白名单", d)
		}
	}
	if ValidEduDomain("教务") {
		t.Fatal("白名单外应拒绝")
	}
}
