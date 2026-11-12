// Package repo policyreg 数据访问层：只做查询与写库，业务规则（白名单/脱敏/预标/ACL 判定）在 service。
// 方法首参 db 为当前会话（全局库或事务 tx），由 service 决定传入哪一个。
// 分页口径冻结：本插件列表上限 200（区别于 data.Page 的 100 钳位）与检索扫描上限
// 1000/5000 为既有约定，保留本地实现。
package repo

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// Repo policyreg 数据访问（无状态，可零值使用）
type Repo struct{}

// ---------- 制度文档 ----------

// FindDoc 取制度（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindDoc(db *gorm.DB, id uint) (*model.PolicyDoc, error) {
	var row model.PolicyDoc
	err := db.First(&row, id).Error
	return &row, err
}

// FindDocByContentHash 按幂等键取制度
func (Repo) FindDocByContentHash(db *gorm.DB, hash string) (*model.PolicyDoc, error) {
	var row model.PolicyDoc
	err := db.Where("content_hash = ?", hash).First(&row).Error
	return &row, err
}

// CreateDoc 落库制度
func (Repo) CreateDoc(db *gorm.DB, in *model.PolicyDoc) error {
	return db.Create(in).Error
}

// DocList 制度列表（格式/密级过滤，倒序，page/pageSize 已由 service 归一）
func (Repo) DocList(db *gorm.DB, format, securityLevel string, page, pageSize int) ([]model.PolicyDoc, int64, error) {
	q := db.Model(&model.PolicyDoc{})
	if format != "" {
		q = q.Where("format = ?", format)
	}
	if securityLevel != "" {
		q = q.Where("security_level = ?", securityLevel)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PolicyDoc
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DocsForSearch 检索扫描集（关键字过滤，倒序，上限 1000——ACL 过滤在此集合内存进行）
func (Repo) DocsForSearch(db *gorm.DB, keyword string) ([]model.PolicyDoc, error) {
	q := db.Model(&model.PolicyDoc{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var list []model.PolicyDoc
	err := q.Order("id DESC").Limit(1000).Find(&list).Error
	return list, err
}

// ---------- 分词标注 ----------

// FindTermTagByDoc 按制度取标注
func (Repo) FindTermTagByDoc(db *gorm.DB, docId uint) (*model.PolicyTermTag, error) {
	var row model.PolicyTermTag
	err := db.Where("doc_id = ?", docId).First(&row).Error
	return &row, err
}

// UpdateTermTagColumns 更新标注指定列
func (Repo) UpdateTermTagColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.PolicyTermTag{}).Where("id = ?", id).Updates(cols).Error
}

// CreateTermTag 落库标注
func (Repo) CreateTermTag(db *gorm.DB, tag *model.PolicyTermTag) error {
	return db.Create(tag).Error
}

// TermTagListByDoc 制度标注列表（id 升序，上限 100）
func (Repo) TermTagListByDoc(db *gorm.DB, docId uint) ([]model.PolicyTermTag, error) {
	var list []model.PolicyTermTag
	err := db.Where("doc_id = ?", docId).Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// TermTagsAll 全量标注扫描（统计聚合，上限 5000）
func (Repo) TermTagsAll(db *gorm.DB) ([]model.PolicyTermTag, error) {
	var list []model.PolicyTermTag
	err := db.Order("id ASC").Limit(5000).Find(&list).Error
	return list, err
}

// ---------- 实体 ----------

// DeleteEntitiesByDoc 清空制度实体（幂等先清）
func (Repo) DeleteEntitiesByDoc(db *gorm.DB, docId uint) error {
	return db.Where("doc_id = ?", docId).Delete(&model.PolicyEntity{}).Error
}

// CreateEntity 落库实体
func (Repo) CreateEntity(db *gorm.DB, e *model.PolicyEntity) error {
	return db.Create(e).Error
}

// EntityList 实体列表（类型过滤，id 升序，上限 500）
func (Repo) EntityList(db *gorm.DB, docId uint, eType string) ([]model.PolicyEntity, error) {
	q := db.Model(&model.PolicyEntity{}).Where("doc_id = ?", docId)
	if eType != "" {
		q = q.Where("e_type = ?", eType)
	}
	var list []model.PolicyEntity
	err := q.Order("id ASC").Limit(500).Find(&list).Error
	return list, err
}

// ---------- 标注（token/relation） ----------

// DeleteTokenLabelsByDoc 清空分词标注（幂等先清）
func (Repo) DeleteTokenLabelsByDoc(db *gorm.DB, docId uint) error {
	return db.Where("doc_id = ?", docId).Delete(&model.PolicyTokenLabel{}).Error
}

// DeleteRelationLabelsByDoc 清空关系标注（幂等先清）
func (Repo) DeleteRelationLabelsByDoc(db *gorm.DB, docId uint) error {
	return db.Where("doc_id = ?", docId).Delete(&model.PolicyRelationLabel{}).Error
}

// CreateTokenLabel 落库分词标注
func (Repo) CreateTokenLabel(db *gorm.DB, l *model.PolicyTokenLabel) error {
	return db.Create(l).Error
}

// CreateRelationLabel 落库关系标注
func (Repo) CreateRelationLabel(db *gorm.DB, l *model.PolicyRelationLabel) error {
	return db.Create(l).Error
}

// FindTokenLabel 取分词标注（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindTokenLabel(db *gorm.DB, id uint) (*model.PolicyTokenLabel, error) {
	var row model.PolicyTokenLabel
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateTokenLabelColumns 更新分词标注指定列
func (Repo) UpdateTokenLabelColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.PolicyTokenLabel{}).Where("id = ?", id).Updates(cols).Error
}

// TokenLabelListByDoc 分词标注列表（seq 升序，上限 2000）
func (Repo) TokenLabelListByDoc(db *gorm.DB, docId uint) ([]model.PolicyTokenLabel, error) {
	var list []model.PolicyTokenLabel
	err := db.Where("doc_id = ?", docId).Order("seq ASC").Limit(2000).Find(&list).Error
	return list, err
}

// RelationLabelListByDoc 关系标注列表（id 升序，上限 500）
func (Repo) RelationLabelListByDoc(db *gorm.DB, docId uint) ([]model.PolicyRelationLabel, error) {
	var list []model.PolicyRelationLabel
	err := db.Where("doc_id = ?", docId).Order("id ASC").Limit(500).Find(&list).Error
	return list, err
}

// ---------- 鉴权 grant ----------

// GrantsAll 全量 grant 扫描（检索角色过滤，上限 500）
func (Repo) GrantsAll(db *gorm.DB) ([]model.PolicyPermGrant, error) {
	var list []model.PolicyPermGrant
	err := db.Limit(500).Find(&list).Error
	return list, err
}

// GrantsByRole 角色 grant 列表（id 升序，上限 200）
func (Repo) GrantsByRole(db *gorm.DB, roleName string) ([]model.PolicyPermGrant, error) {
	q := db.Model(&model.PolicyPermGrant{})
	if roleName != "" {
		q = q.Where("role_name = ?", roleName)
	}
	var list []model.PolicyPermGrant
	err := q.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// FindGrantByRoleCat 按角色×分类取 grant（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindGrantByRoleCat(db *gorm.DB, roleName, docCategory string) (*model.PolicyPermGrant, error) {
	var row model.PolicyPermGrant
	err := db.Where("role_name = ? AND doc_category = ?", roleName, docCategory).First(&row).Error
	return &row, err
}

// UpdateGrantColumns 更新 grant 指定列
func (Repo) UpdateGrantColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.PolicyPermGrant{}).Where("id = ?", id).Updates(cols).Error
}

// CreateGrant 落库 grant
func (Repo) CreateGrant(db *gorm.DB, in *model.PolicyPermGrant) error {
	return db.Create(in).Error
}

// ---------- 提示词与问答 ----------

// FindTemplateByName 取模板（不限启用态，记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindTemplateByName(db *gorm.DB, name string) (*model.PolicyPromptTemplate, error) {
	var row model.PolicyPromptTemplate
	err := db.Where("name = ?", name).First(&row).Error
	return &row, err
}

// FindEnabledTemplateByName 取启用态模板
func (Repo) FindEnabledTemplateByName(db *gorm.DB, name string) (*model.PolicyPromptTemplate, error) {
	var row model.PolicyPromptTemplate
	err := db.Where("name = ? AND enabled = ?", name, true).First(&row).Error
	return &row, err
}

// UpdateTemplateColumns 更新模板指定列
func (Repo) UpdateTemplateColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.PolicyPromptTemplate{}).Where("id = ?", id).Updates(cols).Error
}

// CreateTemplate 落库模板
func (Repo) CreateTemplate(db *gorm.DB, in *model.PolicyPromptTemplate) error {
	return db.Create(in).Error
}

// TemplateList 模板列表（id 升序，上限 100）
func (Repo) TemplateList(db *gorm.DB) ([]model.PolicyPromptTemplate, error) {
	var list []model.PolicyPromptTemplate
	err := db.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// CreateQaLog 落库问答日志
func (Repo) CreateQaLog(db *gorm.DB, log *model.PolicyQaLog) error {
	return db.Create(log).Error
}

// QaLogList 问答日志列表（角色过滤，倒序，page/pageSize 已由 service 归一）
func (Repo) QaLogList(db *gorm.DB, roleName string, page, pageSize int) ([]model.PolicyQaLog, int64, error) {
	q := db.Model(&model.PolicyQaLog{})
	if roleName != "" {
		q = q.Where("role_name = ?", roleName)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PolicyQaLog
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
