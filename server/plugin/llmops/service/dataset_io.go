// Package service llmops：数据集导入导出（采购件⑤）——用途 schema 逐行校验 + 下载
// 口径：JSONL 逐行校验必填字段与类型，坏行报告 {line,reason}（上限 20 条）；
// 导入仅保留有效行重写入库并回填样本数；导出原样返回内容供前端下载。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 数据集行校验错误报告上限
const datasetRowErrMax = 20

// DatasetRowError 坏行报告
type DatasetRowError struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

// datasetSampleSchemas 用途 → 必填字段(类型)
func datasetSampleSchemas(purpose string) map[string]string {
	switch purpose {
	case "sft":
		return map[string]string{"instruction": "string", "output": "string"}
	case "rlhf":
		return map[string]string{"prompt": "string", "chosen": "string", "rejected": "string"}
	case "eval":
		return map[string]string{"question": "string", "expected": "string"}
	default: // exam
		return map[string]string{"type": "string", "stem": "string", "answer": "string"}
	}
}

// ValidateDatasetRows 逐行校验（纯函数，测试锚点）：返回有效行数与坏行报告
func ValidateDatasetRows(content, purpose string) (int, []DatasetRowError) {
	schema := datasetSampleSchemas(purpose)
	valid := 0
	var errs []DatasetRowError
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			errs = append(errs, DatasetRowError{Line: i + 1, Reason: "JSON 解析失败"})
			if len(errs) >= datasetRowErrMax {
				break
			}
			continue
		}
		missing := []string{}
		badType := ""
		for field, typ := range schema {
			v, ok := row[field]
			if !ok || v == nil || fmt.Sprintf("%v", v) == "" {
				missing = append(missing, field)
				continue
			}
			if typ == "string" {
				if _, ok := v.(string); !ok {
					badType = field
				}
			}
		}
		if len(missing) > 0 {
			errs = append(errs, DatasetRowError{Line: i + 1, Reason: "缺少字段: " + strings.Join(missing, ",")})
		} else if badType != "" {
			errs = append(errs, DatasetRowError{Line: i + 1, Reason: "字段须为字符串: " + badType})
		} else {
			valid++
		}
		if len(errs) >= datasetRowErrMax {
			break
		}
	}
	return valid, errs
}

// ImportDatasetContent 导入内容（校验→仅有效行重写入库）
func (s *LlmopsServiceStruct) ImportDatasetContent(id uint, content, purpose string) (int, []DatasetRowError, error) {
	var d model.LlmopsDataset
	if err := global.GVA_DB.First(&d, id).Error; err != nil {
		return 0, nil, fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
	}
	if purpose == "" {
		purpose = d.Purpose
	}
	if !ValidDatasetPurpose(purpose) {
		return 0, nil, fmt.Errorf("[%d]用途非法: %s", ErrCodeDatasetBad, purpose)
	}
	if d.Format == "csv" {
		return 0, nil, fmt.Errorf("[%d]逐行校验仅支持 jsonl 格式", ErrCodeDatasetBad)
	}
	valid, errs := ValidateDatasetRows(content, purpose)
	if valid == 0 {
		reason := "无有效样本"
		if len(errs) > 0 {
			reason = fmt.Sprintf("无有效样本（首错第 %d 行: %s）", errs[0].Line, errs[0].Reason)
		}
		return 0, errs, fmt.Errorf("[%d]%s", ErrCodeDatasetBad, reason)
	}
	// 重写：仅有效行
	var kept []string
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		bad := false
		for _, e := range errs {
			if e.Line == i+1 {
				bad = true
				break
			}
		}
		if !bad {
			kept = append(kept, line)
		}
	}
	newContent := strings.Join(kept, "\n")
	if err := global.GVA_DB.Model(&d).Updates(map[string]any{
		"content": newContent, "sample_count": valid, "purpose": purpose,
	}).Error; err != nil {
		return 0, errs, err
	}
	return valid, errs, nil
}

// ExportDataset 导出内容（下载面）
func (s *LlmopsServiceStruct) ExportDataset(id uint) (string, string, error) {
	var d model.LlmopsDataset
	if err := global.GVA_DB.First(&d, id).Error; err != nil {
		return "", "", fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
	}
	if strings.TrimSpace(d.Content) == "" {
		return "", "", fmt.Errorf("[%d]数据集内容为空", ErrCodeDatasetBad)
	}
	filename := fmt.Sprintf("%s-%s.%s", d.Name, d.Version, d.Format)
	return d.Content, filename, nil
}

// DatasetImportIn 数据集导入请求
type DatasetImportIn struct {
	ID      uint   `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
	Purpose string `json:"purpose"`
}
