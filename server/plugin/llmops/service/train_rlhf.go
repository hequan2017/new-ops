// Package service llmops：RLHF 两阶段编排（要求(4)）——奖励模型训练 → 强化学习训练的链路依赖
// 口径：rl 阶段必须挂 preJobId 且指向已成功的 reward 任务（奖励模型产物即其输出目录）；
// reward 阶段前置可选 SFT 产物；类型依赖校验为纯函数锚点。
package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeRlhfChain = 2258

// ValidateRlhfTypes 两阶段类型依赖校验（纯函数，测试锚点）
// jobType=rl: 必须有前置且为 reward+succeeded；jobType=reward: 前置可选且须为 sft 类
func ValidateRlhfTypes(jobType, preType, preStatus string, hasPre bool) error {
	switch jobType {
	case "rl":
		if !hasPre {
			return fmt.Errorf("[%d]RL 阶段必须挂前置奖励模型任务(preJobId)", ErrCodeRlhfChain)
		}
		if preType != "reward" {
			return fmt.Errorf("[%d]RL 前置任务须为 reward 类型（当前 %s）", ErrCodeRlhfChain, preType)
		}
		if preStatus != "succeeded" {
			return fmt.Errorf("[%d]前置奖励任务未完成（状态 %s），不可启动 RL", ErrCodeRlhfChain, preStatus)
		}
	case "reward":
		if hasPre && preType != "sft_full" && preType != "sft_lora" {
			return fmt.Errorf("[%d]reward 前置须为 SFT 任务（当前 %s）", ErrCodeRlhfChain, preType)
		}
	}
	return nil
}

// ResolveRlhfPreJob 解析 RL 前置奖励任务（存在性+类型+状态三重校验，返回前置任务）
func (s *LlmopsServiceStruct) ResolveRlhfPreJob(jobId, preJobId uint) (*model.LlmopsTrainJob, error) {
	if preJobId == 0 {
		return nil, fmt.Errorf("[%d]RL 阶段必须挂前置奖励模型任务(preJobId)", ErrCodeRlhfChain)
	}
	if preJobId == jobId {
		return nil, fmt.Errorf("[%d]前置任务不可为自身", ErrCodeRlhfChain)
	}
	var pre model.LlmopsTrainJob
	if err := global.GVA_DB.First(&pre, preJobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]前置奖励任务不存在", ErrCodeRlhfChain)
	}
	if err := ValidateRlhfTypes("rl", pre.Type, pre.Status, true); err != nil {
		return nil, err
	}
	return &pre, nil
}
