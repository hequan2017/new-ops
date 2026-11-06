// Package service llmops 产物注册单测（命名 slug/供应商标注/能力守卫）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestArtifactModelName(t *testing.T) {
	cases := map[string]string{
		"qwen-7b-lora":   "qwen-7b-lora-served",
		"中文 模型!":         "artifact-0-served", // 全剔除 → id 兜底（id=0 时按名称兜底约定）
		"  ABC_def 123 ": "ABC_def-123-served",
	}
	for in, want := range cases {
		if got := artifactModelName(in, 0); got != want {
			t.Fatalf("artifactModelName(%q)=%q want %q", in, got, want)
		}
	}
	// 空名按 ID 兜底
	if got := artifactModelName("  ", 7); got != "artifact-7-served" {
		t.Fatalf("空名应按 ID 兜底: %q", got)
	}
	if got := artifactModelName(strings.Repeat("a", 120), 1); len(got) != 96 {
		t.Fatalf("超长应截断 96: %d", len(got))
	}
}

func TestArtifactProviderLabel(t *testing.T) {
	art := &model.LlmopsArtifact{JobType: "compress", Method: "quantize"}
	if got := artifactProviderLabel(art); got != "artifact:compress:quantize" {
		t.Fatalf("压缩产物标注异常: %q", got)
	}
	art = &model.LlmopsArtifact{JobType: "sft_lora"}
	if got := artifactProviderLabel(art); got != "artifact:sft_lora" {
		t.Fatalf("训练产物标注异常: %q", got)
	}
}

// TestRegisterArtifactModel_Guard 能力/URL 守卫先于 DB 触达
func TestRegisterArtifactModel_Guard(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	if _, err := svc.RegisterArtifactAsModel(&ArtifactModelIn{Capability: "vision", BaseUrl: "http://x/v1", ApiKey: "k"}); err == nil || !strings.Contains(err.Error(), "2228") {
		t.Fatalf("非法能力应报 2228: %v", err)
	}
	if _, err := svc.RegisterArtifactAsModel(&ArtifactModelIn{Capability: "chat", BaseUrl: "notaurl", ApiKey: "k"}); err == nil {
		t.Fatal("非法 URL 应拒绝")
	}
}
