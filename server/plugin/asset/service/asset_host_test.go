package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

// 纯逻辑校验测试（不依赖 DB）
func Test_validateHost(t *testing.T) {
	t.Run("合法主机", func(t *testing.T) {
		h := &model.AssetHost{Hostname: "web01", IP: "192.168.112.10"}
		if err := validateHost(h); err != nil {
			t.Fatalf("合法主机不应报错: %v", err)
		}
		if h.Status != model.AssetStatusRunning {
			t.Fatalf("未指定状态时应默认运行中, got %s", h.Status)
		}
	})

	t.Run("主机名必填", func(t *testing.T) {
		h := &model.AssetHost{IP: "192.168.112.10"}
		err := validateHost(h)
		if err == nil || !strings.Contains(err.Error(), "主机名") {
			t.Fatalf("应提示主机名必填, got %v", err)
		}
	})

	t.Run("内网IP非法", func(t *testing.T) {
		h := &model.AssetHost{Hostname: "web01", IP: "999.1.1.1"}
		err := validateHost(h)
		if err == nil || !strings.Contains(err.Error(), "内网IP格式不合法") {
			t.Fatalf("应提示IP不合法, got %v", err)
		}
	})

	t.Run("公网IP非法", func(t *testing.T) {
		h := &model.AssetHost{Hostname: "web01", IP: "10.0.0.1", PublicIP: "not-an-ip"}
		err := validateHost(h)
		if err == nil || !strings.Contains(err.Error(), "公网IP格式不合法") {
			t.Fatalf("应提示公网IP不合法, got %v", err)
		}
	})

	t.Run("非法状态", func(t *testing.T) {
		h := &model.AssetHost{Hostname: "web01", IP: "10.0.0.1", Status: model.AssetStatus("未知状态")}
		err := validateHost(h)
		if err == nil || !strings.Contains(err.Error(), "非法的资产状态") {
			t.Fatalf("应提示状态非法, got %v", err)
		}
	})
}

// Test_resolveHostSort 排序白名单解析：白名单外回退默认 id DESC，防注入
func Test_resolveHostSort(t *testing.T) {
	cases := []struct {
		name       string
		sortField  string
		sortOrder  string
		wantColumn string
		wantOrder  string
	}{
		{"空参数回退默认", "", "", "asset_hosts.id", "DESC"},
		{"白名单内更新时间", "updatedAt", "asc", "asset_hosts.updated_at", "ASC"},
		{"白名单内采集时间", "lastCollectAt", "DESC", "asset_hosts.last_collect_at", "DESC"},
		{"白名单外回退默认", "ip; DROP TABLE asset_hosts;--", "asc", "asset_hosts.id", "DESC"},
		{"方向非法按desc", "updatedAt", "sideways", "asset_hosts.updated_at", "DESC"},
		{"方向大小写不敏感", "id", "ASC", "asset_hosts.id", "ASC"},
	}
	for _, c := range cases {
		col, order := resolveHostSort(c.sortField, c.sortOrder)
		if col != c.wantColumn || order != c.wantOrder {
			t.Fatalf("%s: want (%s %s) got (%s %s)", c.name, c.wantColumn, c.wantOrder, col, order)
		}
	}
}
