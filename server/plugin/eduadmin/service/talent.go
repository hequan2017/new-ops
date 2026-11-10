// Package service eduadmin：人才预测（需求 3.6-14）
// 口径：历史时序线性回归（正规方程 slope/intercept）预测人才需求（预测值下限 0）；
// 流失率=离职/期初在职，3 年移动平均；培训计划建议=规则模板（流失率/增长率阈值），
// 全部纯函数可测，无第三方依赖。
package service

import (
	"fmt"
	"math"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduTalentParam = 2811

// linearRegression 一元线性回归正规方程（纯函数，测试锚点）：y = slope*x + intercept
func linearRegression(x, y []float64) (slope, intercept float64, err error) {
	if len(x) != len(y) || len(x) < 2 {
		return 0, 0, fmt.Errorf("[%d]时序样本须 ≥2 且等长", ErrCodeEduTalentParam)
	}
	n := float64(len(x))
	var sx, sy, sxx, sxy float64
	for i := range x {
		sx += x[i]
		sy += y[i]
		sxx += x[i] * x[i]
		sxy += x[i] * y[i]
	}
	denom := n*sxx - sx*sx
	if denom == 0 {
		return 0, sy / n, nil // 全部 x 相同：退化为均值
	}
	slope = (n*sxy - sx*sy) / denom
	intercept = (sy - slope*sx) / n
	return slope, intercept, nil
}

// movingAverage 简单移动平均（纯函数，测试锚点）：窗口不足返回空
func movingAverage(vals []float64, window int) []float64 {
	if window < 1 || len(vals) < window {
		return nil
	}
	out := make([]float64, 0, len(vals)-window+1)
	for i := 0; i+window <= len(vals); i++ {
		sum := 0.0
		for _, v := range vals[i : i+window] {
			sum += v
		}
		out = append(out, round2(sum/float64(window)))
	}
	return out
}

// TalentForecast 人才预测报告
type TalentForecast struct {
	Years        int      `json:"years"`
	Slope        float64  `json:"slope"`
	Intercept    float64  `json:"intercept"`
	HistoryYears []int    `json:"historyYears"`
	HistoryHead  []int    `json:"historyHeadcount"`
	Forecast     []int    `json:"forecast"` // 未来 years 年需求预测
	TurnoverMA   []float64 `json:"turnoverMA3"`
	Suggestion   []string `json:"suggestion"`
}

// ForecastTalent 人才需求预测 + 流失率移动平均 + 培训计划建议
func (s *EduadminService) ForecastTalent(years int) (*TalentForecast, error) {
	if years < 1 || years > 10 {
		years = 3
	}
	var stats []model.EduTalentStat
	if err := global.GVA_DB.Order("year ASC").Limit(50).Find(&stats).Error; err != nil {
		return nil, err
	}
	if len(stats) < 2 {
		return nil, fmt.Errorf("[%d]历史年度数据须 ≥2 条方可回归", ErrCodeEduTalentParam)
	}
	xs := make([]float64, len(stats))
	ys := make([]float64, len(stats))
	turnover := make([]float64, 0, len(stats))
	prevHead := 0.0
	for i, st := range stats {
		xs[i] = float64(st.Year)
		ys[i] = float64(st.Headcount)
		if prevHead > 0 {
			turnover = append(turnover, float64(st.Departures)/prevHead)
		}
		prevHead = float64(st.Headcount)
	}
	slope, intercept, err := linearRegression(xs, ys)
	if err != nil {
		return nil, err
	}
	lastYear := stats[len(stats)-1].Year
	forecast := make([]int, 0, years)
	for k := 1; k <= years; k++ {
		p := slope*float64(lastYear+k) + intercept
		if p < 0 {
			p = 0
		}
		forecast = append(forecast, int(math.Round(p)))
	}
	ma := movingAverage(turnover, 3)
	lastMA := 0.0
	if len(ma) > 0 {
		lastMA = ma[len(ma)-1]
	}
	growth := 0.0
	if ys[0] > 0 {
		growth = (ys[len(ys)-1] - ys[0]) / ys[0]
	}
	suggestion := talentSuggestions(lastMA, growth, forecast[len(forecast)-1], stats[len(stats)-1].Headcount)
	return &TalentForecast{
		Years: years, Slope: round2(slope), Intercept: round2(intercept),
		HistoryYears: func() []int {
			out := make([]int, len(stats))
			for i := range stats {
				out[i] = stats[i].Year
			}
			return out
		}(),
		HistoryHead: func() []int {
			out := make([]int, len(stats))
			for i := range stats {
				out[i] = stats[i].Headcount
			}
			return out
		}(),
		Forecast:   forecast,
		TurnoverMA: ma,
		Suggestion: suggestion,
	}, nil
}

// talentSuggestions 培训计划建议规则（纯函数，测试锚点）
func talentSuggestions(turnoverMA, growth float64, forecastHead, currentHead int) []string {
	out := make([]string, 0)
	if turnoverMA >= 0.15 {
		out = append(out, fmt.Sprintf("三年流动率均值 %.1f%% 偏高：加强新人导师制与骨干保留培训，梳理薪酬对标", turnoverMA*100))
	} else if turnoverMA > 0 {
		out = append(out, fmt.Sprintf("流动率 %.1f%% 处于正常区间：维持现有培训节奏", turnoverMA*100))
	}
	delta := forecastHead - currentHead
	if delta > 0 {
		out = append(out, fmt.Sprintf("预测需求净增 %d 人：提前启动招聘与岗前培训计划（建议分两批到岗）", delta))
	} else if delta < 0 {
		out = append(out, fmt.Sprintf("预测需求净减 %d 人：控制外聘节奏，转岗培训优先", -delta))
	}
	if growth >= 0.1 {
		out = append(out, "在职规模五年增长超 10%：增设管理梯队与数字化能力专项培训")
	}
	if len(out) == 0 {
		out = append(out, "各项指标平稳：按年度常规培训计划执行")
	}
	return out
}

// SaveEduTalentStat 新建/更新年度统计
func (s *EduadminService) SaveEduTalentStat(in *model.EduTalentStat) (*model.EduTalentStat, error) {
	if in.Year < 2000 || in.Year > 2100 {
		return nil, fmt.Errorf("[%d]年度须在 2000-2100", ErrCodeEduTalentParam)
	}
	var exist model.EduTalentStat
	if err := global.GVA_DB.Where("year = ?", in.Year).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"new_hires": in.NewHires, "departures": in.Departures, "headcount": in.Headcount,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduTalentStats 年度统计列表（升序）
func (s *EduadminService) GetEduTalentStats() ([]model.EduTalentStat, error) {
	var list []model.EduTalentStat
	err := global.GVA_DB.Order("year ASC").Limit(100).Find(&list).Error
	return list, err
}
