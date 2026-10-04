package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
)

func mkShellStep(name, content string) model.PipelineStep {
	return model.PipelineStep{Name: name, Type: model.StepTypeShell, ShellContent: content}
}

func TestValidatePipeline_OK(t *testing.T) {
	p := &model.Pipeline{
		Name: "deploy-web",
		Stages: []model.PipelineStage{
			{Name: "build", Steps: []model.PipelineStep{mkShellStep("go build", "go build ./...")}},
			{Name: "notify", Steps: []model.PipelineStep{{
				Name: "hook", Type: model.StepTypeHTTP, HTTPURL: "https://hooks.example/x",
			}}},
		},
	}
	if err := validatePipeline(p); err != nil {
		t.Fatalf("合法定义不应报错: %v", err)
	}
	// http 步骤缺 method 默认 GET
	if p.Stages[1].Steps[0].HTTPMethod != "GET" {
		t.Fatalf("http 步骤应默认 GET: %s", p.Stages[1].Steps[0].HTTPMethod)
	}
}

func TestValidatePipeline_Errors(t *testing.T) {
	cases := []struct {
		name string
		p    *model.Pipeline
		code int
	}{
		{"缺名称", &model.Pipeline{}, ErrCodePlNameRequired},
		{"阶段缺名称", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: " ", Steps: []model.PipelineStep{mkShellStep("s", "ls")}}}}, ErrCodePlStructInvalid},
		{"阶段无步骤", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: "a"}}}, ErrCodePlStructInvalid},
		{"步骤缺名称", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: "a", Steps: []model.PipelineStep{{Type: model.StepTypeShell, ShellContent: "ls"}}}}}, ErrCodePlStructInvalid},
		{"步骤类型非法", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: "a", Steps: []model.PipelineStep{{Name: "s", Type: "grpc"}}}}}, ErrCodePlStructInvalid},
		{"shell 缺内容", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: "a", Steps: []model.PipelineStep{{Name: "s", Type: model.StepTypeShell}}}}}, ErrCodePlStructInvalid},
		{"http 地址非法", &model.Pipeline{Name: "x", Stages: []model.PipelineStage{{Name: "a", Steps: []model.PipelineStep{{Name: "s", Type: model.StepTypeHTTP, HTTPURL: "ftp://x"}}}}}, ErrCodePlStructInvalid},
	}
	for _, c := range cases {
		err := validatePipeline(c.p)
		pe, ok := err.(*pipelineError)
		if !ok || pe.Code != c.code {
			t.Fatalf("%s: 应报 %d，实际 %v", c.name, c.code, err)
		}
	}
}
