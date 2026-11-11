// Package service eduadmin：办学预警（需求 3.6-25）
// 口径：指标阈值判定（支持 越高越好/越低越好 双向：越高越好时 value<threshold 触发，
// 越低越好时 value>threshold 触发）；<阈值 90% 为高、其余为中；预警按 指标+年度
// 幂等；招生策略优化报告=规则模板（专业结构/宣传投放/专项申请）。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduRunParam = 2826

// runLevel 指标等级（纯函数，测试锚点）：空=正常/中/高
func runLevel(value, threshold float64, higherBetter bool) string {
	if threshold <= 0 {
		return ""
	}
	over := value < threshold
	if !higherBetter {
		over = value > threshold
	}
	if !over {
		return ""
	}
	var ratio float64
	if higherBetter {
		if threshold > 0 {
			ratio = value / threshold
		}
	} else {
		if value > 0 {
			ratio = threshold / value
		}
	}
	if ratio < 0.9 {
		return "高"
	}
	return "中"
}

// admissionStrategy 招生策略优化报告（纯函数，测试锚点）
func admissionStrategy(alerts []model.EduRunAlert) string {
	var sb strings.Builder
	if len(alerts) == 0 {
		return "各项办学指标达标：维持现行招生结构，持续跟踪生源质量与就业去向。"
	}
	sb.WriteString(fmt.Sprintf("本年度共 %d 项办学指标触发预警，策略建议如下：\n", len(alerts)))
	for _, a := range alerts {
		switch {
		case strings.Contains(a.IndexName, "就业"):
			sb.WriteString(fmt.Sprintf("·【%s】就业率 %.0f%% 低于阈值：缩减低就业率专业招生计划，增设订单班与产教融合项目。\n", a.IndexName, a.Value*100))
		case strings.Contains(a.IndexName, "招生"):
			sb.WriteString(fmt.Sprintf("·【%s】招生完成率 %.0f%% 低于阈值：加大生源基地宣传投放，优化平行志愿梯度与调剂预案。\n", a.IndexName, a.Value*100))
		case strings.Contains(a.IndexName, "经费"):
			sb.WriteString(fmt.Sprintf("·【%s】经费到位率 %.0f%% 低于阈值：申请财政专项补足，优先保障教学一线支出。\n", a.IndexName, a.Value*100))
		default:
			sb.WriteString(fmt.Sprintf("·【%s】%.2f 低于阈值 %.2f：制定专项整改计划并按月跟踪。\n", a.IndexName, a.Value, a.Threshold))
		}
	}
	sb.WriteString("以上建议按季度复评执行效果。")
	return sb.String()
}

// ScanEduRunWarning 办学预警扫描（幂等落库）
func (s *EduadminService) ScanEduRunWarning() ([]model.EduRunAlert, string, error) {
	var indexes []model.EduRunIndex
	if err := global.GVA_DB.Order("id ASC").Limit(500).Find(&indexes).Error; err != nil {
		return nil, "", err
	}
	if len(indexes) == 0 {
		return nil, "", fmt.Errorf("[%d]无办学指标数据", ErrCodeEduRunParam)
	}
	created := make([]model.EduRunAlert, 0)
	for _, idx := range indexes {
		level := runLevel(idx.Value, idx.Threshold, idx.HigherBetter)
		if level == "" {
			continue
		}
		var exist model.EduRunAlert
		if err := global.GVA_DB.Where("index_name = ? AND year = ?", idx.Name, idx.Year).First(&exist).Error; err == nil {
			continue
		}
		alert := model.EduRunAlert{
			IndexName: idx.Name, Year: idx.Year, Level: level,
			Value: idx.Value, Threshold: idx.Threshold,
			Advice: envAdviceForRun(idx, level),
		}
		if err := global.GVA_DB.Create(&alert).Error; err != nil {
			return nil, "", err
		}
		created = append(created, alert)
	}
	return created, admissionStrategy(created), nil
}

// envAdviceForRun 单指标建议（模板）
func envAdviceForRun(idx model.EduRunIndex, level string) string {
	if strings.Contains(idx.Name, "就业") {
		return "缩减低就业率专业计划，增设订单班"
	}
	if strings.Contains(idx.Name, "招生") {
		return "加大宣传投放，优化志愿梯度"
	}
	if strings.Contains(idx.Name, "经费") {
		return "申请专项补足，保障教学一线"
	}
	return fmt.Sprintf("低于阈值，制定专项整改计划（%s）", level)
}

// SaveEduRunIndex 新建/更新指标
func (s *EduadminService) SaveEduRunIndex(in *model.EduRunIndex) (*model.EduRunIndex, error) {
	if strings.TrimSpace(in.Name) == "" || in.Year < 2000 || in.Threshold <= 0 {
		return nil, fmt.Errorf("[%d]指标名/年度/阈值必填", ErrCodeEduRunParam)
	}
	var exist model.EduRunIndex
	if err := global.GVA_DB.Where("name = ? AND year = ?", in.Name, in.Year).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"value": in.Value, "threshold": in.Threshold, "higher_better": in.HigherBetter,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduRunIndexes 指标列表
func (s *EduadminService) GetEduRunIndexes(year int) ([]model.EduRunIndex, error) {
	db := global.GVA_DB.Model(&model.EduRunIndex{})
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var list []model.EduRunIndex
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// GetEduRunAlerts 预警列表
func (s *EduadminService) GetEduRunAlerts(year int) ([]model.EduRunAlert, error) {
	db := global.GVA_DB.Model(&model.EduRunAlert{})
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var list []model.EduRunAlert
	err := db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}
