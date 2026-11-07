// Package service oplib 管道编排单测（定义校验/拓扑序/执行链）
package service

import (
	"strings"
	"testing"
)

func TestParsePipelineDef(t *testing.T) {
	// 合法链 + 拓扑序
	def, order, err := ParsePipelineDef(`{
		"nodes":[{"id":"b","operator":"dedupe_lines"},{"id":"a","operator":"text_clean"}],
		"edges":[["a","b"]]}`)
	if err != nil {
		t.Fatalf("合法定义失败: %v", err)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("拓扑序异常: %v", order)
	}
	_ = def
	// 环
	if _, _, err := ParsePipelineDef(`{
		"nodes":[{"id":"a","operator":"x"},{"id":"b","operator":"y"}],
		"edges":[["a","b"],["b","a"]]}`); err == nil || !strings.Contains(err.Error(), "环") {
		t.Fatalf("环应拒绝: %v", err)
	}
	// 多父
	if _, _, err := ParsePipelineDef(`{
		"nodes":[{"id":"a","operator":"x"},{"id":"b","operator":"y"},{"id":"c","operator":"z"}],
		"edges":[["a","c"],["b","c"]]}`); err == nil || !strings.Contains(err.Error(), "多个父") {
		t.Fatalf("多父应拒绝: %v", err)
	}
	// 边引用不存在
	if _, _, err := ParsePipelineDef(`{
		"nodes":[{"id":"a","operator":"x"}],
		"edges":[["a","ghost"]]}`); err == nil {
		t.Fatal("幽灵边应拒绝")
	}
	// 重复 ID / 缺算子
	if _, _, err := ParsePipelineDef(`{"nodes":[{"id":"a","operator":"x"},{"id":"a","operator":"y"}]}`); err == nil {
		t.Fatal("重复 ID 应拒绝")
	}
	if _, _, err := ParsePipelineDef(`{"nodes":[{"id":"a"}]}`); err == nil {
		t.Fatal("缺算子应拒绝")
	}
}

// TestRunPipelineDef 执行链：{input} 渲染 + 默认 text 透传 + 节点失败中断
func TestRunPipelineDef(t *testing.T) {
	def, order, _ := ParsePipelineDef(`{
		"nodes":[
			{"id":"a","operator":"clean","args":{"text":"{input}"}},
			{"id":"b","operator":"count"}],
		"edges":[["a","b"]]}`)
	var calls []string
	executor := func(operator, argsJSON string) (any, error) {
		calls = append(calls, operator+":"+argsJSON)
		if operator == "clean" {
			return "干净文本", nil
		}
		return map[string]any{"chars": 4}, nil
	}
	res, err := RunPipelineDef(def, order, executor, "原始输入")
	if err != nil || res.Status != "done" || res.Output != "map[chars:4]" || len(res.Steps) != 2 {
		t.Fatalf("执行链异常: %+v %v", res, err)
	}
	// 节点 a 渲染 {input}；节点 b 无 args → 默认 text=父输出
	if !strings.Contains(calls[0], `"text":"原始输入"`) {
		t.Fatalf("a 渲染异常: %s", calls[0])
	}
	if !strings.Contains(calls[1], `"text":"干净文本"`) {
		t.Fatalf("b 默认透传异常: %s", calls[1])
	}
	// 节点失败 → failed + 留痕
	def2, order2, _ := ParsePipelineDef(`{"nodes":[{"id":"a","operator":"bad"}]}`)
	res2, err := RunPipelineDef(def2, order2, func(string, string) (any, error) {
		return nil, errTestPipe{}
	}, "x")
	if err == nil || res2.Status != "failed" || len(res2.Steps) != 1 {
		t.Fatalf("失败路径异常: %+v %v", res2, err)
	}
}

type errTestPipe struct{}

func (errTestPipe) Error() string { return "boom" }
