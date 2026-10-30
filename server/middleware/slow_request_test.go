package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestResolveSlowThreshold(t *testing.T) {
	cases := []struct {
		name string
		sec  int
		want time.Duration
	}{
		{"未配置用默认2s", 0, 2 * time.Second},
		{"正数原样生效", 5, 5 * time.Second},
		{"负数关闭", -1, 0},
	}
	for _, c := range cases {
		if got := resolveSlowThreshold(c.sec); got != c.want {
			t.Fatalf("%s: want %v got %v", c.name, c.want, got)
		}
	}
}

func TestSlowRequestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := global.GVA_LOG
	t.Cleanup(func() { global.GVA_LOG = orig })

	core, logs := observer.New(zapcore.WarnLevel)
	global.GVA_LOG = zap.New(core)

	r := gin.New()
	r.GET("/slow", slowRequest(20*time.Millisecond), func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})
	r.GET("/off", slowRequest(-time.Second), func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("slow 用例应 200 got %d", w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/off", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("关闭用例应 200 got %d", w.Code)
	}

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("应恰好 1 条慢请求 warn（关闭用例不计）got %d", len(entries))
	}
	if entries[0].Message != "slow request" {
		t.Fatalf("日志消息 want slow request got %q", entries[0].Message)
	}
	path := ""
	for _, f := range entries[0].Context {
		if f.Key == "path" {
			path = f.String
		}
	}
	if path != "/slow" {
		t.Fatalf("path 字段 want /slow got %q", path)
	}
}
