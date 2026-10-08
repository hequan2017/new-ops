// Package service aiops：LLM 网关（统一 OpenAI 兼容调用、密钥密文管理）与 Pod 诊断
// 网关只认 chat/completions 协议；供应商密钥复用 asset/crypto AES-GCM 加密落库、解密即用即弃。
// 诊断 = k8s 采集（Pod 详情+事件+日志尾部）→ 固定 prompt 模板 → 网关 → 结论留档。
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/aiops/model"
)

// 错误码（aiops 段 2010-2015，agent 已占 2001-2004）
const (
	ErrCodeProviderNotFound   = 2010
	ErrCodeProviderInvalid    = 2011
	ErrCodeLLMCallFailed      = 2012
	ErrCodeDiagnosisFailed    = 2013
	ErrCodeDiagnosisNotFound  = 2014
	ErrCodeProviderNameDup    = 2015
)

var AiopsService = new(AiopsServiceStruct)

type AiopsServiceStruct struct{}

// ---------- 供应商管理 ----------

// GetProviders 供应商列表（密钥永不回显）
func (s *AiopsServiceStruct) GetProviders() ([]model.AiopsLlmProvider, error) {
	var list []model.AiopsLlmProvider
	err := global.GVA_DB.Order("is_default desc, id").Find(&list).Error
	return list, err
}

// SaveProvider 新建/更新供应商（密钥留空表示不改；isDefault=true 时清其它默认）
func (s *AiopsServiceStruct) SaveProvider(p *model.AiopsLlmProvider) error {
	if p == nil || p.Name == "" || p.BaseUrl == "" || p.Model == "" {
		return fmt.Errorf("[%d]名称/地址/模型不能为空", ErrCodeProviderInvalid)
	}
	if p.TimeoutSec <= 0 || p.TimeoutSec > 600 {
		p.TimeoutSec = 60
	}
	if p.IsDefault {
		global.GVA_DB.Model(&model.AiopsLlmProvider{}).Where("is_default = ?", true).Update("is_default", false)
	}
	if p.ID > 0 {
		updates := map[string]any{
			"name": p.Name, "base_url": p.BaseUrl, "model": p.Model,
			"is_default": p.IsDefault, "timeout_sec": p.TimeoutSec, "notes": p.Notes,
		}
		if p.ApiKeyEnc != "" {
			enc, err := crypto.Encrypt(p.ApiKeyEnc)
			if err != nil {
				return fmt.Errorf("密钥加密失败: %w", err)
			}
			updates["api_key_enc"] = enc
		}
		return global.GVA_DB.Model(&model.AiopsLlmProvider{}).Where("id = ?", p.ID).Updates(updates).Error
	}
	var count int64
	global.GVA_DB.Model(&model.AiopsLlmProvider{}).Where("name = ?", p.Name).Count(&count)
	if count > 0 {
		return fmt.Errorf("[%d]供应商名称已存在", ErrCodeProviderNameDup)
	}
	if p.ApiKeyEnc != "" {
		enc, err := crypto.Encrypt(p.ApiKeyEnc)
		if err != nil {
			return fmt.Errorf("密钥加密失败: %w", err)
		}
		p.ApiKeyEnc = enc
	}
	// 首个供应商自动设为默认
	if !p.IsDefault {
		var total int64
		global.GVA_DB.Model(&model.AiopsLlmProvider{}).Count(&total)
		p.IsDefault = total == 0
	}
	return global.GVA_DB.Create(p).Error
}

// DeleteProvider 删除供应商
func (s *AiopsServiceStruct) DeleteProvider(id uint) error {
	return global.GVA_DB.Delete(&model.AiopsLlmProvider{}, id).Error
}

// resolveProvider 解析供应商（id=0 取默认；密文解密即用即弃）
func (s *AiopsServiceStruct) resolveProvider(id uint) (*model.AiopsLlmProvider, string, error) {
	var p model.AiopsLlmProvider
	q := global.GVA_DB
	if id > 0 {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("is_default = ?", true)
	}
	if err := q.First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", fmt.Errorf("[%d]未找到 LLM 供应商配置，请先在 AI 运维页登记", ErrCodeProviderNotFound)
		}
		return nil, "", err
	}
	key := ""
	if p.ApiKeyEnc != "" {
		plain, err := crypto.Decrypt(p.ApiKeyEnc)
		if err != nil {
			return nil, "", fmt.Errorf("密钥解密失败: %w", err)
		}
		key = plain
	}
	return &p, key, nil
}

// ---------- LLM 调用 ----------

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 统一 LLM 调用（OpenAI 兼容 chat/completions；system 可空）
func (s *AiopsServiceStruct) Chat(ctx context.Context, providerID uint, system, user string) (string, string, string, error) {
	p, key, err := s.resolveProvider(providerID)
	if err != nil {
		return "", "", "", err
	}
	messages := make([]chatMessage, 0, 2)
	if system != "" {
		messages = append(messages, chatMessage{Role: "system", Content: system})
	}
	messages = append(messages, chatMessage{Role: "user", Content: user})

	body, err := json.Marshal(chatRequest{
		Model: p.Model, Messages: messages, Temperature: 0.2, Stream: false,
	})
	if err != nil {
		return "", "", "", err
	}
	url := strings.TrimRight(p.BaseUrl, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	timeout := time.Duration(p.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", p.Name, p.Model, fmt.Errorf("[%d]LLM 调用失败: %w", ErrCodeLLMCallFailed, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", p.Name, p.Model, fmt.Errorf("[%d]LLM 响应读取失败: %w", ErrCodeLLMCallFailed, err)
	}
	var cr chatResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", p.Name, p.Model, fmt.Errorf("[%d]LLM 响应非 JSON(HTTP %d): %s",
			ErrCodeLLMCallFailed, resp.StatusCode, strings.TrimSpace(string(raw))[:min(len(raw), 200)])
	}
	if cr.Error != nil && cr.Error.Message != "" {
		return "", p.Name, p.Model, fmt.Errorf("[%d]LLM 返回错误: %s", ErrCodeLLMCallFailed, cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", p.Name, p.Model, fmt.Errorf("[%d]LLM 响应无 choices(HTTP %d)", ErrCodeLLMCallFailed, resp.StatusCode)
	}
	return cr.Choices[0].Message.Content, p.Name, p.Model, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
