// Package service llmops MCP 工具单测（参数模板/参数校验/调用代理 httptest GET+POST）
package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestParseArgsSchema(t *testing.T) {
	schema, err := ParseArgsSchema(`{"city":"城市名","date":"日期"}`)
	if err != nil || len(schema) != 2 || schema["city"] != "城市名" {
		t.Fatalf("模板解析异常: %+v %v", schema, err)
	}
	// 空模板
	schema, err = ParseArgsSchema("")
	if err != nil || len(schema) != 0 {
		t.Fatalf("空模板应得空表: %v", err)
	}
	if _, err := ParseArgsSchema(`[1,2]`); err == nil {
		t.Fatal("数组应拒绝")
	}
}

func TestValidateMcpArgs(t *testing.T) {
	schema := map[string]string{"city": "城市"}
	if err := validateMcpArgs(schema, map[string]any{"city": "北京"}); err != nil {
		t.Fatalf("合法参数应放行: %v", err)
	}
	if err := validateMcpArgs(schema, map[string]any{"hacker": "x"}); err == nil || !strings.Contains(err.Error(), "2288") {
		t.Fatalf("未知参数应报 2288: %v", err)
	}
	if err := validateMcpArgs(schema, nil); err != nil {
		t.Fatalf("空参数应放行: %v", err)
	}
}

// TestMcpAdapter_Execute httptest 全链：GET query / POST body / 未知参数拒
func TestMcpAdapter_Execute(t *testing.T) {
	var gotPath, gotQuery, gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// GET：参数走 query
	tool := model.LlmopsMcpTool{
		Name: "查天气", Desc: "查询城市天气", Endpoint: srv.URL + "/weather", Method: "GET",
		ArgsSchema: `{"city":"城市"}`,
	}
	ad := &mcpToolAdapter{tool: tool}
	out, err := ad.Execute(nil, `{"city":"北京"}`)
	if err != nil || !strings.Contains(out, "HTTP 200") || !strings.Contains(out, `"ok":true`) {
		t.Fatalf("GET 代理异常: %q %v", out, err)
	}
	if gotPath != "/weather" || !strings.Contains(gotQuery, "city=%E5%8C%97%E4%BA%AC") {
		t.Fatalf("GET query 异常: %s %s", gotPath, gotQuery)
	}

	// POST：参数走 JSON body
	tool2 := model.LlmopsMcpTool{
		Name: "建工单", Desc: "创建工单", Endpoint: srv.URL + "/ticket", Method: "POST",
		ArgsSchema: `{"title":"标题"}`,
	}
	ad2 := &mcpToolAdapter{tool: tool2}
	if _, err := ad2.Execute(nil, `{"title":"故障"}`); err != nil {
		t.Fatalf("POST 代理异常: %v", err)
	}
	if gotPath != "/ticket" || !strings.Contains(gotBody, `"title":"故障"`) || !strings.Contains(gotCT, "application/json") {
		t.Fatalf("POST body 异常: %s %s %s", gotPath, gotBody, gotCT)
	}

	// 未知参数
	if _, err := ad.Execute(nil, `{"evil":"x"}`); err == nil || !strings.Contains(err.Error(), "未知参数") {
		t.Fatalf("未知参数应拒: %v", err)
	}
	// 坏参数 JSON
	if _, err := ad.Execute(nil, `no-json`); err == nil {
		t.Fatal("坏参数 JSON 应拒")
	}
	_ = json.Marshal
}
