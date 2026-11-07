// Package service llmops：智能体评测（要求(5) 评测阶段）——评测集运行 + 规则/LLM 双轨打分
// 口径：被测回答经 qa 链单轮（sysPrompt+问题）；评分三型 contains/keyword(词表)/llm(评判模型
// 输出 {"score","reason"})；被测与评判双注入可测；明细留档。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeAgentEvalMiss = 2276
	ErrCodeAgentScorer   = 2277
	agentEvalMaxSamples  = 30
)

// ParseAgentKeywords 词表解析（逗号分隔去空白去重）
func ParseAgentKeywords(csvLine string) []string {
	seen := map[string]bool{}
	var out []string
	for _, k := range strings.Split(csvLine, ",") {
		k = strings.TrimSpace(k)
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

// buildJudgePrompt LLM 评分 prompt（纯函数）
func buildJudgePrompt(question, expected, reply string) string {
	return "你是评测裁判。判断「回答」是否正确满足「期望」。只输出 JSON：{\"score\":0到1的分数,\"reason\":\"一句话理由\"}\n" +
		"问题：" + question + "\n期望：" + expected + "\n回答：" + reply
}

// ParseJudgeOutput 评判输出解析（纯函数：均衡截取+分数夹取 0-1）
func ParseJudgeOutput(text string) (float64, string, error) {
	obj, ok := extractJSONObject(text)
	if !ok {
		return 0, "", fmt.Errorf("[%d]评判输出解析失败", ErrCodeAgentScorer)
	}
	var parsed struct {
		Score  *float64 `json:"score"`
		Reason string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil || parsed.Score == nil {
		return 0, "", fmt.Errorf("[%d]评判输出解析失败", ErrCodeAgentScorer)
	}
	score := *parsed.Score
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score, strings.TrimSpace(parsed.Reason), nil
}

// agentEvalDeps 评测依赖注入
type agentEvalDeps struct {
	replyFn func(question string) (string, error)                              // 被测智能体
	judgeFn func(question, expected, reply string) (float64, string, error)    // LLM 评判（llm 型）
}

// evalAgentSamples 评测核心（注入式，测试锚点）：阈值 0.6（llm 型以评判分为准）
func evalAgentSamples(deps *agentEvalDeps, samples []EvalSample, scorerType string, keywords []string) ([]model.LlmopsAgentEvalResult, int, error) {
	results := make([]model.LlmopsAgentEvalResult, 0, len(samples))
	pass := 0
	for i, s := range samples {
		reply, err := deps.replyFn(s.Question)
		if err != nil {
			return nil, 0, fmt.Errorf("第 %d 条样本被测失败: %w", i+1, err)
		}
		var score float64
		detail := ""
		switch scorerType {
		case "contains":
			score = ScoreAnswer(&EvalMetric{Type: "contains"}, s.Expected, reply)
			detail = "包含匹配"
		case "keyword":
			score = ScoreAnswer(&EvalMetric{Type: "keyword", Keywords: keywords}, s.Expected, reply)
			detail = fmt.Sprintf("词表命中 %d/%d", countKeywordHits(keywords, reply), len(keywords))
		default: // llm
			score, detail, err = deps.judgeFn(s.Question, s.Expected, reply)
			if err != nil {
				return nil, 0, fmt.Errorf("第 %d 条样本评判失败: %w", i+1, err)
			}
		}
		p := score >= 0.6
		if p {
			pass++
		}
		results = append(results, model.LlmopsAgentEvalResult{
			Idx: i + 1, Question: s.Question, Reply: reply, Score: score, Pass: p, Detail: detail,
		})
	}
	return results, pass, nil
}

func countKeywordHits(keywords []string, reply string) int {
	lower := strings.ToLower(reply)
	n := 0
	for _, k := range keywords {
		if strings.Contains(lower, strings.ToLower(k)) {
			n++
		}
	}
	return n
}

// RunAgentEval 执行智能体评测
func (s *LlmopsServiceStruct) RunAgentEval(evalId uint) (*model.LlmopsAgentEval, error) {
	var ev model.LlmopsAgentEval
	if err := global.GVA_DB.First(&ev, evalId).Error; err != nil {
		return nil, fmt.Errorf("[%d]评测任务不存在", ErrCodeAgentEvalMiss)
	}
	if ev.Status == "running" {
		return nil, fmt.Errorf("[%d]评测进行中", ErrCodeAgentScorer)
	}
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, ev.AgentId).Error; err != nil {
		return nil, fmt.Errorf("[%d]被测智能体不存在", ErrCodeAgentNotFound)
	}
	if ev.ScorerType == "llm" && ev.JudgeModelId == 0 {
		return nil, fmt.Errorf("[%d]llm 评分必须指定评判模型", ErrCodeAgentScorer)
	}
	samples, err := ParseEvalSamples(ev.SamplesJSON)
	if err != nil {
		return nil, err
	}
	if len(samples) > agentEvalMaxSamples {
		samples = samples[:agentEvalMaxSamples]
	}
	global.GVA_DB.Model(&ev).Updates(map[string]any{"status": "running", "err_msg": ""})

	keywords := ParseAgentKeywords(ev.Keywords)
	deps := &agentEvalDeps{
		replyFn: func(q string) (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
			defer cancel()
			reply, _, err := runAgentTurn(func(_ string, msgs []ChatMessage) (string, error) {
				content, _, err := s.ChatCompletion(ctx, ag.ModelId, msgs, nil)
				return content, err
			}, "qa", ag.SysPrompt, nil, q, ag.MaxTurns)
			return reply, err
		},
		judgeFn: func(question, expected, reply string) (float64, string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
			defer cancel()
			content, _, err := s.ChatCompletion(ctx, ev.JudgeModelId, []ChatMessage{{Role: "user", Content: buildJudgePrompt(question, expected, reply)}}, nil)
			if err != nil {
				return 0, "", err
			}
			return ParseJudgeOutput(content)
		},
	}
	results, pass, evalErr := evalAgentSamples(deps, samples, ev.ScorerType, keywords)
	if evalErr != nil {
		global.GVA_DB.Model(&ev).Updates(map[string]any{"status": "failed", "err_msg": truncateSample(evalErr.Error())})
		return nil, evalErr
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("eval_id = ?", evalId).Delete(&model.LlmopsAgentEvalResult{}).Error; err != nil {
			return err
		}
		for i := range results {
			results[i].EvalId = evalId
		}
		if err := tx.CreateInBatches(results, 100).Error; err != nil {
			return err
		}
		return tx.Model(&ev).Updates(map[string]any{
			"status": "done", "sample_count": len(samples), "pass_count": pass, "err_msg": "",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	ev.Status, ev.SampleCount, ev.PassCount = "done", len(samples), pass
	return &ev, nil
}

// GetAgentEvals 评测任务列表
func (s *LlmopsServiceStruct) GetAgentEvals() ([]model.LlmopsAgentEval, error) {
	var list []model.LlmopsAgentEval
	err := global.GVA_DB.Order("id desc").Find(&list).Error
	return list, err
}

// SaveAgentEval 保存评测任务（评分配置即检）
func (s *LlmopsServiceStruct) SaveAgentEval(in *model.LlmopsAgentEval) (*model.LlmopsAgentEval, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || in.AgentId == 0 {
		return nil, fmt.Errorf("[%d]名称与被测智能体必填", ErrCodeAgentEvalMiss)
	}
	switch in.ScorerType {
	case "contains", "keyword", "llm":
	default:
		return nil, fmt.Errorf("[%d]评分类型须为 contains/keyword/llm", ErrCodeAgentScorer)
	}
	if in.ScorerType == "llm" && in.JudgeModelId == 0 {
		return nil, fmt.Errorf("[%d]llm 评分必须指定评判模型", ErrCodeAgentScorer)
	}
	if in.ScorerType == "keyword" && len(ParseAgentKeywords(in.Keywords)) == 0 {
		return nil, fmt.Errorf("[%d]keyword 评分必须提供词表", ErrCodeAgentScorer)
	}
	if _, err := ParseEvalSamples(in.SamplesJSON); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsAgentEval{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]评测名已存在", ErrCodeAgentEvalMiss)
	}
	if in.ID > 0 {
		var ev model.LlmopsAgentEval
		if err := global.GVA_DB.First(&ev, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]评测任务不存在", ErrCodeAgentEvalMiss)
		}
		err := global.GVA_DB.Model(&ev).Updates(map[string]any{
			"name": in.Name, "agent_id": in.AgentId, "samples_json": in.SamplesJSON,
			"scorer_type": in.ScorerType, "judge_model_id": in.JudgeModelId,
			"keywords": in.Keywords, "status": "pending",
		}).Error
		if err != nil {
			return nil, err
		}
		return &ev, nil
	}
	in.Status = "pending"
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetAgentEvalResults 评测明细
func (s *LlmopsServiceStruct) GetAgentEvalResults(evalId uint) ([]model.LlmopsAgentEvalResult, error) {
	var list []model.LlmopsAgentEvalResult
	err := global.GVA_DB.Where("eval_id = ?", evalId).Order("idx").Find(&list).Error
	return list, err
}
