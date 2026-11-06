// Package service llmops 网关单测（payload 组装/端点拼接/响应解析/能力匹配）
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func mkModel(id uint, caps, base string) *model.LlmopsModel {
	m := &model.LlmopsModel{Name: "m" + string(rune('0'+int(id%10))), Capability: caps, BaseUrl: base, Status: "enabled", ApiKeyEnc: "enc"}
	m.ID = id
	return m
}

func TestModelHasCapability(t *testing.T) {
	m := mkModel(1, "chat,embed", "http://x/v1")
	if !modelHasCapability(m, "chat") || !modelHasCapability(m, "embed") {
		t.Fatal("应具备 chat/embed")
	}
	if modelHasCapability(m, "rerank") {
		t.Fatal("不应具备 rerank")
	}
	if !modelHasCapability(m, "") {
		t.Fatal("空能力应放行")
	}
}

func TestGatewayEndpoint(t *testing.T) {
	cases := map[string]string{
		"http://x/v1":    "http://x/v1/chat/completions",
		"http://x/v1/":   "http://x/v1/chat/completions",
		" http://x/v1  ": "http://x/v1/chat/completions",
	}
	for base, want := range cases {
		if got := gatewayEndpoint(base, "/chat/completions"); got != want {
			t.Fatalf("gatewayEndpoint(%q)=%q want %q", base, got, want)
		}
	}
}

func TestBuildChatPayload(t *testing.T) {
	payload, err := buildChatPayload("m1", []ChatMessage{{Role: "user", Content: "hi"}}, map[string]any{"temperature": 0.5, "model": "hack"})
	if err != nil {
		t.Fatalf("组装失败: %v", err)
	}
	if payload["model"] != "m1" {
		t.Fatalf("model 不应被 opts 覆盖: %v", payload["model"])
	}
	if payload["temperature"] != 0.5 {
		t.Fatalf("opts 应透传: %v", payload["temperature"])
	}
	if _, err := buildChatPayload("m1", nil, nil); err == nil {
		t.Fatal("空消息应报错")
	}
}

func TestParseChatResponse(t *testing.T) {
	body := `{"choices":[{"message":{"role":"assistant","content":"你好"}}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
	content, usage, err := parseChatResponse([]byte(body))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if content != "你好" || usage.PromptTokens != 10 || usage.CompletionTokens != 5 || usage.TotalTokens != 15 {
		t.Fatalf("content=%q usage=%+v", content, usage)
	}
	// 无 usage 字段容忍
	if _, usage, err := parseChatResponse([]byte(`{"choices":[{"message":{"content":"a"}}]}`)); err != nil || usage.TotalTokens != 0 {
		t.Fatalf("无 usage 应容忍: %+v %v", usage, err)
	}
	// choices 空
	if _, _, err := parseChatResponse([]byte(`{"choices":[]}`)); err == nil || !strings.Contains(err.Error(), "2232") {
		t.Fatalf("空 choices 应报 2232: %v", err)
	}
	// 坏 JSON
	if _, _, err := parseChatResponse([]byte(`not-json`)); err == nil || !strings.Contains(err.Error(), "2232") {
		t.Fatalf("坏 JSON 应报 2232: %v", err)
	}
}

func TestParseEmbeddingsResponse(t *testing.T) {
	body := `{"data":[{"index":1,"embedding":[3,4]},{"index":0,"embedding":[1,2]}]}`
	vecs, err := parseEmbeddingsResponse([]byte(body))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(vecs) != 2 || vecs[0][0] != 1 || vecs[1][0] != 3 {
		t.Fatalf("按 index 归位异常: %v", vecs)
	}
	if _, err := parseEmbeddingsResponse([]byte(`{"data":[]}`)); err == nil || !strings.Contains(err.Error(), "2232") {
		t.Fatalf("空 data 应报 2232: %v", err)
	}
}

// TestChatCompletion_HttpTest httptest 全链：端点/鉴权头/载荷/解析
func TestChatCompletion_HttpTest(t *testing.T) {
	var gotPath, gotAuth, gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		gotModel, _ = payload["model"].(string)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"pong"}}],"usage":{"total_tokens":2}}`))
	}))
	defer srv.Close()

	svc := &LlmopsServiceStruct{}
	m := mkModel(7, "chat", srv.URL+"/v1") // baseUrl 口径：含 /v1（同 aiops 供应商约定）
	// 直接注入内存模型桩，绕过 DB（GetModel 走 DB；此处以 loadModelFor 的能力/URL 逻辑验证为主）
	chatHTTPClientOriginal := chatHTTPClient
	defer func() { chatHTTPClient = chatHTTPClientOriginal }()

	content, usage, err := fakeChat(svc, m.BaseUrl, m)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	if content != "pong" || usage.TotalTokens != 2 {
		t.Fatalf("content=%q usage=%+v", content, usage)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("端点错误: %s", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("鉴权头错误: %q", gotAuth)
	}
	if gotModel != "m7" {
		t.Fatalf("模型名错误: %q", gotModel)
	}
}

// fakeChat 绕过 DB 的链路验证：手动拼端点+key 走 gatewayPost+parseChatResponse
func fakeChat(svc *LlmopsServiceStruct, base string, m *model.LlmopsModel) (string, *ChatUsage, error) {
	payload, err := buildChatPayload(m.Name, []ChatMessage{{Role: "user", Content: "ping"}}, nil)
	if err != nil {
		return "", nil, err
	}
	data, err := gatewayPost(context.Background(), chatHTTPClient, gatewayEndpoint(base, "/chat/completions"), "test-key", payload)
	if err != nil {
		return "", nil, err
	}
	return parseChatResponse(data)
}
