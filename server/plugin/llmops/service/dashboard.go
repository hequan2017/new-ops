// Package service llmops：数据看板（要求(8) 基础管控——各域统计聚合）
// 口径：本插件各域直接计数；exam/oplib 跨插件以表名计数（防 import 环）；
// 趋势取近 7 日 llmops 模型调用按日聚合，缺口日补零。
package service

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// DashboardStats 看板统计
type DashboardStats struct {
	Assets       int64           `json:"assets"`
	Knowledge    int64           `json:"knowledge"`
	Models       int64           `json:"models"`
	TrainJobs    int64           `json:"trainJobs"`
	Agents       int64           `json:"agents"`
	Datasets     int64           `json:"datasets"`
	Exams        int64           `json:"exams"`
	Operators    int64           `json:"operators"`
	CallsToday   int64           `json:"callsToday"`
	AgentCallsToday int64        `json:"agentCallsToday"`
	SheetsSubmitted int64        `json:"sheetsSubmitted"`
	Trend        []TrendPoint    `json:"trend"`
}

// TrendPoint 趋势点
type TrendPoint struct {
	Date  string `json:"date"`
	Calls int64  `json:"calls"`
}

// buildTrend 趋势补零（纯函数，测试锚点）：rows 为 DB 按日聚合结果，补齐 days 窗口缺口
func buildTrend(rows []struct {
	Day  string
	Cnt  int64
}, days int) []TrendPoint {
	if days <= 0 {
		days = 7
	}
	cntMap := map[string]int64{}
	for _, r := range rows {
		cntMap[r.Day] = r.Cnt
	}
	today := time.Now()
	out := make([]TrendPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		out = append(out, TrendPoint{Date: d, Calls: cntMap[d]})
	}
	return out
}

// countTable 表名计数（跨插件防 import 环）
func countTable(table string) int64 {
	var n int64
	global.GVA_DB.Table(table).Where("deleted_at IS NULL").Count(&n)
	return n
}

// GetDashboard 看板聚合
func (s *LlmopsServiceStruct) GetDashboard() (*DashboardStats, error) {
	st := &DashboardStats{}
	counters := []struct {
		dst  *int64
		def  bool // true=本插件 model 计数，false=跨插件表名计数
		key  any
	}{
		{&st.Assets, true, &model.LlmopsDataAsset{}},
		{&st.Knowledge, true, &model.LlmopsKnowledge{}},
		{&st.Models, true, &model.LlmopsModel{}},
		{&st.TrainJobs, true, &model.LlmopsTrainJob{}},
		{&st.Agents, true, &model.LlmopsAgent{}},
		{&st.Datasets, true, &model.LlmopsDataset{}},
	}
	for _, c := range counters {
		if err := global.GVA_DB.Model(c.key).Count(c.dst).Error; err != nil {
			return nil, err
		}
	}
	st.Exams = countTable("exam_exam")
	st.Operators = countTable("oplib_operator")
	// 近 24h 调用量
	since := time.Now().Add(-24 * time.Hour)
	if err := global.GVA_DB.Model(&model.LlmopsModelCallLog{}).Where("created_at > ?", since).Count(&st.CallsToday).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Table("llmops_agent_call_log").Where("created_at > ?", since).Count(&st.AgentCallsToday).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Table("exam_answer_sheet").Where("status = 'submitted' AND deleted_at IS NULL").Count(&st.SheetsSubmitted).Error; err != nil {
		return nil, err
	}
	// 近 7 日模型调用趋势
	var rows []struct {
		Day string
		Cnt int64
	}
	weekAgo := time.Now().AddDate(0, 0, -6)
	if err := global.GVA_DB.Model(&model.LlmopsModelCallLog{}).
		Select(fmt.Sprintf("DATE(created_at) as day, COUNT(*) as cnt")).
		Where("created_at > ?", weekAgo).
		Group("DATE(created_at)").Find(&rows).Error; err != nil {
		return nil, err
	}
	st.Trend = buildTrend(rows, 7)
	return st, nil
}

// timeNowDash 测试辅助（趋势日期基准）
func timeNowDash() time.Time { return time.Now() }
