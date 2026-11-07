// Package service llmops：数据集管理（采购件⑤ AI 开发数据集）
package service

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段末位 2299，数据集域通用）
const ErrCodeDatasetBad = 2299

// DatasetPurposes 用途白名单（要求(5) AI 开发数据集四用途）
var DatasetPurposes = []string{"sft", "rlhf", "eval", "exam"}

// ValidDatasetPurpose 用途白名单
func ValidDatasetPurpose(p string) bool {
	for _, v := range DatasetPurposes {
		if v == p {
			return true
		}
	}
	return false
}

// ValidDatasetFormat 格式白名单
func ValidDatasetFormat(f string) bool {
	return f == "jsonl" || f == "csv"
}

// CountDatasetSamples 样本计数（纯函数，测试锚点）：jsonl 非空行；csv 行-表头
func CountDatasetSamples(content, format string) int {
	if strings.TrimSpace(content) == "" {
		return 0
	}
	if format == "csv" {
		r := csv.NewReader(strings.NewReader(content))
		rows, err := r.ReadAll()
		if err != nil || len(rows) <= 1 {
			return 0
		}
		return len(rows) - 1
	}
	n := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

// GetDatasets 数据集列表（用途过滤）
func (s *LlmopsServiceStruct) GetDatasets(purpose string) ([]model.LlmopsDataset, error) {
	db := global.GVA_DB.Model(&model.LlmopsDataset{})
	if purpose != "" {
		if !ValidDatasetPurpose(purpose) {
			return nil, fmt.Errorf("[%d]不支持的用途: %s", ErrCodeDatasetBad, purpose)
		}
		db = db.Where("purpose = ?", purpose)
	}
	var list []model.LlmopsDataset
	err := db.Order("id desc").Find(&list).Error
	return list, err
}

// SaveDataset 新建/更新数据集（样本数自动计数）
func (s *LlmopsServiceStruct) SaveDataset(in *model.LlmopsDataset) (*model.LlmopsDataset, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidDatasetPurpose(in.Purpose) {
		return nil, fmt.Errorf("[%d]名称必填且用途须为 sft/rlhf/eval/exam", ErrCodeDatasetBad)
	}
	if in.Format == "" {
		in.Format = "jsonl"
	}
	if !ValidDatasetFormat(in.Format) {
		return nil, fmt.Errorf("[%d]格式须为 jsonl/csv", ErrCodeDatasetBad)
	}
	if strings.TrimSpace(in.Version) == "" {
		in.Version = "v1"
	}
	in.SampleCount = CountDatasetSamples(in.Content, in.Format)
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsDataset{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]数据集名已存在", ErrCodeDatasetBad)
	}
	if in.ID > 0 {
		var d model.LlmopsDataset
		if err := global.GVA_DB.First(&d, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
		}
		err := global.GVA_DB.Model(&d).Updates(map[string]any{
			"name": in.Name, "purpose": in.Purpose, "format": in.Format,
			"version": in.Version, "license": in.License,
			"sample_count": in.SampleCount, "content": in.Content, "desc": in.Desc,
		}).Error
		if err != nil {
			return nil, err
		}
		return &d, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteDataset 删除数据集（软删除）
func (s *LlmopsServiceStruct) DeleteDataset(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsDataset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
	}
	return nil
}
