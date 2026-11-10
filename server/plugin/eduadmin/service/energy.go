// Package service eduadmin：能耗优化（需求 3.6-20）
// 口径：楼栋×时段聚合（白天/夜间）；同比异常=同月去年对比涨幅>20% 判异常；
// 节能建议=规则模板（异常楼栋专项诊断/夜间占比>30% 夜间管控/最大耗能楼栋优先改造）。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduEnergyParam = 2819

// buildingUsage 楼栋聚合
type buildingUsage struct {
	Building  string  `json:"building"`
	DayKwh    float64 `json:"dayKwh"`
	NightKwh  float64 `json:"nightKwh"`
	TotalKwh  float64 `json:"totalKwh"`
	NightRate float64 `json:"nightRate"`
	YoYGrowth float64 `json:"yoYGrowth"` // 同比涨幅（无去年记 0）
	Abnormal  bool    `json:"abnormal"`
}

// aggregateEnergy 楼栋聚合（纯函数，测试锚点）：今年值+夜间占比+同比
func aggregateEnergy(current, lastYear []model.EduEnergyRecord) []buildingUsage {
	cur := map[string]*buildingUsage{}
	for _, r := range current {
		b, ok := cur[r.Building]
		if !ok {
			b = &buildingUsage{Building: r.Building}
			cur[r.Building] = b
		}
		if r.Period == "夜间" {
			b.NightKwh += r.Kwh
		} else {
			b.DayKwh += r.Kwh
		}
		b.TotalKwh += r.Kwh
	}
	last := map[string]float64{}
	for _, r := range lastYear {
		last[r.Building] += r.Kwh
	}
	out := make([]buildingUsage, 0, len(cur))
	for _, b := range cur {
		if b.TotalKwh > 0 {
			b.NightRate = round2(b.NightKwh / b.TotalKwh)
		}
		if last[b.Building] > 0 {
			b.YoYGrowth = round2((b.TotalKwh - last[b.Building]) / last[b.Building])
		}
		b.Abnormal = b.YoYGrowth > 0.2
		out = append(out, *b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TotalKwh > out[j].TotalKwh })
	return out
}

// energySuggestions 节能建议模板（纯函数，测试锚点）
func energySuggestions(usages []buildingUsage) []string {
	out := make([]string, 0)
	for _, b := range usages {
		if b.Abnormal {
			out = append(out, fmt.Sprintf("【%s】同比上涨 %.0f%%：开展专项能耗诊断（空调设定/照明/待机设备），必要时加装分项计量", b.Building, b.YoYGrowth*100))
		}
		if b.NightRate > 0.3 {
			out = append(out, fmt.Sprintf("【%s】夜间用电占比 %.0f%%：实施夜间巡查与分时管控，检查实验设备通宵运行", b.Building, b.NightRate*100))
		}
	}
	if len(usages) > 0 && !usages[0].Abnormal {
		out = append(out, fmt.Sprintf("【%s】为当前最大耗能楼栋：优先纳入节能改造（LED 照明/变频空调）", usages[0].Building))
	}
	if len(out) == 0 {
		out = append(out, "各楼栋能耗平稳：维持既有节能措施")
	}
	return out
}

// AnalyzeEduEnergy 能耗分析（当年某月 vs 去年同月）
func (s *EduadminService) AnalyzeEduEnergy(year, month int) ([]buildingUsage, []string, error) {
	if month < 1 || month > 12 || year < 2000 {
		return nil, nil, fmt.Errorf("[%d]合法年月必填", ErrCodeEduEnergyParam)
	}
	var current, last []model.EduEnergyRecord
	if err := global.GVA_DB.Where("year = ? AND month = ?", year, month).Limit(2000).Find(&current).Error; err != nil {
		return nil, nil, err
	}
	if len(current) == 0 {
		return nil, nil, fmt.Errorf("[%d]该月无能耗数据", ErrCodeEduEnergyParam)
	}
	global.GVA_DB.Where("year = ? AND month = ?", year-1, month).Limit(2000).Find(&last)
	usages := aggregateEnergy(current, last)
	return usages, energySuggestions(usages), nil
}

// SaveEduEnergyRecord 新增能耗记录
func (s *EduadminService) SaveEduEnergyRecord(in *model.EduEnergyRecord) (*model.EduEnergyRecord, error) {
	if strings.TrimSpace(in.Building) == "" || in.Month < 1 || in.Month > 12 || in.Year < 2000 {
		return nil, fmt.Errorf("[%d]楼栋与合法年月必填", ErrCodeEduEnergyParam)
	}
	if in.Period != "白天" && in.Period != "夜间" {
		return nil, fmt.Errorf("[%d]时段须为 白天/夜间", ErrCodeEduEnergyParam)
	}
	if in.Kwh < 0 {
		return nil, fmt.Errorf("[%d]电量不可为负", ErrCodeEduEnergyParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduEnergyRecords 能耗记录列表（楼栋过滤）
func (s *EduadminService) GetEduEnergyRecords(building string) ([]model.EduEnergyRecord, error) {
	db := global.GVA_DB.Model(&model.EduEnergyRecord{})
	if building != "" {
		db = db.Where("building = ?", building)
	}
	var list []model.EduEnergyRecord
	err := db.Order("year DESC, month DESC").Limit(500).Find(&list).Error
	return list, err
}
