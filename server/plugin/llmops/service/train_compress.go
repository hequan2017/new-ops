// Package service llmops：模型压缩（要求(4)）——量化/剪枝/蒸馏渲染 + 产物登记
// 口径：compress 任务超参内含 compressSpec（method/bits/sparsity/teacher 等）；
// 本地 dry-run 仅渲染压缩命令（占位执行器，真实压缩后端待接入）；产物手工/后续自动登记。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeCompress = 2262

// CompressSpec 压缩规格
type CompressSpec struct {
	Method      string  `json:"method"`      // quantize/prune/distill
	Bits        int     `json:"bits"`        // quantize: 4/8
	GroupSize   int     `json:"groupSize"`   // quantize: 分组大小
	Sparsity    float64 `json:"sparsity"`    // prune: 0-1
	TeacherModel string `json:"teacherModel"` // distill: 教师模型
	Temperature float64 `json:"temperature"` // distill: 蒸馏温度
}

// CompressSub 超参内压缩规格载体（ParseHyper 忽略未知字段的补充解析入口）
type CompressSub struct {
	CompressSpec CompressSpec `json:"compressSpec"`
}

// ParseCompressSpec 压缩规格解析（纯函数，测试锚点）
func ParseCompressSpec(raw string) (*CompressSpec, error) {
	var cs CompressSub
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &cs); err != nil {
			return nil, fmt.Errorf("[%d]压缩规格解析失败: %w", ErrCodeCompress, err)
		}
	}
	spec := &cs.CompressSpec
	switch spec.Method {
	case "quantize":
		if spec.Bits != 4 && spec.Bits != 8 {
			spec.Bits = 4
		}
		if spec.GroupSize <= 0 {
			spec.GroupSize = 128
		}
	case "prune":
		if spec.Sparsity <= 0 || spec.Sparsity >= 1 {
			return nil, fmt.Errorf("[%d]剪枝稀疏度须在 (0,1): %v", ErrCodeCompress, spec.Sparsity)
		}
	case "distill":
		if err := sanitizeTrainArg(spec.TeacherModel); err != nil {
			return nil, err
		}
		if spec.Temperature <= 0 {
			spec.Temperature = 2.0
		}
	default:
		return nil, fmt.Errorf("[%d]不支持的压缩方法: %s", ErrCodeCompress, spec.Method)
	}
	return spec, nil
}

// RenderCompressCommand 压缩命令渲染（纯函数；模型路径/输出目录沿用超参口径）
func RenderCompressCommand(spec *CompressSpec, p *TrainHyperParams, basePath string) (string, error) {
	if err := sanitizeTrainArg(basePath); err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "python compress.py --method %s --model_path %s --output_dir %s", spec.Method, basePath, p.OutputDir)
	switch spec.Method {
	case "quantize":
		fmt.Fprintf(&sb, " --bits %d --group_size %d", spec.Bits, spec.GroupSize)
	case "prune":
		fmt.Fprintf(&sb, " --sparsity %g", spec.Sparsity)
	case "distill":
		fmt.Fprintf(&sb, " --teacher_model %s --temperature %g", spec.TeacherModel, spec.Temperature)
	}
	return sb.String(), nil
}

// RegisterArtifact 产物登记（路径必填 + 防注入）
func (s *LlmopsServiceStruct) RegisterArtifact(in *model.LlmopsArtifact) (*model.LlmopsArtifact, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, fmt.Errorf("[%d]产物名与路径必填", ErrCodeCompress)
	}
	if err := sanitizeTrainArg(in.Path); err != nil {
		return nil, err
	}
	if in.JobType == "compress" && in.Method != "" {
		if _, err := ParseCompressSpec(fmt.Sprintf(`{"compressSpec":{"method":%q}}`, in.Method)); err != nil {
			return nil, err
		}
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetArtifacts 产物列表
func (s *LlmopsServiceStruct) GetArtifacts() ([]model.LlmopsArtifact, error) {
	var list []model.LlmopsArtifact
	err := global.GVA_DB.Order("id desc").Find(&list).Error
	return list, err
}

// DeleteArtifact 产物删除
func (s *LlmopsServiceStruct) DeleteArtifact(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsArtifact{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]产物不存在", ErrCodeCompress)
	}
	return nil
}
