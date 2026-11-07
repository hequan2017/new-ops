// Package service llmops：智能体工具调用回路（要求(5) 工具选择调用 + 结果评估与自我反思）
// 口径：AgentTool 接口为 F01 技能/F04 MCP 工具的统一接入面；循环协议——模型输出
// {"action":"call","tool":..,"args":..} 或 {"action":"answer","reply":..}；工具观察结果回填上下文
// 形成反思；轮次耗尽强制终答；全程 trace 留痕；全部依赖注入可单测。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// 错误码（llmops 段续用）
const ErrCodeToolLoop = 2271

// AgentTool 智能体可调用工具统一接口（技能/MCP 后续实现）
type AgentTool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, argsJSON string) (string, error)
}

// agentToolRegistry 工具注册表
var (
	toolRegMu    sync.RWMutex
	toolRegistry = map[string]AgentTool{}
)

// RegisterAgentTool 注册工具（重名覆盖）
func RegisterAgentTool(t AgentTool) {
	toolRegMu.Lock()
	defer toolRegMu.Unlock()
	toolRegistry[t.Name()] = t
}

// GetAgentTool 按名取工具
func GetAgentTool(name string) (AgentTool, bool) {
	toolRegMu.RLock()
	defer toolRegMu.RUnlock()
	t, ok := toolRegistry[name]
	return t, ok
}

// ToolDecision 模型决策
type ToolDecision struct {
	Action string          `json:"action"` // call/answer
	Tool   string          `json:"tool"`
	Args   json.RawMessage `json:"args"`
	Reply  string          `json:"reply"`
}

// ToolTrace 工具回路留痕
type ToolTrace struct {
	Turn   int    `json:"turn"`
	Tool   string `json:"tool"`
	Args   string `json:"args"`
	Result string `json:"result"`
	Err    string `json:"err,omitempty"`
}

// buildToolLoopPrompt 工具循环系统指令（纯函数，测试锚点）
func buildToolLoopPrompt(tools map[string]AgentTool) string {
	var sb strings.Builder
	sb.WriteString("你是具备工具调用能力的智能体。可用工具：\n")
	for _, t := range tools {
		fmt.Fprintf(&sb, "- %s: %s\n", t.Name(), t.Description())
	}
	sb.WriteString(`每轮只输出一个 JSON 对象：
{"action":"call","tool":"工具名","args":{...}} 调用工具
{"action":"answer","reply":"最终回答"} 结束并回答
工具结果会回传给你，请评估结果后再决定继续调用或作答。`)
	return sb.String()
}

// ParseToolDecision 决策解析（纯函数：均衡截取 + action 白名单）
func ParseToolDecision(text string) (*ToolDecision, error) {
	obj, ok := extractJSONObject(text)
	if !ok {
		return nil, fmt.Errorf("[%d]工具决策解析失败", ErrCodeToolLoop)
	}
	var d ToolDecision
	if err := json.Unmarshal([]byte(obj), &d); err != nil {
		return nil, fmt.Errorf("[%d]工具决策解析失败: %w", ErrCodeToolLoop, err)
	}
	if d.Action != "call" && d.Action != "answer" {
		return nil, fmt.Errorf("[%d]未知决策 action: %s", ErrCodeToolLoop, d.Action)
	}
	if d.Action == "call" && strings.TrimSpace(d.Tool) == "" {
		return nil, fmt.Errorf("[%d]call 决策缺工具名", ErrCodeToolLoop)
	}
	return &d, nil
}

// runToolLoop 工具回路核心（全依赖注入，测试锚点）：
// 返回最终回复与逐轮 trace；maxTurns 为工具调用轮上限，耗尽后强制终答。
func runToolLoop(fn turnChatFunc, tools map[string]AgentTool, question string, maxTurns int) (string, []ToolTrace, error) {
	if maxTurns <= 0 {
		maxTurns = 1
	}
	sys := buildToolLoopPrompt(tools)
	msgs := []ChatMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: question},
	}
	var trace []ToolTrace
	forceFinal := false
	for turn := 1; ; turn++ {
		if turn > maxTurns+1 { // +1 为强制终答轮
			return "", trace, fmt.Errorf("[%d]工具回路异常退出", ErrCodeToolLoop)
		}
		out, err := fn("", msgs)
		if err != nil {
			return "", trace, err
		}
		d, err := ParseToolDecision(out)
		if err != nil {
			if forceFinal {
				// 强制终答轮模型仍不听话：以其原文作答兜底
				return out, trace, nil
			}
			msgs = append(msgs, ChatMessage{Role: "assistant", Content: out},
				ChatMessage{Role: "user", Content: "输出不符合决策格式，请按 JSON 协议重输。"})
			continue
		}
		if d.Action == "answer" {
			return d.Reply, trace, nil
		}
		if forceFinal {
			return d.Reply, trace, nil // 已超轮次的 call 视作答内容兜底
		}
		// 执行工具
		t, ok := tools[d.Tool]
		entry := ToolTrace{Turn: turn, Tool: d.Tool, Args: string(d.Args)}
		if !ok {
			entry.Err = fmt.Sprintf("未知工具 %s", d.Tool)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), embedHTTPClient.Timeout)
			res, execErr := t.Execute(ctx, string(d.Args))
			cancel()
			entry.Result, entry.Err = res, ""
			if execErr != nil {
				entry.Err = execErr.Error()
			}
		}
		trace = append(trace, entry)
		if turn >= maxTurns {
			forceFinal = true
			obs := entry.Result
			if entry.Err != "" {
				obs = "工具失败: " + entry.Err
			}
			msgs = append(msgs, ChatMessage{Role: "assistant", Content: out},
				ChatMessage{Role: "user", Content: "已达工具调用轮次上限，请基于已有观察直接给出最终回答（answer）。\n最近观察：" + obs})
			continue
		}
		// 观察回填 → 模型自我反思下一轮决策
		obs := "工具 " + d.Tool + " 结果：" + entry.Result
		if entry.Err != "" {
			obs = "工具 " + d.Tool + " 失败：" + entry.Err
		}
		msgs = append(msgs, ChatMessage{Role: "assistant", Content: out},
			ChatMessage{Role: "user", Content: obs + "\n请评估该结果并继续决策。"})
	}
}
