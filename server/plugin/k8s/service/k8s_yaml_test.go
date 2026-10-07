// Package service YAML 编辑下发纯函数单测（diffLines / parseTypedWorkload / validateWorkloadIdentity）
package service

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiffLines(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want []string
	}{
		{"完全一致", []string{"a", "b"}, []string{"a", "b"}, []string{"  a", "  b"}},
		{"新增行", []string{"a", "c"}, []string{"a", "b", "c"}, []string{"  a", "+ b", "  c"}},
		{"删除行", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"  a", "- b", "  c"}},
		{"修改行", []string{"replicas: 1"}, []string{"replicas: 2"}, []string{"- replicas: 1", "+ replicas: 2"}},
		{"全删", []string{"x"}, nil, []string{"- x"}},
		{"全增", nil, []string{"y"}, []string{"+ y"}},
		{"双空", nil, nil, []string{}},
	}
	for _, tc := range cases {
		got := diffLines(tc.a, tc.b)
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Errorf("%s: diffLines = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestDiffLines_OverLimit(t *testing.T) {
	big := make([]string, 1300)
	for i := range big {
		big[i] = "line"
	}
	got := diffLines(big, big)
	if len(got) != 1 || !strings.Contains(got[0], "跳过逐行对比") {
		t.Errorf("超限应返回占位提示，实际 %v", got)
	}
}

func TestParseTypedWorkload_Kind(t *testing.T) {
	if _, _, err := parseTypedWorkload("job", []byte("{}")); err == nil {
		t.Error("非法 kind 应报错")
	}
	dep := "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: d1\n  namespace: default\n"
	obj, meta, err := parseTypedWorkload("deployment", []byte(dep))
	if err != nil {
		t.Fatalf("合法 deployment 解析失败: %v", err)
	}
	if meta.GetName() != "d1" || meta.GetNamespace() != "default" {
		t.Errorf("解析元数据错误: %s/%s", meta.GetNamespace(), meta.GetName())
	}
	if obj == nil {
		t.Error("对象不应为空")
	}
	bad := "metadata:\n  replicas: [oops\n"
	if _, _, err := parseTypedWorkload("deployment", []byte(bad)); err == nil {
		t.Error("非法 YAML 应报错")
	}
}

func TestValidateWorkloadIdentity(t *testing.T) {
	meta := &metav1.ObjectMeta{Namespace: "default", Name: "web"}
	if err := validateWorkloadIdentity(meta, "default", "web"); err != nil {
		t.Errorf("一致对象不应报错: %v", err)
	}
	err := validateWorkloadIdentity(meta, "default", "other")
	if err == nil {
		t.Fatal("名称不一致应报错")
	}
	if !strings.Contains(err.Error(), "default/other") {
		t.Errorf("报错应包含期望对象名: %v", err)
	}
}
