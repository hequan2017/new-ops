// Package service eduadmin：政策数据更新单测（纯函数 + httptest 假端点，无 DB）
package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestParseGovPolicy 负载解析
func TestParseGovPolicy(t *testing.T) {
	rows, err := parseGovPolicy([]byte(`[
		{"title":"关于做好2026年高校招生工作的通知","url":"https://moe.gov.cn/a1","publishedAt":"2026-10-20"},
		{"title":"上海市教育综合改革方案","url":"https://shmec.edu.cn/a2","publishedAt":"2026-10-22"}
	]`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 2 || rows[1].Url == "" {
		t.Fatalf("应解析 2 行: %+v", rows)
	}
	if rows[0].Title == "" || rows[0].Url == "" {
		t.Fatalf("字段应完整: %+v", rows[0])
	}
	if _, err := parseGovPolicy([]byte(`{"a":1}`)); err == nil {
		t.Fatal("对象负载应报错")
	}
}

// TestGovContentHash 幂等键确定性
func TestGovContentHash(t *testing.T) {
	h1 := govContentHash("标题A", "https://moe.gov.cn/1")
	if h1 != govContentHash("标题A", "https://moe.gov.cn/1") {
		t.Fatal("同文同链应同哈希")
	}
	if h1 == govContentHash("标题A", "https://moe.gov.cn/2") {
		t.Fatal("不同链接应不同哈希")
	}
	if len(h1) != 64 {
		t.Fatalf("应为 64 位 hex: %s", h1)
	}
}

// TestPullGovViaHttptest httptest 假端点全链路（HTTP 拉取+解析，无 DB）
func TestPullGovViaHttptest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `[{"title":"通知","url":"https://moe.gov.cn/x","publishedAt":"2026-11-01"}]`)
	}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("拉取失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	rows, err := parseGovPolicy(body)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 1 || rows[0].PublishedAt != "2026-11-01" {
		t.Fatalf("应解析 1 行含发布时间: %+v", rows)
	}
}
