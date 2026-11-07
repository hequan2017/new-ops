// Package service llmops 工具回路单测（决策解析/循环链路 fake 工具+fake 模型/轮次上限强制终答）
package service

import (
	"context"
	"strings"
	"testing"
)

// echoTool 假工具
type echoTool struct{}

func (echoTool) Name() string { return "echo" }
func (echoTool) Description() string { return "原样返回输入" }
func (echoTool) Execute(_ context.Context, argsJSON string) (string, error) {
	return "ECHO:" + argsJSON, nil
}

// boomTool 失败工具
type boomTool struct{}

func (boomTool) Name() string { return "boom" }
func (boomTool) Description() string { return "总是失败" }
func (boomTool) Execute(_ context.Context, _ string) (string, error) {
	return "", contextErr{}
}

func TestParseToolDecision(t *testing.T) {
	d, err := ParseToolDecision(`{"action":"call","tool":"echo","args":{"x":1}}`)
	if err != nil || d.Action != "call" || d.Tool != "echo" || string(d.Args) != `{"x":1}` {
		t.Fatalf("call 解析异常: %+v %v", d, err)
	}
	d, err = ParseToolDecision("前缀 {\"action\":\"answer\",\"reply\":\"答案\"} 后缀")
	if err != nil || d.Action != "answer" || d.Reply != "答案" {
		t.Fatalf("answer 解析异常: %+v %v", d, err)
	}
	if _, err := ParseToolDecision(`{"action":"dance"}`); err == nil {
		t.Fatal("未知 action 应拒绝")
	}
	if _, err := ParseToolDecision(`{"action":"call"}`); err == nil {
		t.Fatal("call 缺工具名应拒绝")
	}
	if _, err := ParseToolDecision("no json"); err == nil {
		t.Fatal("无 JSON 应拒绝")
	}
}

// TestToolLoop_CallThenAnswer call→观察→answer 两轮链路
func TestToolLoop_CallThenAnswer(t *testing.T) {
	tools := map[string]AgentTool{"echo": echoTool{}}
	script := []string{
		`{"action":"call","tool":"echo","args":{"msg":"hi"}}`,
		`{"action":"answer","reply":"工具说 ECHO 了"}`,
	}
	i := 0
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		if i >= len(script) {
			t.Fatalf("脚本越界 turn %d", i)
		}
		out := script[i]
		i++
		// 第 2 轮应包含工具观察回填
		if i == 2 {
			found := false
			for _, m := range msgs {
				if strings.Contains(m.Content, "ECHO:") {
					found = true
				}
			}
			if !found {
				t.Fatal("第 2 轮上下文应含工具观察")
			}
		}
		return out, nil
	}
	reply, trace, err := runToolLoop(fn, tools, "调用 echo", 3)
	if err != nil || reply != "工具说 ECHO 了" || len(trace) != 1 {
		t.Fatalf("链路异常: %q %+v %v", reply, trace, err)
	}
	if trace[0].Result != `ECHO:{"msg":"hi"}` || trace[0].Err != "" {
		t.Fatalf("trace 异常: %+v", trace[0])
	}
}

// TestToolLoop_MaxTurnsForce 轮次耗尽强制终答
func TestToolLoop_MaxTurnsForce(t *testing.T) {
	tools := map[string]AgentTool{"echo": echoTool{}}
	calls := 0
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		calls++
		for _, m := range msgs {
			if m.Role == "user" && strings.Contains(m.Content, "已达工具调用轮次上限") {
				return `{"action":"answer","reply":"基于观察的最终答案"}`, nil
			}
		}
		return `{"action":"call","tool":"echo","args":{}}`, nil
	}
	reply, trace, err := runToolLoop(fn, tools, "一直调用", 2)
	if err != nil || reply != "基于观察的最终答案" || len(trace) != 2 || calls != 3 {
		t.Fatalf("强制终答链路异常: %q trace=%d calls=%d err=%v", reply, len(trace), calls, err)
	}
}

// TestToolLoop_Errors 工具失败回填与未知工具
func TestToolLoop_Errors(t *testing.T) {
	tools := map[string]AgentTool{"boom": boomTool{}, "echo": echoTool{}}
	script := []string{
		`{"action":"call","tool":"boom","args":{}}`,
		`{"action":"call","tool":"ghost","args":{}}`,
		`{"action":"answer","reply":"降级回答"}`,
	}
	i := 0
	var sawBoomObs, sawGhostObs bool
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		out := script[i]
		i++
		for _, m := range msgs {
			if strings.Contains(m.Content, "工具 boom 失败") {
				sawBoomObs = true
			}
			if strings.Contains(m.Content, "未知工具 ghost") {
				sawGhostObs = true
			}
		}
		return out, nil
	}
	reply, trace, err := runToolLoop(fn, tools, "混合", 5)
	if err != nil || reply != "降级回答" || len(trace) != 2 {
		t.Fatalf("错误链路异常: %q %+v %v", reply, trace, err)
	}
	if !sawBoomObs || !sawGhostObs {
		t.Fatalf("失败/未知工具观察未回填: %v %v", sawBoomObs, sawGhostObs)
	}
	if trace[0].Err == "" || trace[1].Err == "" {
		t.Fatalf("trace 应记录错误: %+v", trace)
	}
}

// TestToolRegistry 注册表读写
func TestToolRegistry(t *testing.T) {
	RegisterAgentTool(echoTool{})
	if _, ok := GetAgentTool("echo"); !ok {
		t.Fatal("echo 应已注册")
	}
	if _, ok := GetAgentTool("nope"); ok {
		t.Fatal("未注册工具不应命中")
	}
}
