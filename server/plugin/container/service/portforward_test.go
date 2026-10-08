// Package service 端口转发规则纯函数单测
package service

import (
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
)

func TestValidatePortRules(t *testing.T) {
	cases := []struct {
		name    string
		rules   []PortRule
		wantErr bool
	}{
		{"空规则", nil, true},
		{"正常 tcp", []PortRule{{HostPort: "8080", ContainerPort: 80, Proto: "tcp"}}, false},
		{"空协议默认 tcp", []PortRule{{HostPort: "53", ContainerPort: 53}}, false},
		{"udp 合法", []PortRule{{HostPort: "53", ContainerPort: 53, Proto: "UDP"}}, false},
		{"非法协议", []PortRule{{HostPort: "80", ContainerPort: 80, Proto: "sctp"}}, true},
		{"容器端口 0", []PortRule{{HostPort: "80", ContainerPort: 0}}, true},
		{"宿主端口越界", []PortRule{{HostPort: "70000", ContainerPort: 80}}, true},
		{"宿主端口非数字", []PortRule{{HostPort: "http", ContainerPort: 80}}, true},
		{"hostIP 非法", []PortRule{{HostIP: "999.1.1.1", HostPort: "80", ContainerPort: 80}}, true},
		{"hostIP 合法", []PortRule{{HostIP: "192.168.1.10", HostPort: "80", ContainerPort: 80}}, false},
		{"重复宿主端口", []PortRule{
			{HostPort: "8080", ContainerPort: 80},
			{HostPort: "8080", ContainerPort: 81},
		}, true},
		{"同端口不同协议不冲突", []PortRule{
			{HostPort: "53", ContainerPort: 53, Proto: "tcp"},
			{HostPort: "53", ContainerPort: 53, Proto: "udp"},
		}, false},
	}
	for _, tc := range cases {
		err := validatePortRules(tc.rules)
		if (err != nil) != tc.wantErr {
			t.Fatalf("%s: wantErr=%v got=%v", tc.name, tc.wantErr, err)
		}
	}
}

func TestPortRulesToBindings(t *testing.T) {
	exposed, bindings := portRulesToBindings([]PortRule{
		{HostIP: "127.0.0.1", HostPort: "8080", ContainerPort: 80, Proto: "tcp"},
		{HostPort: "5353", ContainerPort: 53, Proto: "udp"},
	})
	if len(exposed) != 2 {
		t.Fatalf("exposed 应为 2，got %d", len(exposed))
	}
	b80, ok := bindings[nat.Port("80/tcp")]
	if !ok || len(b80) != 1 || b80[0].HostIP != "127.0.0.1" || b80[0].HostPort != "8080" {
		t.Fatalf("80/tcp 绑定错误: %+v", b80)
	}
	b53, ok := bindings[nat.Port("53/udp")]
	if !ok || len(b53) != 1 || b53[0].HostPort != "5353" {
		t.Fatalf("53/udp 绑定错误: %+v", b53)
	}
}

func TestBindingsToPortRules(t *testing.T) {
	pm := nat.PortMap{
		nat.Port("80/tcp"):   []nat.PortBinding{{HostIP: "", HostPort: "8080"}},
		nat.Port("53/udp"):   []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "5353"}},
		nat.Port("9090/tcp"): nil,
	}
	rules := bindingsToPortRules(pm)
	if len(rules) != 3 {
		t.Fatalf("应解析出 3 条规则，got %d", len(rules))
	}
	m := map[string]PortRule{}
	for _, r := range rules {
		m[strconv.Itoa(r.ContainerPort)+"/"+r.Proto] = r
	}
	if r := m["80/tcp"]; r.HostPort != "8080" {
		t.Fatalf("80/tcp 规则错误: %+v", r)
	}
	if r := m["53/udp"]; r.HostIP != "0.0.0.0" || r.HostPort != "5353" {
		t.Fatalf("53/udp 规则错误: %+v", r)
	}
	if r := m["9090/tcp"]; r.HostPort != "" {
		t.Fatalf("无绑定端口规则错误: %+v", r)
	}
}
