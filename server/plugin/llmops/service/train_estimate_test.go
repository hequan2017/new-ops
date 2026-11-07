// Package service llmops 训练估算单测（经验模型口径）
package service

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestEstimateTrain(t *testing.T) {
	// 100 样本 × 32 字符 × 3 epochs = 6000 tokens（100*32/1.6*3）
	est, err := EstimateTrain(100, 32, 3, 1, 6.0)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if est.Tokens != 6000 {
		t.Fatalf("token 估算异常: %d", est.Tokens)
	}
	// 时长 = 6000 / (3000*0.85) 秒 ≈ 2.353s ≈ 0.000654h
	wantHours := 6000.0 / (3000 * 0.85) / 3600
	if !approx(est.Hours, wantHours) {
		t.Fatalf("时长估算异常: %v want %v", est.Hours, wantHours)
	}
	if !approx(est.Cost, wantHours*1*6.0) {
		t.Fatalf("费用估算异常: %v", est.Cost)
	}
	// 多卡线性加速：效率因子作用于每卡，4 卡时长 = 单卡口径/4
	est4, _ := EstimateTrain(100, 32, 3, 4, 6.0)
	if !approx(est4.Hours, est.Hours/4) {
		t.Fatalf("多卡加速异常: %v want %v", est4.Hours, est.Hours/4)
	}
	// 费用不变：线性加速口径下时长降幅恰好抵消卡数增量（诚实模型结论）
	if !approx(est4.Cost, est.Cost) {
		t.Fatalf("线性模型费用应不变: %v vs %v", est4.Cost, est.Cost)
	}
}

func TestEstimateTrain_Guards(t *testing.T) {
	if _, err := EstimateTrain(0, 32, 3, 1, 6.0); err == nil {
		t.Fatal("零样本应拒绝")
	}
	// 缺省 epochs/gpu/单价
	est, err := EstimateTrain(10, 16, 0, 0, 0)
	if err != nil {
		t.Fatalf("缺省应补齐: %v", err)
	}
	if est.Tokens != int64(10*16/1.6*3) {
		t.Fatalf("缺省 epochs=3 未生效: %d", est.Tokens)
	}
}
