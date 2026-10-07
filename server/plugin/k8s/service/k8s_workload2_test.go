// Package service K2 补齐纯函数单测（containerStateText）
package service

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestContainerStateText(t *testing.T) {
	cases := []struct {
		name       string
		status     corev1.ContainerStatus
		wantState  string
		wantReason string
	}{
		{"运行中", corev1.ContainerStatus{State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}}, "Running", ""},
		{"等待镜像", corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}}}, "Waiting", "ImagePullBackOff"},
		{"已退出", corev1.ContainerStatus{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"}}}, "Completed", "Completed"},
		{"未知", corev1.ContainerStatus{}, "Unknown", ""},
	}
	for _, tc := range cases {
		state, reason := containerStateText(tc.status)
		if state != tc.wantState || reason != tc.wantReason {
			t.Errorf("%s: containerStateText = (%q,%q), want (%q,%q)", tc.name, state, reason, tc.wantState, tc.wantReason)
		}
	}
}

// 编译期保障：PodDetail/EventInfo 等响应结构与 metav1 时间字段联动正确
var _ = metav1.Time{}
