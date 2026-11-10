// Package service eduadmin：数据驾驶舱（需求 3.6-24）
// 口径：四域（人事/财务/资产/后勤）直接计数+来源分布；趋势=近 7 日记录按日聚合
// 缺口日补零（对齐 llmops L01 口径）；钻取=按月份/来源/域过滤下钻（按月聚合）。
package service

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduDashboardParam = 2825

// EduTrendPoint 趋势点
type EduTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// buildEduTrend 趋势补零（纯函数，测试锚点）
func buildEduTrend(rows []struct {
	Day string
	Cnt int64
}, days int) []EduTrendPoint {
	if days <= 0 {
		days = 7
	}
	cntMap := map[string]int64{}
	for _, r := range rows {
		cntMap[r.Day] = r.Cnt
	}
	today := time.Now()
	out := make([]EduTrendPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		out = append(out, EduTrendPoint{Date: d, Count: cntMap[d]})
	}
	return out
}

// DashboardStats 驾驶舱统计
type DashboardStats struct {
	DomainCounts map[string]int64   `json:"domainCounts"`
	Total        int64              `json:"total"`
	SourceCounts map[string]int64   `json:"sourceCounts"`
	Trend        []EduTrendPoint    `json:"trend"`
	MonthCounts  map[string]int64   `json:"monthCounts"` // 钻取：按月聚合（month 参数为空时仅近 6 个月）
}

// GetEduDashboard 驾驶舱聚合（钻取参数：domain/source/month）
func (s *EduadminService) GetEduDashboard(domain, source, month string) (*DashboardStats, error) {
	if domain != "" && !ValidEduDomain(domain) {
		return nil, fmt.Errorf("[%d]域须为人事/财务/资产/后勤", ErrCodeEduDashboardParam)
	}
	if month != "" {
		if _, err := time.ParseInLocation("2006-01", month, time.Local); err != nil {
			return nil, fmt.Errorf("[%d]月份格式须为 yyyy-MM", ErrCodeEduDashboardParam)
		}
	}
	st := &DashboardStats{
		DomainCounts: map[string]int64{},
		SourceCounts: map[string]int64{},
		MonthCounts:  map[string]int64{},
	}
	db := global.GVA_DB.Model(&model.EduRecord{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	if source != "" {
		db = db.Where("source LIKE ?", "%"+source+"%")
	}
	// 四域计数
	var rows []struct {
		Domain string
		Cnt    int64
	}
	if err := db.Select("domain, COUNT(*) as cnt").Where("deleted_at IS NULL").Group("domain").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		st.DomainCounts[r.Domain] = r.Cnt
		st.Total += r.Cnt
	}
	// 来源分布
	var srcRows []struct {
		Source string
		Cnt    int64
	}
	if err := db.Select("source, COUNT(*) as cnt").Where("deleted_at IS NULL").Group("source").Find(&srcRows).Error; err != nil {
		return nil, err
	}
	for _, r := range srcRows {
		st.SourceCounts[r.Source] = r.Cnt
	}
	// 趋势：有 month 钻取时按月聚合近 6 个月，否则近 7 日补零
	if month != "" {
		var monthRows []struct {
			Month string
			Cnt   int64
		}
		if err := db.Select("DATE_FORMAT(created_at, '%Y-%m') as month, COUNT(*) as cnt").
			Where("deleted_at IS NULL AND created_at >= ?", month+"-01").
			Group("DATE_FORMAT(created_at, '%Y-%m')").Find(&monthRows).Error; err != nil {
			return nil, err
		}
		base, _ := time.ParseInLocation("2006-01", month, time.Local)
		for i := 0; i < 6; i++ {
			key := base.AddDate(0, i, 0).Format("2006-01")
			var cnt int64
			for _, r := range monthRows {
				if r.Month == key {
					cnt = r.Cnt
				}
			}
			st.MonthCounts[key] = cnt
		}
	} else {
		var trendRows []struct {
			Day string
			Cnt int64
		}
		weekAgo := time.Now().AddDate(0, 0, -6)
		if err := db.Select("DATE(created_at) as day, COUNT(*) as cnt").
			Where("deleted_at IS NULL AND created_at > ?", weekAgo).
			Group("DATE(created_at)").Find(&trendRows).Error; err != nil {
			return nil, err
		}
		st.Trend = buildEduTrend(trendRows, 7)
	}
	return st, nil
}

// timeNowDash 测试辅助（趋势日期基准）
func timeNowDash() time.Time { return time.Now() }
