// Package service artsci：自动出题与组卷（需求 3.3-9）
// 口径：出题=模板规则生成（ModelId>0 时 chat 生成 JSON，失败自动回落模板）；
// 组卷=按规则（知识点×题型×题量×单题分值）顺序抽题，知识点覆盖由规则组合表达，
// 分值合计服务端计算；题量不足报 2514。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// 错误码（artsci 段 2500-2599）
const (
	ErrCodeQGenParam = 2513
	ErrCodePaperInsu = 2514
)

// questionChatFn 出题网关调用点（测试替换锚点）
var questionChatFn = gradingChatFn

// ValidQType 题型白名单
func ValidQType(t string) bool {
	switch t {
	case "单选", "多选", "判断", "简答":
		return true
	}
	return false
}

// templateQuestion 模板规则出题（纯函数，测试锚点）
func templateQuestion(kp, qType string) (stem, answer string, options []string) {
	switch qType {
	case "单选":
		return "下列关于「" + kp + "」的表述，哪一项最为准确？",
			"A", []string{"对「" + kp + "」的准确表述（正确项）", "与「" + kp + "」无关的干扰项", "与「" + kp + "」相悖的表述", "对「" + kp + "」的片面表述"}
	case "多选":
		return "下列关于「" + kp + "」的表述中，正确的有哪些？",
			"AB", []string{"「" + kp + "」的正确表述一", "「" + kp + "」的正确表述二", "与「" + kp + "」相悖的表述", "与「" + kp + "」无关的表述"}
	case "判断":
		return "「" + kp + "」是本领域的核心概念之一，教学中需要重点掌握。这一说法是否正确？", "对", nil
	default: // 简答
		return "请简述「" + kp + "」的核心要点，并举例说明其典型应用场景。",
			"参考要点：围绕「" + kp + "」的定义与原理、主要特征、典型应用展开，言之有理即可。", nil
	}
}

// genQuestionOut chat 生成解析结构
type genQuestionOut struct {
	Stem    string   `json:"stem"`
	Answer  string   `json:"answer"`
	Options []string `json:"options"`
}

// GenerateQuestion 生成单题并入库（llm 失败回落模板）
func (s *ArtsciService) GenerateQuestion(kp, qType, difficulty string, modelId uint, score float64) (*model.ArtsciQuestion, error) {
	kp = strings.TrimSpace(kp)
	if kp == "" || !ValidQType(qType) {
		return nil, fmt.Errorf("[%d]知识点必填且题型须在白名单内", ErrCodeQGenParam)
	}
	if difficulty != "易" && difficulty != "中" && difficulty != "难" {
		difficulty = "中"
	}
	q := &model.ArtsciQuestion{KnowledgePoint: kp, QType: qType, Difficulty: difficulty, Score: score}
	if modelId > 0 {
		prompt := "生成一道" + qType + "题，知识点：" + kp + "，难度：" + difficulty + "。只输出 JSON：\n" +
			"{\"stem\":\"题干\",\"answer\":\"参考答案\",\"options\":[\"选项A\",\"选项B\"]}（判断/简答题 options 给空数组）"
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		content, err := questionChatFn(ctx, modelId, prompt)
		cancel()
		if err == nil {
			var out genQuestionOut
			l, r := strings.Index(content, "{"), strings.LastIndex(content, "}")
			if l >= 0 && r > l && json.Unmarshal([]byte(content[l:r+1]), &out) == nil &&
				strings.TrimSpace(out.Stem) != "" && strings.TrimSpace(out.Answer) != "" {
				q.Stem, q.Answer, q.Source = out.Stem, out.Answer, "llm"
				if len(out.Options) > 0 {
					b, _ := json.Marshal(out.Options)
					q.Options = string(b)
				}
			}
		}
	}
	if q.Source != "llm" {
		stem, answer, options := templateQuestion(kp, qType)
		q.Stem, q.Answer, q.Source = stem, answer, "template"
		if len(options) > 0 {
			b, _ := json.Marshal(options)
			q.Options = string(b)
		}
	}
	if err := global.GVA_DB.Create(q).Error; err != nil {
		return nil, err
	}
	return q, nil
}

// paperRule 组卷规则
type paperRule struct {
	KnowledgePoint string  `json:"knowledgePoint"`
	QType          string  `json:"qType"`
	Count          int     `json:"count"`
	ScorePer       float64 `json:"scorePer"`
}

// validatePaperRules 规则校验与归一（纯函数，测试锚点）
func validatePaperRules(raw string) ([]paperRule, float64, error) {
	var rules []paperRule
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return nil, 0, fmt.Errorf("[%d]规则须为 JSON 数组", ErrCodeQGenParam)
	}
	if len(rules) == 0 {
		return nil, 0, fmt.Errorf("[%d]抽题规则不能为空", ErrCodeQGenParam)
	}
	total := 0.0
	for i := range rules {
		if rules[i].Count < 1 || rules[i].Count > 100 || rules[i].ScorePer <= 0 || !ValidQType(rules[i].QType) {
			return nil, 0, fmt.Errorf("[%d]第 %d 条规则参数非法（题量 1-100/单题分>0/题型白名单）", ErrCodeQGenParam, i+1)
		}
		total += float64(rules[i].Count) * rules[i].ScorePer
	}
	return rules, total, nil
}

// pickQuestions 候选抽题（顺序取前 count，纯函数，测试锚点）
func pickQuestions(candidates []model.ArtsciQuestion, count int) []model.ArtsciQuestion {
	if len(candidates) > count {
		return candidates[:count]
	}
	return candidates
}

// BuildPaper 规则组卷并入库
func (s *ArtsciService) BuildPaper(name, rulesJSON string) (*model.ArtsciPaper, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("[%d]试卷名必填", ErrCodeQGenParam)
	}
	rules, total, err := validatePaperRules(rulesJSON)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, r := range rules {
		db := global.GVA_DB.Model(&model.ArtsciQuestion{}).Where("q_type = ?", r.QType)
		if strings.TrimSpace(r.KnowledgePoint) != "" {
			db = db.Where("knowledge_point = ?", r.KnowledgePoint)
		}
		var candidates []model.ArtsciQuestion
		if err := db.Order("id").Limit(r.Count).Find(&candidates).Error; err != nil {
			return nil, err
		}
		picked := pickQuestions(candidates, r.Count)
		if len(picked) < r.Count {
			return nil, fmt.Errorf("[%d]题量不足：%s/%s 仅 %d 题，需 %d 题", ErrCodePaperInsu, r.KnowledgePoint, r.QType, len(picked), r.Count)
		}
		for _, q := range picked {
			ids = append(ids, strconv.FormatUint(uint64(q.ID), 10))
		}
	}
	paper := &model.ArtsciPaper{Name: name, QuestionIds: strings.Join(ids, ","), TotalScore: total, RulesJSON: rulesJSON}
	if err := global.GVA_DB.Create(paper).Error; err != nil {
		return nil, err
	}
	return paper, nil
}

// GetQuestions 题目列表（知识点/题型过滤，倒序分页）
func (s *ArtsciService) GetQuestions(kp, qType string, page, pageSize int) ([]model.ArtsciQuestion, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ArtsciQuestion{})
	if kp != "" {
		db = db.Where("knowledge_point = ?", kp)
	}
	if qType != "" {
		if !ValidQType(qType) {
			return nil, 0, fmt.Errorf("[%d]不支持的题型: %s", ErrCodeQGenParam, qType)
		}
		db = db.Where("q_type = ?", qType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ArtsciQuestion
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetPapers 试卷列表（倒序分页）
func (s *ArtsciService) GetPapers(page, pageSize int) ([]model.ArtsciPaper, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ArtsciPaper{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ArtsciPaper
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
