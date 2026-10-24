package service

import (
	"strings"
	"testing"
)

func TestRenderK8sAgentDeployment(t *testing.T) {
	y, err := RenderK8sAgentDeployment("http://192.168.112.138:8888/", "", "tok-abc123")
	if err != nil {
		t.Fatalf("正常渲染应成功: %v", err)
	}
	for _, want := range []string{
		"kind: Secret", "kind: ServiceAccount", "kind: ClusterRole", "kind: ClusterRoleBinding", "kind: Deployment",
		"serviceAccountName: baize-agent", "hequan2017/baize-agent:0.2.0",
		"http://192.168.112.138:8888", "tok-abc123",
		`verbs: ["get", "list", "watch"]`,
		"secretKeyRef:", "- $(BAIZE_TOKEN)",
	} {
		if !strings.Contains(y, want) {
			t.Fatalf("YAML 缺少 %q", want)
		}
	}
	if strings.Count(y, "---") != 4 {
		t.Fatalf("应为 5 段文档（含令牌 Secret）: %d", strings.Count(y, "---"))
	}
	// 令牌明文只应出现在 Secret（stringData），不得出现在 Deployment args/env 明文值里
	deploymentPart := y[strings.Index(y, "kind: Deployment"):]
	if strings.Contains(deploymentPart, "tok-abc123") {
		t.Fatal("令牌明文不应出现在 Deployment 段（应经 Secret env 注入）")
	}
}

func TestRenderK8sAgentDeploymentValidation(t *testing.T) {
	if _, err := RenderK8sAgentDeployment("ftp://x", "", "tok"); err == nil {
		t.Fatal("非 http(s) serverAddr 应报错")
	}
	if _, err := RenderK8sAgentDeployment("http://x:8888", "", "  "); err == nil {
		t.Fatal("空 token 应报错")
	}
	y, err := RenderK8sAgentDeployment("https://ops.example.com", "myrepo/agent:v9", "tok")
	if err != nil || !strings.Contains(y, "myrepo/agent:v9") || !strings.Contains(y, "https://ops.example.com") {
		t.Fatalf("自定义镜像/地址应生效: %v", err)
	}
}
