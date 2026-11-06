// Package service llmops 模型部署单测（端口/环境规格、容器名归一、健康探测）
package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildPortSpecs(t *testing.T) {
	specs, err := buildPortSpecs("8000:8000, 11434:11434/tcp")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(specs) != 2 || specs[0] != "8000:8000/tcp" || specs[1] != "11434:11434/tcp" {
		t.Fatalf("缺省 proto 应补 /tcp: %v", specs)
	}
	if specs, err := buildPortSpecs(""); err != nil || len(specs) != 0 {
		t.Fatalf("空串应为空列表: %v %v", specs, err)
	}
	if _, err := buildPortSpecs("8000"); err == nil || !strings.Contains(err.Error(), "2236") {
		t.Fatalf("缺容器端口应报 2236: %v", err)
	}
	if _, err := buildPortSpecs("a:b"); err == nil {
		t.Fatal("非数字端口应报错")
	}
}

func TestBuildEnvSpecs(t *testing.T) {
	envs := buildEnvSpecs("K1=V1, BAD, =EMPTY, K2=a=b,  K3=V3 ")
	if len(envs) != 3 || envs[0] != "K1=V1" || envs[1] != "K2=a=b" || envs[2] != "K3=V3" {
		t.Fatalf("环境变量规格异常: %v", envs)
	}
}

func TestDeployContainerName(t *testing.T) {
	cases := map[string]string{
		"preset-vllm-openai": "llmops-preset-vllm-openai",
		"中文 部署名!":            "llmops-model", // 全部字符被剔除后 slug 为空 → model 兜底
		"  ABC_def 123 ":     "llmops-abc_def-123",
		"":                   "llmops-model",
	}
	for in, want := range cases {
		if got := deployContainerName(in); got != want {
			t.Fatalf("deployContainerName(%q)=%q want %q", in, got, want)
		}
	}
	long := strings.Repeat("a", 100)
	if got := deployContainerName(long); len(got) != 40 {
		t.Fatalf("超长应截断 40: %d", len(got))
	}
}

// TestProbeModelService 探测契约：任意 HTTP 响应视为存活；拒连报错
func TestProbeModelService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound) // 404 也算服务存活
	}))
	if err := probeModelService(srv.URL); err != nil {
		t.Fatalf("404 应视为存活: %v", err)
	}
	srv.Close()
	if err := probeModelService(srv.URL); err == nil {
		t.Fatal("拒连应报错")
	}
	if err := probeModelService(""); err == nil {
		t.Fatal("空地址应报错")
	}
}
