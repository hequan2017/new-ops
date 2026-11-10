// Package service eduadmin：数据脱敏与分级单测（纯函数，无 DB）
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// TestHashSensitive SHA-256 确定性与不可逆（标准向量）
func TestHashSensitive(t *testing.T) {
	sum := sha256.Sum256([]byte("310101199001011234"))
	want := hex.EncodeToString(sum[:])
	if got := HashSensitive("310101199001011234"); got != want {
		t.Fatalf("SHA-256 应匹配标准向量: %s != %s", got, want)
	}
	if HashSensitive("a") == HashSensitive("b") {
		t.Fatal("不同输入应不同哈希")
	}
	if strings.Contains(want, "310101") {
		t.Fatal("哈希不应包含原文片段（不可逆）")
	}
}

// TestHideAmountDecimals 金额取整
func TestHideAmountDecimals(t *testing.T) {
	if v := HideAmountDecimals(12345.67); v != 12345 {
		t.Fatalf("应截断小数: %v", v)
	}
	if v := HideAmountDecimals(-88.9); v != -88 {
		t.Fatalf("负数应向零截断: %v", v)
	}
}

// TestClassifyPayload 四级分级取最高
func TestClassifyPayload(t *testing.T) {
	cases := map[string]string{
		`{"note":"公开通知，教室调整"}`:                    "公开",
		`{"部门":"计算机学院","项目":"重点实验室"}`:   "内部",
		`{"salary":9000,"perfNote":"绩效优秀"}`:        "敏感",
		`{"idCard":"310101199001011234","name":"张三"}`: "机密",
		`{"bankCard":"6222021234567890123"}`:          "机密",
	}
	for payload, want := range cases {
		if got := classifyPayload(payload); got != want {
			t.Fatalf("%s 应分级 %s，得 %s", payload, want, got)
		}
	}
}

// TestMaskPayload 脱敏转换集成：哈希字段+金额字段+分级
func TestMaskPayload(t *testing.T) {
	res, err := MaskPayload(`{"name":"李四","idCard":"310101199001011234","bankCard":"6222021234567890123","amount":1234.56,"note":"薪酬调整"}`)
	if err != nil {
		t.Fatalf("脱敏失败: %v", err)
	}
	if len(res.HashedFields) != 2 {
		t.Fatalf("应哈希 2 个字段: %+v", res.HashedFields)
	}
	if !strings.Contains(res.Masked, HashSensitive("310101199001011234")) {
		t.Fatal("身份证应被哈希替换")
	}
	if strings.Contains(res.Masked, "310101199001011234") {
		t.Fatal("原文身份证不应残留")
	}
	if len(res.AmountFields) != 1 || !strings.Contains(res.Masked, `"amount":1234`) {
		t.Fatalf("金额应取整: %s", res.Masked)
	}
	if res.Level != "机密" {
		t.Fatalf("含身份证应机密: %s", res.Level)
	}
	if _, err := MaskPayload(`not-json`); err == nil {
		t.Fatal("非 JSON 对象应报错")
	}
}
