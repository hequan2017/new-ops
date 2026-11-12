package errx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// errChainWrap 链式包裹：模拟 repo→service 的错误透传
func errChainWrap(cause error) error {
	return Wrap(1002, "导入失败", cause)
}

func TestNewAndCodeOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"AppError 直返", New(1605, "资源不足"), 1605},
		{"Wrap 链上取码", errChainWrap(errors.New("json bad")), 1002},
		{"普通错误为零码", errors.New("boom"), 0},
		{"nil 为零码", nil, 0},
		{"wrapped 标准链", Wrap(1203, "执行失败", errors.New("ssh dial")), 1203},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CodeOf(tc.err); got != tc.want {
				t.Fatalf("CodeOf=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestWrapMessageAndUnwrap(t *testing.T) {
	cause := errors.New("dial tcp refused")
	err := Wrap(1701, "SQL 审核失败", cause)
	if err.Error() != "SQL 审核失败: dial tcp refused" {
		t.Fatalf("消息口径不符: %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("Unwrap 链断裂，errors.Is 未命中底层原因")
	}
	// 无因构造不追加冒号
	if got := New(1, "仅消息").Error(); got != "仅消息" {
		t.Fatalf("无因消息口径不符: %q", got)
	}
}

func TestRespond(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name     string
		err      error
		fallback string
		wantCode int
		wantMsg  string
	}{
		{"AppError 带分段码", New(2903, "互认校验失败"), "默认失败", 2903, "互认校验失败"},
		{"普通错误退化通用码", errors.New("boom"), "默认失败", 7, "默认失败"},
	}
	for i, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		Respond(c, tc.err, tc.fallback)
		if w.Code != http.StatusOK {
			t.Fatalf("case %d(%s): HTTP 状态应为 200，实际 %d", i, tc.name, w.Code)
		}
		var got struct {
			Code int    `json:"code"`
			Data any    `json:"data"`
			Msg  string `json:"msg"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("case %d(%s): %v", i, tc.name, err)
		}
		if got.Code != tc.wantCode || got.Msg != tc.wantMsg {
			t.Fatalf("case %d(%s): code=%d msg=%q want code=%d msg=%q",
				i, tc.name, got.Code, got.Msg, tc.wantCode, tc.wantMsg)
		}
	}
}
