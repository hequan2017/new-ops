// Package service 白泽 GPU 算力：节点/规格/实例与防超卖引擎（M5）
// gpu 插件错误码段：1600-1699（DEV_PLAN 3.6）
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"gorm.io/gorm"
)

// 错误码
const (
	ErrCodeGpuNameRequired  = 1601
	ErrCodeGpuDuplicate     = 1602
	ErrCodeGpuNotFound      = 1603
	ErrCodeGpuStructInvalid = 1604
	ErrCodeGpuOversold      = 1605
	ErrCodeGpuNodeOffline   = 1606
)

func newGpuErr(code int, msg string) error { return &gpuError{Code: code, Msg: msg} }

type gpuError struct {
	Code int
	Msg  string
}

func (e *gpuError) Error() string { return e.Msg }

// GpuService GPU 服务
type GpuService struct{}

// canAllocate 配额校验（纯逻辑，可单测）：total - used >= want
func canAllocate(total, used, want int) bool {
	return want > 0 && total-used >= want
}

// nodeUsage 节点当前占用量（运行中实例汇总）
func nodeUsage(tx *gorm.DB, nodeID uint) (gpu, cpu, mem int, err error) {
	type agg struct {
		Gpu int
		Cpu int
		Mem int
	}
	var a agg
	err = tx.Model(&model.GpuInstance{}).
		Select("COALESCE(SUM(gpu_allocated),0) as gpu, COALESCE(SUM(cpu_allocated),0) as cpu, COALESCE(SUM(mem_alloc_gb),0) as mem").
		Where("node_id = ? AND status = ?", nodeID, model.InstRunning).
		Scan(&a).Error
	return a.Gpu, a.Cpu, a.Mem, err
}

// CreateNode 注册算力节点
func (s *GpuService) CreateNode(n *model.GpuNode) error {
	if strings.TrimSpace(n.Name) == "" || n.GpuTotal <= 0 {
		return newGpuErr(ErrCodeGpuNameRequired, "节点名称与 GPU 卡数必填")
	}
	var count int64
	global.GVA_DB.Model(&model.GpuNode{}).Where("name = ?", n.Name).Count(&count)
	if count > 0 {
		return newGpuErr(ErrCodeGpuDuplicate, "节点已存在: "+n.Name)
	}
	if n.Status == "" {
		n.Status = model.NodeOnline
	}
	return global.GVA_DB.Create(n).Error
}

// UpdateNode 更新节点
func (s *GpuService) UpdateNode(n *model.GpuNode) error {
	if n == nil || n.ID == 0 {
		return newGpuErr(ErrCodeGpuNameRequired, "节点ID不能为空")
	}
	// 总量不可低于当前运行中占用
	gpuUsed, cpuUsed, memUsed, err := nodeUsage(global.GVA_DB, n.ID)
	if err != nil {
		return err
	}
	if n.GpuTotal < gpuUsed || n.CpuTotal < cpuUsed || n.MemTotalGb < memUsed {
		return newGpuErr(ErrCodeGpuOversold, "总量不可低于运行中实例的已分配量")
	}
	return global.GVA_DB.Model(&model.GpuNode{}).Where("id = ?", n.ID).Updates(map[string]any{
		"name": n.Name, "gpu_model": n.GpuModel, "gpu_total": n.GpuTotal,
		"cpu_total": n.CpuTotal, "mem_total_gb": n.MemTotalGb,
		"status": n.Status, "notes": n.Notes,
	}).Error
}

// DeleteNode 删除节点（有运行中实例拒绝）
func (s *GpuService) DeleteNode(id uint) error {
	var running int64
	global.GVA_DB.Model(&model.GpuInstance{}).
		Where("node_id = ? AND status = ?", id, model.InstRunning).Count(&running)
	if running > 0 {
		return newGpuErr(ErrCodeGpuOversold, fmt.Sprintf("仍有 %d 个运行中实例，不可删除", running))
	}
	return global.GVA_DB.Delete(&model.GpuNode{}, id).Error
}

// GetNodeList 节点列表（含余量计算）
type NodeWithUsage struct {
	model.GpuNode
	GpuUsed int `json:"gpuUsed"`
	CpuUsed int `json:"cpuUsed"`
	MemUsed int `json:"memUsed"`
}

func (s *GpuService) GetNodeList() ([]NodeWithUsage, error) {
	var nodes []model.GpuNode
	if err := global.GVA_DB.Order("id DESC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	out := make([]NodeWithUsage, 0, len(nodes))
	for _, n := range nodes {
		gpu, cpu, mem, _ := nodeUsage(global.GVA_DB, n.ID)
		out = append(out, NodeWithUsage{GpuNode: n, GpuUsed: gpu, CpuUsed: cpu, MemUsed: mem})
	}
	return out, nil
}

// CreateSpec 创建规格
func (s *GpuService) CreateSpec(sp *model.GpuSpec) error {
	if strings.TrimSpace(sp.Name) == "" || sp.GpuCount <= 0 {
		return newGpuErr(ErrCodeGpuNameRequired, "规格名称与 GPU 卡数必填")
	}
	var count int64
	global.GVA_DB.Model(&model.GpuSpec{}).Where("name = ?", sp.Name).Count(&count)
	if count > 0 {
		return newGpuErr(ErrCodeGpuDuplicate, "规格已存在: "+sp.Name)
	}
	return global.GVA_DB.Create(sp).Error
}

// DeleteSpec 删除规格
func (s *GpuService) DeleteSpec(id uint) error {
	return global.GVA_DB.Delete(&model.GpuSpec{}, id).Error
}

// GetSpecList 规格列表
func (s *GpuService) GetSpecList() ([]model.GpuSpec, error) {
	var list []model.GpuSpec
	err := global.GVA_DB.Order("id DESC").Find(&list).Error
	return list, err
}

// StartInstance 开通实例：事务内行锁节点 → 防超卖校验 → 台账落库
// 容器实际创建联动 container 插件（后续场次），当前为资源分配台账模式。
func (s *GpuService) StartInstance(nodeID, specID uint, name, creator string, userID uint) (*model.GpuInstance, error) {
	var spec model.GpuSpec
	if err := global.GVA_DB.First(&spec, specID).Error; err != nil {
		return nil, newGpuErr(ErrCodeGpuNotFound, "规格不存在")
	}
	if name == "" {
		name = fmt.Sprintf("gpu-%s-%d", spec.Name, time.Now().Unix()%100000)
	}
	var inst *model.GpuInstance
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 行锁节点（防并发超卖的关键）
		var node model.GpuNode
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&node, nodeID).Error; err != nil {
			return newGpuErr(ErrCodeGpuNotFound, "节点不存在")
		}
		if node.Status != model.NodeOnline {
			return newGpuErr(ErrCodeGpuNodeOffline, "节点离线，不可开通")
		}
		gpuUsed, cpuUsed, memUsed, err := nodeUsage(tx, nodeID)
		if err != nil {
			return err
		}
		if !canAllocate(node.GpuTotal, gpuUsed, spec.GpuCount) ||
			!canAllocate(node.CpuTotal, cpuUsed, spec.CpuCores) ||
			!canAllocate(node.MemTotalGb, memUsed, spec.MemGb) {
			return newGpuErr(ErrCodeGpuOversold,
				fmt.Sprintf("节点 %s 资源不足：GPU 余 %d 需 %d，CPU 余 %d 需 %d，内存余 %dG 需 %dG",
					node.Name, node.GpuTotal-gpuUsed, spec.GpuCount,
					node.CpuTotal-cpuUsed, spec.CpuCores,
					node.MemTotalGb-memUsed, spec.MemGb))
		}
		inst = &model.GpuInstance{
			NodeID: node.ID, NodeName: node.Name,
			SpecID: spec.ID, SpecName: spec.Name,
			Name:         name,
			GpuAllocated: spec.GpuCount, CpuAllocated: spec.CpuCores, MemAllocGb: spec.MemGb,
			Status: model.InstRunning, Creator: creator, UserID: userID,
		}
		return tx.Create(inst).Error
	})
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// ReleaseInstance 销毁实例：标记状态释放配额
func (s *GpuService) ReleaseInstance(id uint) error {
	var inst model.GpuInstance
	if err := global.GVA_DB.First(&inst, id).Error; err != nil {
		return newGpuErr(ErrCodeGpuNotFound, "实例不存在")
	}
	if inst.Status != model.InstRunning {
		return newGpuErr(ErrCodeGpuStructInvalid, "实例非运行中状态")
	}
	now := time.Now()
	return global.GVA_DB.Model(&model.GpuInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status": model.InstReleased, "released_at": &now,
	}).Error
}

// GetInstanceList 实例列表（status 过滤）
func (s *GpuService) GetInstanceList(status string) ([]model.GpuInstance, error) {
	db := global.GVA_DB.Model(&model.GpuInstance{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []model.GpuInstance
	err := db.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}
