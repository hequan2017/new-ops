// Package service llmops 数据集单测（用途/格式白名单、样本计数）
package service

import (
	"strings"
	"testing"
)

func TestValidDatasetPurpose(t *testing.T) {
	for _, ok := range DatasetPurposes {
		if !ValidDatasetPurpose(ok) {
			t.Fatalf("合法用途 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "SFT", "train", "sft;"} {
		if ValidDatasetPurpose(bad) {
			t.Fatalf("非法用途 %q 被放行", bad)
		}
	}
}

func TestValidDatasetFormat(t *testing.T) {
	if !ValidDatasetFormat("jsonl") || !ValidDatasetFormat("csv") {
		t.Fatal("合法格式应放行")
	}
	for _, bad := range []string{"", "JSONL", "parquet", "jsonl;"} {
		if ValidDatasetFormat(bad) {
			t.Fatalf("非法格式 %q 被放行", bad)
		}
	}
}

func TestCountDatasetSamples(t *testing.T) {
	// jsonl：非空行
	jsonl := "{\"a\":1}\n\n{\"a\":2}\n  \n{\"a\":3}"
	if got := CountDatasetSamples(jsonl, "jsonl"); got != 3 {
		t.Fatalf("jsonl 计数异常: %d", got)
	}
	// csv：行-表头
	csv := "a,b\n1,2\n3,4\n5,6"
	if got := CountDatasetSamples(csv, "csv"); got != 3 {
		t.Fatalf("csv 计数异常: %d", got)
	}
	// 空内容 / csv 仅表头
	if got := CountDatasetSamples("", "jsonl"); got != 0 {
		t.Fatalf("空应 0: %d", got)
	}
	if got := CountDatasetSamples("a,b", "csv"); got != 0 {
		t.Fatalf("仅表头应 0: %d", got)
	}
	// 未匹配格式按 jsonl 口径
	if got := CountDatasetSamples("x\ny", "weird"); got != 2 {
		t.Fatalf("未匹配格式应回退 jsonl: %d", got)
	}
	_ = strings.TrimSpace
}
