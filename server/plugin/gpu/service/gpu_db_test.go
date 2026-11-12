package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// newTestSvc 内存库注入服务（repo 层 GORM 路径真测）
func newTestSvc(t *testing.T) *GpuService {
	t.Helper()
	db, err := data.OpenMemory(&model.GpuNode{}, &model.GpuSpec{}, &model.GpuInstance{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	return NewGpuService(db)
}

func seedNodeSpec(t *testing.T, s *GpuService, gpuTotal int, specGpu int) (*model.GpuNode, *model.GpuSpec) {
	t.Helper()
	n := &model.GpuNode{Name: "节点A", GpuModel: "A100", GpuTotal: gpuTotal, CpuTotal: 64, MemTotalGb: 256}
	if err := s.CreateNode(n); err != nil {
		t.Fatalf("seed node: %v", err)
	}
	sp := &model.GpuSpec{Name: "规格1", GpuCount: specGpu, CpuCores: 8, MemGb: 32}
	if err := s.CreateSpec(sp); err != nil {
		t.Fatalf("seed spec: %v", err)
	}
	return n, sp
}

func TestCreateNodeDuplicate(t *testing.T) {
	s := newTestSvc(t)
	n, _ := seedNodeSpec(t, s, 4, 1)
	n2 := &model.GpuNode{Name: n.Name, GpuModel: "V100", GpuTotal: 2}
	if err := s.CreateNode(n2); errx.CodeOf(err) != ErrCodeGpuDuplicate {
		t.Fatalf("重名应 1602，实际 err=%v code=%d", err, errx.CodeOf(err))
	}
	// 必填校验
	if err := s.CreateNode(&model.GpuNode{Name: " ", GpuTotal: 1}); errx.CodeOf(err) != ErrCodeGpuNameRequired {
		t.Fatalf("空名应 1601，实际 code=%d", errx.CodeOf(err))
	}
}

func TestStartInstanceAntiOversell(t *testing.T) {
	s := newTestSvc(t)
	n, sp := seedNodeSpec(t, s, 4, 3)

	inst, err := s.StartInstance(n.ID, sp.ID, "", "tester", 1)
	if err != nil {
		t.Fatalf("首次开通应成功: %v", err)
	}
	if inst.GpuAllocated != 3 || inst.Status != model.InstRunning || inst.Name == "" {
		t.Fatalf("实例字段口径不符: %+v", inst)
	}

	// 余 1 需 3 → 防超卖拒绝
	if _, err = s.StartInstance(n.ID, sp.ID, "b", "tester", 1); errx.CodeOf(err) != ErrCodeGpuOversold {
		t.Fatalf("超卖应 1605，实际 code=%d", errx.CodeOf(err))
	}

	// 释放后可再开通
	if err = s.ReleaseInstance(inst.ID); err != nil {
		t.Fatalf("释放应成功: %v", err)
	}
	if _, err = s.StartInstance(n.ID, sp.ID, "c", "tester", 1); err != nil {
		t.Fatalf("释放后再开通应成功: %v", err)
	}
}

func TestStartInstanceSpecAndNodeState(t *testing.T) {
	s := newTestSvc(t)
	n, sp := seedNodeSpec(t, s, 4, 1)

	// 规格不存在
	if _, err := s.StartInstance(n.ID, 999, "", "t", 1); errx.CodeOf(err) != ErrCodeGpuNotFound {
		t.Fatalf("规格不存在应 1603，实际 code=%d", errx.CodeOf(err))
	}
	// 节点离线
	n.Status = model.NodeOffline
	if err := s.UpdateNode(n); err != nil {
		t.Fatalf("置离线应成功: %v", err)
	}
	if _, err := s.StartInstance(n.ID, sp.ID, "", "t", 1); errx.CodeOf(err) != ErrCodeGpuNodeOffline {
		t.Fatalf("离线节点应 1606，实际 code=%d", errx.CodeOf(err))
	}
}

func TestUpdateNodeShrinkRejected(t *testing.T) {
	s := newTestSvc(t)
	n, sp := seedNodeSpec(t, s, 4, 3)
	if _, err := s.StartInstance(n.ID, sp.ID, "", "t", 1); err != nil {
		t.Fatalf("开通: %v", err)
	}
	n.GpuTotal = 2 // 缩到低于已用 3
	if err := s.UpdateNode(n); errx.CodeOf(err) != ErrCodeGpuOversold {
		t.Fatalf("缩容应 1605，实际 code=%d", errx.CodeOf(err))
	}
	n.GpuTotal = 4
	if err := s.UpdateNode(n); err != nil {
		t.Fatalf("维持总量应成功: %v", err)
	}
	// 删除有运行中实例的节点应拒绝
	if err := s.DeleteNode(n.ID); errx.CodeOf(err) != ErrCodeGpuOversold {
		t.Fatalf("有运行中实例删除应 1605，实际 code=%d", errx.CodeOf(err))
	}
}

func TestReleaseInstanceOnce(t *testing.T) {
	s := newTestSvc(t)
	n, sp := seedNodeSpec(t, s, 2, 1)
	inst, err := s.StartInstance(n.ID, sp.ID, "x", "t", 1)
	if err != nil {
		t.Fatalf("开通: %v", err)
	}
	if err = s.ReleaseInstance(inst.ID); err != nil {
		t.Fatalf("首释放: %v", err)
	}
	// 重复释放：非运行中状态
	if err = s.ReleaseInstance(inst.ID); errx.CodeOf(err) != ErrCodeGpuStructInvalid {
		t.Fatalf("重复释放应 1604，实际 code=%d", errx.CodeOf(err))
	}
	// 不存在的实例
	if err = s.ReleaseInstance(999); errx.CodeOf(err) != ErrCodeGpuNotFound {
		t.Fatalf("不存在应 1603，实际 code=%d", errx.CodeOf(err))
	}
}

func TestGetInstanceListPaging(t *testing.T) {
	s := newTestSvc(t)
	db := s.dbh()
	for i := 0; i < 3; i++ {
		inst := &model.GpuInstance{Name: "inst", Status: model.InstRunning, Creator: "t"}
		if err := db.Create(inst).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// 标准分页
	list, total, err := s.GetInstanceList("", 0, 0, 1, 2)
	if err != nil || total != 3 || len(list) != 2 {
		t.Fatalf("分页口径不符: err=%v total=%d len=%d", err, total, len(list))
	}
	// 状态过滤
	list, total, err = s.GetInstanceList(model.InstReleased, 0, 0, 0, 0)
	if err != nil || total != 0 || len(list) != 0 {
		t.Fatalf("状态过滤不符: err=%v total=%d", err, total)
	}
	// page<=0 全量
	list, total, err = s.GetInstanceList("", 0, 0, 0, 0)
	if err != nil || total != 3 || len(list) != 3 {
		t.Fatalf("全量口径不符: err=%v total=%d len=%d", err, total, len(list))
	}
}
