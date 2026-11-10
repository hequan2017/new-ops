// Package service eduadmin：绩效考评单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestPerfTotal 教师/行政权重与学时折算
func TestPerfTotal(t *testing.T) {
	teacher := model.EduPerfRecord{StaffType: "教师", TeachingHours: 160, TeachingScore: 90, ResearchScore: 80, ServiceScore: 70}
	// 0.5*90*1 + 0.3*80 + 0.2*70 = 45+24+14 = 83
	if v, err := perfTotal(teacher); err != nil || v != 83 {
		t.Fatalf("教师满分学时应 83: %v %v", v, err)
	}
	teacher.TeachingHours = 80 // 折算 0.5 → 0.5*90*0.5+24+14 = 60.5
	if v, _ := perfTotal(teacher); v != 60.5 {
		t.Fatalf("半学时应 60.5: %v", v)
	}
	admin := model.EduPerfRecord{StaffType: "行政", TeachingHours: 0, TeachingScore: 60, ResearchScore: 60, ServiceScore: 90}
	// 0.3*60*0 + 0.1*60 + 0.6*90 = 0+6+54 = 60
	if v, err := perfTotal(admin); err != nil || v != 60 {
		t.Fatalf("行政零学时应 60: %v %v", v, err)
	}
	if _, err := perfTotal(model.EduPerfRecord{StaffType: "外聘"}); err == nil {
		t.Fatal("白名单外类型应报错")
	}
}

// TestPerfGrade 四档等级
func TestPerfGrade(t *testing.T) {
	cases := map[float64]string{95: "优秀", 80: "良好", 60: "合格", 59.9: "不合格"}
	for score, want := range cases {
		if got := perfGrade(score); got != want {
			t.Fatalf("%v 应 %s，得 %s", score, want, got)
		}
	}
}

// TestRenderPerfReport 模板填充
func TestRenderPerfReport(t *testing.T) {
	rec := model.EduPerfRecord{StaffName: "王教授", StaffType: "教师", Year: 2025, TeachingHours: 200, TeachingScore: 92, ResearchScore: 88, ServiceScore: 76}
	total, _ := perfTotal(rec) // 0.5*92+0.3*88+0.2*76 = 46+26.4+15.2 = 87.6 → 良好
	report := renderPerfReport(rec, total, perfGrade(total))
	for _, want := range []string{"王教授 2025 年度绩效报告（教师）", "全年 200 学时", "综合得分 87.6", "等级 良好"} {
		if !strings.Contains(report, want) {
			t.Fatalf("报告缺 %q: %s", want, report)
		}
	}
	low := model.EduPerfRecord{StaffName: "李", StaffType: "教师", Year: 2025, TeachingHours: 40, TeachingScore: 40, ResearchScore: 40, ServiceScore: 40}
	lowTotal, _ := perfTotal(low) // 0.5*40*0.25 + 12+8 = 5+20=25
	text := renderPerfReport(low, lowTotal, perfGrade(lowTotal))
	if !strings.Contains(text, "改进计划") {
		t.Fatalf("不合格报告应含改进建议: %s", text)
	}
}
