// Package service eduadmin：环境监测单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestEnvAlertLevel 阈值三档
func TestEnvAlertLevel(t *testing.T) {
	if envAlertLevel(90, 75) != "高" {
		t.Fatal("超阈值应为高")
	}
	if envAlertLevel(65, 75) != "中" { // ≥0.8×75=60
		t.Fatal("接近阈值应为中")
	}
	if envAlertLevel(30, 75) != "" {
		t.Fatal("正常应为空")
	}
	if envAlertLevel(90, 0) != "" {
		t.Fatal("阈值未配置应为空")
	}
}

// TestEnvAdvice 健康建议模板
func TestEnvAdvice(t *testing.T) {
	if a := envAdvice("空气", "PM2.5", "高"); !strings.Contains(a, "通风") {
		t.Fatalf("空气高应含通风: %s", a)
	}
	if a := envAdvice("水质", "浊度", "高"); !strings.Contains(a, "停用") {
		t.Fatalf("水质高应含停用: %s", a)
	}
	if a := envAdvice("空气", "CO2", "中"); !strings.Contains(a, "监测频次") {
		t.Fatalf("空气中应含加密监测: %s", a)
	}
}
