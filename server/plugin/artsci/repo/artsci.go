// Package repo artsci 数据访问层：只做查询与写库，业务规则（白名单/schema 校验/
// 状态机/出题组卷规则）留在 service。方法首参 db 为当前会话（全局库或事务 tx），
// 由 service 决定传入哪一个；本包禁止 import global。
package repo

import (
	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// Repo artsci 数据访问（无状态，可零值使用）
type Repo struct{}

// normPage 分页归一：page<1 视为第一页（不是全量请求），pageSize 缺省 20、上限 200。
// 分页口径冻结：与 data.Page（钳位 [1,100]、page<=0 全量上限 500）不等价——
// 本插件列表 page<=0 语义为「回到第一页」且上限 200 为既有约定，保留本地实现。
func normPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	return page, pageSize
}

// paged 列表分页共用（Count → 归一分页 Find，id 倒序）；db 传入 Model+Where 链
func paged[T any](db *gorm.DB, page, pageSize int, dest *[]T) (int64, error) {
	page, pageSize = normPage(page, pageSize)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(dest).Error
	return total, err
}

// ---------- 工具注册表 ----------

// ToolCountByName 按名查工具数（重名校验；excludeId>0 时排除自身——更新路径）
func (Repo) ToolCountByName(db *gorm.DB, name string, excludeId uint) (int64, error) {
	q := db.Model(&model.ArtsciTool{}).Where("name = ?", name)
	if excludeId > 0 {
		q = q.Where("id <> ?", excludeId)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// FindTool 取工具（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindTool(db *gorm.DB, id uint) (*model.ArtsciTool, error) {
	var row model.ArtsciTool
	err := db.First(&row, id).Error
	return &row, err
}

// CreateTool 落库工具
func (Repo) CreateTool(db *gorm.DB, in *model.ArtsciTool) error {
	return db.Create(in).Error
}

// UpdateToolColumns 更新工具指定列
func (Repo) UpdateToolColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ArtsciTool{}).Where("id = ?", id).Updates(cols).Error
}

// DeleteTool 软删除工具，返回受影响行数（0=不存在）
func (Repo) DeleteTool(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ArtsciTool{}, id)
	return res.RowsAffected, res.Error
}

// ToolList 工具列表（学科/关键字过滤，subject 学科、keyword 命中名称或描述，subject,id 排序）
func (Repo) ToolList(db *gorm.DB, subject, keyword string) ([]model.ArtsciTool, error) {
	q := db.Model(&model.ArtsciTool{})
	if subject != "" {
		q = q.Where("subject = ?", subject)
	}
	if keyword != "" {
		q = q.Where("name LIKE ? OR desc LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var list []model.ArtsciTool
	err := q.Order("subject, id").Find(&list).Error
	return list, err
}

// ---------- 执行留痕 ----------

// CreateRunLog 落库执行留痕
func (Repo) CreateRunLog(db *gorm.DB, log *model.ArtsciRunLog) error {
	return db.Create(log).Error
}

// RunLogList 留痕列表（工具过滤，id 倒序分页）
func (Repo) RunLogList(db *gorm.DB, toolId uint, page, pageSize int) ([]model.ArtsciRunLog, int64, error) {
	q := db.Model(&model.ArtsciRunLog{})
	if toolId > 0 {
		q = q.Where("tool_id = ?", toolId)
	}
	var list []model.ArtsciRunLog
	total, err := paged(q, page, pageSize, &list)
	return list, total, err
}

// RunLogsForStats 统计扫描集（id 升序，上限 2000——成功率聚合在此集合内存进行）
func (Repo) RunLogsForStats(db *gorm.DB, limit int) ([]model.ArtsciRunLog, error) {
	var list []model.ArtsciRunLog
	err := db.Order("id ASC").Limit(limit).Find(&list).Error
	return list, err
}

// ---------- 作业批改 ----------

// CreateGrading 落库批改记录
func (Repo) CreateGrading(db *gorm.DB, rec *model.ArtsciGrading) error {
	return db.Create(rec).Error
}

// UpdateGradingColumns 更新批改记录指定列
func (Repo) UpdateGradingColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ArtsciGrading{}).Where("id = ?", id).Updates(cols).Error
}

// GradingList 批改记录列表（id 倒序分页）
func (Repo) GradingList(db *gorm.DB, page, pageSize int) ([]model.ArtsciGrading, int64, error) {
	var list []model.ArtsciGrading
	total, err := paged(db.Model(&model.ArtsciGrading{}), page, pageSize, &list)
	return list, total, err
}

// ---------- 出题与组卷 ----------

// CreateQuestion 落库题目
func (Repo) CreateQuestion(db *gorm.DB, q *model.ArtsciQuestion) error {
	return db.Create(q).Error
}

// QuestionCandidates 抽题候选（题型必过滤、知识点可选，id 升序取前 count）
func (Repo) QuestionCandidates(db *gorm.DB, kp, qType string, count int) ([]model.ArtsciQuestion, error) {
	q := db.Model(&model.ArtsciQuestion{}).Where("q_type = ?", qType)
	if kp != "" {
		q = q.Where("knowledge_point = ?", kp)
	}
	var list []model.ArtsciQuestion
	err := q.Order("id").Limit(count).Find(&list).Error
	return list, err
}

// CreatePaper 落库试卷
func (Repo) CreatePaper(db *gorm.DB, p *model.ArtsciPaper) error {
	return db.Create(p).Error
}

// QuestionList 题目列表（知识点/题型过滤，id 倒序分页；题型白名单校验在 service）
func (Repo) QuestionList(db *gorm.DB, kp, qType string, page, pageSize int) ([]model.ArtsciQuestion, int64, error) {
	q := db.Model(&model.ArtsciQuestion{})
	if kp != "" {
		q = q.Where("knowledge_point = ?", kp)
	}
	if qType != "" {
		q = q.Where("q_type = ?", qType)
	}
	var list []model.ArtsciQuestion
	total, err := paged(q, page, pageSize, &list)
	return list, total, err
}

// PaperList 试卷列表（id 倒序分页）
func (Repo) PaperList(db *gorm.DB, page, pageSize int) ([]model.ArtsciPaper, int64, error) {
	var list []model.ArtsciPaper
	total, err := paged(db.Model(&model.ArtsciPaper{}), page, pageSize, &list)
	return list, total, err
}

// ---------- 文生图场景 ----------

// FindSceneTemplate 按场景名取模板（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindSceneTemplate(db *gorm.DB, scene string) (*model.ArtsciSceneTemplate, error) {
	var row model.ArtsciSceneTemplate
	err := db.Where("scene = ?", scene).First(&row).Error
	return &row, err
}

// SceneTemplateList 场景模板全量（id 升序）
func (Repo) SceneTemplateList(db *gorm.DB) ([]model.ArtsciSceneTemplate, error) {
	var list []model.ArtsciSceneTemplate
	err := db.Order("id").Find(&list).Error
	return list, err
}

// CreateImageTask 落库文生图任务（queued 初始态）
func (Repo) CreateImageTask(db *gorm.DB, t *model.ArtsciImageTask) error {
	return db.Create(t).Error
}

// UpdateImageTaskColumns 更新文生图任务指定列（状态机推进 queued→done/failed）
func (Repo) UpdateImageTaskColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ArtsciImageTask{}).Where("id = ?", id).Updates(cols).Error
}

// ImageTaskList 任务列表（场景过滤，id 倒序分页）
func (Repo) ImageTaskList(db *gorm.DB, scene string, page, pageSize int) ([]model.ArtsciImageTask, int64, error) {
	q := db.Model(&model.ArtsciImageTask{})
	if scene != "" {
		q = q.Where("scene = ?", scene)
	}
	var list []model.ArtsciImageTask
	total, err := paged(q, page, pageSize, &list)
	return list, total, err
}

// ---------- 图像资产与解读 ----------

// FindImageAsset 取图像资产（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindImageAsset(db *gorm.DB, id uint) (*model.ArtsciImageAsset, error) {
	var row model.ArtsciImageAsset
	err := db.First(&row, id).Error
	return &row, err
}

// CreateImageAsset 落库图像资产
func (Repo) CreateImageAsset(db *gorm.DB, a *model.ArtsciImageAsset) error {
	return db.Create(a).Error
}

// ImageAssetList 资产列表（名称关键字过滤，id 倒序分页）
func (Repo) ImageAssetList(db *gorm.DB, keyword string, page, pageSize int) ([]model.ArtsciImageAsset, int64, error) {
	q := db.Model(&model.ArtsciImageAsset{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var list []model.ArtsciImageAsset
	total, err := paged(q, page, pageSize, &list)
	return list, total, err
}

// CreateInterpretation 落库解读记录
func (Repo) CreateInterpretation(db *gorm.DB, r *model.ArtsciInterpretation) error {
	return db.Create(r).Error
}

// InterpretationList 解读记录列表（资产过滤，id 倒序分页）
func (Repo) InterpretationList(db *gorm.DB, assetId uint, page, pageSize int) ([]model.ArtsciInterpretation, int64, error) {
	q := db.Model(&model.ArtsciInterpretation{})
	if assetId > 0 {
		q = q.Where("asset_id = ?", assetId)
	}
	var list []model.ArtsciInterpretation
	total, err := paged(q, page, pageSize, &list)
	return list, total, err
}
