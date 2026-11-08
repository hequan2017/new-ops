// Package service artsci：文生图场景单测（纯函数+适配器契约，无 DB）
package service

import (
	"context"
	"strings"
	"testing"
)

// TestRenderPrompt 占位符渲染
func TestRenderPrompt(t *testing.T) {
	out := renderPrompt("场景 {{scene}}：{{text}}，再提一次 {{text}}", "历史场景再现", "淞沪抗战")
	if out != "场景 历史场景再现：淞沪抗战，再提一次 淞沪抗战" {
		t.Fatalf("渲染不符: %s", out)
	}
	if out2 := renderPrompt("无占位符模板", "x", "y"); out2 != "无占位符模板" {
		t.Fatalf("无占位符应原样: %s", out2)
	}
}

// TestNoopImageGen 默认适配器契约：未配置明确报 2515（替换点由后续模型接入实现）
func TestNoopImageGen(t *testing.T) {
	if _, ok := ImageGen.(NoopImageGenAdapter); !ok {
		t.Fatalf("默认适配器应为 NoopImageGenAdapter")
	}
	_, err := ImageGen.Generate(context.Background(), "p")
	if err == nil || !strings.Contains(err.Error(), "2515") {
		t.Fatalf("未配置应报 2515: %v", err)
	}
}

// TestImageGenAdapterInterface 编译期接口契约 + 假适配器替换点
func TestImageGenAdapterInterface(t *testing.T) {
	var _ ImageGenAdapter = NoopImageGenAdapter{}
	var _ ImageGenAdapter = fakeImageGen{}
	url, err := fakeImageGen{}.Generate(context.Background(), "p")
	if err != nil || url != "http://fake/img.png" {
		t.Fatalf("替换点未生效: %q %v", url, err)
	}
}

type fakeImageGen struct{}

func (fakeImageGen) Generate(context.Context, string) (string, error) {
	return "http://fake/img.png", nil
}

// TestSceneNames 五类场景白名单（需求 3.3-10/11/12/14/15）
func TestSceneNames(t *testing.T) {
	want := []string{"历史场景再现", "文学作品可视化", "地理地貌展示", "图像作文训练", "跨学科学习"}
	if len(SceneNames) != len(want) {
		t.Fatalf("场景数应为 5")
	}
	for i, s := range want {
		if SceneNames[i] != s {
			t.Fatalf("场景 %d 应为 %s", i, s)
		}
	}
}
