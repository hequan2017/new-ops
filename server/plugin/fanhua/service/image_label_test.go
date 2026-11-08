// Package service fanhua：图片标注单测（纯函数，无 DB）
package service

import "testing"

// TestValidateImageSplit 用途白名单
func TestValidateImageSplit(t *testing.T) {
	for _, s := range []string{"train", "test", "none"} {
		if !validateImageSplit(s) {
			t.Fatalf("%s 应合法", s)
		}
	}
	if validateImageSplit("val") || validateImageSplit("") {
		t.Fatal("白名单外应拒绝")
	}
}

// TestValidateTagsJSON 标签 JSON 校验
func TestValidateTagsJSON(t *testing.T) {
	tags, err := validateTagsJSON(`["弄堂","石库门"]`)
	if err != nil || len(tags) != 2 {
		t.Fatalf("合法标签应通过: %v %v", tags, err)
	}
	if _, err = validateTagsJSON(`["a",""]`); err == nil {
		t.Fatal("空串标签应拒绝")
	}
	if _, err = validateTagsJSON(`not json`); err == nil {
		t.Fatal("非 JSON 应拒绝")
	}
}
