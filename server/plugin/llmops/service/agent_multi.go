// Package service llmops：多智能体模式（要求(5) multi 构建模式）
// 口径：子智能体引用编排两种模式——sequential 串联流水线（输出接力）/ converge 并行汇聚
// （单子失败不拖垮，主模型综合）；def 存于智能体 configJSON；自引用在服务层防护。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeMultiDef  = 2274
	ErrCodeMultiSub  = 2275
	multiMaxSubs     = 10
)

// SubAgentRef 子智能体引用步骤
type SubAgentRef struct {
	AgentId uint   `json:"agentId"`
	Role    string `json:"role"`
	Prompt  string `json:"prompt"` // {input} 占位
}

// MultiDef 多智能体编排定义
type MultiDef struct {
	Pattern string        `json:"pattern"` // sequential/converge
	Subs    []SubAgentRef `json:"subs"`
}

// ParseMultiDef 解析编排定义（纯函数）
func ParseMultiDef(raw string) (*MultiDef, error) {
	var def MultiDef
	if err := json.Unmarshal([]byte(raw), &def); err != nil {
		return nil, fmt.Errorf("[%d]多智能体定义解析失败: %w", ErrCodeMultiDef, err)
	}
	if def.Pattern != "sequential" && def.Pattern != "converge" {
		return nil, fmt.Errorf("[%d]pattern 须为 sequential/converge", ErrCodeMultiDef)
	}
	if len(def.Subs) == 0 || len(def.Subs) > multiMaxSubs {
		return nil, fmt.Errorf("[%d]子智能体数量须为 1-%d", ErrCodeMultiDef, multiMaxSubs)
	}
	for i := range def.Subs {
		if def.Subs[i].AgentId == 0 {
			return nil, fmt.Errorf("[%d]第 %d 个子智能体 ID 为空", ErrCodeMultiDef, i+1)
		}
	}
	return &def, nil
}

// SubTrace 子智能体执行留痕
type SubTrace struct {
	AgentId uint   `json:"agentId"`
	Role    string `json:"role"`
	In      string `json:"in"`
	Out     string `json:"out,omitempty"`
	Err     string `json:"err,omitempty"`
}

// MultiRunResult 多智能体运行结果
type MultiRunResult struct {
	Pattern   string     `json:"pattern"`
	Output    string     `json:"output"`
	Subs      []SubTrace `json:"subs"`
}

// subRunner 子智能体执行注入点
type subRunner func(sub SubAgentRef, input string) (string, error)

// runMultiAgent 编排核心（全依赖注入，测试锚点）
func runMultiAgent(def *MultiDef, run subRunner, aggregateFn turnChatFunc, input string) (*MultiRunResult, error) {
	res := &MultiRunResult{Pattern: def.Pattern}
	switch def.Pattern {
	case "sequential":
		current := input
		for _, sub := range def.Subs {
			out, err := run(sub, current)
			st := SubTrace{AgentId: sub.AgentId, Role: sub.Role, In: current}
			if err != nil {
				st.Err = err.Error()
				res.Subs = append(res.Subs, st)
				return nil, fmt.Errorf("[%d]子智能体 %s 失败: %w", ErrCodeMultiSub, sub.Role, err)
			}
			st.Out = out
			res.Subs = append(res.Subs, st)
			current = out
		}
		res.Output = current
		return res, nil
	default: // converge：并行 + 汇聚
		traces := make([]SubTrace, len(def.Subs))
		outputs := make([]string, len(def.Subs))
		var wg sync.WaitGroup
		for i := range def.Subs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						traces[i].Err = fmt.Sprintf("panic: %v", r)
					}
				}()
				out, err := run(def.Subs[i], input)
				if err != nil {
					traces[i].Err = err.Error()
					return
				}
				traces[i].Out = out
				outputs[i] = out
			}(i)
		}
		wg.Wait()
		var parts []string
		for i := range def.Subs {
			traces[i].AgentId = def.Subs[i].AgentId
			traces[i].Role = def.Subs[i].Role
			traces[i].In = input
			if traces[i].Err == "" && outputs[i] != "" {
				parts = append(parts, fmt.Sprintf("【%s】%s", def.Subs[i].Role, outputs[i]))
			}
		}
		res.Subs = traces
		if len(parts) == 0 {
			return nil, fmt.Errorf("[%d]全部子智能体执行失败", ErrCodeMultiSub)
		}
		// 汇聚：主模型综合（无 aggregateFn 时简单拼接）
		if aggregateFn == nil {
			res.Output = strings.Join(parts, "\n\n")
			return res, nil
		}
		prompt := "综合以下多方观点给出最终结论：\n" + strings.Join(parts, "\n\n")
		final, err := aggregateFn(prompt, []ChatMessage{{Role: "user", Content: prompt}})
		if err != nil {
			return nil, err
		}
		res.Output = final
		return res, nil
	}
}

// AgentMultiRun 多智能体运行（服务入口：自引用防护 + 子智能体装载）
func (s *LlmopsServiceStruct) AgentMultiRun(agentId uint, input string) (*MultiRunResult, error) {
	var ag model.LlmopsAgent
	if err := global.GVA_DB.First(&ag, agentId).Error; err != nil {
		return nil, fmt.Errorf("[%d]智能体不存在", ErrCodeAgentNotFound)
	}
	if ag.Mode != "multi" {
		return nil, fmt.Errorf("[%d]仅 multi 模式支持多智能体运行（当前 %s）", ErrCodeMultiDef, ag.Mode)
	}
	def, err := ParseMultiDef(ag.ConfigJSON)
	if err != nil {
		return nil, err
	}
	for _, sub := range def.Subs {
		if sub.AgentId == agentId {
			return nil, fmt.Errorf("[%d]子智能体不可引用自身", ErrCodeMultiSub)
		}
	}
	// 装载子智能体（存在性校验 + 闭包构造）
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	run := func(sub SubAgentRef, input string) (string, error) {
		var subAg model.LlmopsAgent
		if err := global.GVA_DB.First(&subAg, sub.AgentId).Error; err != nil {
			return "", fmt.Errorf("子智能体 %d 不存在", sub.AgentId)
		}
		msgs := make([]ChatMessage, 0, 2)
		if subAg.SysPrompt != "" {
			msgs = append(msgs, ChatMessage{Role: "system", Content: subAg.SysPrompt})
		}
		msgs = append(msgs, ChatMessage{Role: "user", Content: renderTemplate(sub.Prompt, input)})
		content, _, err := s.ChatCompletion(ctx, subAg.ModelId, msgs, nil)
		return content, err
	}
	return runMultiAgent(def, run, nil, input)
}
