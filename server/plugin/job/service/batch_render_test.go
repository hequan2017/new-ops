package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/job/model"
)

// TestRenderCommandForHost 批量渲染核心纯逻辑（替代逐主机 Pluck 的 N+1 版本）
func TestRenderCommandForHost(t *testing.T) {
	vgIndex := map[uint]model.JobVariableGroup{
		7: {Name: "web", Variables: `[{"key":"port","value":"8080"}]`},
		9: {Name: "db", Variables: `[{"key":"port","value":"3306"}]`},
	}
	groups := map[uint][]uint{
		101: {7, 9},  // 多组：首个命中生效
		102: {8},     // 组无变量组：原样
		// 103: 无组 → 原样
	}
	base := "echo {{port}}"

	got, err := renderCommandForHost(base, 101, groups, vgIndex)
	if err != nil || !strings.Contains(got, "8080") || strings.Contains(got, "{{port}}") {
		t.Fatalf("多组首个命中应渲染: %q %v", got, err)
	}
	got, err = renderCommandForHost(base, 102, groups, vgIndex)
	if err != nil || got != base {
		t.Fatalf("未命中应原样: %q %v", got, err)
	}
	got, err = renderCommandForHost(base, 103, groups, vgIndex)
	if err != nil || got != base {
		t.Fatalf("无组应原样: %q %v", got, err)
	}
}
