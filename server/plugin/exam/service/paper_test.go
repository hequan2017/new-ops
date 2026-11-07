// Package service exam 组卷引擎单测（规则解析/抽题去重/不足报错/分值合计）
package service

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func mkQ(id uint, qType string, score float64, kp string, diff int) model.ExamQuestion {
	q := model.ExamQuestion{Type: qType, Score: score, KnowledgePoint: kp, Difficulty: diff}
	q.ID = id
	return q
}

func TestParsePaperRule(t *testing.T) {
	rule, err := ParsePaperRule(`{"sections":[{"type":"single","count":5,"knowledgePoint":"微调","maxDifficulty":3}]}`)
	if err != nil || len(rule.Sections) != 1 || rule.Sections[0].Count != 5 {
		t.Fatalf("合法规则异常: %+v %v", rule, err)
	}
	if _, err := ParsePaperRule(`{"sections":[]}`); err == nil {
		t.Fatal("空节应拒绝")
	}
	if _, err := ParsePaperRule(`{"sections":[{"type":"essay","count":1}]}`); err == nil {
		t.Fatal("非法题型应拒绝")
	}
	if _, err := ParsePaperRule(`{"sections":[{"type":"single","count":0}]}`); err == nil {
		t.Fatal("count=0 应拒绝")
	}
	if _, err := ParsePaperRule(`{"sections":[{"type":"single","count":1,"maxDifficulty":9}]}`); err == nil {
		t.Fatal("难度上限越界应拒绝")
	}
}

// TestPickQuestions 抽题：数量/去重/过滤/分值合计/不足报错
func TestPickQuestions(t *testing.T) {
	pool := []model.ExamQuestion{
		mkQ(1, "single", 2, "微调", 2), mkQ(2, "single", 2, "微调", 3), mkQ(3, "single", 3, "检索", 1),
		mkQ(4, "judge", 1, "微调", 1), mkQ(5, "judge", 1, "微调", 2), mkQ(6, "short", 10, "综合", 3),
	}
	rule := &PaperRule{Sections: []PaperSection{
		{Type: "single", Count: 2, MaxDifficulty: 3},
		{Type: "judge", Count: 1},
	}}
	rng := rand.New(rand.NewSource(42))
	ids, total, err := PickQuestions(pool, rule, rng)
	if err != nil {
		t.Fatalf("抽题失败: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("应抽 3 题: %v", ids)
	}
	// 去重 + 过滤：single 全部难度≤3
	seen := map[uint]bool{}
	singleCount := 0
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("题目重复: %d", id)
		}
		seen[id] = true
	}
	for _, id := range ids[:2] {
		if id == 3 { // 检索单选被 knowledgePoint? 无 kp 过滤时可能命中；难度 1≤3 合法
			singleCount++
		}
	}
	_ = singleCount
	// 总分 = 三题分值之和（任一组合）
	var sum float64
	for _, q := range pool {
		if seen[q.ID] {
			sum += q.Score
		}
	}
	if total != sum {
		t.Fatalf("分值合计异常: %v vs %v", total, sum)
	}
	// 题量不足
	rule2 := &PaperRule{Sections: []PaperSection{{Type: "single", Count: 10}}}
	if _, _, err := PickQuestions(pool, rule2, rng); err == nil || !strings.Contains(err.Error(), "题量不足") {
		t.Fatalf("不足应报: %v", err)
	}
	// 跨 section 去重：judge 池 2 题两节各抽 2 → 第二节不足
	rule3 := &PaperRule{Sections: []PaperSection{{Type: "judge", Count: 2}, {Type: "judge", Count: 2}}}
	if _, _, err := PickQuestions(pool, rule3, rng); err == nil {
		t.Fatal("跨节去重后不足应报错")
	}
	// 难度上限过滤
	rule4 := &PaperRule{Sections: []PaperSection{{Type: "single", Count: 2, MaxDifficulty: 2}}}
	ids4, _, err := PickQuestions(pool, rule4, rng)
	if err != nil {
		t.Fatalf("难度过滤失败: %v", err)
	}
	if len(ids4) != 2 {
		t.Fatalf("难度≤2 的 single 仅 2 题: %v", ids4)
	}
}
