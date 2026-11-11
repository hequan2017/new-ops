// Package service crossauth：SSO 单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
	"time"
)

// TestTicketExpired 票据过期判定
func TestTicketExpired(t *testing.T) {
	now := time.Now().Unix()
	if ticketExpired(now, now+60) {
		t.Fatal("未过期应 false")
	}
	if !ticketExpired(now+61, now+60) {
		t.Fatal("过期应 true")
	}
	if ticketExpired(now, now) {
		t.Fatal("恰到期时点应 false（now > expired 才过期）")
	}
}

// TestBuildSsoRedirect 跳转 URL 组装
func TestBuildSsoRedirect(t *testing.T) {
	u, err := buildSsoRedirect("https://ai.example.com/cb?src=sso", "TICKET123", "app-ai")
	if err != nil {
		t.Fatalf("组装失败: %v", err)
	}
	if !strings.HasPrefix(u, "https://ai.example.com/cb?") {
		t.Fatalf("应保留原回调参数: %s", u)
	}
	if !strings.Contains(u, "ticket=TICKET123") || !strings.Contains(u, "appKey=app-ai") || !strings.Contains(u, "src=sso") {
		t.Fatalf("参数不齐: %s", u)
	}
	if _, err := buildSsoRedirect("not-a-url", "t", "k"); err == nil {
		t.Fatal("非法回调应报错")
	}
}
