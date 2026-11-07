// Package service llmops：Prompt 在线优化（要求(7)）——单条/批量改写，保留原版对比
// 口径：chat 模型改写输出优化后全文（非 JSON，直接透传）；批量 ≤10 条逐条执行，
// 单条失败不拖垮批次（逐条留错）；原版随结果一并返回供对比。
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodePromptOpt      = 2297
	promptOptBatchMax     = 10
)

// buildOptimizePrompt 优化指令（纯函数，测试锚点）
func buildOptimizePrompt(content, requirement string) string {
	req := strings.TrimSpace(requirement)
	if req == "" {
		req = "通用优化：提升清晰度、结构化与可执行性，保留全部占位符变量"
	}
	return "你是 Prompt 优化专家。按以下需求改写 Prompt，只输出优化后的 Prompt 全文（保留 {{var}} 占位符原样）：\n" +
		"优化需求：" + req + "\n\n原 Prompt：\n" + content
}

// OptimizeItem 单条优化项（templateId 与 content 二选一）
type OptimizeItem struct {
	TemplateId  uint   `json:"templateId"`
	Content     string `json:"content"`
	Requirement string `json:"requirement"`
}

// OptimizeResult 单条优化结果（原版+优化版对比；失败带 Err）
type OptimizeResult struct {
	Original  string `json:"original"`
	Optimized string `json:"optimized"`
	Err       string `json:"err,omitempty"`
}

// resolveOptimizeContent 解析待优化内容（模板优先）
func (s *LlmopsServiceStruct) resolveOptimizeContent(item *OptimizeItem) (string, error) {
	if item.TemplateId > 0 {
		var tpl model.LlmopsPromptTemplate
		if err := global.GVA_DB.First(&tpl, item.TemplateId).Error; err != nil {
			return "", fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
		}
		return tpl.Content, nil
	}
	c := strings.TrimSpace(item.Content)
	if c == "" {
		return "", fmt.Errorf("[%d]templateId 与 content 至少提供一个", ErrCodePromptOpt)
	}
	return c, nil
}

// optimizeWith 批量优化核心（chatFn 注入，测试锚点）：单条失败留错不中断
func optimizeWith(fn chatFunc, items []OptimizeItem, resolve func(*OptimizeItem) (string, error)) []OptimizeResult {
	results := make([]OptimizeResult, 0, len(items))
	for i := range items {
		original, err := resolve(&items[i])
		res := OptimizeResult{}
		if err != nil {
			res.Err = err.Error()
			results = append(results, res)
			continue
		}
		res.Original = original
		out, callErr := fn(buildOptimizePrompt(original, items[i].Requirement))
		if callErr != nil {
			res.Err = truncateSample(callErr.Error())
		} else {
			res.Optimized = strings.TrimSpace(out)
			if res.Optimized == "" {
				res.Err = "模型返回为空"
			}
		}
		results = append(results, res)
	}
	return results
}

// OptimizePrompts 单条/批量优化入口（ModelId=0 时走 chat 能力调度）
func (s *LlmopsServiceStruct) OptimizePrompts(items []OptimizeItem, modelId uint) ([]OptimizeResult, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("[%d]优化项不能为空", ErrCodePromptOpt)
	}
	if len(items) > promptOptBatchMax {
		return nil, fmt.Errorf("[%d]批量上限 %d 条", ErrCodePromptOpt, promptOptBatchMax)
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	fn := func(prompt string) (string, error) {
		target := modelId
		if target == 0 {
			m, err := s.SelectModel("chat")
			if err != nil {
				return "", err
			}
			target = m.ID
		}
		content, _, err := s.ChatCompletion(ctx, target, []ChatMessage{{Role: "user", Content: prompt}}, nil)
		return content, err
	}
	return optimizeWith(fn, items, s.resolveOptimizeContent), nil
}
