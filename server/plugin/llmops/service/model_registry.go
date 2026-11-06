// Package service llmops：模型注册管理（要求(3) 统一接入——注册/纳管/密钥密文/能力标签）
package service

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeModelNotFound   = 2227
	ErrCodeModelCapBad     = 2228
	ErrCodeModelNameDup    = 2229
	ErrCodeModelURLBad     = 2230
)

// ModelCapabilities 能力标签白名单
var ModelCapabilities = []string{"chat", "embed", "rerank"}

// ModelIn 模型保存入参（apiKey 明文仅入参态）
type ModelIn struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	BaseUrl    string `json:"baseUrl"`
	ApiKey     string `json:"apiKey"`
	Capability string `json:"capability"`
	CtxLen     int    `json:"ctxLen"`
	Status     string `json:"status"`
	DefaultFor string `json:"defaultFor"`
	Notes      string `json:"notes"`
}

// ValidCapability 能力标签白名单
func ValidCapability(c string) bool {
	for _, v := range ModelCapabilities {
		if v == c {
			return true
		}
	}
	return false
}

// ParseCapabilities 能力标签串解析：去空白、去重；存在非法项返回错误
func ParseCapabilities(raw string) ([]string, error) {
	seen := make(map[string]bool)
	var out []string
	for _, part := range strings.Split(raw, ",") {
		c := strings.ToLower(strings.TrimSpace(part))
		if c == "" {
			continue
		}
		if !ValidCapability(c) {
			return nil, fmt.Errorf("[%d]不支持的能力标签: %s", ErrCodeModelCapBad, c)
		}
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("[%d]至少需要一个能力标签(chat/embed/rerank)", ErrCodeModelCapBad)
	}
	return out, nil
}

// ValidateModelBaseURL 模型服务地址校验（http/https + 带 host）
func ValidateModelBaseURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return fmt.Errorf("[%d]模型服务地址非法: %s", ErrCodeModelURLBad, raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("[%d]模型服务地址仅支持 http/https", ErrCodeModelURLBad)
	}
	return nil
}

// GetModels 模型注册列表（密钥不回显）
func (s *LlmopsServiceStruct) GetModels() ([]model.LlmopsModel, error) {
	var list []model.LlmopsModel
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SaveModel 新建/更新模型注册（密钥留空不改；DefaultFor 互斥清位）
func (s *LlmopsServiceStruct) SaveModel(in *ModelIn) (*model.LlmopsModel, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("[%d]模型注册名不能为空", ErrCodeModelNotFound)
	}
	if err := ValidateModelBaseURL(in.BaseUrl); err != nil {
		return nil, err
	}
	caps, err := ParseCapabilities(in.Capability)
	if err != nil {
		return nil, err
	}
	if in.DefaultFor != "" && !ValidCapability(in.DefaultFor) {
		return nil, fmt.Errorf("[%d]不支持的能力标签: %s", ErrCodeModelCapBad, in.DefaultFor)
	}
	hasCap := false
	for _, c := range caps {
		if c == in.DefaultFor {
			hasCap = true
		}
	}
	if in.DefaultFor != "" && !hasCap {
		return nil, fmt.Errorf("[%d]默认调度能力 %s 须在能力标签内", ErrCodeModelCapBad, in.DefaultFor)
	}
	if in.CtxLen <= 0 {
		in.CtxLen = 8192
	}
	if in.Status == "" {
		in.Status = "enabled"
	}

	var count int64
	q := global.GVA_DB.Model(&model.LlmopsModel{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]模型注册名已存在", ErrCodeModelNameDup)
	}

	capStr := strings.Join(caps, ",")
	if in.ID > 0 {
		var m model.LlmopsModel
		if err := global.GVA_DB.First(&m, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]模型不存在", ErrCodeModelNotFound)
		}
		updates := map[string]any{
			"name": in.Name, "provider": in.Provider, "base_url": in.BaseUrl,
			"capability": capStr, "ctx_len": in.CtxLen, "status": in.Status,
			"default_for": in.DefaultFor, "notes": in.Notes,
		}
		if in.ApiKey != "" {
			enc, err := crypto.Encrypt(in.ApiKey)
			if err != nil {
				return nil, fmt.Errorf("密钥加密失败: %w", err)
			}
			updates["api_key_enc"] = enc
		}
		if err := global.GVA_DB.Model(&m).Updates(updates).Error; err != nil {
			return nil, err
		}
	} else {
		if in.ApiKey == "" {
			return nil, fmt.Errorf("[%d]新建模型必须提供 API Key", ErrCodeModelNotFound)
		}
		enc, err := crypto.Encrypt(in.ApiKey)
		if err != nil {
			return nil, fmt.Errorf("密钥加密失败: %w", err)
		}
		m := &model.LlmopsModel{
			Name: in.Name, Provider: in.Provider, BaseUrl: in.BaseUrl, ApiKeyEnc: enc,
			Protocol: "openai", Capability: capStr, CtxLen: in.CtxLen,
			Status: in.Status, DefaultFor: in.DefaultFor, Notes: in.Notes,
		}
		if err := global.GVA_DB.Create(m).Error; err != nil {
			return nil, err
		}
	}
	// DefaultFor 互斥：同能力其它模型清位
	if in.DefaultFor != "" {
		clear := global.GVA_DB.Model(&model.LlmopsModel{}).
			Where("default_for = ? AND name <> ?", in.DefaultFor, in.Name).Update("default_for", "")
		if clear.Error != nil {
			return nil, clear.Error
		}
	}
	return s.loadModelByName(in.Name)
}

// loadModelByName 按注册名取模型
func (s *LlmopsServiceStruct) loadModelByName(name string) (*model.LlmopsModel, error) {
	var m model.LlmopsModel
	err := global.GVA_DB.Where("name = ?", name).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("[%d]模型不存在", ErrCodeModelNotFound)
		}
		return nil, err
	}
	return &m, nil
}

// GetModel 模型详情（密钥不回显）
func (s *LlmopsServiceStruct) GetModel(id uint) (*model.LlmopsModel, error) {
	var m model.LlmopsModel
	err := global.GVA_DB.First(&m, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("[%d]模型不存在", ErrCodeModelNotFound)
		}
		return nil, err
	}
	return &m, nil
}

// DeleteModel 删除模型注册（软删除）
func (s *LlmopsServiceStruct) DeleteModel(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]模型不存在", ErrCodeModelNotFound)
	}
	return nil
}
