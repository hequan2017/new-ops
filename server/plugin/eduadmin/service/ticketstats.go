// Package service eduadmin：工单统计报表（需求 3.6-36）
// 口径：总量/办结率（办结+评价+归档/总量）/超时率（超时督办数/总量）/满意度
// （好评/全部评价）/高频诉求类型 Top5/部门办理效率（分派数与闭环数）；
// CSV 导出复用 exam I07 口径（strings.Builder+表头，纯函数）。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduTicketStatsParam = 2843

// rateOf 比例（分母 0 → 0）
func rateOf(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return round2(float64(num) / float64(den))
}

// typeCount 诉求类型计数
type typeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// deptEff 部门办理效率
type deptEff struct {
	Dept     string `json:"dept"`
	Dispatch int    `json:"dispatch"`
}

// TicketStats 工单统计
type TicketStats struct {
	Total         int         `json:"total"`
	Closed        int         `json:"closed"`
	CloseRate     float64     `json:"closeRate"`
	UrgedNear     int         `json:"urgedNear"`
	UrgedOver     int         `json:"urgedOver"`
	OverdueRate   float64     `json:"overdueRate"`
	Reviews       int         `json:"reviews"`
	GoodRate      float64     `json:"goodRate"`
	TopTypes      []typeCount `json:"topTypes"`
	DeptEffList   []deptEff   `json:"deptEfficiency"`
	ChannelCounts map[string]int `json:"channelCounts"` // 渠道维度（网页/企微/小程序）
}

// ticketStats 统计聚合（纯函数，测试锚点）
func ticketStats(tickets []model.EduTicket, urges []model.EduTicketUrge, reviews []model.EduTicketReview, logs []model.EduDispatchLog) *TicketStats {
	st := &TicketStats{
		Total: len(tickets), TopTypes: []typeCount{}, DeptEffList: []deptEff{},
		ChannelCounts: map[string]int{},
	}
	for _, t := range tickets {
		switch t.Status {
		case "办结", "评价", "归档":
			st.Closed++
		}
		if t.Channel != "" {
			st.ChannelCounts[t.Channel]++
		}
	}
	for _, u := range urges {
		switch u.Level {
		case "临近":
			st.UrgedNear++
		case "超时":
			st.UrgedOver++
		}
	}
	good := 0
	for _, r := range reviews {
		st.Reviews++
		if r.Score == "好评" {
			good++
		}
	}
	st.CloseRate = rateOf(st.Closed, st.Total)
	st.OverdueRate = rateOf(st.UrgedOver, st.Total)
	st.GoodRate = rateOf(good, st.Reviews)
	// 高频诉求类型 Top5
	typeMap := map[string]int{}
	for _, t := range tickets {
		typeMap[t.DemandType]++
	}
	types := make([]typeCount, 0, len(typeMap))
	for k, v := range typeMap {
		types = append(types, typeCount{Type: k, Count: v})
	}
	sort.Slice(types, func(i, j int) bool {
		if types[i].Count != types[j].Count {
			return types[i].Count > types[j].Count
		}
		return types[i].Type < types[j].Type
	})
	if len(types) > 5 {
		types = types[:5]
	}
	st.TopTypes = types
	// 部门办理效率（自动+人工分派记录计数）
	deptMap := map[string]int{}
	for _, l := range logs {
		if l.Dept != "" {
			deptMap[l.Dept]++
		}
	}
	depts := make([]deptEff, 0, len(deptMap))
	for k, v := range deptMap {
		depts = append(depts, deptEff{Dept: k, Dispatch: v})
	}
	sort.Slice(depts, func(i, j int) bool {
		if depts[i].Dispatch != depts[j].Dispatch {
			return depts[i].Dispatch > depts[j].Dispatch
		}
		return depts[i].Dept < depts[j].Dept
	})
	st.DeptEffList = depts
	return st
}

// ticketStatsCSV 统计 CSV（纯函数，测试锚点，对齐 I07 口径）
func ticketStatsCSV(st *TicketStats) string {
	var sb strings.Builder
	sb.WriteString("指标,数值\n")
	sb.WriteString(fmt.Sprintf("工单总量,%d\n", st.Total))
	sb.WriteString(fmt.Sprintf("已办结,%d\n", st.Closed))
	sb.WriteString(fmt.Sprintf("办结率,%.0f%%\n", st.CloseRate*100))
	sb.WriteString(fmt.Sprintf("超时督办,%d\n", st.UrgedOver))
	sb.WriteString(fmt.Sprintf("超时率,%.0f%%\n", st.OverdueRate*100))
	sb.WriteString(fmt.Sprintf("评价数,%d\n", st.Reviews))
	sb.WriteString(fmt.Sprintf("满意率,%.0f%%\n", st.GoodRate*100))
	for _, t := range st.TopTypes {
		sb.WriteString(fmt.Sprintf("高频诉求 %s,%d\n", t.Type, t.Count))
	}
	for _, d := range st.DeptEffList {
		sb.WriteString(fmt.Sprintf("部门分派 %s,%d\n", d.Dept, d.Dispatch))
	}
	channels := make([]string, 0, len(st.ChannelCounts))
	for c := range st.ChannelCounts {
		channels = append(channels, c)
	}
	sort.Strings(channels)
	for _, c := range channels {
		sb.WriteString(fmt.Sprintf("渠道 %s,%d\n", c, st.ChannelCounts[c]))
	}
	return sb.String()
}

// GetEduTicketStats 工单统计聚合
func (s *EduadminService) GetEduTicketStats() (*TicketStats, error) {
	var tickets []model.EduTicket
	if err := global.GVA_DB.Order("id ASC").Limit(5000).Find(&tickets).Error; err != nil {
		return nil, err
	}
	var urges []model.EduTicketUrge
	if err := global.GVA_DB.Limit(5000).Find(&urges).Error; err != nil {
		return nil, err
	}
	var reviews []model.EduTicketReview
	if err := global.GVA_DB.Limit(5000).Find(&reviews).Error; err != nil {
		return nil, err
	}
	var logs []model.EduDispatchLog
	if err := global.GVA_DB.Limit(5000).Find(&logs).Error; err != nil {
		return nil, err
	}
	return ticketStats(tickets, urges, reviews, logs), nil
}

// ExportEduTicketStats 统计 CSV 导出
func (s *EduadminService) ExportEduTicketStats() (string, error) {
	st, err := s.GetEduTicketStats()
	if err != nil {
		return "", err
	}
	return ticketStatsCSV(st), nil
}
