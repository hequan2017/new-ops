// Package service eduadmin：智能分派单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestMatchDispatchRule 关键词命中+优先级+零命中
func TestMatchDispatchRule(t *testing.T) {
	rules := []model.EduDispatchRule{
		{Keywords: "成绩,选课", Dept: "教务处", Priority: 10},
		{Keywords: "宿舍,空调", Dept: "后勤保障处", Priority: 10},
		{Keywords: "宿舍,选课,空调", Dept: "信息化办公室", Priority: 5},
	}
	// 双命中 + 高优先级 → 教务处
	c := matchDispatchRule(rules, "我的选课成绩有问题")
	if c == nil || c.Rule.Dept != "教务处" {
		t.Fatalf("应命中教务处: %+v", c)
	}
	if c.Hits != 2 {
		t.Fatalf("应命中 2 词: %d", c.Hits)
	}
	// 单命中 → 后勤
	c = matchDispatchRule(rules, "宿舍空调不制冷")
	if c == nil || c.Rule.Dept != "后勤保障处" {
		t.Fatalf("应命中后勤: %+v", c)
	}
	// 同命中数比优先级：3 号规则 priority 5 vs 2 号 10 → 后勤
	c = matchDispatchRule(rules, "空调坏了")
	if c == nil || c.Rule.Dept != "后勤保障处" {
		t.Fatalf("同分应取高优先级: %+v", c)
	}
	// 零命中
	if matchDispatchRule(rules, "我想咨询的天空") != nil {
		t.Fatal("零命中应返回 nil")
	}
}
