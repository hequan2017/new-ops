// Package service artsci：作业批改单测（假模型/纯函数，无 DB）
package service

import (
	"context"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// TestRuleGrade 规则口径：分数有界、要点覆盖计入、缺失建议
func TestRuleGrade(t *testing.T) {
	rubric := "论点明确，论据充分，结构完整"
	score, comment, sugg := ruleGrade(rubric, "本文论点明确，论据充分，只是结构一般。论点明确再次出现。")
	if score < 60 || score > 100 {
		t.Fatalf("分数应落在 [60,100]，实际 %v", score)
	}
	if !strings.Contains(comment, "2/3") {
		t.Fatalf("要点覆盖应为 2/3：%s", comment)
	}
	if len(sugg) == 0 || !strings.Contains(sugg[0], "结构完整") {
		t.Fatalf("缺失要点应进建议: %v", sugg)
	}
	// 长文封顶
	_, _, _ = ruleGrade(rubric, strings.Repeat("长", 2000))
	longScore, _, _ := ruleGrade(rubric, strings.Repeat("论点明确论据充分结构完整", 400))
	if longScore > 100 {
		t.Fatalf("分数应封顶 100，实际 %v", longScore)
	}
}

// TestParseGradingJSON 输出解析：干净 JSON/带前后缀/越界/非 JSON
func TestParseGradingJSON(t *testing.T) {
	s, c, sg, err := parseGradingJSON(`{"score":88,"comment":"好","suggestions":["A"]}`)
	if err != nil || s != 88 || c != "好" || len(sg) != 1 {
		t.Fatalf("干净 JSON 解析不符: %v %v %v %v", s, c, sg, err)
	}
	s, _, _, err = parseGradingJSON(`前置说明{"score":72,"comment":"可以","suggestions":[]}后缀`)
	if err != nil || s != 72 {
		t.Fatalf("带前后缀应可解析: %v %v", s, err)
	}
	if _, _, _, err = parseGradingJSON(`{"score":120,"comment":"x"}`); err == nil {
		t.Fatal("score 越界应报错")
	}
	if _, _, _, err = parseGradingJSON(`完全不是 JSON`); err == nil {
		t.Fatal("非 JSON 应报错")
	}
}

// TestGradeEngineLLM 假模型：成功路径 + 坏输出重试 1 次后成功 + 两次失败返回错误
func TestGradeEngineLLM(t *testing.T) {
	rec := &model.ArtsciGrading{Rubric: "要点a，要点b", StudentText: "覆盖了要点a", ModelId: 9}
	calls := 0
	fake := func(ctx context.Context, modelId uint, prompt string) (string, error) {
		calls++
		return `{"score":90,"comment":"优秀","suggestions":["保持"]}`, nil
	}
	if err := gradeEngine(rec, fake); err != nil {
		t.Fatalf("成功路径不应报错: %v", err)
	}
	if rec.Score != 90 || rec.Engine != "llm" || !rec.Ok || calls != 1 {
		t.Fatalf("成功路径字段不符: %+v calls=%d", rec, calls)
	}
	// 第一次坏输出、第二次好输出 → 重试后成功
	rec2 := &model.ArtsciGrading{Rubric: "要点a", StudentText: "内容", ModelId: 9}
	attempt := 0
	flaky := func(ctx context.Context, modelId uint, prompt string) (string, error) {
		attempt++
		if attempt == 1 {
			return "模型胡言乱语", nil
		}
		return `{"score":75,"comment":"尚可","suggestions":[]}`, nil
	}
	if err := gradeEngine(rec2, flaky); err != nil {
		t.Fatalf("重试后应成功: %v", err)
	}
	if rec2.Score != 75 || attempt != 2 {
		t.Fatalf("重试口径不符: score=%v attempt=%d", rec2.Score, attempt)
	}
	// 两次都坏 → 报 2511
	rec3 := &model.ArtsciGrading{Rubric: "要点a", StudentText: "内容", ModelId: 9}
	if err := gradeEngine(rec3, func(ctx context.Context, modelId uint, prompt string) (string, error) {
		return "仍不是 JSON", nil
	}); err == nil || !strings.Contains(err.Error(), "[2511]") {
		t.Fatalf("两次失败应报 2511: %v", err)
	}
}

// TestGradeEngineGuard 必填校验
func TestGradeEngineGuard(t *testing.T) {
	rec := &model.ArtsciGrading{Rubric: "标准", StudentText: "  "}
	if err := gradeEngine(rec, nil); err == nil || !strings.Contains(err.Error(), "[2510]") {
		t.Fatalf("空作答应报 2510: %v", err)
	}
}
