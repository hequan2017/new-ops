// Package repo 白泽公告数据访问层：只做查询与写库，业务语义（分页特例、数据源口径）在 service。
// 方法首参 db 为当前会话（全局库或事务 tx），由 service 决定传入哪一个。
package repo

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/announcement/model"
	"github.com/hequan2017/new-ops/server/plugin/announcement/model/request"
)

// Repo 公告数据访问（无状态，可零值使用）
type Repo struct{}

// Create 落库公告
func (Repo) Create(db *gorm.DB, info *model.Info) error {
	return db.Create(info).Error
}

// Delete 按 ID 删除公告
func (Repo) Delete(db *gorm.DB, id string) error {
	return db.Delete(&model.Info{}, "id = ?", id).Error
}

// DeleteByIds 批量删除公告
func (Repo) DeleteByIds(db *gorm.DB, ids []string) error {
	return db.Delete(&[]model.Info{}, "id in ?", ids).Error
}

// Update 更新公告
func (Repo) Update(db *gorm.DB, info model.Info) error {
	return db.Model(&model.Info{}).Where("id = ?", info.ID).Updates(&info).Error
}

// FindByID 按 ID 取公告
func (Repo) FindByID(db *gorm.DB, id string) (*model.Info, error) {
	var row model.Info
	err := db.Where("id = ?", id).First(&row).Error
	return &row, err
}

// List 分页列表（口径冻结：PageSize==0 视为全量不分页——模板既有约定，data.Page 的
// page<=0→上限 500 语义与此不同，故保留本地实现；StartCreatedAt/EndCreatedAt 范围过滤）
func (Repo) List(db *gorm.DB, search request.InfoSearch) ([]model.Info, int64, error) {
	limit := search.PageSize
	offset := search.PageSize * (search.Page - 1)
	q := db.Model(&model.Info{})
	if search.StartCreatedAt != nil && search.EndCreatedAt != nil {
		q = q.Where("created_at BETWEEN ? AND ?", search.StartCreatedAt, search.EndCreatedAt)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit != 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var list []model.Info
	if err := q.Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UserOptions 数据源查询：sys_users 昵称选项（label/value）
func (Repo) UserOptions(db *gorm.DB) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := db.Table("sys_users").
		Select("nick_name as label,id as value").
		Scan(&rows).Error
	return rows, err
}
