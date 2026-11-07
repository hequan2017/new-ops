// Package service exam 分布式调度单测（择优节点/心跳超时判定）
package service

import (
	"strings"
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func mkNode(key, region string, load int, status string) model.ExamNode {
	return model.ExamNode{NodeKey: key, Region: region, Load: load, Status: status}
}

func TestPickNode(t *testing.T) {
	nodes := []model.ExamNode{
		mkNode("n1", "北京", 10, "online"),
		mkNode("n2", "上海", 2, "online"),
		mkNode("n3", "北京", 5, "online"),
		mkNode("n4", "广州", 0, "offline"),
	}
	// 无偏好 → 负载最低（上海 n2）
	got, err := pickNode(nodes, "")
	if err != nil || got != "n2" {
		t.Fatalf("无偏好应取最低负载: %q %v", got, err)
	}
	// region 优先：北京 → n3(load 5) 而非 n1(load 10)
	got, _ = pickNode(nodes, "北京")
	if got != "n3" {
		t.Fatalf("region 优先异常: %q", got)
	}
	// 同负载 → 名称序
	tied := []model.ExamNode{mkNode("b", "x", 3, "online"), mkNode("a", "x", 3, "online")}
	got, _ = pickNode(tied, "")
	if got != "a" {
		t.Fatalf("同负载应取名称序: %q", got)
	}
	// 全离线
	if _, err := pickNode([]model.ExamNode{mkNode("n", "x", 0, "offline")}, ""); err == nil || !strings.Contains(err.Error(), "2417") {
		t.Fatalf("全离线应报 2417: %v", err)
	}
	// 空列表
	if _, err := pickNode(nil, ""); err == nil {
		t.Fatal("空候选应拒绝")
	}
}

func TestIsStale(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)
	if isStale(now.Add(-4*time.Minute), now, 5) {
		t.Fatal("4 分钟不应超时(阈值 5)")
	}
	if !isStale(now.Add(-6*time.Minute), now, 5) {
		t.Fatal("6 分钟应超时")
	}
	// 恰在阈值不算超时（> 严格）
	if isStale(now.Add(-5*time.Minute), now, 5) {
		t.Fatal("恰 5 分钟不应超时")
	}
}
