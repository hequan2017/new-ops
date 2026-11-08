// Package service fanhua：价值观健康度单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

func healthWords() []model.FanhuaSensitiveWord {
	return []model.FanhuaSensitiveWord{
		{Word: "赌博", Action: "block", Weight: 30, Enabled: true},
		{Word: "傻子", Action: "replace", Replacement: "反应较慢的人", Weight: 5, Enabled: true},
		{Word: "已停用", Action: "block", Weight: 50, Enabled: false},
	}
}

// TestComputeHealth 加权扣减与拦截
func TestComputeHealth(t *testing.T) {
	res := computeHealth("参与赌博是违法行为", healthWords())
	if res.Health != 70 || !res.Blocked {
		t.Fatalf("block 词应扣 30 且拦截: %+v", res)
	}
	if len(res.Hits) != 1 || res.Hits[0].Word != "赌博" {
		t.Fatalf("命中项不符: %+v", res.Hits)
	}
	// 停用词不扣
	res2 := computeHealth("已停用词表内容", healthWords())
	if res2.Health != 100 || res2.Blocked {
		t.Fatalf("停用词不应扣分: %+v", res2)
	}
}

// TestComputeHealthReplace replace 替换输出
func TestComputeHealthReplace(t *testing.T) {
	res := computeHealth("他真是傻子吗", healthWords())
	if res.Blocked {
		t.Fatal("replace 词不应整体拦截")
	}
	if !strings.Contains(res.Sanitized, "反应较慢的人") || strings.Contains(res.Sanitized, "傻子") {
		t.Fatalf("替换输出不符: %s", res.Sanitized)
	}
	if res.Health != 95 {
		t.Fatalf("应扣 5 分: %d", res.Health)
	}
}

// TestComputeHealthFloor 下限 0 与情感分类
func TestComputeHealthFloor(t *testing.T) {
	text := "赌博赌博赌博赌博赌博" // 5×30=150 → 下限 0
	res := computeHealth(text, healthWords())
	if res.Health != 0 {
		t.Fatalf("应下限 0: %d", res.Health)
	}
	if res.Sentiment == "" {
		t.Fatal("应输出情感分类")
	}
	// 空命中
	empty := computeHealth("今天天气不错", healthWords())
	if empty.Health != 100 || len(empty.Hits) != 0 {
		t.Fatalf("无命中应满分: %+v", empty)
	}
}
