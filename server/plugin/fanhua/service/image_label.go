// Package service fanhua：图片标注任务（需求 3.4-13/14）
// 口径：图片资产登记带 train/test 用途字段；标签一图一条 upsert（JSON 数组校验），
// 写标签即置 Labeled；列表支持用途/标注状态过滤。
package service

import (
	"encoding/json"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeImgLabelParam    = 2615
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
		return nil, errx.New(ErrCodeImgLabelParam, "标签须为字符串 JSON 数组")
	}
	for _, t := range tags {
		if strings.TrimSpace(t) == "" {
			return nil, errx.New(ErrCodeImgLabelParam, "标签不能为空串")
		}
	}
	return tags, nil
}

// SaveImageAsset 登记图片资产
func (s *FanhuaService) SaveImageAsset(in *model.FanhuaImageAsset) (*model.FanhuaImageAsset, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, errx.New(ErrCodeImgLabelParam, "名称与路径必填")
	}
	if !validateImageSplit(in.Split) {
		in.Split = "none"
	}
	if in.Source != "upload" && in.Source != "url" {
		in.Source = "upload"
	}
	if err := s.repo.CreateImageAsset(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "登记图片资产失败", err)
	}
	return in, nil
}

// LabelImage 写标签（一图一条 upsert）
func (s *FanhuaService) LabelImage(assetId uint, tagsJSON string) (*model.FanhuaImageLabel, error) {
	if _, err := validateTagsJSON(tagsJSON); err != nil {
		return nil, err
	}
	db := s.dbh()
	if _, err := s.repo.FindImageAsset(db, assetId); err != nil {
		return nil, errx.New(ErrCodeImgLabelNotFound, "图片资产不存在")
	}
	label, err := s.repo.FindImageLabelByAsset(db, assetId)
	if err != nil {
		// 无既有标签 → 新建一条
		label = &model.FanhuaImageLabel{AssetId: assetId, Tags: tagsJSON, Labeled: true}
		if err := s.repo.CreateImageLabel(db, label); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "写图片标签失败", err)
		}
		return label, nil
	}
	if err := s.repo.UpdateImageLabelColumns(db, label.ID, map[string]any{"tags": tagsJSON, "labeled": true}); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "更新图片标签失败", err)
	}
	return label, nil
}

// GetImageAssets 图片资产列表（用途/关键字过滤，倒序分页）
func (s *FanhuaService) GetImageAssets(split, keyword string, page, pageSize int) ([]model.FanhuaImageAsset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	if split != "" && !validateImageSplit(split) {
		return nil, 0, errx.New(ErrCodeImgLabelParam, "用途须为 train/test/none")
	}
	list, total, err := s.repo.ImageAssetList(s.dbh(), split, strings.TrimSpace(keyword), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询图片资产失败", err)
	}
	return list, total, nil
}

// GetImageLabels 标签列表（标注状态过滤，倒序分页）
func (s *FanhuaService) GetImageLabels(labeled string, page, pageSize int) ([]map[string]any, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := s.dbh()
	labels, total, err := s.repo.ImageLabelList(db, labeled, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询图片标签失败", err)
	}
	out := make([]map[string]any, 0, len(labels))
	for _, l := range labels {
		asset, _ := s.repo.FindImageAsset(db, l.AssetId)
		out = append(out, map[string]any{"label": l, "asset": *asset})
	}
	return out, total, nil
}
