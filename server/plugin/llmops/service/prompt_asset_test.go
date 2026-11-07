// Package service llmops Prompt 配套资产单测（类型白名单）
package service

import "testing"

func TestValidPromptAssetKind(t *testing.T) {
	for _, ok := range []string{"rule", "term", "cot"} {
		if !ValidPromptAssetKind(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "RULE", "term,cot", "faq"} {
		if ValidPromptAssetKind(bad) {
			t.Fatalf("非法类型 %q 被放行", bad)
		}
	}
}
