// Package service llmops：模型网关调用器（OpenAI 兼容 chat/completions 与 embeddings 统一底座）
// 口径：按模型注册 ID 解析 baseUrl/密钥（AES-GCM 解密即用即弃）；能力不匹配在入口即拒；
// 响应解析为纯函数单测锚点；后续检索(C03/C08)/智能体(E02)/Prompt 优化(G03) 统一经此调用。
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeGatewayCall  = 2231
	ErrCodeGatewayParse = 2232
	ErrCodeGatewayCap   = 2233
)

var (
	chatHTTPClient = &http.Client{Timeout: 120 * time.Second}
	embedHTTPClient = &http.Client{Timeout: 60 * time.Second}
)

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatUsage token 用量
type ChatUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// modelHasCapability 模型是否具备能力（纯函数）
func modelHasCapability(m *model.LlmopsModel, capability string) bool {
	if capability == "" {
		return true
	}
	for _, c := range strings.Split(m.Capability, ",") {
		if strings.TrimSpace(c) == capability {
			return true
		}
	}
	return false
}

// gatewayEndpoint 拼接网关端点（baseUrl 去尾斜杠 + /chat/completions 或 /embeddings）
func gatewayEndpoint(baseUrl, path string) string {
	return strings.TrimRight(strings.TrimSpace(baseUrl), "/") + path
}

// buildChatPayload 组装 chat/completions 请求体（opts 透传 temperature/max_tokens 等）
func buildChatPayload(modelName string, messages []ChatMessage, opts map[string]any) (map[string]any, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("[%d]消息不能为空", ErrCodeGatewayCall)
	}
	payload := map[string]any{
		"model":    modelName,
		"messages": messages,
	}
	for k, v := range opts {
		if k == "model" || k == "messages" { // 核心字段不覆盖
			continue
		}
		payload[k] = v
	}
	return payload, nil
}

// gatewayPost 统一 POST（解密密钥即用即弃；非 200 报 2231）
func gatewayPost(ctx context.Context, client *http.Client, endpoint, apiKey string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("[%d]请求体序列化失败: %w", ErrCodeGatewayCall, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("[%d]请求构造失败: %w", ErrCodeGatewayCall, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[%d]网关调用失败: %w", ErrCodeGatewayCall, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("[%d]网关响应读取失败: %w", ErrCodeGatewayCall, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%d]网关调用失败: HTTP %d %s", ErrCodeGatewayCall, resp.StatusCode, truncateSample(string(data)))
	}
	return data, nil
}

// parseChatResponse 解析 chat/completions 响应（纯函数）
func parseChatResponse(data []byte) (string, *ChatUsage, error) {
	var parsed struct {
		Choices []struct {
			Message ChatMessage `json:"message"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", nil, fmt.Errorf("[%d]响应解析失败: %w", ErrCodeGatewayParse, err)
	}
	if len(parsed.Choices) == 0 {
		return "", nil, fmt.Errorf("[%d]响应解析失败: choices 为空", ErrCodeGatewayParse)
	}
	usage := &ChatUsage{}
	if parsed.Usage != nil {
		usage.PromptTokens = parsed.Usage.PromptTokens
		usage.CompletionTokens = parsed.Usage.CompletionTokens
		usage.TotalTokens = parsed.Usage.TotalTokens
	}
	return parsed.Choices[0].Message.Content, usage, nil
}

// parseEmbeddingsResponse 解析 embeddings 响应（纯函数）
func parseEmbeddingsResponse(data []byte) ([][]float64, error) {
	var parsed struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("[%d]响应解析失败: %w", ErrCodeGatewayParse, err)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("[%d]响应解析失败: data 为空", ErrCodeGatewayParse)
	}
	out := make([][]float64, len(parsed.Data))
	for _, d := range parsed.Data {
		if d.Index >= 0 && d.Index < len(out) {
			out[d.Index] = d.Embedding
		}
	}
	return out, nil
}

// loadModelFor 解析模型并校验能力（密文→明文即用即弃）
func (s *LlmopsServiceStruct) loadModelFor(modelID uint, capability string) (*model.LlmopsModel, string, error) {
	m, err := s.GetModel(modelID)
	if err != nil {
		return nil, "", err
	}
	if m.Status != "enabled" {
		return nil, "", fmt.Errorf("[%d]模型 %s 未启用", ErrCodeGatewayCap, m.Name)
	}
	if !modelHasCapability(m, capability) {
		return nil, "", fmt.Errorf("[%d]模型 %s 不具备 %s 能力", ErrCodeGatewayCap, m.Name, capability)
	}
	apiKey, err := crypto.Decrypt(m.ApiKeyEnc)
	if err != nil {
		return nil, "", fmt.Errorf("密钥解密失败: %w", err)
	}
	return m, apiKey, nil
}

// ChatCompletion 对话调用（要求(3) 在线调用底座）
func (s *LlmopsServiceStruct) ChatCompletion(ctx context.Context, modelID uint, messages []ChatMessage, opts map[string]any) (string, *ChatUsage, error) {
	m, apiKey, err := s.loadModelFor(modelID, "chat")
	if err != nil {
		return "", nil, err
	}
	payload, err := buildChatPayload(m.Name, messages, opts)
	if err != nil {
		return "", nil, err
	}
	data, err := gatewayPost(ctx, chatHTTPClient, gatewayEndpoint(m.BaseUrl, "/chat/completions"), apiKey, payload)
	if err != nil {
		return "", nil, err
	}
	return parseChatResponse(data)
}

// Embeddings 向量化调用
func (s *LlmopsServiceStruct) Embeddings(ctx context.Context, modelID uint, input []string) ([][]float64, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("[%d]输入不能为空", ErrCodeGatewayCall)
	}
	m, apiKey, err := s.loadModelFor(modelID, "embed")
	if err != nil {
		return nil, err
	}
	data, err := gatewayPost(ctx, embedHTTPClient, gatewayEndpoint(m.BaseUrl, "/embeddings"), apiKey, map[string]any{
		"model": m.Name, "input": input,
	})
	if err != nil {
		return nil, err
	}
	return parseEmbeddingsResponse(data)
}
