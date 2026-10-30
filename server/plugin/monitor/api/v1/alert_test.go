package api

import (
	"strings"
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
)

func TestParseEventRange(t *testing.T) {
	cases := []struct {
		name      string
		startAt   string
		endAt     string
		wantStart bool // 是否有起始过滤
		wantEnd   bool
		wantErr   bool
	}{
		{"均为空不加过滤", "", "", false, false, false},
		{"完整时间格式", "2026-10-01 08:00:00", "2026-10-31 23:59:59", true, true, false},
		{"纯日期结束含全天", "", "2026-10-31", false, true, false},
		{"非法格式报错", "10/01/2026", "", false, false, true},
	}
	for _, c := range cases {
		start, end, err := parseEventRange(c.startAt, c.endAt)
		if (err != nil) != c.wantErr {
			t.Fatalf("%s: err=%v wantErr=%v", c.name, err, c.wantErr)
		}
		if c.wantErr {
			continue
		}
		if (start != nil) != c.wantStart || (end != nil) != c.wantEnd {
			t.Fatalf("%s: start/end 存在性不符 (start=%v end=%v)", c.name, start, end)
		}
	}
	// 纯日期结束边界：应落在当日 23:59:59
	_, end, _ := parseEventRange("", "2026-10-31")
	if got := end.Format("2006-01-02 15:04:05"); got != "2026-10-31 23:59:59" {
		t.Fatalf("纯日期结束应取当日 23:59:59, got %s", got)
	}
}

func TestAlertEventsCSV(t *testing.T) {
	resolved := time.Date(2026, 10, 31, 8, 0, 0, 0, time.Local)
	got := alertEventsCSV([]model.MonitorAlertEvent{
		{Summary: "CPU 超阈值, 持续 5 分钟", FiredAt: resolved, Value: 95.5, Notified: true},
	})
	s := string(got)
	if !strings.HasPrefix(s, "\uFEFF") {
		t.Fatal("CSV 应带 BOM 前缀")
	}
	for _, want := range []string{"触发时间,恢复时间,资产ID,摘要,触发值,已通知", `"CPU 超阈值, 持续 5 分钟"`, "是", "2026-10-31 08:00:00"} {
		if !strings.Contains(s, want) {
			t.Fatalf("CSV 缺少内容 %q", want)
		}
	}
	if got := alertEventsCSV(nil); !strings.Contains(string(got), "触发时间") {
		t.Fatal("空列表也应输出表头")
	}
}
