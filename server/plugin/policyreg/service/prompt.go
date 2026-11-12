// Package service policyreg：提示词库与问答（需求 3.8-8）
// 口径：{{var}} 解析渲染自实现（对齐 G01/N12/O09 口径，缺失变量报错）；
// 问答=检索（角色过滤制度 top3）+模板注入+chat 可选（policyChatFn 注入锚点，
// 默认 llmops 网关，未配置模型时规则口径 fallback，假模型单测）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const ErrCodePrPromptParam = 3009

// varPatternPr {{var}} 占位
var varPatternPr = regexp.MustCompile(`\{\{(\w+)\}\}`)

// parsePrVars 占位解析（纯函数，测试锚点）
func parsePrVars(tpl string) []string {
	seen := map[string]bool{}
	var vars []string
	for _, m := range varPatternPr.FindAllStringSubmatch(tpl, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			vars = append(vars, m[1])
		}
	}
	return vars
}

// renderPrTemplate 模板渲染（纯函数，测试锚点）：缺失变量报错
func renderPrTemplate(tpl string, vars map[string]string) (string, error) {
	var missing []string
	for _, v := range parsePrVars(tpl) {
		if _, ok := vars[v]; !ok {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return "", errx.New(ErrCodePrPromptParam, "缺失占位变量: "+strings.Join(missing, ","))
	}
	return varPatternPr.ReplaceAllStringFunc(tpl, func(m string) string {
		name := varPatternPr.FindStringSubmatch(m)[1]
		return vars[name]
	}), nil
}

// policyChatFn chat 网关调用点（测试替换锚点；默认走 llmops 网关，先例 exam J03/N13）
var policyChatFn = func(ctx context.Context, modelId uint, prompt string) (string, error) {
	return "", fmt.Errorf("未配置 chat 模型（modelId=%d）", modelId)
}

// rulePolicyAnswer 规则口径回答（纯函数，测试锚点）：检索命中组合
func rulePolicyAnswer(question string, docs []model.PolicyDoc) string {
	var sb strings.Builder
	if len(docs) == 0 {
		return "未检索到相关制度，请换个问法或联系制度管理部门。"
	}
	sb.WriteString(fmt.Sprintf("就「%s」检索到 %d 项相关制度：\n", question, len(docs)))
	for i, d := range docs {
		summary := d.MaskedText
		if summary == "" {
			summary = d.RawText
		}
		if len([]rune(summary)) > 60 {
			summary = string([]rune(summary)[:60]) + "…"
		}
		fmt.Fprintf(&sb, "%d.《%s》（%s/%s）：%s\n", i+1, d.Name, d.SecurityLevel, d.Format, summary)
	}
	sb.WriteString("以上为制度原文初筛，正式引用请以发文机关公布版本为准。")
	return sb.String()
}

// PolicyQa 制度问答（检索+模板注入+chat 可选）
func (s *PolicyService) PolicyQa(roleName, question string, modelId uint) (*model.PolicyQaLog, error) {
	start := time.Now()
	if strings.TrimSpace(question) == "" {
		return nil, errx.New(ErrCodePrPromptParam, "问题必填")
	}
	docs, _, err := s.SearchPolicyDocsByRole(roleName, "", 1, 3)
	if err != nil {
		return nil, err
	}
	// 模板注入
	prompt := ""
	if tpl, tErr := s.repo.FindEnabledTemplateByName(s.dbh(), "制度问答-标准"); tErr == nil {
		docDigest := rulePolicyAnswer(question, docs)
		rendered, rErr := renderPrTemplate(tpl.Content, map[string]string{
			"question": question, "role": roleName, "refs": docDigest,
		})
		if rErr == nil {
			prompt = rendered
		}
	}
	source := "规则"
	answer := ""
	if prompt != "" && modelId > 0 {
		if content, chatErr := policyChatFn(context.Background(), modelId, prompt); chatErr == nil && strings.TrimSpace(content) != "" {
			answer = content
			source = "chat"
		}
	}
	if answer == "" {
		answer = rulePolicyAnswer(question, docs)
	}
	log := &model.PolicyQaLog{
		Question: question, Answer: answer, Source: source,
		LatencyMs: time.Since(start).Milliseconds(), RoleName: roleName,
	}
	if err := s.repo.CreateQaLog(s.dbh(), log); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "落问答日志失败", err)
	}
	return log, nil
}

// SavePolicyPrompt 新建/更新模板（vars 自动解析）
func (s *PolicyService) SavePolicyPrompt(in *model.PolicyPromptTemplate) (*model.PolicyPromptTemplate, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Content) == "" {
		return nil, errx.New(ErrCodePrPromptParam, "模板名与内容必填")
	}
	vars, _ := json.Marshal(parsePrVars(in.Content))
	in.Vars = string(vars)
	if exist, err := s.repo.FindTemplateByName(s.dbh(), in.Name); err == nil {
		if err := s.repo.UpdateTemplateColumns(s.dbh(), exist.ID, map[string]any{
			"content": in.Content, "vars": in.Vars, "enabled": in.Enabled,
		}); err != nil {
			return nil, errx.Wrap(ErrCodePrStore, "更新模板失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateTemplate(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "创建模板失败", err)
	}
	return in, nil
}

// RenderPolicyPromptByName 按模板名渲染
func (s *PolicyService) RenderPolicyPromptByName(name string, vars map[string]string) (map[string]any, error) {
	tpl, err := s.repo.FindTemplateByName(s.dbh(), name)
	if err != nil {
		return nil, errx.New(ErrCodePrPromptParam, "模板不存在: "+name)
	}
	rendered, err := renderPrTemplate(tpl.Content, vars)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": tpl.Name, "rendered": rendered, "vars": tpl.Vars}, nil
}

// GetPolicyPrompts 模板列表
func (s *PolicyService) GetPolicyPrompts() ([]model.PolicyPromptTemplate, error) {
	list, err := s.repo.TemplateList(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "查询模板列表失败", err)
	}
	return list, nil
}

// GetPolicyQaLogs 问答日志列表（角色过滤，倒序分页）
func (s *PolicyService) GetPolicyQaLogs(roleName string, page, pageSize int) ([]model.PolicyQaLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	list, total, err := s.repo.QaLogList(s.dbh(), strings.TrimSpace(roleName), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodePrStore, "查询问答日志失败", err)
	}
	return list, total, nil
}
