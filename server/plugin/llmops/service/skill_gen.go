// Package service llmops：零代码创建技能（要求(6) 自然语言零代码创建专属技能）
// 口径：自然语言描述 → chat 生成 {name,desc,def} JSON → ParseSkillDef 校验 → custom 型入库；
// 解析失败带严格指令重试一次；chat 调用注入可测。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeSkillGen = 2284

// buildSkillGenPrompt 技能生成 prompt（纯函数，测试锚点）
func buildSkillGenPrompt(desc string) string {
	return "根据技能需求描述，生成技能定义。只输出 JSON 对象：\n" +
		`{"name":"技能名(≤12字)","desc":"一句话说明","def":{"kind":"prompt","system":"系统指令","template":"任务模板，用 {input} 占位用户输入，可用 {参数名} 占位补充参数","params":{"参数名":{"desc":"说明","default":"默认值"}}}}` +
		"\n需求描述：" + desc
}

// skillGenOut 生成输出结构
type skillGenOut struct {
	Name string          `json:"name"`
	Desc string          `json:"desc"`
	Def  json.RawMessage `json:"def"`
}

// ParseSkillGenOutput 生成结果解析（纯函数：均衡截取 + 名称清理 + def 校验）
func ParseSkillGenOutput(text string) (*model.LlmopsSkill, error) {
	obj, ok := extractJSONObject(text)
	if !ok {
		return nil, fmt.Errorf("[%d]技能生成解析失败", ErrCodeSkillGen)
	}
	var out skillGenOut
	if err := json.Unmarshal([]byte(obj), &out); err != nil {
		return nil, fmt.Errorf("[%d]技能生成解析失败: %w", ErrCodeSkillGen, err)
	}
	name := strings.TrimSpace(out.Name)
	name = strings.Trim(name, "\"'，。 ")
	r := []rune(name)
	if len(r) == 0 {
		return nil, fmt.Errorf("[%d]生成技能名为空", ErrCodeSkillGen)
	}
	if len(r) > 32 {
		name = string(r[:32])
	}
	defRaw := strings.TrimSpace(string(out.Def))
	if defRaw == "" {
		return nil, fmt.Errorf("[%d]生成技能定义缺失", ErrCodeSkillGen)
	}
	if _, err := ParseSkillDef(defRaw); err != nil {
		return nil, err
	}
	return &model.LlmopsSkill{
		Name: name, Type: "custom", Desc: strings.TrimSpace(out.Desc), DefJSON: defRaw, Status: "enabled",
	}, nil
}

// CreateSkillFromDescription 零代码创建（chat 注入，重试一次）
func (s *LlmopsServiceStruct) CreateSkillFromDescription(desc string) (*model.LlmopsSkill, error) {
	if strings.TrimSpace(desc) == "" {
		return nil, fmt.Errorf("[%d]需求描述不能为空", ErrCodeSkillGen)
	}
	ctx, cancel := context.WithTimeout(context.Background(), chatHTTPClient.Timeout)
	defer cancel()
	genCall := func(strict bool) (*model.LlmopsSkill, error) {
		prompt := buildSkillGenPrompt(desc)
		if strict {
			prompt += "\n注意：只输出纯 JSON 对象，禁止任何其它文字或代码块围栏。"
		}
		m, err := s.SelectModel("chat")
		if err != nil {
			return nil, err
		}
		content, _, err := s.ChatCompletion(ctx, m.ID, []ChatMessage{{Role: "user", Content: prompt}}, nil)
		if err != nil {
			return nil, err
		}
		return ParseSkillGenOutput(content)
	}
	sk, err := genCall(false)
	if err != nil {
		sk, err = genCall(true)
		if err != nil {
			return nil, err
		}
	}
	return s.SaveSkill(sk)
}
