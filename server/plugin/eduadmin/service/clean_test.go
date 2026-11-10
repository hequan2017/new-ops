// Package service eduadmin：数据清洗单测（纯函数，无 DB）
package service

import (
	"math"
	"testing"
)

// TestMojibakeRate 乱码率：合法中文 0、含非法字节 >0.3
func TestMojibakeRate(t *testing.T) {
	if v := mojibakeRate("正常中文内容，无乱码"); v != 0 {
		t.Fatalf("合法文本乱码率应为 0: %v", v)
	}
	bad := string([]byte{0xff, 0xfe, 0xfd}) + string([]byte{0x80, 0x81}) // 5 个非法 rune
	if v := mojibakeRate(bad); v < 0.3 {
		t.Fatalf("乱码文本乱码率应 >0.3: %v", v)
	}
	if v := mojibakeRate(""); v != 0 {
		t.Fatalf("空文本乱码率应为 0: %v", v)
	}
}

// TestFillMissing 均值/前值两策略填充
func TestFillMissing(t *testing.T) {
	rows := []map[string]any{
		{"name": "a", "salary": 100.0},
		{"name": "b"},
		{"name": "c", "salary": 200.0},
		{"name": "d"},
	}
	n, err := fillMissing(rows, []string{"salary"}, "均值")
	if err != nil || n != 2 {
		t.Fatalf("均值应填充 2 处: %d %v", n, err)
	}
	if rows[1]["salary"] != 150.0 || rows[3]["salary"] != 150.0 {
		t.Fatalf("均值应为 150: %+v", rows)
	}

	rows2 := []map[string]any{
		{"k": 1.0},
		{},
		{"k": "3"}, // 字符串数值可解析
		{},
	}
	n2, err := fillMissing(rows2, []string{"k"}, "前值")
	if err != nil || n2 != 2 {
		t.Fatalf("前值应填充 2 处: %d %v", n2, err)
	}
	if rows2[1]["k"] != 1.0 || rows2[3]["k"] != 3.0 {
		t.Fatalf("前值应为上一数值: %+v", rows2)
	}

	if _, err := fillMissing(rows2, []string{"k"}, "中位数"); err == nil {
		t.Fatal("策略白名单外应报错")
	}
}

// TestSigmaOutliers 3σ：极端值检出、σ=0 无异常（单点离群需 n≥11 才可能超 3σ，
// 理论上界 z=√(n-1)，故构造 12 个正常值 + 1 个极端值）
func TestSigmaOutliers(t *testing.T) {
	vals := []float64{10, 11, 10, 12, 11, 10, 11, 12, 10, 11, 12, 11, 1000}
	out := sigmaOutliers(vals)
	if len(out) != 1 || !out[12] {
		t.Fatalf("应仅检出极端值下标 12: %+v", out)
	}
	if o := sigmaOutliers([]float64{5, 5, 5, 5}); len(o) != 0 {
		t.Fatalf("σ=0 应无异常: %+v", o)
	}
	if o := sigmaOutliers([]float64{42}); len(o) != 0 {
		t.Fatalf("单值应无异常: %+v", o)
	}
}

// TestOutlierIndexes NaN 剔除后的下标映射（10 正常 + 1 极端 + 2 NaN）
func TestOutlierIndexes(t *testing.T) {
	vals := []float64{1, 2, 1.5, 1.8, 1.2, 0.9, 1.1, 800, 1.3, 1.6, 1.4, math.NaN(), math.NaN()}
	idx := outlierIndexes(vals)
	if len(idx) != 1 || idx[0] != 7 {
		t.Fatalf("应映射回原下标 7: %v", idx)
	}
}
