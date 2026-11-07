// Package service llmops：智能体工作流模式（要求(5) workflow 构建模式）
// 口径：节点 JSON 编排（llm/tool/condition/human）顺序执行；{input} 占位模板渲染；
// condition 支持 contains/not_contains/len>/len</empty/not_empty 六式，真假可跳转指定节点；
// human 节点挂起等待人工（状态 needs_human）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// 错误码（llmops 段续用）
const ErrCodeWf = 2273

// WfNode 工作流节点
type WfNode struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // llm/tool/condition/human
	Name    string `json:"name"`
	Prompt  string `json:"prompt"` // llm：{input} 占位
	Tool    string `json:"tool"`   // tool：工具名
	Args    string `json:"args"`   // tool：args 模板（{input} 占位）
	Expr    string `json:"expr"`   // condition：表达式
	OnTrue  string `json:"onTrue"` // condition：真→跳转节点 ID（空=顺序继续）
	OnFalse string `json:"onFalse"` // condition：假→跳转节点 ID（空=顺序继续）
}

// WfDef 工作流定义
type WfDef struct {
	Nodes []WfNode `json:"nodes"`
}

// ParseWfDef 解析工作流定义并校验（纯函数）
func ParseWfDef(raw string) (*WfDef, error) {
	var def WfDef
	if err := json.Unmarshal([]byte(raw), &def); err != nil {
		return nil, fmt.Errorf("[%d]工作流定义解析失败: %w", ErrCodeWf, err)
	}
	if len(def.Nodes) == 0 {
		return nil, fmt.Errorf("[%d]工作流至少需要一个节点", ErrCodeWf)
	}
	seen := map[string]bool{}
	for i := range def.Nodes {
		n := &def.Nodes[i]
		if n.ID == "" || seen[n.ID] {
			return nil, fmt.Errorf("[%d]节点 ID 空/重复: %q", ErrCodeWf, n.ID)
		}
		seen[n.ID] = true
		switch n.Type {
		case "llm":
			if strings.TrimSpace(n.Prompt) == "" {
				return nil, fmt.Errorf("[%d]llm 节点 %s 缺 prompt", ErrCodeWf, n.ID)
			}
		case "tool":
			if strings.TrimSpace(n.Tool) == "" {
				return nil, fmt.Errorf("[%d]tool 节点 %s 缺工具名", ErrCodeWf, n.ID)
			}
		case "condition":
			if strings.TrimSpace(n.Expr) == "" {
				return nil, fmt.Errorf("[%d]condition 节点 %s 缺表达式", ErrCodeWf, n.ID)
			}
		case "human":
		default:
			return nil, fmt.Errorf("[%d]不支持的节点类型: %s", ErrCodeWf, n.Type)
		}
	}
	return &def, nil
}

// renderTemplate 模板渲染（{input} 占位，纯函数）
func renderTemplate(tpl, input string) string {
	return strings.ReplaceAll(tpl, "{input}", input)
}

// evalCondition 条件表达式求值（纯函数，测试锚点）
func evalCondition(expr, input string) (bool, error) {
	expr = strings.TrimSpace(expr)
	switch {
	case expr == "empty":
		return strings.TrimSpace(input) == "", nil
	case expr == "not_empty":
		return strings.TrimSpace(input) != "", nil
	case strings.HasPrefix(expr, "contains:"):
		return strings.Contains(input, strings.TrimPrefix(expr, "contains:")), nil
	case strings.HasPrefix(expr, "not_contains:"):
		return !strings.Contains(input, strings.TrimPrefix(expr, "not_contains:")), nil
	case strings.HasPrefix(expr, "len>"):
		n, err := strconv.Atoi(strings.TrimPrefix(expr, "len>"))
		if err != nil {
			return false, fmt.Errorf("[%d]表达式非法: %s", ErrCodeWf, expr)
		}
		return len([]rune(input)) > n, nil
	case strings.HasPrefix(expr, "len<"):
		n, err := strconv.Atoi(strings.TrimPrefix(expr, "len<"))
		if err != nil {
			return false, fmt.Errorf("[%d]表达式非法: %s", ErrCodeWf, expr)
		}
		return len([]rune(input)) < n, nil
	default:
		return false, fmt.Errorf("[%d]不支持的表达式: %s", ErrCodeWf, expr)
	}
}

// WfStepTrace 节点执行留痕
type WfStepTrace struct {
	NodeId  string `json:"nodeId"`
	Type    string `json:"type"`
	In      string `json:"in"`
	Out     string `json:"out,omitempty"`
	Skipped bool   `json:"skipped"`
}

// WfRunResult 工作流运行结果
type WfRunResult struct {
	Status string         `json:"status"` // done/needs_human/failed
	Output string         `json:"output"`
	Steps  []WfStepTrace  `json:"steps"`
}

// RunWf 工作流顺序执行引擎（chatFn/tool 注入，测试锚点）
func RunWf(fn turnChatFunc, tools map[string]AgentTool, def *WfDef, input string) (*WfRunResult, error) {
	byId := make(map[string]int, len(def.Nodes))
	for i := range def.Nodes {
		byId[def.Nodes[i].ID] = i
	}
	res := &WfRunResult{Status: "done", Output: input}
	current := input
	idx := 0
	guard := 0
	for idx < len(def.Nodes) {
		guard++
		if guard > 4*len(def.Nodes)+8 { // 跳转环防护
			return nil, fmt.Errorf("[%d]工作流跳转成环", ErrCodeWf)
		}
		n := &def.Nodes[idx]
		step := WfStepTrace{NodeId: n.ID, Type: n.Type, In: current}
		switch n.Type {
		case "llm":
			msgs := []ChatMessage{{Role: "user", Content: renderTemplate(n.Prompt, current)}}
			out, err := fn(renderTemplate(n.Prompt, current), msgs)
			if err != nil {
				res.Status = "failed"
				res.Steps = append(res.Steps, step)
				return res, err
			}
			step.Out = out
			current = out
		case "tool":
			t, ok := tools[n.Tool]
			if !ok {
				return nil, fmt.Errorf("[%d]节点 %s 引用未注册工具 %s", ErrCodeWf, n.ID, n.Tool)
			}
			ctx, cancel := context.WithTimeout(context.Background(), embedHTTPClient.Timeout)
			out, execErr := t.Execute(ctx, renderTemplate(n.Args, current))
			cancel()
			if execErr != nil {
				return nil, fmt.Errorf("[%d]节点 %s 工具执行失败: %w", ErrCodeWf, n.ID, execErr)
			}
			step.Out = out
			current = out
		case "condition":
			b, err := evalCondition(n.Expr, current)
			if err != nil {
				return nil, err
			}
			step.Out = fmt.Sprintf("%v", b)
			next := n.OnFalse
			if b {
				next = n.OnTrue
			}
			if next == "" {
				idx++
			} else {
				target, ok := byId[next]
				if !ok {
					return nil, fmt.Errorf("[%d]节点 %s 跳转目标不存在: %s", ErrCodeWf, n.ID, next)
				}
				idx = target
			}
			res.Steps = append(res.Steps, step)
			continue
		case "human":
			// 人工节点：挂起等待，输出标记人工确认内容
			step.Out = "[需人工确认] " + current
			res.Status = "needs_human"
			res.Output = step.Out
			res.Steps = append(res.Steps, step)
			return res, nil
		}
		res.Steps = append(res.Steps, step)
		res.Output = current
		idx++
	}
	return res, nil
}
