// Package service K1 Node 管理纯函数单测（drainDecisions/summarizeNodeStatus/cordonPatch）
package service

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func boolPtr(b bool) *bool { return &b }

func nodeWith(ready bool, unschedulable bool) *corev1.Node {
	condStatus := corev1.ConditionTrue
	if !ready {
		condStatus = corev1.ConditionFalse
	}
	return &corev1.Node{
		Spec: corev1.NodeSpec{Unschedulable: unschedulable},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: condStatus}},
		},
	}
}

func TestSummarizeNodeStatus(t *testing.T) {
	cases := []struct {
		name string
		node *corev1.Node
		want string
	}{
		{"就绪可调度", nodeWith(true, false), "Ready"},
		{"就绪已隔离", nodeWith(true, true), "Ready,SchedulingDisabled"},
		{"未就绪可调度", nodeWith(false, false), "NotReady"},
		{"未就绪已隔离", nodeWith(false, true), "NotReady,SchedulingDisabled"},
	}
	for _, tc := range cases {
		if got := summarizeNodeStatus(tc.node); got != tc.want {
			t.Errorf("%s: summarizeNodeStatus = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestCordonPatch(t *testing.T) {
	var body struct {
		Spec struct {
			Unschedulable bool `json:"unschedulable"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(cordonPatch(true), &body); err != nil {
		t.Fatalf("cordonPatch(true) 非法 JSON: %v", err)
	}
	if !body.Spec.Unschedulable {
		t.Error("cordonPatch(true) 应置 unschedulable=true")
	}
	if err := json.Unmarshal(cordonPatch(false), &body); err != nil {
		t.Fatalf("cordonPatch(false) 非法 JSON: %v", err)
	}
	if body.Spec.Unschedulable {
		t.Error("cordonPatch(false) 应置 unschedulable=false")
	}
}

func podWith(ns, name, controllerKind string, mirror bool) corev1.Pod {
	p := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}
	if mirror {
		p.Annotations = map[string]string{"kubernetes.io/config.mirror": "abc123"}
	}
	if controllerKind != "" {
		p.OwnerReferences = []metav1.OwnerReference{{Kind: controllerKind, Name: "owner", Controller: boolPtr(true)}}
	}
	return p
}

func TestDrainDecisions(t *testing.T) {
	pods := []corev1.Pod{
		podWith("kube-system", "kube-proxy-abc", "DaemonSet", false),
		podWith("kube-system", "etcd-localhost", "", true),
		podWith("default", "web-7d9", "ReplicaSet", false),
		podWith("default", "job-runner", "Job", false),
		podWith("default", "one-off", "", false),
	}
	evict, skip := drainDecisions(pods)

	if len(skip) != 2 {
		t.Fatalf("应跳过 2 个（DaemonSet+镜像 Pod），实际 %d：%+v", len(skip), skip)
	}
	for _, s := range skip {
		if s.Pod != "kube-proxy-abc" && s.Pod != "etcd-localhost" {
			t.Errorf("不应跳过 %s", s.Pod)
		}
	}
	if len(evict) != 3 {
		t.Fatalf("应驱逐 3 个（ReplicaSet/Job/独立），实际 %d：%+v", len(evict), evict)
	}
	for _, e := range evict {
		if e.Controller == "" {
			t.Errorf("%s 驱逐项 Controller 不应为空（独立 Pod 应标注）", e.Pod)
		}
	}
}

func TestDrainDecisions_Empty(t *testing.T) {
	evict, skip := drainDecisions(nil)
	if len(evict) != 0 || len(skip) != 0 {
		t.Errorf("空列表应返回空决策，实际 evict=%d skip=%d", len(evict), len(skip))
	}
}
