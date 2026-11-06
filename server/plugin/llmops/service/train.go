// Package service llmops：训练与调优任务（要求(4)）——任务 CRUD + 状态机
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeTrainNotFound  = 2252
	ErrCodeTrainTypeBad   = 2253
	ErrCodeTrainTrans     = 2254
	ErrCodeTrainNameDup   = 2255
	ErrCodeTrainHyper     = 2256
)

// TrainTypes 任务类型白名单
var TrainTypes = []string{"sft_full", "sft_lora", "reward", "rl", "eval", "compress"}

// ValidTrainType 任务类型白名单
func ValidTrainType(t string) bool {
	for _, v := range TrainTypes {
		if v == t {
			return true
		}
	}
	return false
}

// trainTransitions 状态机合法流转
var trainTransitions = map[string][]string{
	"draft":    {"queued", "failed"}, // queued 提交；failed 表示直接废弃置败
	"queued":   {"running", "failed"},
	"running":  {"succeeded", "failed", "stopped"},
	"stopped":  {"queued"},
	"failed":   {"queued"}, // 可重新排队
	"succeeded": {},
}

// CanTransition 状态机流转校验（纯函数）
func CanTransition(from, to string) bool {
	for _, t := range trainTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// ValidateHyperJSON 超参 JSON 合法性（须为对象）
func ValidateHyperJSON(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return fmt.Errorf("[%d]超参须为 JSON 对象: %w", ErrCodeTrainHyper, err)
	}
	return nil
}

// TrainIn 训练任务保存入参
type TrainIn struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	DatasetId uint   `json:"datasetId"`
	BaseModel string `json:"baseModel"`
	HyperJSON string `json:"hyperJSON"`
	PreJobId  uint   `json:"preJobId"`
}

// GetTrainJobs 任务列表（类型/状态过滤，分页）
func (s *LlmopsServiceStruct) GetTrainJobs(page, pageSize int, jobType, status string) ([]model.LlmopsTrainJob, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	db := global.GVA_DB.Model(&model.LlmopsTrainJob{})
	if jobType != "" {
		if !ValidTrainType(jobType) {
			return nil, 0, fmt.Errorf("[%d]不支持的任务类型: %s", ErrCodeTrainTypeBad, jobType)
		}
		db = db.Where("type = ?", jobType)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsTrainJob
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveTrainJob 新建/更新任务（草稿态可改；名称唯一；超参即检）
func (s *LlmopsServiceStruct) SaveTrainJob(in *TrainIn) (*model.LlmopsTrainJob, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidTrainType(in.Type) {
		return nil, fmt.Errorf("[%d]名称必填且类型须为 sft_full/sft_lora/reward/rl/eval/compress", ErrCodeTrainTypeBad)
	}
	if err := ValidateHyperJSON(in.HyperJSON); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsTrainJob{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]任务名已存在", ErrCodeTrainNameDup)
	}
	if in.ID > 0 {
		var job model.LlmopsTrainJob
		if err := global.GVA_DB.First(&job, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
		}
		if job.Status != "draft" {
			return nil, fmt.Errorf("[%d]仅草稿态可编辑（当前 %s）", ErrCodeTrainTrans, job.Status)
		}
		err := global.GVA_DB.Model(&job).Updates(map[string]any{
			"name": in.Name, "type": in.Type, "dataset_id": in.DatasetId,
			"base_model": in.BaseModel, "hyper_json": in.HyperJSON, "pre_job_id": in.PreJobId,
		}).Error
		if err != nil {
			return nil, err
		}
		return &job, nil
	}
	job := &model.LlmopsTrainJob{
		Name: in.Name, Type: in.Type, DatasetId: in.DatasetId,
		BaseModel: in.BaseModel, HyperJSON: in.HyperJSON, PreJobId: in.PreJobId, Status: "draft",
	}
	if err := global.GVA_DB.Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

// TransitionTrainJob 状态流转（状态机校验 + 失败/停止写原因）
func (s *LlmopsServiceStruct) TransitionTrainJob(id uint, to, reason string) (*model.LlmopsTrainJob, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	if !CanTransition(job.Status, to) {
		return nil, fmt.Errorf("[%d]非法流转 %s → %s", ErrCodeTrainTrans, job.Status, to)
	}
	updates := map[string]any{"status": to}
	if reason != "" {
		updates["err_msg"] = truncateSample(reason)
	}
	if err := global.GVA_DB.Model(&job).Updates(updates).Error; err != nil {
		return nil, err
	}
	job.Status = to
	return &job, nil
}

// DeleteTrainJob 删除任务（仅终态/草稿可删）
func (s *LlmopsServiceStruct) DeleteTrainJob(id uint) error {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, id).Error; err != nil {
		return fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	if job.Status == "running" || job.Status == "queued" {
		return fmt.Errorf("[%d]运行/排队中任务不可删除，请先停止", ErrCodeTrainTrans)
	}
	return global.GVA_DB.Delete(&job).Error
}
