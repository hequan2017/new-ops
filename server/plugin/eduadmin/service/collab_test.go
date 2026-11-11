// Package service eduadmin：跨校协同单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestScheduleCheck 调度规则
func TestScheduleCheck(t *testing.T) {
	res := &model.EduShareResource{ResourceName: "高光谱仪", OwnerSchool: "上海大学", Capacity: 3, Status: "可共享"}
	if ok, _ := scheduleCheck(res, "同济大学"); !ok {
		t.Fatal("他校申请可共享资源应通过")
	}
	if ok, _ := scheduleCheck(res, "上海大学"); ok {
		t.Fatal("归属校自用应拒绝")
	}
	res.Status = "占用"
	if ok, _ := scheduleCheck(res, "同济大学"); ok {
		t.Fatal("占用中应拒绝")
	}
	res.Status = "可共享"
	res.Capacity = 0
	if ok, _ := scheduleCheck(res, "同济大学"); ok {
		t.Fatal("容量用尽应拒绝")
	}
	if ok, _ := scheduleCheck(res, ""); ok {
		t.Fatal("申请校为空应拒绝")
	}
}

// TestJointAlloc 统筹分配（占比+下限 10% 重归一）
func TestJointAlloc(t *testing.T) {
	bases := map[string]float64{"甲校": 600, "乙校": 300, "丙校": 100}
	out := jointAlloc(bases, 1000)
	if out["甲校"] != 600 || out["乙校"] != 300 || out["丙校"] != 100 {
		t.Fatalf("无下限触发应保持原占比: %+v", out)
	}
	// 丙校占比 0.1 恰好达下限（不触发），甲 0.6/乙 0.3
	bases2 := map[string]float64{"甲校": 800, "乙校": 150, "丙校": 50}
	out2 := jointAlloc(bases2, 1000)
	// 丙校 0.05 < 0.10 触发下限；甲/乙按原比例分摊 0.9
	if out2["丙校"] != 100 {
		t.Fatalf("丙校应抬升到 100: %+v", out2)
	}
	if out2["甲校"]+out2["乙校"]+out2["丙校"] != 1000 {
		t.Fatalf("总额应守恒: %+v", out2)
	}
	if jointAlloc(nil, 100) != nil {
		t.Fatal("空基数应返回空")
	}
}
