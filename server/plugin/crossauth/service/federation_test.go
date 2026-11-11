// Package service crossauth：身份对接单测（纯函数 + httptest 假端点，无外部服务/无 DB）
package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestParseCaPull 拉取负载解析
func TestParseCaPull(t *testing.T) {
	rows, err := parseCaPull([]byte(`[
		{"name":"张三","userNo":"2026001","dept":"计算机学院","identityStatus":"在读"},
		{"name":"李四","userNo":"T0001","dept":"财务处","identityStatus":"在职"}
	]`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 2 || rows[1].UserNo != "T0001" {
		t.Fatalf("应解析 2 行: %+v", rows)
	}
	if _, err := parseCaPull([]byte(`[]`)); err != nil {
		t.Fatalf("空数组应合法: %v", err)
	}
	if _, err := parseCaPull([]byte(`bad`)); err == nil {
		t.Fatal("非法负载应报错")
	}
}

// TestValidCaIdentityStatus 身份状态白名单
func TestValidCaIdentityStatus(t *testing.T) {
	for _, s := range CaIdentityStatuses {
		if !ValidCaIdentityStatus(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if ValidCaIdentityStatus("毕业") {
		t.Fatal("白名单外应拒绝")
	}
}

// TestCaPullViaHttptest httptest 假端点（鉴权+拉取+解析全链路）
func TestCaPullViaHttptest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ca-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `[{"name":"王五","userNo":"2026002","dept":"机械学院","identityStatus":"在读"}]`)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Authorization", "Bearer ca-secret")
	resp, err := caHttpClient.Do(req)
	if err != nil {
		t.Fatalf("拉取失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	rows, err := parseCaPull(body)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 1 || rows[0].IdentityStatus != "在读" {
		t.Fatalf("应解析 1 行在读用户: %+v", rows)
	}
}
