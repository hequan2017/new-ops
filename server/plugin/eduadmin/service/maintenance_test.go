// Package service eduadmin：设施维护单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestRouteTeam 班组路由：关键字 > 等级 > 默认
func TestRouteTeam(t *testing.T) {
	cases := []struct {
		level, description, want string
	}{
		{"严重", "空调压缩机烧毁", "机电班组"},
		{"轻微", "门锁松动", "综合维修班组"},
		{"常规", "走廊灯更换", "综合维修班组"},
		{"常规", "教学楼漏水", "水电班组"},
		{"严重", "宿舍断电", "水电班组"},
		{"轻微", "水管渗水", "水电班组"},
	}
	for _, c := range cases {
		if got := routeTeam(c.level, c.description); got != c.want {
			t.Fatalf("level=%s desc=%s 应 %s，得 %s", c.level, c.description, c.want, got)
		}
	}
}
