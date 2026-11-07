// Package service llmops：MCP 工具封装（要求(6) 业务系统 API→标准工具，双向交互）
// 口径：API 注册（endpoint/method/参数模板/固定头值 AES-GCM 密文不回显）→ mcpToolAdapter
// 接入 E03 工具注册表；调用代理校验参数键 ⊆ 模板键，GET 走 query、POST 走 JSON body，
// 10s 超时 + 2MB 响应上限 + 4000 rune 截断。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"

	"github.com/hequan2017/new-ops/server/global"
)

// 错误码（llmops 段续用）
const (
	ErrCodeMcpNotFound  = 2286
	ErrCodeMcpBad       = 2287
	ErrCodeMcpArgs      = 2288
	mcpCallTimeout      = 10 * time.Second
	mcpMaxRespBytes     int64 = 2 << 20
	mcpResultMaxRunes         = 4000
)

var mcpHTTPClient = &http.Client{Timeout: mcpCallTimeout}

// ValidMcpMethod 方法白名单
func ValidMcpMethod(m string) bool {
	return m == "GET" || m == "POST"
}

// ParseArgsSchema 参数模板解析（纯函数：字段→说明）
func ParseArgsSchema(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}
	var schema map[string]string
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, fmt.Errorf("[%d]参数模板须为 JSON 对象(字段→说明): %w", ErrCodeMcpBad, err)
	}
	return schema, nil
}

// validateMcpArgs 调用参数校验（纯函数：键 ⊆ 模板键）
func validateMcpArgs(schema map[string]string, args map[string]any) error {
	for k := range args {
		if _, ok := schema[k]; !ok {
			return fmt.Errorf("[%d]未知参数: %s", ErrCodeMcpArgs, k)
		}
	}
	return nil
}

// GetMcpTools MCP 工具列表
func (s *LlmopsServiceStruct) GetMcpTools() ([]model.LlmopsMcpTool, error) {
	var list []model.LlmopsMcpTool
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SaveMcpTool 新建/更新 MCP 工具（headersJSON 明文入参→逐值加密落库）
func (s *LlmopsServiceStruct) SaveMcpTool(in *model.LlmopsMcpTool, headersJSON string) (*model.LlmopsMcpTool, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidMcpMethod(in.Method) {
		return nil, fmt.Errorf("[%d]名称必填且 method 须为 GET/POST", ErrCodeMcpBad)
	}
	if err := ValidateModelBaseURL(in.Endpoint); err != nil {
		return nil, fmt.Errorf("[%d]%w", ErrCodeMcpBad, err)
	}
	if _, err := ParseArgsSchema(in.ArgsSchema); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsMcpTool{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]工具名已存在", ErrCodeMcpBad)
	}
	if in.ID > 0 {
		var t model.LlmopsMcpTool
		if err := global.GVA_DB.First(&t, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]工具不存在", ErrCodeMcpNotFound)
		}
		updates := map[string]any{
			"name": in.Name, "desc": in.Desc, "endpoint": in.Endpoint,
			"method": in.Method, "args_schema": in.ArgsSchema, "status": in.Status,
		}
		if strings.TrimSpace(headersJSON) != "" {
			enc, err := encryptHeaderValues(headersJSON)
			if err != nil {
				return nil, err
			}
			updates["headers_json"] = enc
		}
		if err := global.GVA_DB.Model(&t).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &t, nil
	}
	if strings.TrimSpace(headersJSON) != "" {
		enc, err := encryptHeaderValues(headersJSON)
		if err != nil {
			return nil, err
		}
		in.HeadersJSON = enc
	}
	if in.Status == "" {
		in.Status = "enabled"
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// encryptHeaderValues 固定头逐值加密（纯 JSON 对象入参）
func encryptHeaderValues(headersJSON string) (string, error) {
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return "", fmt.Errorf("[%d]headers 须为 JSON 对象(键→值)", ErrCodeMcpBad)
	}
	enc := make(map[string]string, len(headers))
	for k, v := range headers {
		e, err := crypto.Encrypt(v)
		if err != nil {
			return "", fmt.Errorf("header %s 加密失败: %w", k, err)
		}
		enc[k] = e
	}
	b, _ := json.Marshal(enc)
	return string(b), nil
}

// decryptHeaderValues 固定头解密（调用期即用即弃）
func decryptHeaderValues(headersJSON string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(headersJSON) == "" {
		return out
	}
	var enc map[string]string
	if json.Unmarshal([]byte(headersJSON), &enc) != nil {
		return out
	}
	for k, v := range enc {
		if plain, err := crypto.Decrypt(v); err == nil {
			out[k] = plain
		}
	}
	return out
}

// DeleteMcpTool 删除 MCP 工具
func (s *LlmopsServiceStruct) DeleteMcpTool(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsMcpTool{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]工具不存在", ErrCodeMcpNotFound)
	}
	return nil
}

// mcpToolAdapter MCP 工具→AgentTool 适配器
type mcpToolAdapter struct {
	tool model.LlmopsMcpTool
}

func (a *mcpToolAdapter) Name() string        { return "mcp:" + a.tool.Name }
func (a *mcpToolAdapter) Description() string { return a.tool.Desc }

// EndpointURL 暴露目标地址（F06 沙箱 URL 白名单受检面）
func (a *mcpToolAdapter) EndpointURL() string { return a.tool.Endpoint }

// Execute 调用代理：参数校验 → HTTP 调用 → 截断返回
func (a *mcpToolAdapter) Execute(_ context.Context, argsJSON string) (string, error) {
	schema, _ := ParseArgsSchema(a.tool.ArgsSchema)
	args := map[string]any{}
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("[%d]参数须为 JSON 对象: %w", ErrCodeMcpArgs, err)
		}
	}
	if err := validateMcpArgs(schema, args); err != nil {
		return "", err
	}
	var req *http.Request
	var err error
	if a.tool.Method == "GET" {
		q := url.Values{}
		for k, v := range args {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		req, err = http.NewRequest(http.MethodGet, a.tool.Endpoint+"?"+q.Encode(), nil)
	} else {
		body, _ := json.Marshal(args)
		req, err = http.NewRequest(http.MethodPost, a.tool.Endpoint, strings.NewReader(string(body)))
	}
	if err != nil {
		return "", fmt.Errorf("[%d]请求构造失败: %w", ErrCodeMcpBad, err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range decryptHeaderValues(a.tool.HeadersJSON) {
		req.Header.Set(k, v)
	}
	resp, err := mcpHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("[%d]调用失败: %w", ErrCodeMcpBad, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, mcpMaxRespBytes))
	if err != nil {
		return "", fmt.Errorf("[%d]响应读取失败: %w", ErrCodeMcpBad, err)
	}
	out := fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, truncateSample(string(data)))
	if len([]rune(out)) > mcpResultMaxRunes {
		out = string([]rune(out)[:mcpResultMaxRunes])
	}
	return out, nil
}

// RegisterMcpToolAdapters 启用 MCP 工具注册进工具注册表（启动期 best-effort）
func (s *LlmopsServiceStruct) RegisterMcpToolAdapters() {
	var tools []model.LlmopsMcpTool
	if err := global.GVA_DB.Where("status = 'enabled'").Find(&tools).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops MCP 适配器装载失败: %v", err))
		return
	}
	for i := range tools {
		RegisterAgentTool(&mcpToolAdapter{tool: tools[i]})
	}
}
