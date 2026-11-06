// Package service llmops 模型压缩单测（规格解析三方法/渲染快照/产物登记校验）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestParseCompressSpec(t *testing.T) {
	// quantize 缺省补齐
	spec, err := ParseCompressSpec(`{"compressSpec":{"method":"quantize"}}`)
	if err != nil || spec.Method != "quantize" || spec.Bits != 4 || spec.GroupSize != 128 {
		t.Fatalf("quantize 缺省异常: %+v %v", spec, err)
	}
	// bits 非法收敛 4
	spec, _ = ParseCompressSpec(`{"compressSpec":{"method":"quantize","bits":3}}`)
	if spec.Bits != 4 {
		t.Fatalf("bits=3 应收敛 4: %d", spec.Bits)
	}
	// prune 越界拒绝
	if _, err := ParseCompressSpec(`{"compressSpec":{"method":"prune","sparsity":1.5}}`); err == nil || !strings.Contains(err.Error(), "2262") {
		t.Fatalf("稀疏度越界应报 2262: %v", err)
	}
	// distill 教师模型注入拦截
	if _, err := ParseCompressSpec(`{"compressSpec":{"method":"distill","teacherModel":"m;rm -rf /"}}`); err == nil {
		t.Fatal("教师模型注入应拦截")
	}
	// 非法方法
	if _, err := ParseCompressSpec(`{"compressSpec":{"method":"magic"}}`); err == nil {
		t.Fatal("非法方法应拒绝")
	}
	// 坏 JSON
	if _, err := ParseCompressSpec(`{bad`); err == nil {
		t.Fatal("坏 JSON 应拒绝")
	}
}

func TestRenderCompressCommand(t *testing.T) {
	p, _ := ParseHyper(`{"outputDir":"/data/out"}`)
	// quantize
	cmd, err := RenderCompressCommand(&CompressSpec{Method: "quantize", Bits: 4, GroupSize: 128}, p, "/models/qwen-7b")
	if err != nil || !strings.Contains(cmd, "--method quantize --model_path /models/qwen-7b") ||
		!strings.Contains(cmd, "--bits 4 --group_size 128") || !strings.Contains(cmd, "--output_dir /data/out") {
		t.Fatalf("quantize 渲染异常: %v\n%s", err, cmd)
	}
	// prune
	cmd, _ = RenderCompressCommand(&CompressSpec{Method: "prune", Sparsity: 0.5}, p, "/m")
	if !strings.Contains(cmd, "--sparsity 0.5") {
		t.Fatalf("prune 渲染异常: %s", cmd)
	}
	// distill
	cmd, _ = RenderCompressCommand(&CompressSpec{Method: "distill", TeacherModel: "/m/teacher", Temperature: 2}, p, "/m")
	for _, want := range []string{"--teacher_model /m/teacher", "--temperature 2"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("distill 缺 %q: %s", want, cmd)
		}
	}
	// 路径注入拦截
	if _, err := RenderCompressCommand(&CompressSpec{Method: "prune", Sparsity: 0.5}, p, "/m;reboot"); err == nil {
		t.Fatal("模型路径注入应拦截")
	}
}

// TestRegisterArtifactGuard 产物登记校验（路径必填+压缩方法白名单，校验先于 DB 触达）
func TestRegisterArtifactGuard(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	if _, err := svc.RegisterArtifact(nil); err == nil {
		t.Fatal("nil 应拒绝")
	}
	if _, err := svc.RegisterArtifact(&model.LlmopsArtifact{Name: "x", Path: ""}); err == nil {
		t.Fatal("路径为空应拒绝")
	}
	if _, err := svc.RegisterArtifact(&model.LlmopsArtifact{Name: "x", JobType: "compress", Method: "magic", Path: "/p"}); err == nil || !strings.Contains(err.Error(), "压缩方法") {
		t.Fatalf("非法压缩方法应拒绝: %v", err)
	}
	// 路径注入
	if _, err := svc.RegisterArtifact(&model.LlmopsArtifact{Name: "x", Path: "/p;id"}); err == nil {
		t.Fatal("路径注入应拒绝")
	}
}
