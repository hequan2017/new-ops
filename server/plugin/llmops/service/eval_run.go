// Package service llmops：模型评估（要求(4)）——样本调用生成 + 指标判分 + 结果留档
// 口径：指标白名单 exact/contains/keyword；chat 调用以 chatFn 注入（单测用假模型）；
// 单轮同步批量上限 50 样本；明细全量落库。
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
	ErrCodeEvalNotFound = 2259
	ErrCodeEvalMetric   = 2260
	ErrCodeEvalSample   = 2261
	evalMaxSamples      = 50
)

// EvalMetric 指标定义
type EvalMetric struct {
	Type      string   `json:"type"` // exact/contains/keyword
	Threshold float64  `json:"threshold"`
	Keywords  []string `json:"keywords"`
}

// EvalSample 评测样本
type EvalSample struct {
	Question string `json:"question"`
	Expected string `json:"expected"`
}

// ParseEvalMetric 指标定义解析（白名单+阈值夹取+keyword 必须带词表）
func ParseEvalMetric(raw string) (*EvalMetric, error) {
	m := &EvalMetric{Type: "contains", Threshold: 1}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), m); err != nil {
			return nil, fmt.Errorf("[%d]指标定义解析失败: %w", ErrCodeEvalMetric, err)
		}
	}
	switch m.Type {
	case "exact", "contains":
	case "keyword":
		if len(m.Keywords) == 0 {
			return nil, fmt.Errorf("[%d]keyword 指标必须提供词表", ErrCodeEvalMetric)
		}
	default:
		return nil, fmt.Errorf("[%d]不支持的指标类型: %s", ErrCodeEvalMetric, m.Type)
	}
	if m.Threshold <= 0 || m.Threshold > 1 {
		m.Threshold = 0.6
	}
	return m, nil
}

// ParseEvalSamples 样本解析（非空校验+上限）
func ParseEvalSamples(raw string) ([]EvalSample, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("[%d]评测样本不能为空", ErrCodeEvalSample)
	}
	var samples []EvalSample
	if err := json.Unmarshal([]byte(raw), &samples); err != nil {
		return nil, fmt.Errorf("[%d]样本解析失败: %w", ErrCodeEvalSample, err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("[%d]评测样本不能为空", ErrCodeEvalSample)
	}
	if len(samples) > evalMaxSamples {
		return nil, fmt.Errorf("[%d]样本数超上限 %d", ErrCodeEvalSample, evalMaxSamples)
	}
	for i, s := range samples {
		if strings.TrimSpace(s.Question) == "" {
			return nil, fmt.Errorf("[%d]第 %d 条样本问题为空", ErrCodeEvalSample, i+1)
		}
	}
	return samples, nil
}

// ScoreAnswer 判分（纯函数，测试锚点）：exact 归一全等/contains 包含/keyword 命中率
func ScoreAnswer(m *EvalMetric, expected, actual string) float64 {
	a := strings.ToLower(strings.TrimSpace(actual))
	switch m.Type {
	case "exact":
		if a == strings.ToLower(strings.TrimSpace(expected)) {
			return 1
		}
		return 0
	case "contains":
		if strings.Contains(a, strings.ToLower(strings.TrimSpace(expected))) {
			return 1
		}
		return 0
	default: // keyword
		if len(m.Keywords) == 0 {
			return 0
		}
		hit := 0
		for _, k := range m.Keywords {
			if strings.Contains(a, strings.ToLower(strings.TrimSpace(k))) {
				hit++
			}
		}
		return float64(hit) / float64(len(m.Keywords))
	}
}

// chatFunc 单轮问答注入点
type chatFunc func(question string) (string, error)

// evalSamples 批量评测（注入 chatFn；返回明细与通过数）
func evalSamples(fn chatFunc, samples []EvalSample, m *EvalMetric) ([]model.LlmopsEvalResult, int, error) {
	results := make([]model.LlmopsEvalResult, 0, len(samples))
	pass := 0
	for i, s := range samples {
		actual, err := fn(s.Question)
		if err != nil {
			return nil, 0, fmt.Errorf("第 %d 条样本调用失败: %w", i+1, err)
		}
		score := ScoreAnswer(m, s.Expected, actual)
		p := score >= m.Threshold
		if p {
			pass++
		}
		results = append(results, model.LlmopsEvalResult{
			Idx: i + 1, Question: s.Question, Expected: s.Expected,
			Actual: actual, Score: score, Pass: p,
		})
	}
	return results, pass, nil
}

// RunEval 执行评估（同步小批量；状态机 pending/failed→running→done/failed）
func (s *LlmopsServiceStruct) RunEval(runId uint) (*model.LlmopsEvalRun, error) {
	var run model.LlmopsEvalRun
	if err := global.GVA_DB.First(&run, runId).Error; err != nil {
		return nil, fmt.Errorf("[%d]评估任务不存在", ErrCodeEvalNotFound)
	}
	if run.Status == "running" {
		return nil, fmt.Errorf("[%d]评估进行中", ErrCodeEvalMetric)
	}
	metric, err := ParseEvalMetric(run.MetricJSON)
	if err != nil {
		return nil, err
	}
	samples, err := ParseEvalSamples(run.SamplesJSON)
	if err != nil {
		return nil, err
	}
	global.GVA_DB.Model(&run).Updates(map[string]any{"status": "running", "err_msg": ""})

	chat := func(q string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
		defer cancel()
		content, _, err := s.ChatCompletion(ctx, run.ModelId, []ChatMessage{{Role: "user", Content: q}}, nil)
		return content, err
	}
	results, pass, evalErr := evalSamples(chat, samples, metric)
	if evalErr != nil {
		global.GVA_DB.Model(&run).Updates(map[string]any{"status": "failed", "err_msg": truncateSample(evalErr.Error())})
		return nil, evalErr
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("run_id = ?", runId).Delete(&model.LlmopsEvalResult{}).Error; err != nil {
			return err
		}
		for i := range results {
			results[i].RunId = runId
		}
		if err := tx.CreateInBatches(results, 100).Error; err != nil {
			return err
		}
		return tx.Model(&run).Updates(map[string]any{
			"status": "done", "sample_count": len(samples), "pass_count": pass, "err_msg": "",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	run.Status, run.SampleCount, run.PassCount = "done", len(samples), pass
	return &run, nil
}

// GetEvalRuns 评估任务列表
func (s *LlmopsServiceStruct) GetEvalRuns(page, pageSize int) ([]model.LlmopsEvalRun, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	db := global.GVA_DB.Model(&model.LlmopsEvalRun{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsEvalRun
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveEvalRun 新建/更新评估任务（名称唯一；指标/样本即检）
func (s *LlmopsServiceStruct) SaveEvalRun(in *model.LlmopsEvalRun) (*model.LlmopsEvalRun, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || in.ModelId == 0 {
		return nil, fmt.Errorf("[%d]名称与被测模型必填", ErrCodeEvalNotFound)
	}
	if _, err := ParseEvalMetric(in.MetricJSON); err != nil {
		return nil, err
	}
	if _, err := ParseEvalSamples(in.SamplesJSON); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsEvalRun{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]评估名已存在", ErrCodeEvalNotFound)
	}
	if in.ID > 0 {
		var run model.LlmopsEvalRun
		if err := global.GVA_DB.First(&run, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]评估任务不存在", ErrCodeEvalNotFound)
		}
		err := global.GVA_DB.Model(&run).Updates(map[string]any{
			"name": in.Name, "model_id": in.ModelId, "dataset_id": in.DatasetId,
			"samples_json": in.SamplesJSON, "metric_json": in.MetricJSON, "status": "pending",
		}).Error
		if err != nil {
			return nil, err
		}
		return &run, nil
	}
	in.Status = "pending"
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEvalResults 评估明细
func (s *LlmopsServiceStruct) GetEvalResults(runId uint) ([]model.LlmopsEvalResult, error) {
	var list []model.LlmopsEvalResult
	err := global.GVA_DB.Where("run_id = ?", runId).Order("idx").Find(&list).Error
	return list, err
}
