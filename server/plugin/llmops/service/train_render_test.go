// Package service llmops 训练渲染器单测（白名单缺省/命令快照/防注入）
package service

import (
	"strings"
	"testing"
)

func TestParseHyper(t *testing.T) {
	// 空超参 → 全缺省
	p, err := ParseHyper("")
	if err != nil {
		t.Fatalf("空超参应放行: %v", err)
	}
	if p.LearningRate != 2e-5 || p.Epochs != 3 || p.BatchSize != 8 || p.MixedPrecision != "bf16" || p.LoraR != 8 {
		t.Fatalf("缺省值异常: %+v", p)
	}
	// 显式超参 + 未知字段忽略
	p, err = ParseHyper(`{"learningRate":0.001,"epochs":5,"unknownField":"x","outputDir":"/data/out"}`)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if p.LearningRate != 0.001 || p.Epochs != 5 || p.OutputDir != "/data/out" {
		t.Fatalf("显式超参异常: %+v", p)
	}
	// 非法精度
	if _, err := ParseHyper(`{"mixedPrecision":"int8"}`); err == nil || !strings.Contains(err.Error(), "bf16/fp16/off") {
		t.Fatalf("非法精度应拒绝: %v", err)
	}
	// 坏 JSON
	if _, err := ParseHyper(`{bad`); err == nil {
		t.Fatal("坏 JSON 应拒绝")
	}
}

func TestRenderTrainCommand(t *testing.T) {
	p, _ := ParseHyper(`{"learningRate":0.001,"gpuNum":4}`)
	// 全参快照锚点
	cmd, err := RenderTrainCommand("sft_full", p, "/data/datasets/9", "/models/qwen-7b")
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	for _, want := range []string{
		"torchrun --nproc_per_node=4", "train_sft_full.py", "--model_name_or_path /models/qwen-7b",
		"--dataset_path /data/datasets/9", "--learning_rate 0.001", "--epochs 3", "--precision bf16",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("缺 %q:\n%s", want, cmd)
		}
	}
	// LoRA 附加参数；全参不带
	lora, _ := RenderTrainCommand("sft_lora", p, "/d", "/m")
	if !strings.Contains(lora, "--lora_r 8 --lora_alpha 16 --lora_dropout 0.05") {
		t.Fatalf("LoRA 参数缺失:\n%s", lora)
	}
	if strings.Contains(cmd, "--lora_r") {
		t.Fatal("全参不应带 LoRA 参数")
	}
	// reward：配对损失口径；rl：奖励模型路径锚点 + 缺路径拒绝
	reward, err := RenderTrainCommandEx("reward", p, "/d", "/m", "")
	if err != nil || !strings.Contains(reward, "train_reward.py --model_name_or_path /m") || !strings.Contains(reward, "--loss sequence_pairwise") {
		t.Fatalf("reward 渲染异常: %v\n%s", err, reward)
	}
	rl, err := RenderTrainCommandEx("rl", p, "/d", "/m", "/data/outputs/reward-1")
	if err != nil || !strings.Contains(rl, "--reward_model_path /data/outputs/reward-1") {
		t.Fatalf("rl 渲染异常: %v\n%s", err, rl)
	}
	if _, err := RenderTrainCommandEx("rl", p, "/d", "/m", ""); err == nil || !strings.Contains(err.Error(), "奖励模型路径") {
		t.Fatalf("rl 缺奖励模型路径应拒绝: %v", err)
	}
	// 未接入类型
	if _, err := RenderTrainCommand("eval", p, "/d", "/m"); err == nil {
		t.Fatal("eval 渲染应被拦截")
	}
}

// TestSanitizeTrainArg 注入防护契约
func TestSanitizeTrainArg(t *testing.T) {
	for _, bad := range []string{"a;rm -rf /", "x&&y", "`id`", "$(whoami)", "a|b", "a<b", "line\nbreak", `quo"te`, "back\\slash"} {
		if err := sanitizeTrainArg(bad); err == nil {
			t.Fatalf("应拒绝: %q", bad)
		}
	}
	for _, ok := range []string{"/data/outputs", "qwen-7b-instruct", "/models/llama3.1-8b"} {
		if err := sanitizeTrainArg(ok); err != nil {
			t.Fatalf("应放行: %q %v", ok, err)
		}
	}
	// 经超参注入路径的整链防护
	if _, err := ParseHyper(`{"outputDir":"/tmp/x; touch /tmp/pwned"}`); err == nil {
		t.Fatal("超参注入 outputDir 应被拦截")
	}
}
