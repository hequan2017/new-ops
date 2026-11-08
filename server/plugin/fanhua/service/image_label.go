// Package service fanhua：图片标注任务（需求 3.4-13/14）
// 口径：图片资产登记带 train/test 用途字段；标签一图一条 upsert（JSON 数组校验），
// 写标签即置 Labeled；列表支持用途/标注状态过滤。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeImgLabelParam   = 2615
	ErrCodeImgLabelNotFound = 2616
)

// validateImageSplit 用途白名单（纯函数）
func validateImageSplit(s string) bool {
	return s == "train" || s == "test" || s == "none"
}

// validateTagsJSON 标签须为非空字符串 JSON 数组（纯函数）
func validateTagsJSON(raw string) ([]string, error) {
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil, fmt.Errorf("[%d]标签须为字符串 JSON 数组", ErrCodeImgLabelParam)
	}
	for _, t := range tags {
		if strings.TrimSpace(t) == "" {
			return nil, fmt.Errorf("[%d]标签不能为空串", ErrCodeImgLabelParam)
		}
	}
	return tags, nil
}

// SaveImageAsset 登记图片资产
func (s *FanhuaService) SaveImageAsset(in *model.FanhuaImageAsset) (*model.FanhuaImageAsset, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, fmt.Errorf("[%d]名称与路径必填", ErrCodeImgLabelParam)
	}
	if !validateImageSplit(in.Split) {
		in.Split = "none"
	}
	if in.Source != "upload" && in.Source != "url" {
		in.Source = "upload"
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// LabelImage 写标签（一图一条 upsert）
func (s *FanhuaService) LabelImage(assetId uint, tagsJSON string) (*model.FanhuaImageLabel, error) {
	if _, err := validateTagsJSON(tagsJSON); err != nil {
		return nil, err
	}
	var asset model.FanhuaImageAsset
	if err := global.GVA_DB.First(&asset, assetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]图片资产不存在", ErrCodeImgLabelNotFound)
	}
	var label model.FanhuaImageLabel
	err := global.GVA_DB.Where("asset_id = ?", assetId).First(&label).Error
	if err != nil {
		label = model.FanhuaImageLabel{AssetId: assetId, Tags: tagsJSON, Labeled: true}
		err = global.GVA_DB.Create(&label).Error
		return &label, err
	}
	err = global.GVA_DB.Model(&label).Updates(map[string]any{"tags": tagsJSON, "labeled": true}).Error
	return &label, err
}

// GetImageAssets 图片资产列表（用途/关键字过滤，倒序分页）
func (s *FanhuaService) GetImageAssets(split, keyword string, page, pageSize int) ([]model.FanhuaImageAsset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaImageAsset{})
	if split != "" {
		if !validateImageSplit(split) {
			return nil, 0, fmt.Errorf("[%d]用途须为 train/test/none", ErrCodeImgLabelParam)
		}
		db = db.Where("split = ?", split)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("name LIKE ?", "%"+kw+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaImageAsset
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetImageLabels 标签列表（标注状态过滤，倒序分页）
func (s *FanhuaService) GetImageLabels(labeled string, page, pageSize int) ([]map[string]any, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaImageLabel{})
	if labeled == "yes" {
		db = db.Where("labeled = ?", true)
	} else if labeled == "no" {
		db = db.Where("labeled = ?", false)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var labels []model.FanhuaImageLabel
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&labels).Error; err != nil {
		return nil, 0, err
	}
	out := make([]map[string]any, 0, len(labels))
	for _, l := range labels {
		var asset model.FanhuaImageAsset
		_ = global.GVA_DB.First(&asset, l.AssetId).Error
		out = append(out, map[string]any{"label": l, "asset": asset})
	}
	return out, total, nil
}
