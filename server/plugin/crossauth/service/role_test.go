// Package service crossauth：角色模板单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestValidatePermJSON 权限 JSON 校验
func TestValidatePermJSON(t *testing.T) {
	if err := validatePermJSON(`{"grants":[{"resource":"教学资源库","actions":["read"]}]}`); err != nil {
		t.Fatalf("合法权限应通过: %v", err)
	}
	if err := validatePermJSON(`{"foo":1}`); err == nil {
		t.Fatal("缺 grants 应报错")
	}
	if err := validatePermJSON(`not-json`); err == nil {
		t.Fatal("非法 JSON 应报错")
	}
}
