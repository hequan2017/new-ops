// Package service 白泽 GPU 算力：节点/规格/实例与防超卖引擎（M5）
// gpu 插件错误码段：1600-1699
package service

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"github.com/hequan2017/new-ops/server/plugin/gpu/repo"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码
const (
	ErrCodeGpuStore         = 1600
	ErrCodeGpuNameRequired  = 1601
	ErrCodeGpuDuplicate     = 1602
	ErrCodeGpuNotFound      = 1603
	ErrCodeGpuStructInvalid = 1604
	ErrCodeGpuOversold      = 1605
	ErrCodeGpuNodeOffline   = 1606
)

// GpuService GPU 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type GpuService struct {
	db   *gorm.DB
	repo *repo.Repo
}

// NewGpuService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewGpuService(db *gorm.DB) *GpuService {
	return &GpuService{db: db, repo: new(repo.Repo)}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *GpuService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// canAllocate 配额校验（纯逻辑，可单测）：total - used >= want
func canAllocate(total, used, want int) bool {
	return want > 0 && total-used >= want
}

// CreateNode 注册算力节点
func (s *GpuService) CreateNode(n *model.GpuNode) error {
	if strings.TrimSpace(n.Name) == "" || n.GpuTotal <= 0 {
		return errx.New(ErrCodeGpuNameRequired, "节点名称与 GPU 卡数必填")
	}
	db := s.dbh()
	count, err := s.repo.NodeCountByName(db, n.Name)
	if err != nil {
		return errx.Wrap(ErrCodeGpuStore, "查询节点失败", err)
	}
	if count > 0 {
		return errx.New(ErrCodeGpuDuplicate, "节点已存在: "+n.Name)
	}
	if n.Status == "" {
		n.Status = model.NodeOnline
	}
	if err := s.repo.CreateNode(db, n); err != nil {
		return errx.Wrap(ErrCodeGpuStore, "创建节点失败", err)
	}
	return nil
}

// UpdateNode 更新节点
func (s *GpuService) UpdateNode(n *model.GpuNode) error {
	if n == nil || n.ID == 0 {
		return errx.New(ErrCodeGpuNameRequired, "节点ID不能为空")
	}
	// 缩容校验须与 StartInstance 同一把行锁互斥：原 SUM 为无锁读，校验与更新之间存在窗口——
	// 并发开通插入新实例并提交后，仍可把总量改到低于实际已用量，打破防超卖不变量
	err := s.dbh().Transaction(func(tx *gorm.DB) error {
		row, err := s.repo.FindNodeForUpdate(tx, n.ID)
		if err != nil {
			return errx.New(ErrCodeGpuNotFound, "节点不存在")
		}
		usage, err := s.repo.NodeUsage(tx, n.ID)
		if err != nil {
			return errx.Wrap(ErrCodeGpuStore, "查询节点占用失败", err)
		}
		if n.GpuTotal < usage.Gpu || n.CpuTotal < usage.Cpu || n.MemTotalGb < usage.Mem {
			return errx.New(ErrCodeGpuOversold, "总量不可低于运行中实例的已分配量")
		}
		return s.repo.UpdateNodeColumns(tx, row.ID, map[string]any{
			"name": n.Name, "gpu_model": n.GpuModel, "gpu_total": n.GpuTotal,
			"cpu_total": n.CpuTotal, "mem_total_gb": n.MemTotalGb,
			"status": n.Status, "notes": n.Notes,
		})
	})
	return err
}

// DeleteNode 删除节点（有运行中实例拒绝）
func (s *GpuService) DeleteNode(id uint) error {
	db := s.dbh()
	running, err := s.repo.RunningInstanceCount(db, id)
	if err != nil {
		return errx.Wrap(ErrCodeGpuStore, "查询运行中实例失败", err)
	}
	if running > 0 {
		return errx.New(ErrCodeGpuOversold, fmt.Sprintf("仍有 %d 个运行中实例，不可删除", running))
	}
	if err := s.repo.DeleteNode(db, id); err != nil {
		return errx.Wrap(ErrCodeGpuStore, "删除节点失败", err)
	}
	return nil
}

// NodeWithUsage 节点列表行（含余量）
type NodeWithUsage struct {
	model.GpuNode
	GpuUsed int `json:"gpuUsed"`
	CpuUsed int `json:"cpuUsed"`
	MemUsed int `json:"memUsed"`
}

// GetNodeList 节点列表（含余量计算）
func (s *GpuService) GetNodeList() ([]NodeWithUsage, error) {
	db := s.dbh()
	nodes, err := s.repo.NodeList(db)
	if err != nil {
		return nil, errx.Wrap(ErrCodeGpuStore, "查询节点列表失败", err)
	}
	// 一次 GROUP BY 批量取占用（原逐节点查询的 N+1）
	ids := make([]uint, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	usage, err := s.repo.NodeUsageBatch(db, ids)
	if err != nil {
		return nil, errx.Wrap(ErrCodeGpuStore, "查询节点占用聚合失败", err)
	}
	out := make([]NodeWithUsage, 0, len(nodes))
	for _, n := range nodes {
		u := usage[n.ID]
		out = append(out, NodeWithUsage{GpuNode: n, GpuUsed: u.Gpu, CpuUsed: u.Cpu, MemUsed: u.Mem})
	}
	return out, nil
}

// CreateSpec 创建规格
func (s *GpuService) CreateSpec(sp *model.GpuSpec) error {
	if strings.TrimSpace(sp.Name) == "" || sp.GpuCount <= 0 {
		return errx.New(ErrCodeGpuNameRequired, "规格名称与 GPU 卡数必填")
	}
	db := s.dbh()
	count, err := s.repo.SpecCountByName(db, sp.Name)
	if err != nil {
		return errx.Wrap(ErrCodeGpuStore, "查询规格失败", err)
	}
	if count > 0 {
		return errx.New(ErrCodeGpuDuplicate, "规格已存在: "+sp.Name)
	}
	if err := s.repo.CreateSpec(db, sp); err != nil {
		return errx.Wrap(ErrCodeGpuStore, "创建规格失败", err)
	}
	return nil
}

// DeleteSpec 删除规格
func (s *GpuService) DeleteSpec(id uint) error {
	if err := s.repo.DeleteSpec(s.dbh(), id); err != nil {
		return errx.Wrap(ErrCodeGpuStore, "删除规格失败", err)
	}
	return nil
}

// GetSpecList 规格列表
func (s *GpuService) GetSpecList() ([]model.GpuSpec, error) {
	list, err := s.repo.SpecList(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeGpuStore, "查询规格列表失败", err)
	}
	return list, nil
}

// StartInstance 开通实例：事务内行锁节点 → 防超卖校验 → 台账落库
// 容器实际创建联动 container 插件（后续场次），当前为资源分配台账模式。
func (s *GpuService) StartInstance(nodeID, specID uint, name, creator string, userID uint) (*model.GpuInstance, error) {
	db := s.dbh()
	spec, err := s.repo.FindSpec(db, specID)
	if err != nil {
		return nil, errx.New(ErrCodeGpuNotFound, "规格不存在")
	}
	if name == "" {
		name = fmt.Sprintf("gpu-%s-%d", spec.Name, time.Now().Unix()%100000)
	}
	var inst *model.GpuInstance
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 行锁节点（防并发超卖的关键）——gorm:query_option 是 jinzhu/gorm v1 语法，
		// gorm.io/gorm 下被静默忽略（锁根本没加），必须走 clause.Locking（repo 内 WithRowLock）
		node, err := s.repo.FindNodeForUpdate(tx, nodeID)
		if err != nil {
			return errx.New(ErrCodeGpuNotFound, "节点不存在")
		}
		if node.Status != model.NodeOnline {
			return errx.New(ErrCodeGpuNodeOffline, "节点离线，不可开通")
		}
		usage, err := s.repo.NodeUsage(tx, node.ID)
		if err != nil {
			return errx.Wrap(ErrCodeGpuStore, "查询节点占用失败", err)
		}
		if !canAllocate(node.GpuTotal, usage.Gpu, spec.GpuCount) ||
			!canAllocate(node.CpuTotal, usage.Cpu, spec.CpuCores) ||
			!canAllocate(node.MemTotalGb, usage.Mem, spec.MemGb) {
			return errx.New(ErrCodeGpuOversold,
				fmt.Sprintf("节点 %s 资源不足：GPU 余 %d 需 %d，CPU 余 %d 需 %d，内存余 %dG 需 %dG",
					node.Name, node.GpuTotal-usage.Gpu, spec.GpuCount,
					node.CpuTotal-usage.Cpu, spec.CpuCores,
					node.MemTotalGb-usage.Mem, spec.MemGb))
		}
		inst = &model.GpuInstance{
			NodeID: node.ID, NodeName: node.Name,
			SpecID: spec.ID, SpecName: spec.Name,
			Name:         name,
			GpuAllocated: spec.GpuCount, CpuAllocated: spec.CpuCores, MemAllocGb: spec.MemGb,
			Status: model.InstRunning, Creator: creator, UserID: userID,
		}
		return s.repo.CreateInstance(tx, inst)
	}); err != nil {
		return nil, err
	}
	return inst, nil
}

// ReleaseInstance 销毁实例：标记状态释放配额
func (s *GpuService) ReleaseInstance(id uint) error {
	db := s.dbh()
	inst, err := s.repo.FindInstance(db, id)
	if err != nil {
		return errx.New(ErrCodeGpuNotFound, "实例不存在")
	}
	if inst.Status != model.InstRunning {
		return errx.New(ErrCodeGpuStructInvalid, "实例非运行中状态")
	}
	// 条件更新替代检查后更新：并发重复释放（双击/重试）都过状态检查时，
	// 后提交者会覆盖前者的 released_at 致审计时间戳失真；携带状态条件天然幂等
	affected, err := s.repo.ReleaseInstanceIfRunning(db, id, time.Now())
	if err != nil {
		return errx.Wrap(ErrCodeGpuStore, "销毁实例失败", err)
	}
	if affected == 0 {
		return errx.New(ErrCodeGpuStructInvalid, "实例非运行中状态（可能已被并发释放）")
	}
	return nil
}

// GetInstanceList 实例台账分页（nodeId/userId/status 过滤；page<=0 视为全量兼容旧调用，上限 500）
// 原硬编码 Limit(500) 无分页无过滤：台账含已销毁记录持续增长，第 501 条起静默不可见
func (s *GpuService) GetInstanceList(status string, nodeID, userID uint, page, pageSize int) (list []model.GpuInstance, total int64, err error) {
	list, total, err = s.repo.InstanceList(s.dbh(), repo.InstanceFilter{
		Status: status, NodeID: nodeID, UserID: userID, Page: page, Size: pageSize,
	})
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeGpuStore, "查询实例台账失败", err)
	}
	return list, total, nil
}
