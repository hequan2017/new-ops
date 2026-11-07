// Package service llmops 调用日志导出单测（CSV 形状/错误字段清洗）
package service

import (
	"strings"
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestCallLogsToCSV(t *testing.T) {
	t1 := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)
	rows := []model.LlmopsModelCallLog{
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: t1}, ModelName: "deepseek-chat", Protocol: "chat", LatencyMs: 120,
			PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15, Ok: true},
		{GVA_MODEL: global.GVA_MODEL{CreatedAt: t1}, ModelName: "embed-v1", Protocol: "embed", LatencyMs: 30, Ok: false,
			Error: "超限,重试\n失败"},
	}
	csv := callLogsToCSV(rows)
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 3 {
		t.Fatalf("表头+两行: %v", lines)
	}
	if !strings.HasPrefix(csv, "时间,模型,协议,时延ms,输入tokens,输出tokens,总tokens,结果,错误") {
		t.Fatalf("表头异常: %q", csv)
	}
	if !strings.Contains(csv, "2026-11-08 10:00:00,deepseek-chat,chat,120,10,5,15,成功,") {
		t.Fatalf("成功行异常: %q", csv)
	}
	// 错误字段逗号转中文、换行转空格（CSV 列完整性）
	if !strings.Contains(csv, "超限，重试 失败") {
		t.Fatalf("错误清洗异常: %q", csv)
	}
	if !strings.Contains(csv, "embed-v1,embed,30,0,0,0,失败,") {
		t.Fatalf("失败行异常: %q", csv)
	}
	// 空输入 → 仅表头
	if got := callLogsToCSV(nil); len(strings.Split(strings.TrimSpace(got), "\n")) != 1 {
		t.Fatalf("空输入应仅表头: %q", got)
	}
}
