// Package service llmops 连接器单测（鉴权头构造/字段映射/类型白名单）
package service

import "testing"

func TestBuildAuthHeaders(t *testing.T) {
	cases := []struct {
		authType, value string
		want            map[string]string
	}{
		{"bearer", "tok123", map[string]string{"Authorization": "Bearer tok123"}},
		{"basic", "dXNlcjpwdw==", map[string]string{"Authorization": "Basic dXNlcjpwdw=="}},
		{"apikey", "key-1", map[string]string{"X-API-Key": "key-1"}},
		{"none", "", map[string]string{}},
		{"unknown", "x", map[string]string{}},
	}
	for _, c := range cases {
		got := buildAuthHeaders(c.authType, c.value)
		if len(got) != len(c.want) {
			t.Fatalf("buildAuthHeaders(%q) 长度异常: %v", c.authType, got)
		}
		for k, v := range c.want {
			if got[k] != v {
				t.Fatalf("buildAuthHeaders(%q)[%s]=%q want %q", c.authType, k, got[k], v)
			}
		}
	}
}

func TestParseFieldMap(t *testing.T) {
	m, err := ParseFieldMap(`{"userName":"姓名","dept":"部门"}`)
	if err != nil || len(m) != 2 || m["userName"] != "姓名" {
		t.Fatalf("映射解析异常: %+v %v", m, err)
	}
	m, err = ParseFieldMap("")
	if err != nil || len(m) != 0 {
		t.Fatalf("空映射应得空表: %v", err)
	}
	if _, err := ParseFieldMap(`[1]`); err == nil {
		t.Fatal("数组应拒绝")
	}
}

func TestValidConnectorTypeAndAuth(t *testing.T) {
	for _, ok := range []string{"erp", "crm", "oa", "scm"} {
		if !ValidConnectorType(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	if ValidConnectorType("wms") || ValidConnectorType("ERP") {
		t.Fatal("非法类型应拒绝")
	}
	for _, ok := range []string{"none", "bearer", "basic", "apikey"} {
		if !ValidAuthType(ok) {
			t.Fatalf("合法鉴权 %q 被拒绝", ok)
		}
	}
	if ValidAuthType("oauth") || ValidAuthType("") {
		t.Fatal("非法鉴权应拒绝")
	}
}
