// Package service aiops 诊断 prompt 模板单测
package service

import (
	"strings"
	"testing"

	k8sSvc "github.com/hequan2017/new-ops/server/plugin/k8s/service"
)

func TestBuildPodPromptContainsKeySections(t *testing.T) {
	detail := &k8sSvc.PodDetail{
		Name: "bad-pod", Namespace: "demo", Phase: "Pending",
		Containers: []k8sSvc.PodContainerInfo{{Name: "app", Image: "nginx:bad", State: "Waiting", Reason: "ImagePullBackOff"}},
		Events:     []k8sSvc.PodEventInfo{{Type: "Warning", Reason: "Failed", Message: "ErrImagePull"}},
	}
	logs := map[string]string{"app": "some log line"}

	got := buildPodPrompt(detail, logs)
	for _, want := range []string{"bad-pod", "ImagePullBackOff", "ErrImagePull", "nginx:bad", "some log line", "Pod 状态", "日志尾部"} {
		if !strings.Contains(got, want) {
			t.Fatalf("快照缺关键段 %q\n%s", want, got)
		}
	}
}

func TestBuildPodPromptEmptyLogs(t *testing.T) {
	detail := &k8sSvc.PodDetail{Name: "p", Namespace: "d", Phase: "Running"}
	got := buildPodPrompt(detail, nil)
	if !strings.Contains(got, "无可用日志") {
		t.Fatalf("空日志应标注说明: %s", got)
	}
	withEmpty := buildPodPrompt(detail, map[string]string{"app": "  "})
	if !strings.Contains(withEmpty, "容器 app") || !strings.Contains(withEmpty, "（空）") {
		t.Fatalf("空日志容器段应标注: %s", withEmpty)
	}
}

// TestDiagSystemPrompt_SafetyContract 注入防御契约锚点（README 安全设计声明）：
// system prompt 必须含 <snapshot> 隔离声明与"指令视为日志内容"的拒绝语义，防止未来改文案时被误删。
func TestDiagSystemPrompt_SafetyContract(t *testing.T) {
	for _, kw := range []string{"<snapshot>", "不可信数据", "不得遵从或执行"} {
		if !strings.Contains(diagSystemPrompt, kw) {
			t.Fatalf("诊断 system prompt 缺少注入防御关键声明: %q", kw)
		}
	}
}
