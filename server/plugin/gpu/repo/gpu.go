// Package repo 白泽 GPU 算力数据访问层：只做查询与写库；
// 业务规则（校验/防超卖判定/状态机/事务与行锁编排）留在 service。
// 方法首参 db 为当前会话（全局库或事务 tx），由 service 决定传入哪一个。
package repo

import (
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
)

// Repo GPU 数据访问（无状态，可零值使用）
type Repo struct{}

// NodeUsage 节点占用聚合结果（运行中实例三列汇总）
type NodeUsage struct {
	Gpu int
	Cpu int
	Mem int
}

// nodeUsageAgg GROUP BY 扫描行（单节点与批量共用列口径）
type nodeUsageAgg struct {
	NodeID uint
	Gpu    int
	Cpu    int
	Mem    int
}

const usageSelect = "COALESCE(SUM(gpu_allocated),0) as gpu, COALESCE(SUM(cpu_allocated),0) as cpu, COALESCE(SUM(mem_alloc_gb),0) as mem"

// ---------- 节点 ----------

// NodeCountByName 按名查节点数（重名校验）
func (Repo) NodeCountByName(db *gorm.DB, name string) (int64, error) {
	var count int64
	err := db.Model(&model.GpuNode{}).Where("name = ?", name).Count(&count).Error
	return count, err
}

// CreateNode 落库节点
func (Repo) CreateNode(db *gorm.DB, n *model.GpuNode) error {
	return db.Create(n).Error
}

// FindNodeForUpdate 行锁取节点（防超卖互斥关键，sqlite 单测豁免锁子句）
func (Repo) FindNodeForUpdate(db *gorm.DB, id uint) (*model.GpuNode, error) {
	var row model.GpuNode
	err := data.WithRowLock(db).First(&row, id).Error
	return &row, err
}

// UpdateNodeColumns 更新节点指定列
func (Repo) UpdateNodeColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.GpuNode{}).Where("id = ?", id).Updates(cols).Error
}

// DeleteNode 删除节点
func (Repo) DeleteNode(db *gorm.DB, id uint) error {
	return db.Delete(&model.GpuNode{}, id).Error
}

// NodeList 节点全量（id 倒序）
func (Repo) NodeList(db *gorm.DB) ([]model.GpuNode, error) {
	var list []model.GpuNode
	err := db.Order("id DESC").Find(&list).Error
	return list, err
}

// RunningInstanceCount 节点运行中实例数（删除前守卫）
func (Repo) RunningInstanceCount(db *gorm.DB, nodeID uint) (int64, error) {
	var count int64
	err := db.Model(&model.GpuInstance{}).
		Where("node_id = ? AND status = ?", nodeID, model.InstRunning).Count(&count).Error
	return count, err
}

// ---------- 占用聚合 ----------

// NodeUsage 单节点运行中占用量
func (Repo) NodeUsage(db *gorm.DB, nodeID uint) (NodeUsage, error) {
	var a nodeUsageAgg
	err := db.Model(&model.GpuInstance{}).
		Select(usageSelect).
		Where("node_id = ? AND status = ?", nodeID, model.InstRunning).
		Scan(&a).Error
	return NodeUsage{Gpu: a.Gpu, Cpu: a.Cpu, Mem: a.Mem}, err
}

// NodeUsageBatch 多节点占用聚合（一次 GROUP BY，避免列表 N+1）
func (Repo) NodeUsageBatch(db *gorm.DB, nodeIDs []uint) (map[uint]NodeUsage, error) {
	out := make(map[uint]NodeUsage, len(nodeIDs))
	if len(nodeIDs) == 0 {
		return out, nil
	}
	var rows []nodeUsageAgg
	err := db.Model(&model.GpuInstance{}).
		Select("node_id, "+usageSelect).
		Where("node_id IN ? AND status = ?", nodeIDs, model.InstRunning).
		Group("node_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.NodeID] = NodeUsage{Gpu: r.Gpu, Cpu: r.Cpu, Mem: r.Mem}
	}
	return out, nil
}

// ---------- 规格 ----------

// SpecCountByName 按名查规格数（重名校验）
func (Repo) SpecCountByName(db *gorm.DB, name string) (int64, error) {
	var count int64
	err := db.Model(&model.GpuSpec{}).Where("name = ?", name).Count(&count).Error
	return count, err
}

// CreateSpec 落库规格
func (Repo) CreateSpec(db *gorm.DB, sp *model.GpuSpec) error {
	return db.Create(sp).Error
}

// DeleteSpec 删除规格
func (Repo) DeleteSpec(db *gorm.DB, id uint) error {
	return db.Delete(&model.GpuSpec{}, id).Error
}

// SpecList 规格全量（id 倒序）
func (Repo) SpecList(db *gorm.DB) ([]model.GpuSpec, error) {
	var list []model.GpuSpec
	err := db.Order("id DESC").Find(&list).Error
	return list, err
}

// FindSpec 取规格（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindSpec(db *gorm.DB, id uint) (*model.GpuSpec, error) {
	var row model.GpuSpec
	err := db.First(&row, id).Error
	return &row, err
}

// ---------- 实例 ----------

// InstanceFilter 实例台账过滤条件（status/nodeId/userId + 分页；page<=0 全量上限 500）
type InstanceFilter struct {
	Status string
	NodeID uint
	UserID uint
	Page   int
	Size   int
}

// InstanceList 台账分页（data.Page 统一口径）
func (Repo) InstanceList(db *gorm.DB, f InstanceFilter) ([]model.GpuInstance, int64, error) {
	q := db.Model(&model.GpuInstance{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.NodeID > 0 {
		q = q.Where("node_id = ?", f.NodeID)
	}
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	q = q.Order("id DESC")
	var list []model.GpuInstance
	total, err := data.Page(q, f.Page, f.Size, &list)
	return list, total, err
}

// CreateInstance 落库实例（须在已行锁节点的调用方事务内）
func (Repo) CreateInstance(db *gorm.DB, inst *model.GpuInstance) error {
	return db.Create(inst).Error
}

// FindInstance 取实例（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindInstance(db *gorm.DB, id uint) (*model.GpuInstance, error) {
	var row model.GpuInstance
	err := db.First(&row, id).Error
	return &row, err
}

// ReleaseInstanceIfRunning 条件释放：仅当仍为运行中时置已销毁（并发幂等），
// 返回受影响行数（0=已被并发释放）。
func (Repo) ReleaseInstanceIfRunning(db *gorm.DB, id uint, now time.Time) (int64, error) {
	res := db.Model(&model.GpuInstance{}).
		Where("id = ? AND status = ?", id, model.InstRunning).
		Updates(map[string]any{"status": model.InstReleased, "released_at": &now})
	return res.RowsAffected, res.Error
}
