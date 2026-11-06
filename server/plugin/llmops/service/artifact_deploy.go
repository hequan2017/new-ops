// Package service llmops：产物→预测服务注册（要求(4) 微调后模型部署为预测服务）
// 口径：产物在模型注册表登记为可调度模型（provider 标注 artifact 来源与压缩方法），
// baseUrl 为其预测服务地址（如 vLLM），密钥/能力沿用 B01 全部校验口径。
package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeArtifactReg = 2263

// ArtifactModelIn 产物注册为模型服务入参
type ArtifactModelIn struct {
	ArtifactId uint   `json:"artifactId" binding:"required"`
	Capability string `json:"capability" binding:"required"` // chat/embed/rerank
	BaseUrl    string `json:"baseUrl" binding:"required"`
	ApiKey     string `json:"apiKey" binding:"required"`
}

var reNameSlug = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// artifactModelName 产物→注册名（slug 归一 + -served 后缀，限 96 字符）
func artifactModelName(name string, id uint) string {
	slug := reNameSlug.ReplaceAllString(strings.TrimSpace(name), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = fmt.Sprintf("artifact-%d", id)
	}
	full := fmt.Sprintf("%s-served", slug)
	if len(full) > 96 {
		full = full[:96]
	}
	return full
}

// artifactProviderLabel 产物→供应商标注（含压缩方法）
func artifactProviderLabel(art *model.LlmopsArtifact) string {
	if art.JobType == "compress" && art.Method != "" {
		return fmt.Sprintf("artifact:%s:%s", art.JobType, art.Method)
	}
	return fmt.Sprintf("artifact:%s", art.JobType)
}

// RegisterArtifactAsModel 产物一键注册为模型服务（走 SaveModel 全量校验）
func (s *LlmopsServiceStruct) RegisterArtifactAsModel(in *ArtifactModelIn) (*model.LlmopsModel, error) {
	if !ValidCapability(in.Capability) {
		return nil, fmt.Errorf("[%d]不支持的能力标签: %s", ErrCodeModelCapBad, in.Capability)
	}
	if err := ValidateModelBaseURL(in.BaseUrl); err != nil {
		return nil, err
	}
	var art model.LlmopsArtifact
	if err := s.GetArtifactById(in.ArtifactId, &art); err != nil {
		return nil, err
	}
	mi := &ModelIn{
		Name:       artifactModelName(art.Name, art.ID),
		Provider:   artifactProviderLabel(&art),
		BaseUrl:    in.BaseUrl,
		ApiKey:     in.ApiKey,
		Capability: in.Capability,
		Notes:      fmt.Sprintf("产物 #%d(%s) 一键注册", art.ID, art.Path),
	}
	return s.SaveModel(mi)
}
