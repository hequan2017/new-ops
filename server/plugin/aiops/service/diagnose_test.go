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
