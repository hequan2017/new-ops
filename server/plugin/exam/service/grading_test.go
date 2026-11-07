// Package service exam 判分引擎单测（客观三型/多选漏选比例/LLM 解析/注入式整卷）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func TestGradeObjective_SingleJudge(t *testing.T) {
	if s, _ := GradeObjective("single", "监督微调", " 监督微调 ", 2, true); s != 2 {
		t.Fatalf("single 命中应满分: %v", s)
	}
	if s, _ := GradeObjective("single", "监督微调", "强化学习", 2, true); s != 0 {
		t.Fatalf("single 不符应零分: %v", s)
	}
	// judge 同义词归一
	if s, _ := GradeObjective("judge", "对", "true", 1, false); s != 1 {
		t.Fatalf("judge 同义应命中: %v", s)
	}
	if s, _ := GradeObjective("judge", "对", "错误", 1, false); s != 0 {
		t.Fatalf("judge 相反应零分: %v", s)
	}
}

func TestGradeObjective_Multiple(t *testing.T) {
	// 全对
	if s, _ := GradeObjective("multiple", "ABD", "A,B,D", 3, false); s != 3 {
		t.Fatalf("全对应满分: %v", s)
	}
	// 漏选：比例 2/3
	if s, d := GradeObjective("multiple", "ABD", "AB", 3, true); s != 2 || !strings.Contains(d, "漏选") {
		t.Fatalf("漏选应按比例: %v %q", s, d)
	}
	// 漏选：不给比例 → 零分
	if s, _ := GradeObjective("multiple", "ABD", "AB", 3, false); s != 0 {
		t.Fatalf("漏选不得分口径: %v", s)
	}
	// 错选 → 零分（即使开着比例）
	if s, d := GradeObjective("multiple", "ABD", "ABC", 3, true); s != 0 || !strings.Contains(d, "错选") {
		t.Fatalf("错选应零分: %v %q", s, d)
	}
	// 未作答
	if s, _ := GradeObjective("multiple", "ABD", "", 3, true); s != 0 {
		t.Fatalf("未作答应零分: %v", s)
	}
}

func TestParseGradeScore(t *testing.T) {
	score, reason, err := ParseGradeScore(`{"score":0.85,"reason":"要点齐全"}`)
	if err != nil || score != 0.85 || reason != "要点齐全" {
		t.Fatalf("解析异常: %v %q %v", score, reason, err)
	}
	// 围栏 + 夹取
	score, _, _ = ParseGradeScore("```json\n{\"score\":1.5}\n```")
	if score != 1 {
		t.Fatalf("越界应夹取: %v", score)
	}
	if _, _, err := ParseGradeScore("no json"); err == nil {
		t.Fatal("非 JSON 应拒绝")
	}
	if _, _, err := ParseGradeScore(`{"reason":"缺分"}`); err == nil {
		t.Fatal("缺 score 应拒绝")
	}
}

// TestGradeQuestions 注入式整卷：客观+主观混合
func TestGradeQuestions(t *testing.T) {
	mk := func(id uint, qType, answer string, score float64) model.ExamQuestion {
		q := model.ExamQuestion{Type: qType, Answer: answer, Score: score, Stem: "题干" + answer}
		q.ID = id
		return q
	}
	qs := []model.ExamQuestion{
		mk(1, "single", "A", 2),
		mk(2, "multiple", "AB", 3),
		mk(3, "short", "要点一要点二", 10),
	}
	answers := map[string]string{"1": "A", "2": "A", "3": "答了要点一"}
	judge := func(_, expected, answer string) (float64, string, error) {
		if strings.Contains(answer, "要点一") {
			return 0.5, "命中一半", nil
		}
		return 0, "未命中", nil
	}
	grades, total := gradeQuestions(qs, answers, true, judge)
	if grades[0].Score != 2 || grades[1].Score != 1.5 || grades[2].Score != 5 {
		t.Fatalf("逐题得分异常: %+v", grades)
	}
	if grades[2].Method != "llm" || grades[2].Detail != "命中一半" {
		t.Fatalf("主观判分元数据异常: %+v", grades[2])
	}
	if total != 8.5 {
		t.Fatalf("总分异常: %v", total)
	}
	// judge=nil：主观零分不 panic
	grades2, _ := gradeQuestions(qs, answers, true, nil)
	if grades2[2].Score != 0 || grades2[2].Detail != "未配置评判模型" {
		t.Fatalf("无评判模型口径异常: %+v", grades2[2])
	}
}
