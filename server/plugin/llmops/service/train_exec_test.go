// Package service llmops 训练执行单测（加速参数渲染/日志追加）
package service

import (
	"strings"
	"testing"
)

// TestRenderAccelFlags 训练加速参数渲染（ZeRO/梯度检查点）
func TestRenderAccelFlags(t *testing.T) {
	p, err := ParseHyper(`{"zeroStage":2,"gradientCheckpointing":true,"gpuNum":8}`)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	cmd, err := RenderTrainCommand("sft_full", p, "/d", "/m")
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	for _, want := range []string{"--deepspeed --zero_stage 2", "--gradient_checkpointing", "--nproc_per_node=8"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("缺加速参数 %q:\n%s", want, cmd)
		}
	}
	// 零值不渲染
	p2, _ := ParseHyper("")
	cmd2, _ := RenderTrainCommand("sft_full", p2, "/d", "/m")
	if strings.Contains(cmd2, "--deepspeed") || strings.Contains(cmd2, "--gradient_checkpointing") {
		t.Fatalf("未启用不应渲染:\n%s", cmd2)
	}
	// 越界拒绝
	if _, err := ParseHyper(`{"zeroStage":4}`); err == nil || !strings.Contains(err.Error(), "0-3") {
		t.Fatalf("zeroStage 越界应拒绝: %v", err)
	}
}

// TestAppendTrainLog 日志追加格式（时间戳段落/首段无前导换行）
func TestAppendTrainLog(t *testing.T) {
	first := appendTrainLog("", "START", "cmd line")
	if strings.HasPrefix(first, "\n") || !strings.Contains(first, "START") || !strings.Contains(first, "cmd line") {
		t.Fatalf("首段格式异常: %q", first)
	}
	second := appendTrainLog(first, "STOP", "手动停止")
	if !strings.Contains(second, first) || !strings.Contains(second, "手动停止") {
		t.Fatalf("追加应保留原日志: %q", second)
	}
	if strings.Count(second, "\n[") != 1 || strings.Count(second, "] ") != 2 {
		t.Fatalf("应为两个段落(首段无前导换行): %q", second)
	}
}
