// Package repo fanhua 数据访问层：只做查询与写库，业务规则（白名单/健康度/加权树/情感规则）
// 留在 service。方法首参 db 为当前会话（全局库或事务 tx），由 service 决定传入哪一个。
//
// 分页口径冻结：本插件列表为 page 下限 1（无全量模式）、pageSize 钳位 [1,200] 缺省 20
// （区别于 data.Page 的 page<=0 全量上限 500 / pageSize 钳位 100），语义不等价，
// 保留本地 Offset/Limit 实现；page/pageSize 已由 service 归一。
package repo

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// Repo fanhua 数据访问（无状态，可零值使用）
type Repo struct{}

// ---------- 共享词表 ----------

// EnabledWords 启用态敏感词扫描（价值观审查/健康度/问答前置共用，上限 5000）
func (Repo) EnabledWords(db *gorm.DB) ([]model.FanhuaSensitiveWord, error) {
	var words []model.FanhuaSensitiveWord
	err := db.Where("enabled = ?", true).Limit(5000).Find(&words).Error
	return words, err
}

// ---------- 语料 ----------

// CreateCorporaBatch 批量落库语料
func (Repo) CreateCorporaBatch(db *gorm.DB, list []model.FanhuaCorpus, batchSize int) error {
	return db.CreateInBatches(list, batchSize).Error
}

// FindCorpus 取语料（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindCorpus(db *gorm.DB, id uint) (*model.FanhuaCorpus, error) {
	var row model.FanhuaCorpus
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateCorpusColumns 更新语料指定列
func (Repo) UpdateCorpusColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaCorpus{}).Where("id = ?", id).Updates(cols).Error
}

// UpdateCorpusContent 回写清洗后正文
func (Repo) UpdateCorpusContent(db *gorm.DB, corpus *model.FanhuaCorpus, content string) error {
	return db.Model(corpus).Update("content", content).Error
}

// CreateCorpus 落库语料
func (Repo) CreateCorpus(db *gorm.DB, in *model.FanhuaCorpus) error {
	return db.Create(in).Error
}

// DeleteCorpus 删除语料（软删除），返回影响行数供 service 判不存在
func (Repo) DeleteCorpus(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaCorpus{}, id)
	return res.RowsAffected, res.Error
}

// CorpusList 语料列表（来源/语种/关键字过滤，倒序分页，page/pageSize 已由 service 归一）
func (Repo) CorpusList(db *gorm.DB, source, lang, keyword string, page, pageSize int) ([]model.FanhuaCorpus, int64, error) {
	q := db.Model(&model.FanhuaCorpus{})
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if lang != "" {
		q = q.Where("lang = ?", lang)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaCorpus
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CorporaForExport 导出扫描集（来源/语种过滤，倒序，上限 5000）
func (Repo) CorporaForExport(db *gorm.DB, source, lang string) ([]model.FanhuaCorpus, error) {
	q := db.Model(&model.FanhuaCorpus{})
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if lang != "" {
		q = q.Where("lang = ?", lang)
	}
	var list []model.FanhuaCorpus
	err := q.Order("id DESC").Limit(5000).Find(&list).Error
	return list, err
}

// CorporaForLabeling 标注构建扫描集（id 升序，limit 由 service 归一）
func (Repo) CorporaForLabeling(db *gorm.DB, limit int) ([]model.FanhuaCorpus, error) {
	var list []model.FanhuaCorpus
	err := db.Order("id").Limit(limit).Find(&list).Error
	return list, err
}

// ---------- 平行语料 ----------

// ParallelsScan 全量平行语料扫描（检索打分集，上限 5000）
func (Repo) ParallelsScan(db *gorm.DB) ([]model.FanhuaParallel, error) {
	var list []model.FanhuaParallel
	err := db.Limit(5000).Find(&list).Error
	return list, err
}

// FindParallel 取平行语料（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindParallel(db *gorm.DB, id uint) (*model.FanhuaParallel, error) {
	var row model.FanhuaParallel
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateParallelColumns 更新平行语料指定列
func (Repo) UpdateParallelColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaParallel{}).Where("id = ?", id).Updates(cols).Error
}

// CreateParallel 落库平行语料
func (Repo) CreateParallel(db *gorm.DB, in *model.FanhuaParallel) error {
	return db.Create(in).Error
}

// DeleteParallel 删除平行语料（软删除），返回影响行数
func (Repo) DeleteParallel(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaParallel{}, id)
	return res.RowsAffected, res.Error
}

// ParallelList 平行语料列表（关键字过滤，倒序分页）
func (Repo) ParallelList(db *gorm.DB, keyword string, page, pageSize int) ([]model.FanhuaParallel, int64, error) {
	q := db.Model(&model.FanhuaParallel{})
	if keyword != "" {
		q = q.Where("sh LIKE ? OR pu LIKE ? OR en LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaParallel
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ParallelsForExport 导出扫描集（来源过滤，倒序，上限 5000）
func (Repo) ParallelsForExport(db *gorm.DB, source string) ([]model.FanhuaParallel, error) {
	q := db.Model(&model.FanhuaParallel{})
	if source != "" {
		q = q.Where("source = ?", source)
	}
	var list []model.FanhuaParallel
	err := q.Order("id DESC").Limit(5000).Find(&list).Error
	return list, err
}

// ---------- 实体与标注 ----------

// EntitiesScan 全量实体扫描（词典匹配候选集，上限 5000）
func (Repo) EntitiesScan(db *gorm.DB) ([]model.FanhuaEntity, error) {
	var list []model.FanhuaEntity
	err := db.Limit(5000).Find(&list).Error
	return list, err
}

// CreateEntityHit 落库标注记录
func (Repo) CreateEntityHit(db *gorm.DB, rec *model.FanhuaEntityHit) error {
	return db.Create(rec).Error
}

// FindEntity 取实体（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindEntity(db *gorm.DB, id uint) (*model.FanhuaEntity, error) {
	var row model.FanhuaEntity
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateEntityColumns 更新实体指定列
func (Repo) UpdateEntityColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaEntity{}).Where("id = ?", id).Updates(cols).Error
}

// CreateEntity 落库实体
func (Repo) CreateEntity(db *gorm.DB, in *model.FanhuaEntity) error {
	return db.Create(in).Error
}

// DeleteEntity 删除实体（软删除），返回影响行数
func (Repo) DeleteEntity(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaEntity{}, id)
	return res.RowsAffected, res.Error
}

// EntityList 实体列表（分类过滤，分类+id 排序分页）
func (Repo) EntityList(db *gorm.DB, category string, page, pageSize int) ([]model.FanhuaEntity, int64, error) {
	q := db.Model(&model.FanhuaEntity{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaEntity
	err := q.Order("category, id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// EntityHitList 标注记录列表（倒序分页）
func (Repo) EntityHitList(db *gorm.DB, page, pageSize int) ([]model.FanhuaEntityHit, int64, error) {
	q := db.Model(&model.FanhuaEntityHit{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaEntityHit
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ---------- 文本标注 ----------

// CreateTextLabelsBatch 批量落库标注
func (Repo) CreateTextLabelsBatch(db *gorm.DB, labels []model.FanhuaTextLabel, batchSize int) error {
	return db.CreateInBatches(labels, batchSize).Error
}

// FindTextLabel 取标注（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindTextLabel(db *gorm.DB, id uint) (*model.FanhuaTextLabel, error) {
	var row model.FanhuaTextLabel
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateTextLabelColumns 更新标注指定列
func (Repo) UpdateTextLabelColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaTextLabel{}).Where("id = ?", id).Updates(cols).Error
}

// TextLabelList 标注列表（划分/情感过滤，倒序分页；情感兼容自动预标列）
func (Repo) TextLabelList(db *gorm.DB, split, sentiment string, page, pageSize int) ([]model.FanhuaTextLabel, int64, error) {
	q := db.Model(&model.FanhuaTextLabel{})
	if split != "" {
		q = q.Where("split = ?", split)
	}
	if sentiment != "" {
		q = q.Where("sentiment = ? OR (sentiment = '' AND auto_sentiment = ?)", sentiment, sentiment)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaTextLabel
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// LabelDistributionRow 标注分布聚合行（划分×情感计数）
type LabelDistributionRow struct {
	Split     string `json:"split"`
	Sentiment string `json:"sentiment"`
	Count     int64  `json:"count"`
}

// LabelDistribution 标注分布聚合（人工标签空缺回退自动预标）
func (Repo) LabelDistribution(db *gorm.DB) ([]LabelDistributionRow, error) {
	var rows []LabelDistributionRow
	err := db.Model(&model.FanhuaTextLabel{}).
		Select("split, COALESCE(NULLIF(sentiment, ''), auto_sentiment) as sentiment, COUNT(*) as count").
		Group("split, sentiment").Find(&rows).Error
	return rows, err
}

// ---------- 图片资产与标签 ----------

// CreateImageAsset 落库图片资产
func (Repo) CreateImageAsset(db *gorm.DB, in *model.FanhuaImageAsset) error {
	return db.Create(in).Error
}

// FindImageAsset 取图片资产（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindImageAsset(db *gorm.DB, id uint) (*model.FanhuaImageAsset, error) {
	var row model.FanhuaImageAsset
	err := db.First(&row, id).Error
	return &row, err
}

// FindImageLabelByAsset 按资产取标签（一图一条；无标签返回 gorm.ErrRecordNotFound）
func (Repo) FindImageLabelByAsset(db *gorm.DB, assetId uint) (*model.FanhuaImageLabel, error) {
	var row model.FanhuaImageLabel
	err := db.Where("asset_id = ?", assetId).First(&row).Error
	return &row, err
}

// CreateImageLabel 落库图片标签
func (Repo) CreateImageLabel(db *gorm.DB, in *model.FanhuaImageLabel) error {
	return db.Create(in).Error
}

// UpdateImageLabelColumns 更新图片标签指定列
func (Repo) UpdateImageLabelColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaImageLabel{}).Where("id = ?", id).Updates(cols).Error
}

// ImageAssetList 图片资产列表（用途/关键字过滤，倒序分页）
func (Repo) ImageAssetList(db *gorm.DB, split, keyword string, page, pageSize int) ([]model.FanhuaImageAsset, int64, error) {
	q := db.Model(&model.FanhuaImageAsset{})
	if split != "" {
		q = q.Where("split = ?", split)
	}
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaImageAsset
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ImageLabelList 图片标签列表（标注状态过滤，倒序分页）
func (Repo) ImageLabelList(db *gorm.DB, labeled string, page, pageSize int) ([]model.FanhuaImageLabel, int64, error) {
	q := db.Model(&model.FanhuaImageLabel{})
	if labeled == "yes" {
		q = q.Where("labeled = ?", true)
	} else if labeled == "no" {
		q = q.Where("labeled = ?", false)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaImageLabel
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ---------- 语音条目 ----------

// VoiceEntriesScan 语音条目扫描（组合检索候选集，上限 2000）
func (Repo) VoiceEntriesScan(db *gorm.DB, category, speakerType string) ([]model.FanhuaVoiceEntry, error) {
	q := db.Model(&model.FanhuaVoiceEntry{}).Limit(2000)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if speakerType != "" {
		q = q.Where("speaker_type = ?", speakerType)
	}
	var list []model.FanhuaVoiceEntry
	err := q.Find(&list).Error
	return list, err
}

// FindVoiceEntry 取语音条目（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindVoiceEntry(db *gorm.DB, id uint) (*model.FanhuaVoiceEntry, error) {
	var row model.FanhuaVoiceEntry
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateVoiceEntryColumns 更新语音条目指定列
func (Repo) UpdateVoiceEntryColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaVoiceEntry{}).Where("id = ?", id).Updates(cols).Error
}

// CreateVoiceEntry 落库语音条目
func (Repo) CreateVoiceEntry(db *gorm.DB, in *model.FanhuaVoiceEntry) error {
	return db.Create(in).Error
}

// DeleteVoiceEntry 删除语音条目（软删除），返回影响行数
func (Repo) DeleteVoiceEntry(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaVoiceEntry{}, id)
	return res.RowsAffected, res.Error
}

// VoiceEntryList 语音条目列表（分类/声部过滤，倒序分页）
func (Repo) VoiceEntryList(db *gorm.DB, category, speakerType string, page, pageSize int) ([]model.FanhuaVoiceEntry, int64, error) {
	q := db.Model(&model.FanhuaVoiceEntry{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if speakerType != "" {
		q = q.Where("speaker_type = ?", speakerType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaVoiceEntry
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ---------- 敏感词/合规模板/冲突案例 ----------

// FindSensitiveWord 取敏感词（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindSensitiveWord(db *gorm.DB, id uint) (*model.FanhuaSensitiveWord, error) {
	var row model.FanhuaSensitiveWord
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateSensitiveWordColumns 更新敏感词指定列
func (Repo) UpdateSensitiveWordColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaSensitiveWord{}).Where("id = ?", id).Updates(cols).Error
}

// CreateSensitiveWord 落库敏感词
func (Repo) CreateSensitiveWord(db *gorm.DB, in *model.FanhuaSensitiveWord) error {
	return db.Create(in).Error
}

// DeleteSensitiveWord 删除敏感词（软删除），返回影响行数
func (Repo) DeleteSensitiveWord(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaSensitiveWord{}, id)
	return res.RowsAffected, res.Error
}

// WordList 敏感词列表（动作过滤，倒序分页）
func (Repo) WordList(db *gorm.DB, action string, page, pageSize int) ([]model.FanhuaSensitiveWord, int64, error) {
	q := db.Model(&model.FanhuaSensitiveWord{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaSensitiveWord
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ComplianceTemplatesAll 合规表达模板全量（id 升序）
func (Repo) ComplianceTemplatesAll(db *gorm.DB) ([]model.FanhuaComplianceTemplate, error) {
	var list []model.FanhuaComplianceTemplate
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ConflictCasesByLang 冲突案例列表（语种过滤，id 升序，上限 200）
func (Repo) ConflictCasesByLang(db *gorm.DB, lang string) ([]model.FanhuaConflictCase, error) {
	q := db.Model(&model.FanhuaConflictCase{})
	if lang != "" {
		q = q.Where("lang = ?", lang)
	}
	var list []model.FanhuaConflictCase
	err := q.Order("id").Limit(200).Find(&list).Error
	return list, err
}

// ---------- 城市治理（指标/数据点/照片） ----------

// GovIndicatorsAll 指标全量（id 升序，建树候选集）
func (Repo) GovIndicatorsAll(db *gorm.DB) ([]model.FanhuaGovIndicator, error) {
	var list []model.FanhuaGovIndicator
	err := db.Order("id").Find(&list).Error
	return list, err
}

// FindGovIndicator 取指标（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindGovIndicator(db *gorm.DB, id uint) (*model.FanhuaGovIndicator, error) {
	var row model.FanhuaGovIndicator
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateGovIndicatorColumns 更新指标指定列
func (Repo) UpdateGovIndicatorColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaGovIndicator{}).Where("id = ?", id).Updates(cols).Error
}

// CreateGovIndicator 落库指标
func (Repo) CreateGovIndicator(db *gorm.DB, in *model.FanhuaGovIndicator) error {
	return db.Create(in).Error
}

// DeleteGovIndicator 删除指标（软删除），返回影响行数
func (Repo) DeleteGovIndicator(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaGovIndicator{}, id)
	return res.RowsAffected, res.Error
}

// CreateGovDataPoint 落库数据点
func (Repo) CreateGovDataPoint(db *gorm.DB, in *model.FanhuaGovDataPoint) error {
	return db.Create(in).Error
}

// GovDataPointsByRegion 区域数据点扫描（画像/融合打分集）
func (Repo) GovDataPointsByRegion(db *gorm.DB, region, period string) ([]model.FanhuaGovDataPoint, error) {
	q := db.Model(&model.FanhuaGovDataPoint{}).Where("region = ?", region)
	if period != "" {
		q = q.Where("period = ?", period)
	}
	var list []model.FanhuaGovDataPoint
	err := q.Find(&list).Error
	return list, err
}

// GovDataPointList 数据点列表（区域/期间过滤，倒序分页）
func (Repo) GovDataPointList(db *gorm.DB, region, period string, page, pageSize int) ([]model.FanhuaGovDataPoint, int64, error) {
	q := db.Model(&model.FanhuaGovDataPoint{})
	if region != "" {
		q = q.Where("region = ?", region)
	}
	if period != "" {
		q = q.Where("period = ?", period)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaGovDataPoint
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GovPhotosByRegion 区域照片扫描（融合查询照片源，倒序上限 20）
func (Repo) GovPhotosByRegion(db *gorm.DB, region, period string) ([]model.FanhuaGovPhoto, error) {
	q := db.Model(&model.FanhuaGovPhoto{}).Where("region = ?", region)
	if period != "" {
		q = q.Where("shot_at LIKE ?", period+"%")
	}
	var list []model.FanhuaGovPhoto
	err := q.Order("id DESC").Limit(20).Find(&list).Error
	return list, err
}

// GovPhotoCountByRegion 区域照片计数（不分期间）
func (Repo) GovPhotoCountByRegion(db *gorm.DB, region string) (int64, error) {
	var total int64
	err := db.Model(&model.FanhuaGovPhoto{}).Where("region = ?", region).Count(&total).Error
	return total, err
}

// CreateGovPhoto 落库照片元数据
func (Repo) CreateGovPhoto(db *gorm.DB, in *model.FanhuaGovPhoto) error {
	return db.Create(in).Error
}

// ---------- 提示词模板 ----------

// FindPromptTemplate 取模板（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindPromptTemplate(db *gorm.DB, id uint) (*model.FanhuaPromptTemplate, error) {
	var row model.FanhuaPromptTemplate
	err := db.First(&row, id).Error
	return &row, err
}

// FindPromptTemplateByName 按模板名取模板（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindPromptTemplateByName(db *gorm.DB, name string) (*model.FanhuaPromptTemplate, error) {
	var row model.FanhuaPromptTemplate
	err := db.Where("name = ?", name).First(&row).Error
	return &row, err
}

// UpdatePromptTemplateColumns 更新模板指定列
func (Repo) UpdatePromptTemplateColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.FanhuaPromptTemplate{}).Where("id = ?", id).Updates(cols).Error
}

// CreatePromptTemplate 落库模板
func (Repo) CreatePromptTemplate(db *gorm.DB, in *model.FanhuaPromptTemplate) error {
	return db.Create(in).Error
}

// DeletePromptTemplate 删除模板（软删除），返回影响行数
func (Repo) DeletePromptTemplate(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.FanhuaPromptTemplate{}, id)
	return res.RowsAffected, res.Error
}

// PromptTemplateList 模板列表（场景过滤，id 升序，上限 500）
func (Repo) PromptTemplateList(db *gorm.DB, scene string) ([]model.FanhuaPromptTemplate, error) {
	q := db.Model(&model.FanhuaPromptTemplate{})
	if scene != "" {
		q = q.Where("scene = ?", scene)
	}
	var list []model.FanhuaPromptTemplate
	err := q.Order("id").Limit(500).Find(&list).Error
	return list, err
}

// ---------- 问答/审查日志 ----------

// CreateDialectQaLog 落库问答日志
func (Repo) CreateDialectQaLog(db *gorm.DB, log *model.FanhuaDialectQaLog) error {
	return db.Create(log).Error
}

// DialectQaLogList 问答日志列表（倒序分页）
func (Repo) DialectQaLogList(db *gorm.DB, page, pageSize int) ([]model.FanhuaDialectQaLog, int64, error) {
	q := db.Model(&model.FanhuaDialectQaLog{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaDialectQaLog
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CreateValuesAuditLog 落库审查日志
func (Repo) CreateValuesAuditLog(db *gorm.DB, log *model.FanhuaValuesAuditLog) error {
	return db.Create(log).Error
}

// ValuesAuditLogList 审查日志列表（等级/预警过滤，倒序分页）
func (Repo) ValuesAuditLogList(db *gorm.DB, riskLevel string, alarmOnly bool, page, pageSize int) ([]model.FanhuaValuesAuditLog, int64, error) {
	q := db.Model(&model.FanhuaValuesAuditLog{})
	if riskLevel != "" {
		q = q.Where("risk_level = ?", riskLevel)
	}
	if alarmOnly {
		q = q.Where("alarm = ?", true)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaValuesAuditLog
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
