package main

import "testing"

// Test_buildWSURL server 基地址 → WS 地址转换（http/ws、https/wss、尾斜杠容错）
func Test_buildWSURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"http 转 ws", "http://127.0.0.1:8888", "ws://127.0.0.1:8888/agent/ws"},
		{"https 转 wss", "https://ops.example.com", "wss://ops.example.com/agent/ws"},
		{"尾斜杠容错", "http://127.0.0.1:8888/", "ws://127.0.0.1:8888/agent/ws"},
		{"带路径前缀", "http://gw.example.com/ops", "ws://gw.example.com/ops/agent/ws"},
	}
	for _, c := range cases {
		if got := buildWSURL(c.in); got != c.want {
			t.Fatalf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}
