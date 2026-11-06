// Package service llmops：多模型调度（按能力选模型：DefaultFor 优先，其余按近 N 次成功调用时延均值）
package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 调度口径常量
const (
	schedulerLatencyWindow = 10 // 近 N 次成功调用统计窗口
)

// PickByLatency 时延择优（纯函数，测试锚点）：无调用记录视为 0 优先；同分取 ID 小者
func PickByLatency(candidates []model.LlmopsModel, avgLatency map[uint]int64) uint {
	best := uint(0)
	bestScore := int64(-1)
	for _, m := range candidates {
		score := avgLatency[m.ID] // 无记录 → 0
		if bestScore < 0 || score < bestScore || (score == bestScore && m.ID < best) {
			best, bestScore = m.ID, score
		}
	}
	return best
}

// AvgLatencies 近 N 次成功调用时延均值（毫秒；无记录不计入返回 map）
func (s *LlmopsServiceStruct) AvgLatencies(modelIds []uint) map[uint]int64 {
	out := make(map[uint]int64, len(modelIds))
	if len(modelIds) == 0 {
		return out
	}
	// 取每模型近 N 次 ok 记录的均值：先查窗口内记录再聚合（兼容无窗口函数的 MySQL 版本）
	var logs []model.LlmopsModelCallLog
	if err := global.GVA_DB.
		Where("model_id IN ? AND ok = ?", modelIds, true).
		Order("id desc").
		Find(&logs).Error; err != nil {
		return out
	}
	sum := make(map[uint]int64, len(modelIds))
	cnt := make(map[uint]int64, len(modelIds))
	for _, l := range logs { // 倒序遍历，每模型计满窗口即止
		if cnt[l.ModelId] >= schedulerLatencyWindow {
			continue
		}
		sum[l.ModelId] += l.LatencyMs
		cnt[l.ModelId]++
	}
	for id, c := range cnt {
		if c > 0 {
			out[id] = sum[id] / c
		}
	}
	return out
}

// SelectModel 按能力调度模型：DefaultFor 优先 → 其余按近 N 次成功调用时延均值择优
func (s *LlmopsServiceStruct) SelectModel(capability string) (*model.LlmopsModel, error) {
	if !ValidCapability(capability) {
		return nil, fmt.Errorf("[%d]不支持的能力标签: %s", ErrCodeModelCapBad, capability)
	}
	var candidates []model.LlmopsModel
	err := global.GVA_DB.Where("status = 'enabled'").
		Where("capability LIKE ?", "%"+capability+"%").
		Order("id").Find(&candidates).Error
	if err != nil {
		return nil, err
	}
	// LIKE 收敛后再精确校验能力标签（避免 chat 误匹配 chats 等）
	var exact []model.LlmopsModel
	for _, m := range candidates {
		if modelHasCapability(&m, capability) {
			exact = append(exact, m)
		}
	}
	if len(exact) == 0 {
		return nil, fmt.Errorf("[%d]无可用 %s 模型，请先注册", ErrCodeGatewayCap, capability)
	}
	// DefaultFor 优先
	var ids []uint
	for _, m := range exact {
		if m.DefaultFor == capability {
			return &m, nil
		}
		ids = append(ids, m.ID)
	}
	latency := s.AvgLatencies(ids)
	bestId := PickByLatency(exact, latency)
	for _, m := range exact {
		if m.ID == bestId {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("[%d]无可用 %s 模型", ErrCodeGatewayCap, capability)
}
