// Package repo ziqiang 数据访问层：只做查询与写库，业务规则（白名单校验/成对校验/
// 链式哈希/bigram 融合排序/模板渲染）留在 service。方法首参 db 为当前会话
// （全局库或事务 tx），由 service 决定传入哪一个；本包禁止 import global。
// 分页口径冻结：与 data.Page（钳位 [1,100]、page<=0 全量上限 500）不等价——
// 本插件列表 page<1 语义为「回到第一页」且上限 200 为既有约定，保留本地实现；
// 检索/核验扫描上限 2000/10000 同为既有口径，保留本地 Limit。
package repo

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// Repo ziqiang 数据访问（无状态，可零值使用）
type Repo struct{}

// normPage 分页归一：page<1 视为第一页（不是全量请求），pageSize 缺省 20、上限 200。
// 口径冻结理由见包注释（与 data.Page 不等价）。
func normPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	return page, pageSize
}

// paged 列表分页共用（Count → 归一分页 Find）；db 传入 Model+Where+Order 链，
// Order 先于 Count 对各主流方言均合法（Count 忽略排序语义）。
func paged[T any](db *gorm.DB, page, pageSize int, dest *[]T) (int64, error) {
	page, pageSize = normPage(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(dest).Error
	return total, err
}

// ---------- 文本鉴权语料与对比样本对 ----------

// CreateTexts 批量落库语料（批大小 200 为既有口径）
func (Repo) CreateTexts(db *gorm.DB, list []model.ZqTextCorpus) error {
	return db.CreateInBatches(list, 200).Error
}

// FindText 取语料（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindText(db *gorm.DB, id uint) (*model.ZqTextCorpus, error) {
	var row model.ZqTextCorpus
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateTextColumns 更新语料指定列
func (Repo) UpdateTextColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ZqTextCorpus{}).Where("id = ?", id).Updates(cols).Error
}

// CreateText 落库语料
func (Repo) CreateText(db *gorm.DB, in *model.ZqTextCorpus) error {
	return db.Create(in).Error
}

// DeleteText 删除语料（软删除，返回影响行数供 service 判不存在）
func (Repo) DeleteText(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ZqTextCorpus{}, id)
	return res.RowsAffected, res.Error
}

// TextList 语料列表（类型/AIGC 过滤，关键字 title/content 双列 LIKE，id 倒序分页；
// aigcFlag nil 表示不过滤，keyword 须由 service 预先 TrimSpace）
func (Repo) TextList(db *gorm.DB, ztype string, aigcFlag *bool, keyword string, page, pageSize int) ([]model.ZqTextCorpus, int64, error) {
	q := db.Model(&model.ZqTextCorpus{})
	if ztype != "" {
		q = q.Where("type = ?", ztype)
	}
	if aigcFlag != nil {
		q = q.Where("aigc_flag = ?", *aigcFlag)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var list []model.ZqTextCorpus
	total, err := paged(q.Order("id DESC"), page, pageSize, &list)
	return list, total, err
}

// CreateComparePair 落库对比样本对
func (Repo) CreateComparePair(db *gorm.DB, in *model.ZqComparePair) error {
	return db.Create(in).Error
}

// ComparePairList 对比样本对列表（id 倒序分页）
func (Repo) ComparePairList(db *gorm.DB, page, pageSize int) ([]model.ZqComparePair, int64, error) {
	var list []model.ZqComparePair
	total, err := paged(db.Model(&model.ZqComparePair{}).Order("id DESC"), page, pageSize, &list)
	return list, total, err
}

// ---------- 图像鉴权语料 ----------

// CreateImageCorpus 落库图像语料
func (Repo) CreateImageCorpus(db *gorm.DB, in *model.ZqImageCorpus) error {
	return db.Create(in).Error
}

// FindImageCorpus 取图像语料（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindImageCorpus(db *gorm.DB, id uint) (*model.ZqImageCorpus, error) {
	var row model.ZqImageCorpus
	err := db.First(&row, id).Error
	return &row, err
}

// ImageCorpusList 图像语料列表（图库/名称关键字过滤，id 倒序分页）
func (Repo) ImageCorpusList(db *gorm.DB, gallery, keyword string, page, pageSize int) ([]model.ZqImageCorpus, int64, error) {
	q := db.Model(&model.ZqImageCorpus{})
	if gallery != "" {
		q = q.Where("gallery = ?", gallery)
	}
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var list []model.ZqImageCorpus
	total, err := paged(q.Order("id DESC"), page, pageSize, &list)
	return list, total, err
}

// ImageHashScan 已计算 aHash 的图像语料扫描（问答图像维度候选集，上限 2000，id 升序）
func (Repo) ImageHashScan(db *gorm.DB) ([]model.ZqImageCorpus, error) {
	var list []model.ZqImageCorpus
	err := db.Model(&model.ZqImageCorpus{}).
		Where("hash_hex <> ''").Order("id ASC").Limit(2000).Find(&list).Error
	return list, err
}

// ---------- 实体图谱 ----------

// FindEntity 取实体（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindEntity(db *gorm.DB, id uint) (*model.ZqEntity, error) {
	var row model.ZqEntity
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateEntityColumns 更新实体指定列
func (Repo) UpdateEntityColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ZqEntity{}).Where("id = ?", id).Updates(cols).Error
}

// CreateEntity 落库实体
func (Repo) CreateEntity(db *gorm.DB, in *model.ZqEntity) error {
	return db.Create(in).Error
}

// DeleteEntity 删除实体（软删除，返回影响行数）
func (Repo) DeleteEntity(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ZqEntity{}, id)
	return res.RowsAffected, res.Error
}

// EntityList 实体扫描（分类过滤，id 升序；全图/BFS 以 limit=1000 复用）
func (Repo) EntityList(db *gorm.DB, category string, limit int) ([]model.ZqEntity, error) {
	q := db.Model(&model.ZqEntity{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var list []model.ZqEntity
	err := q.Order("id").Limit(limit).Find(&list).Error
	return list, err
}

// CreateEdge 落库关系边
func (Repo) CreateEdge(db *gorm.DB, in *model.ZqEdge) error {
	return db.Create(in).Error
}

// DeleteEdge 删除关系边（返回影响行数）
func (Repo) DeleteEdge(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ZqEdge{}, id)
	return res.RowsAffected, res.Error
}

// EdgeList 关系边扫描（id 升序；全图/BFS 以 limit=5000 复用）
func (Repo) EdgeList(db *gorm.DB, limit int) ([]model.ZqEdge, error) {
	var list []model.ZqEdge
	err := db.Model(&model.ZqEdge{}).Order("id").Limit(limit).Find(&list).Error
	return list, err
}

// ---------- 知识库（法规/判例） ----------

// FindRegulation 取法规（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindRegulation(db *gorm.DB, id uint) (*model.ZqRegulation, error) {
	var row model.ZqRegulation
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateRegulationColumns 更新法规指定列
func (Repo) UpdateRegulationColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ZqRegulation{}).Where("id = ?", id).Updates(cols).Error
}

// CreateRegulation 落库法规
func (Repo) CreateRegulation(db *gorm.DB, in *model.ZqRegulation) error {
	return db.Create(in).Error
}

// DeleteRegulation 删除法规（软删除，返回影响行数）
func (Repo) DeleteRegulation(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ZqRegulation{}, id)
	return res.RowsAffected, res.Error
}

// RegulationScan 法规检索扫描集（辖区过滤，id 升序，上限 2000——bigram 打分在此集合内存进行）
func (Repo) RegulationScan(db *gorm.DB, jurisdiction string) ([]model.ZqRegulation, error) {
	q := db.Model(&model.ZqRegulation{})
	if jurisdiction != "" {
		q = q.Where("jurisdiction = ?", jurisdiction)
	}
	var list []model.ZqRegulation
	err := q.Order("id ASC").Limit(2000).Find(&list).Error
	return list, err
}

// RegulationList 法规列表（辖区/分类过滤，id 倒序分页）
func (Repo) RegulationList(db *gorm.DB, jurisdiction, category string, page, pageSize int) ([]model.ZqRegulation, int64, error) {
	q := db.Model(&model.ZqRegulation{})
	if jurisdiction != "" {
		q = q.Where("jurisdiction = ?", jurisdiction)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var list []model.ZqRegulation
	total, err := paged(q.Order("id DESC"), page, pageSize, &list)
	return list, total, err
}

// CreateCase 落库判例
func (Repo) CreateCase(db *gorm.DB, in *model.ZqCase) error {
	return db.Create(in).Error
}

// CaseScan 判例检索扫描集（辖区过滤，id 升序，上限 2000）
func (Repo) CaseScan(db *gorm.DB, jurisdiction string) ([]model.ZqCase, error) {
	q := db.Model(&model.ZqCase{})
	if jurisdiction != "" {
		q = q.Where("jurisdiction = ?", jurisdiction)
	}
	var list []model.ZqCase
	err := q.Order("id ASC").Limit(2000).Find(&list).Error
	return list, err
}

// CaseList 判例列表（辖区过滤，id 倒序分页）
func (Repo) CaseList(db *gorm.DB, jurisdiction string, page, pageSize int) ([]model.ZqCase, int64, error) {
	q := db.Model(&model.ZqCase{})
	if jurisdiction != "" {
		q = q.Where("jurisdiction = ?", jurisdiction)
	}
	var list []model.ZqCase
	total, err := paged(q.Order("id DESC"), page, pageSize, &list)
	return list, total, err
}

// ---------- 区块链存证 ----------

// LatestEvidence 取当前最高块（空链返回 gorm.ErrRecordNotFound，service 视为零值→创世块）
func (Repo) LatestEvidence(db *gorm.DB, out *model.ZqEvidence) error {
	return db.Model(&model.ZqEvidence{}).Order("height DESC").First(out).Error
}

// CreateEvidence 落库存证块
func (Repo) CreateEvidence(db *gorm.DB, in *model.ZqEvidence) error {
	return db.Create(in).Error
}

// UpdateEvidenceColumns 更新存证块指定列（仅测试/运维修正场景使用）
func (Repo) UpdateEvidenceColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ZqEvidence{}).Where("id = ?", id).Updates(cols).Error
}

// EvidenceChainScan 存证链扫描（bizKey 非空时仅该业务块，块高升序，上限 10000）
func (Repo) EvidenceChainScan(db *gorm.DB, bizKey string) ([]model.ZqEvidence, error) {
	q := db.Model(&model.ZqEvidence{})
	if bizKey != "" {
		q = q.Where("biz_key = ?", bizKey)
	}
	var list []model.ZqEvidence
	err := q.Order("height ASC").Limit(10000).Find(&list).Error
	return list, err
}

// CountEvidence 存证块总量
func (Repo) CountEvidence(db *gorm.DB) (int64, error) {
	var total int64
	err := db.Model(&model.ZqEvidence{}).Count(&total).Error
	return total, err
}

// EvidenceList 存证列表（业务号过滤，块高升序分页）
func (Repo) EvidenceList(db *gorm.DB, bizKey string, page, pageSize int) ([]model.ZqEvidence, int64, error) {
	q := db.Model(&model.ZqEvidence{})
	if bizKey != "" {
		q = q.Where("biz_key = ?", bizKey)
	}
	var list []model.ZqEvidence
	total, err := paged(q.Order("height ASC"), page, pageSize, &list)
	return list, total, err
}

// FindEvidence 取存证块（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindEvidence(db *gorm.DB, id uint) (*model.ZqEvidence, error) {
	var row model.ZqEvidence
	err := db.First(&row, id).Error
	return &row, err
}

// ---------- 提示词模板 ----------

// FindPromptTemplate 取模板（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindPromptTemplate(db *gorm.DB, id uint) (*model.ZqPromptTemplate, error) {
	var row model.ZqPromptTemplate
	err := db.First(&row, id).Error
	return &row, err
}

// FindPromptTemplateByName 按模板名取模板
func (Repo) FindPromptTemplateByName(db *gorm.DB, name string) (*model.ZqPromptTemplate, error) {
	var row model.ZqPromptTemplate
	err := db.Where("name = ?", name).First(&row).Error
	return &row, err
}

// UpdatePromptTemplateColumns 更新模板指定列
func (Repo) UpdatePromptTemplateColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ZqPromptTemplate{}).Where("id = ?", id).Updates(cols).Error
}

// CreatePromptTemplate 落库模板
func (Repo) CreatePromptTemplate(db *gorm.DB, in *model.ZqPromptTemplate) error {
	return db.Create(in).Error
}

// DeletePromptTemplate 删除模板（软删除，返回影响行数）
func (Repo) DeletePromptTemplate(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ZqPromptTemplate{}, id)
	return res.RowsAffected, res.Error
}

// PromptTemplateList 模板列表（场景过滤，场景+步号+id 升序，上限 500）
func (Repo) PromptTemplateList(db *gorm.DB, scene string) ([]model.ZqPromptTemplate, error) {
	q := db.Model(&model.ZqPromptTemplate{})
	if scene != "" {
		q = q.Where("scene = ?", scene)
	}
	var list []model.ZqPromptTemplate
	err := q.Order("scene, step, id").Limit(500).Find(&list).Error
	return list, err
}

// ---------- 知识产权问答 ----------

// PatentScan 专利语料扫描（问答文本维度候选集，id 升序，上限 2000）
func (Repo) PatentScan(db *gorm.DB) ([]model.ZqTextCorpus, error) {
	var list []model.ZqTextCorpus
	err := db.Model(&model.ZqTextCorpus{}).
		Where("type = ?", "专利").Order("id ASC").Limit(2000).Find(&list).Error
	return list, err
}
