// Package service eduadmin：智能招聘单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestMatchResume 匹配打分三档
func TestMatchResume(t *testing.T) {
	position := model.EduPosition{Title: "讲师", RequiredSkill: "机器学习,深度学习,论文写作", MinYears: 3}

	strong := model.EduResume{Name: "甲", Years: 5, Skills: "机器学习,深度学习,论文写作,Python"}
	m := matchResume(strong, position)
	if m.Level != "高" || m.Overlap != 1 {
		t.Fatalf("全技能+超年限应高: %+v", m)
	}
	if m.Score != 1 {
		t.Fatalf("满分应为 1: %v", m.Score)
	}

	mid := model.EduResume{Name: "乙", Years: 3, Skills: "机器学习"}
	m = matchResume(mid, position)
	if m.Level != "中" || m.Overlap != 0.33 {
		t.Fatalf("1/3 技能+达标年限应中档(0.7*0.33+0.3=0.53): %+v", m)
	}

	// 2/3 技能 + 达标年限 → (0.7*0.67)+(0.3*1)=0.77 高
	better := model.EduResume{Name: "丙", Years: 3, Skills: "机器学习,深度学习"}
	m = matchResume(better, position)
	if m.Level != "高" {
		t.Fatalf("2/3 技能应高: %+v", m)
	}
	// 1/3 技能 + 无年限 → 0.7*0.33=0.23 低
	weak := model.EduResume{Name: "丁", Years: 0, Skills: "机器学习"}
	m = matchResume(weak, position)
	if m.Level != "低" || m.Score >= 0.45 {
		t.Fatalf("弱候选应低: %+v", m)
	}
	// 空岗位要求 → 重合 0，年限退化 min=1
	emptyPos := model.EduPosition{Title: "行政", RequiredSkill: "", MinYears: 0}
	m = matchResume(strong, emptyPos)
	if m.Overlap != 0 || m.Score != 0.3 {
		t.Fatalf("空要求应仅经验项 0.3: %+v", m)
	}
}

// TestSplitSkills 技能集合解析
func TestSplitSkills(t *testing.T) {
	set := splitSkills("机器学习, 深度学习,,Python")
	if len(set) != 3 {
		t.Fatalf("应解析 3 个去重技能: %+v", set)
	}
	if !set["机器学习"] || !set["python"] {
		t.Fatalf("应小写归一: %+v", set)
	}
}
