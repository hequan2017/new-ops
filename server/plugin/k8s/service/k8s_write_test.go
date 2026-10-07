package service

import (
	"strings"
	"testing"
)

func Test_validateReplicas(t *testing.T) {
	if err := validateReplicas(3); err != nil {
		t.Fatalf("合法副本数不应报错: %v", err)
	}
	if err := validateReplicas(0); err != nil {
		t.Fatalf("0 副本（缩到 0）应合法: %v", err)
	}
	err := validateReplicas(-1)
	if err == nil || !strings.Contains(err.Error(), "0-500") {
		t.Fatalf("负数应报范围错误: %v", err)
	}
	err = validateReplicas(501)
	if err == nil || !strings.Contains(err.Error(), "0-500") {
		t.Fatalf("超上限应报范围错误: %v", err)
	}
}

func Test_validateNsName(t *testing.T) {
	if err := validateNsName("default", "web"); err != nil {
		t.Fatalf("合法参数不应报错: %v", err)
	}
	err := validateNsName("", "web")
	if err == nil || !strings.Contains(err.Error(), "命名空间") {
		t.Fatalf("空命名空间应报错: %v", err)
	}
	err = validateNsName("default", "")
	if err == nil || !strings.Contains(err.Error(), "资源名称") {
		t.Fatalf("空名称应报错: %v", err)
	}
}

func Test_timeNowUTC_format(t *testing.T) {
	// restartedAt 注解要求 RFC3339 样式时间戳
	ts := timeNowUTC()
	if !strings.Contains(ts, "T") || !strings.HasSuffix(ts, "Z") {
		t.Fatalf("时间戳格式错误: %s", ts)
	}
}
