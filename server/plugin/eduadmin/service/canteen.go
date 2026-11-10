// Package service eduadmin：食堂优化（需求 3.6-21）
// 口径：一卡通消费按 weekday 均值预测用餐人次（纯函数）；建议规则=
// 预测量×1.15 备货畅销档口、峰值餐段提示加开窗口、低谷档口推新品/合并。
package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduMealParam = 2820

// mealForecast 预测结果
type mealForecast struct {
	Date        string  `json:"date"`
	Weekday     string  `json:"weekday"`
	ForecastCnt float64 `json:"forecastCnt"` // 该 weekday 历史日均值
	TopStall    string  `json:"topStall"`
	TopMealTime string  `json:"topMealTime"`
	Suggestions []string `json:"suggestions"`
}

// weekdayOf 日期→星期（纯函数）
func weekdayOf(date string) (time.Weekday, error) {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return 0, fmt.Errorf("[%d]日期格式须为 yyyy-mm-dd", ErrCodeEduMealParam)
	}
	return t.Weekday(), nil
}

// weekdayMean 同星期日均用餐人次（纯函数，测试锚点）：按日期聚合总量再按星期平均
func weekdayMean(records []model.EduMealRecord, target time.Weekday) float64 {
	byDate := map[string]int{}
	for _, r := range records {
		byDate[r.Date] += r.Count
	}
	var sum float64
	var n float64
	for d, cnt := range byDate {
		wd, err := weekdayOf(d)
		if err != nil || wd != target {
			continue
		}
		sum += float64(cnt)
		n++
	}
	if n == 0 {
		return 0
	}
	return math.Round(sum/n*100) / 100
}

// mealSuggestions 食材与菜品建议（纯函数，测试锚点）
func mealSuggestions(records []model.EduMealRecord, forecastCnt float64) []string {
	stallSum := map[string]int{}
	mealSum := map[string]int{}
	for _, r := range records {
		stallSum[r.Stall] += r.Count
		mealSum[r.MealTime] += r.Count
	}
	out := make([]string, 0)
	if forecastCnt > 0 {
		out = append(out, fmt.Sprintf("预计用餐 %.0f 人次：按 %.0f 人次备货（含 15%% 余量），重点保障畅销档口食材", forecastCnt, forecastCnt*1.15))
	}
	topMeal, topMealCnt := "", 0
	for m, c := range mealSum {
		if c > topMealCnt {
			topMeal, topMealCnt = m, c
		}
	}
	if topMeal != "" {
		out = append(out, fmt.Sprintf("【%s】为用餐高峰：建议加开 1 个窗口并错峰出餐", topMeal))
	}
	stalls := make([]string, 0, len(stallSum))
	for s := range stallSum {
		stalls = append(stalls, s)
	}
	sort.Strings(stalls)
	lowStall, lowCnt := "", 1<<30
	for _, s := range stalls {
		if stallSum[s] < lowCnt {
			lowStall, lowCnt = s, stallSum[s]
		}
	}
	if lowStall != "" && len(stallSum) > 1 {
		out = append(out, fmt.Sprintf("【%s】消费垫底：调整菜品结构（推新品/组合套餐）或并入相邻档口", lowStall))
	}
	return out
}

// ForecastEduMeals 用餐预测
func (s *EduadminService) ForecastEduMeals(date string) (*mealForecast, error) {
	wd, err := weekdayOf(date)
	if err != nil {
		return nil, err
	}
	var records []model.EduMealRecord
	if err := global.GVA_DB.Order("id ASC").Limit(5000).Find(&records).Error; err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("[%d]无一卡通消费数据", ErrCodeEduMealParam)
	}
	fc := &mealForecast{
		Date: date, Weekday: weekdayName(wd),
		ForecastCnt: weekdayMean(records, wd),
	}
	fc.Suggestions = mealSuggestions(records, fc.ForecastCnt)
	// 峰值档口
	stallSum := map[string]int{}
	for _, r := range records {
		stallSum[r.Stall] += r.Count
	}
	topStall, topCnt := "", 0
	for st, c := range stallSum {
		if c > topCnt {
			topStall, topCnt = st, c
		}
	}
	fc.TopStall = topStall
	mealSum := map[string]int{}
	for _, r := range records {
		mealSum[r.MealTime] += r.Count
	}
	topMeal, topMealCnt := "", 0
	for m, c := range mealSum {
		if c > topMealCnt {
			topMeal, topMealCnt = m, c
		}
	}
	fc.TopMealTime = topMeal
	return fc, nil
}

func weekdayName(w time.Weekday) string {
	names := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return names[int(w)]
}

// SaveEduMealRecord 新增消费记录
func (s *EduadminService) SaveEduMealRecord(in *model.EduMealRecord) (*model.EduMealRecord, error) {
	if _, err := weekdayOf(in.Date); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Stall) == "" {
		return nil, fmt.Errorf("[%d]档口必填", ErrCodeEduMealParam)
	}
	switch in.MealTime {
	case "早餐", "午餐", "晚餐":
	default:
		return nil, fmt.Errorf("[%d]餐段须为 早餐/午餐/晚餐", ErrCodeEduMealParam)
	}
	if in.Count < 0 || in.Amount < 0 {
		return nil, fmt.Errorf("[%d]人次与金额不可为负", ErrCodeEduMealParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduMealRecords 消费记录列表（日期过滤）
func (s *EduadminService) GetEduMealRecords(date string) ([]model.EduMealRecord, error) {
	db := global.GVA_DB.Model(&model.EduMealRecord{})
	if date != "" {
		db = db.Where("date = ?", date)
	}
	var list []model.EduMealRecord
	err := db.Order("date DESC").Limit(500).Find(&list).Error
	return list, err
}
