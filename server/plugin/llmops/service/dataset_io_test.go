// Package service llmops 数据集导入单测（purpose schema/逐行校验/坏行报告）
package service

import "testing"

func TestDatasetSampleSchemas(t *testing.T) {
	if s := datasetSampleSchemas("sft"); s["instruction"] != "string" || s["output"] != "string" {
		t.Fatalf("sft schema 异常: %v", s)
	}
	if s := datasetSampleSchemas("rlhf"); s["chosen"] != "string" || s["rejected"] != "string" {
		t.Fatalf("rlhf schema 异常: %v", s)
	}
	if s := datasetSampleSchemas("eval"); s["question"] != "string" {
		t.Fatalf("eval schema 异常: %v", s)
	}
	if s := datasetSampleSchemas("exam"); s["type"] != "string" || s["stem"] != "string" {
		t.Fatalf("exam schema 异常: %v", s)
	}
}

func TestValidateDatasetRows(t *testing.T) {
	good := `{"instruction":"什么是 SFT","output":"有监督微调"}
{"instruction":"LoRA 是","output":"低秩适配"}
{"question":"缺 output 的坏行"}
not-json
{"instruction":123,"output":"类型错"}`
	valid, errs := ValidateDatasetRows(good, "sft")
	if valid != 2 {
		t.Fatalf("有效行应 2: %d", valid)
	}
	if len(errs) != 3 {
		t.Fatalf("坏行应 3: %+v", errs)
	}
	if errs[0].Line != 3 || !containsStr(errs[0].Reason, "output") {
		t.Fatalf("缺字段报告异常: %+v", errs[0])
	}
	if errs[1].Line != 4 || !containsStr(errs[1].Reason, "JSON") {
		t.Fatalf("坏 JSON 报告异常: %+v", errs[1])
	}
	if errs[2].Line != 5 || !containsStr(errs[2].Reason, "字符串") {
		t.Fatalf("类型错报告异常: %+v", errs[2])
	}
	// 全坏
	valid2, errs2 := ValidateDatasetRows("bad\nbad2", "eval")
	if valid2 != 0 || len(errs2) != 2 {
		t.Fatalf("全坏应 0 有效: %d %+v", valid2, errs2)
	}
	// 空行跳过
	valid3, _ := ValidateDatasetRows("\n\n{\"question\":\"q\",\"expected\":\"e\"}\n", "eval")
	if valid3 != 1 {
		t.Fatalf("空行应跳过: %d", valid3)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
