// Package service eduadmin：工单统计单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestTicketStats 统计聚合
func TestTicketStats(t *testing.T) {
	tickets := []model.EduTicket{
		{Status: "办结", DemandType: "教务", Channel: "网页"},
		{Status: "评价", DemandType: "教务", Channel: "网页"},
		{Status: "归档", DemandType: "后勤", Channel: "企微"},
		{Status: "受理", DemandType: "后勤", Channel: "小程序"},
		{Status: "办理", DemandType: "信息化", Channel: "网页"},
	}
	urges := []model.EduTicketUrge{{Level: "临近"}, {Level: "超时"}}
	reviews := []model.EduTicketReview{{Score: "好评"}, {Score: "好评"}, {Score: "差评"}}
	logs := []model.EduDispatchLog{{Dept: "教务处"}, {Dept: "教务处"}, {Dept: "后勤保障处"}}
	st := ticketStats(tickets, urges, reviews, logs)
	if st.Total != 5 || st.Closed != 3 || st.CloseRate != 0.6 {
		t.Fatalf("总量/办结/办结率不符: %+v", st)
	}
	if st.UrgedNear != 1 || st.UrgedOver != 1 || st.OverdueRate != 0.2 {
		t.Fatalf("督办计数不符: %+v", st)
	}
	if st.Reviews != 3 || st.GoodRate != 0.67 {
		t.Fatalf("满意率应 2/3=0.67: %+v", st)
	}
	// 教务与后勤并列 2，字典序「后勤」居首
	if len(st.TopTypes) != 3 || st.TopTypes[0].Type != "后勤" || st.TopTypes[0].Count != 2 {
		t.Fatalf("高频诉求并列 2 应字典序居首: %+v", st.TopTypes)
	}
	if len(st.DeptEffList) != 2 || st.DeptEffList[0].Dept != "教务处" {
		t.Fatalf("部门效率应教务处居首: %+v", st.DeptEffList)
	}
	if rateOf(1, 0) != 0 {
		t.Fatal("零分母应 0")
	}
	// 渠道维度
	if st.ChannelCounts["网页"] != 3 || st.ChannelCounts["企微"] != 1 || st.ChannelCounts["小程序"] != 1 {
		t.Fatalf("渠道计数不符: %+v", st.ChannelCounts)
	}
}

// TestTicketStatsCSV CSV 导出（I07 口径）
func TestTicketStatsCSV(t *testing.T) {
	st := ticketStats([]model.EduTicket{{Status: "办结", DemandType: "教务"}}, nil, []model.EduTicketReview{{Score: "好评"}}, []model.EduDispatchLog{{Dept: "教务处"}})
	csv := ticketStatsCSV(st)
	if !strings.HasPrefix(csv, "指标,数值\n") {
		t.Fatalf("应有表头: %s", csv)
	}
	for _, want := range []string{"工单总量,1", "办结率,100%", "满意率,100%", "高频诉求 教务,1", "部门分派 教务处,1"} {
		if !strings.Contains(csv, want) {
			t.Fatalf("CSV 缺 %s: %s", want, csv)
		}
	}
}
