// Package service eduadmin：数据同步单测（纯函数 + httptest 假端点，无外部服务/无 DB）
package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestParsePull 拉取负载解析
func TestParsePull(t *testing.T) {
	rows, err := parsePull([]byte(`[
		{"externalKey":"HR-100","payload":{"name":"张三","dept":"教务处"}},
		{"externalKey":"HR-101","payload":{"name":"李四"}}
	]`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 2 || rows[0].ExternalKey != "HR-100" {
		t.Fatalf("应解析 2 行: %+v", rows)
	}
	if _, err := parsePull([]byte(`not-json`)); err == nil {
		t.Fatal("非法负载应报错")
	}
}

// TestValidEduSyncSource 五类来源白名单
func TestValidEduSyncSource(t *testing.T) {
	for _, s := range EduSyncSources {
		if !ValidEduSyncSource(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if ValidEduSyncSource("招生") {
		t.Fatal("白名单外应拒绝")
	}
}

// TestPullViaHttptest httptest 假端点：鉴权头+负载拉取+解析全链路（模拟适配器拉取）
func TestPullViaHttptest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `[
			{"externalKey":"SYNC-1","payload":{"k":"v1"}},
			{"externalKey":"SYNC-2","payload":{"k":"v2"}}
		]`)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := httpClientSync.Do(req)
	if err != nil {
		t.Fatalf("拉取失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	rows, err := parsePull(body)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 2 || rows[1].ExternalKey != "SYNC-2" {
		t.Fatalf("应解析 2 行: %+v", rows)
	}

	// 无凭据 → 401
	resp2, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("二次请求失败: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("无凭据应 401: %d", resp2.StatusCode)
	}
}
