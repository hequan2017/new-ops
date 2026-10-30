package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	HealthDetail(c)

	if w.Code != http.StatusOK {
		t.Fatalf("状态码 want 200 got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应应为合法 JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("status want ok got %v", resp["status"])
	}
	if v, _ := resp["version"].(string); v == "" {
		t.Fatalf("version want 非空 got %v", resp["version"])
	}
	if bv, _ := resp["base_version"].(string); bv == "" {
		t.Fatalf("base_version want 非空 got %v", resp["base_version"])
	}
	if up, ok := resp["uptime"].(float64); !ok || up < 0 {
		t.Fatalf("uptime want 非负数值 got %v", resp["uptime"])
	}
	if gv, _ := resp["go_version"].(string); !strings.HasPrefix(gv, "go") {
		t.Fatalf("go_version want go 前缀 got %v", resp["go_version"])
	}
	if ap, _ := resp["app"].(string); ap == "" {
		t.Fatalf("app want 非空 got %v", resp["app"])
	}
}
