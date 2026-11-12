// Package repo 白泽工单引擎数据访问层：只做查询与写库；
// 业务规则（定义校验/状态机/乐观锁决策/审批约束）留在 service。
// 方法首参 db 为当前会话（全局库或事务 tx），由 service 决定传入哪一个。
package repo

import (
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
)

// Repo 工单数据访问（无状态，可零值使用）
type Repo struct{}

// ---------- 定义 ----------

// DefinitionCountByName 按名查定义数（重名校验）
func (Repo) DefinitionCountByName(db *gorm.DB, name string) (int64, error) {
	var count int64
	err := db.Model(&model.WfDefinition{}).Where("name = ?", name).Count(&count).Error
	return count, err
}

// CreateDefinition 落库定义
func (Repo) CreateDefinition(db *gorm.DB, d *model.WfDefinition) error {
	return db.Create(d).Error
}

// RunningInstanceCount 定义下进行中实例数（更新/删除守卫）
func (Repo) RunningInstanceCount(db *gorm.DB, definitionID uint) (int64, error) {
	var count int64
	err := db.Model(&model.WfInstance{}).
		Where("definition_id = ? AND status = ?", definitionID, model.InstanceRunning).Count(&count).Error
	return count, err
}

// UpdateDefinitionColumns 更新定义指定列
func (Repo) UpdateDefinitionColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.WfDefinition{}).Where("id = ?", id).Updates(cols).Error
}

// UpdateDefinitionEnabled 启停开关位（返回受影响行数，0=定义不存在）
func (Repo) UpdateDefinitionEnabled(db *gorm.DB, id uint, enabled bool) (int64, error) {
	res := db.Model(&model.WfDefinition{}).Where("id = ?", id).Update("enabled", enabled)
	return res.RowsAffected, res.Error
}

// DeleteDefinition 删除定义
func (Repo) DeleteDefinition(db *gorm.DB, id uint) error {
	return db.Delete(&model.WfDefinition{}, id).Error
}

// DefinitionList 定义列表（keyword 模糊，id 倒序，上限 200）
func (Repo) DefinitionList(db *gorm.DB, keyword string) ([]*model.WfDefinition, error) {
	q := db.Model(&model.WfDefinition{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var list []*model.WfDefinition
	err := q.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// FindDefinition 取定义（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindDefinition(db *gorm.DB, id uint) (*model.WfDefinition, error) {
	var row model.WfDefinition
	err := db.First(&row, id).Error
	return &row, err
}

// ---------- 实例与流转 ----------

// CreateInstance 落库工单实例
func (Repo) CreateInstance(db *gorm.DB, ins *model.WfInstance) error {
	return db.Create(ins).Error
}

// CreateActionLog 落库流转记录
func (Repo) CreateActionLog(db *gorm.DB, log *model.WfActionLog) error {
	return db.Create(log).Error
}

// FindInstance 取实例（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindInstance(db *gorm.DB, id uint) (*model.WfInstance, error) {
	var row model.WfInstance
	err := db.First(&row, id).Error
	return &row, err
}

// InstanceOwnerCount 实例属主校验（非 888 用户读日志前的归属检查）
func (Repo) InstanceOwnerCount(db *gorm.DB, instanceID, userID uint) (int64, error) {
	var count int64
	err := db.Model(&model.WfInstance{}).
		Where("id = ? AND user_id = ?", instanceID, userID).Count(&count).Error
	return count, err
}

// CancelIfRunning 乐观锁撤回：仅 running 可撤（并发流转/收档竞态防护），
// 返回受影响行数（0=状态已变化）。
func (Repo) CancelIfRunning(db *gorm.DB, instanceID uint, now time.Time) (int64, error) {
	res := db.Model(&model.WfInstance{}).
		Where("id = ? AND status = ?", instanceID, model.InstanceRunning).
		Updates(map[string]any{"status": model.InstanceCanceled, "finished_at": &now})
	return res.RowsAffected, res.Error
}

// TransitionIfState 乐观锁流转：Where 携带读取时的 state+status（两人同时审批/
// 审批与撤回竞态只有一个提交生效，同时防发版钩子重复触发），返回受影响行数。
func (Repo) TransitionIfState(db *gorm.DB, instanceID uint, fromState string, updates map[string]any) (int64, error) {
	res := db.Model(&model.WfInstance{}).
		Where("id = ? AND state = ? AND status = ?", instanceID, fromState, model.InstanceRunning).
		Updates(updates)
	return res.RowsAffected, res.Error
}

// ApproveLogCount 实例的 approve 流转记录数（发版钩子"须经过审批"守卫）
func (Repo) ApproveLogCount(db *gorm.DB, instanceID uint) (int64, error) {
	var count int64
	err := db.Model(&model.WfActionLog{}).
		Where("instance_id = ? AND action = ?", instanceID, "approve").Count(&count).Error
	return count, err
}

// ActionLogList 流转记录（id 升序，上限 100）
func (Repo) ActionLogList(db *gorm.DB, instanceID uint) ([]*model.WfActionLog, error) {
	var list []*model.WfActionLog
	err := db.Where("instance_id = ?", instanceID).Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// InstanceList 工单分页（scoped=false 即 888 全量；page 已由 service 归一 ≥1，
// pageSize 已归一 [1,100] 缺省 10，data.Page 按 page>0 精确分页）
func (Repo) InstanceList(db *gorm.DB, page, pageSize int, creator, status string, userID uint, scoped bool) ([]*model.WfInstance, int64, error) {
	q := db.Model(&model.WfInstance{})
	if scoped {
		q = q.Where("user_id = ?", userID)
	}
	if creator != "" {
		q = q.Where("creator LIKE ?", "%"+creator+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("id DESC")
	var list []*model.WfInstance
	total, err := data.Page(q, page, pageSize, &list)
	return list, total, err
}
