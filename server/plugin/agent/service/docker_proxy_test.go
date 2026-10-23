package service

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestRoundTripOffline agent 未在线时应即时报错（不悬挂）
func TestRoundTripOffline(t *testing.T) {
	tr := &AgentDockerTransport{AssetID: 999999}
	req, _ := http.NewRequest(http.MethodGet, "http://docker.proxy/_ping", nil)
	resp, err := tr.RoundTrip(req)
	if err == nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("离线应返回错误")
	}
	if !strings.Contains(err.Error(), "agent 未在线") {
		t.Fatalf("错误应来自注册表: %v", err)
	}
	if resp != nil {
		t.Fatal("离线不应返回响应")
	}
}

// TestRoundTripBodyLimit 超限请求体应拒绝
func TestRoundTripBodyLimit(t *testing.T) {
	tr := &AgentDockerTransport{AssetID: 999999}
	req, _ := http.NewRequest(http.MethodPost, "http://docker.proxy/x", strings.NewReader(strings.Repeat("a", 5<<20)))
	if _, err := tr.RoundTrip(req); err == nil || !strings.Contains(err.Error(), "超限") {
		t.Fatalf("超限应报错: %v", err)
	}
}

var _ io.Reader = (*strings.Reader)(nil)
