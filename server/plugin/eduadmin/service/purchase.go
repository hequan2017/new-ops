// Package service eduadmin：智能采购（需求 3.6-19）
// 口径：历史按月聚合→3 期移动平均预测下期需求量（复用 talent.go movingAverage）；
// 采购建议=量（预测，下限 0）/时点（历史峰值月提前 1 个月）/预算（预测量×最近单价），
// 纯函数可测。
package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduPurchaseParam = 2818

// purchaseSuggestion 采购建议
type purchaseSuggestion struct {
	Item         string   `json:"item"`
	History      []float64 `json:"history"` // 按月升序数量
	ForecastQty  float64  `json:"forecastQty"`
	OrderMonth   int      `json:"orderMonth"`  // 建议下单月份
	UnitPrice    float64  `json:"unitPrice"`
	Budget       float64  `json:"budget"`
	Summary      string   `json:"summary"`
}

// aggregateMonthly 按月聚合数量（纯函数，测试锚点）：升序返回逐月总量
func aggregateMonthly(list []model.EduPurchase) ([]int, []float64) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Year != list[j].Year {
			return list[i].Year < list[j].Year
		}
		return list[i].Month < list[j].Month
	})
	months := make([]int, 0, len(list))
	qtys := make([]float64, 0, len(list))
	for _, p := range list {
		m := p.Year*12 + p.Month
		if len(months) > 0 && months[len(months)-1] == m {
			qtys[len(qtys)-1] += p.Qty
			continue
		}
		months = append(months, m)
		qtys = append(qtys, p.Qty)
	}
	return months, qtys
}

// suggestPurchase 采购建议（纯函数，测试锚点）：窗口取 min(3, len/2) 的移动平均
func suggestPurchase(item string, list []model.EduPurchase) (*purchaseSuggestion, error) {
	months, qtys := aggregateMonthly(list)
	if len(qtys) == 0 {
		return nil, fmt.Errorf("[%d]该品目无采购历史", ErrCodeEduPurchaseParam)
	}
	_ = months
	window := 3
	if window > len(qtys) {
		window = len(qtys)
	}
	ma := movingAverage(qtys, window)
	forecast := qtys[len(qtys)-1]
	if len(ma) > 0 {
		forecast = ma[len(ma)-1]
	}
	if forecast < 0 {
		forecast = 0
	}
	forecast = math.Round(forecast*100) / 100
	// 最近单价：按年月取最新一条
	recent := make([]model.EduPurchase, len(list))
	copy(recent, list)
	sort.Slice(recent, func(i, j int) bool {
		if recent[i].Year != recent[j].Year {
			return recent[i].Year > recent[j].Year
		}
		return recent[i].Month > recent[j].Month
	})
	unitPrice := recent[0].UnitPrice
	// 历史峰值月：取历史出现最多的采购月
	countByMonth := map[int]int{}
	for _, p := range list {
		countByMonth[p.Month]++
	}
	peakMonth := 1
	maxCount := 0
	for m, c := range countByMonth {
		if c > maxCount || (c == maxCount && m < peakMonth) {
			maxCount = c
			peakMonth = m
		}
	}
	orderMonth := peakMonth - 1
	if orderMonth < 1 {
		orderMonth = 12
	}
	budget := math.Round(forecast*unitPrice*100) / 100
	return &purchaseSuggestion{
		Item: item, History: qtys, ForecastQty: forecast, OrderMonth: orderMonth,
		UnitPrice: unitPrice, Budget: budget,
		Summary: fmt.Sprintf("建议 %d 月下单 %g 件（近 %d 期移动平均），按最近单价 %g 元/件，预算约 %g 元",
			orderMonth, forecast, window, unitPrice, budget),
	}, nil
}

// SuggestEduPurchase 智能采购建议（DB 编排）
func (s *EduadminService) SuggestEduPurchase(item string) (*purchaseSuggestion, error) {
	if strings.TrimSpace(item) == "" {
		return nil, fmt.Errorf("[%d]品目必填", ErrCodeEduPurchaseParam)
	}
	var list []model.EduPurchase
	if err := global.GVA_DB.Where("item = ?", item).Order("year ASC, month ASC").Limit(500).Find(&list).Error; err != nil {
		return nil, err
	}
	return suggestPurchase(item, list)
}

// SaveEduPurchase 新增采购记录
func (s *EduadminService) SaveEduPurchase(in *model.EduPurchase) (*model.EduPurchase, error) {
	if strings.TrimSpace(in.Item) == "" || in.Month < 1 || in.Month > 12 || in.Year < 2000 {
		return nil, fmt.Errorf("[%d]品目与合法年月必填", ErrCodeEduPurchaseParam)
	}
	if in.Qty <= 0 || in.UnitPrice < 0 {
		return nil, fmt.Errorf("[%d]数量须为正数且单价非负", ErrCodeEduPurchaseParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduPurchases 采购记录列表（品目过滤，年月升序）
func (s *EduadminService) GetEduPurchases(item string) ([]model.EduPurchase, error) {
	db := global.GVA_DB.Model(&model.EduPurchase{})
	if item != "" {
		db = db.Where("item = ?", item)
	}
	var list []model.EduPurchase
	err := db.Order("year ASC, month ASC").Limit(500).Find(&list).Error
	return list, err
}
