// Package service artsci：图像解读与分析（需求 3.3-13）
// 口径：解读=结构化两段（表面细节/文化含义）；ModelId>0 走 chat JSON（失败重试回落），
// ModelId=0 元数据口径 scaffold（名称/MIME/来源/视角引导）；复用 gradingChatFn 网关调用点。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// 错误码（artsci 段 2500-2599）
const (
	ErrCodeInterpParam = 2517
	ErrCodeInterpLLM   = 2518
)

// parseInterpJSON 解析解读输出
func parseInterpJSON(content string) (surface, cultural string, err error) {
	l, r := strings.Index(content, "{"), strings.LastIndex(content, "}")
	if l < 0 || r <= l {
		return "", "", fmt.Errorf("输出不含 JSON 对象")
	}
	var out struct {
		Surface  string `json:"surface"`
		Cultural string `json:"cultural"`
	}
	if err := json.Unmarshal([]byte(content[l:r+1]), &out); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(out.Surface) == "" || strings.TrimSpace(out.Cultural) == "" {
		return "", "", fmt.Errorf("surface/cultural 字段必填")
	}
	return out.Surface, out.Cultural, nil
}

// ruleInterpretation 元数据口径解读（纯函数，测试锚点）
func ruleInterpretation(asset *model.ArtsciImageAsset, subject string) (surface, cultural string) {
	surface = fmt.Sprintf("图像《%s》：格式 %s，大小 %d 字节，来源 %s（元数据口径编目）。", asset.Name, asset.MIME, asset.SizeByte, asset.Source)
	cultural = fmt.Sprintf("「%s」视角引导：结合《%s》的元数据（%s）中的年代背景与创作语境，从画面符号、社会环境与历史脉络三个层面展开文化含义解读。",
		subject, asset.Name, asset.MetaJSON)
	return surface, cultural
}

// InterpretImage 图像解读（无 DB 部分为 gradeEngine 同构编排；DB：资产加载+记录落库）
func (s *ArtsciService) InterpretImage(assetId uint, subject string, modelId uint) (*model.ArtsciInterpretation, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("[%d]解读视角必填", ErrCodeInterpParam)
	}
	var asset model.ArtsciImageAsset
	if err := global.GVA_DB.First(&asset, assetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]图像资产不存在", ErrCodeInterpParam)
	}
	rec := &model.ArtsciInterpretation{AssetId: assetId, Subject: subject, ModelId: modelId}
	if modelId > 0 {
		prompt := "你是图像解读助手。针对图像《" + asset.Name + "》，从「" + subject + "」视角解读，只输出 JSON：\n" +
			"{\"surface\":\"表面细节描述（构图/色彩/元素）\",\"cultural\":\"文化、社会与历史含义\"}\n图像元数据：" + asset.MetaJSON
		var lastErr error
		for attempt := 0; attempt < 2; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			content, err := gradingChatFn(ctx, modelId, prompt)
			cancel()
			if err != nil {
				lastErr = fmt.Errorf("[%d]模型调用失败: %w", ErrCodeInterpLLM, err)
				continue
			}
			surface, cultural, perr := parseInterpJSON(content)
			if perr == nil {
				rec.Engine, rec.Ok, rec.SurfaceDesc, rec.CulturalMeaning = "llm", true, surface, cultural
				if err := global.GVA_DB.Create(rec).Error; err != nil {
					return nil, err
				}
				return rec, nil
			}
			lastErr = fmt.Errorf("[%d]解读输出解析失败: %w", ErrCodeInterpLLM, perr)
		}
		rec.Ok, rec.ErrMsg = false, truncateRunText(lastErr.Error())
		_ = global.GVA_DB.Create(rec).Error
		return rec, lastErr
	}
	// 元数据口径
	surface, cultural := ruleInterpretation(&asset, subject)
	rec.Engine, rec.Ok = "rule", true
	rec.SurfaceDesc, rec.CulturalMeaning = surface, cultural
	if err := global.GVA_DB.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// RegisterImageAsset 图像资产登记
func (s *ArtsciService) RegisterImageAsset(in *model.ArtsciImageAsset) (*model.ArtsciImageAsset, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, fmt.Errorf("[%d]名称与路径必填", ErrCodeInterpParam)
	}
	if in.Source != "upload" && in.Source != "url" && in.Source != "generated" {
		in.Source = "upload"
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetImageAssets 图像资产列表（关键字过滤，倒序分页）
func (s *ArtsciService) GetImageAssets(keyword string, page, pageSize int) ([]model.ArtsciImageAsset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ArtsciImageAsset{})
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("name LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ArtsciImageAsset
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetInterpretations 解读记录列表（资产过滤，倒序分页）
func (s *ArtsciService) GetInterpretations(assetId uint, page, pageSize int) ([]model.ArtsciInterpretation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ArtsciInterpretation{})
	if assetId > 0 {
		db = db.Where("asset_id = ?", assetId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ArtsciInterpretation
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
