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
		"kind: ServiceAccount", "kind: ClusterRole", "kind: ClusterRoleBinding", "kind: Deployment",
		"serviceAccountName: baize-agent", "hequan2017/baize-agent:0.2.0",
		"http://192.168.112.138:8888", "tok-abc123",
		`verbs: ["get", "list", "watch"]`,
	} {
		if !strings.Contains(y, want) {
			t.Fatalf("YAML 缺少 %q", want)
		}
	}
	if strings.Count(y, "---") != 3 {
		t.Fatalf("应为 4 段文档: %d", strings.Count(y, "---"))
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
