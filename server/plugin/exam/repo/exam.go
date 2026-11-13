// Package repo exam 数据访问层：只做查询与写库，业务规则（四态状态机/判分引擎/
// 防作弊抽查/组卷规则/乐观锁）留在 service。方法首参 db 为当前会话（全局库或事务 tx），
// 由 service 决定传入哪一个；本包禁止 import global。
// 行锁统一走 data.WithRowLock（sqlite 单测内存库豁免 FOR UPDATE，生产（MySQL 系列）
// 行为与手写 clause.Locking 完全一致）。
package repo

import (
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
)

// Repo exam 数据访问（无状态，可零值使用）
type Repo struct{}

// ---------- 题库 ----------

// QuestionList 题目列表（题型/知识点过滤，id 倒序）。page/pageSize 已由 service 归一
// （page≤0 回第一页、pageSize 缺省 20 / 上限 100）——归一后与 data.Page 语义等价，
// 采用共享分页（区别于 artsci 的 200 上限冻结口径）。
func (Repo) QuestionList(db *gorm.DB, page, pageSize int, qType, knowledgePoint string) ([]model.ExamQuestion, int64, error) {
	q := db.Model(&model.ExamQuestion{})
	if qType != "" {
		q = q.Where("type = ?", qType)
	}
	if knowledgePoint != "" {
		q = q.Where("knowledge_point = ?", knowledgePoint)
	}
	var list []model.ExamQuestion
	total, err := data.Page(q.Order("id desc"), page, pageSize, &list)
	return list, total, err
}

// FindQuestion 取题目（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindQuestion(db *gorm.DB, id uint) (*model.ExamQuestion, error) {
	var row model.ExamQuestion
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateQuestionColumns 更新题目指定列
func (Repo) UpdateQuestionColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamQuestion{}).Where("id = ?", id).Updates(cols).Error
}

// CreateQuestion 落库题目
func (Repo) CreateQuestion(db *gorm.DB, q *model.ExamQuestion) error {
	return db.Create(q).Error
}

// CreateQuestionsBatch 批量落库题目（JSONL 导入，500/批）
func (Repo) CreateQuestionsBatch(db *gorm.DB, qs []model.ExamQuestion) error {
	return db.CreateInBatches(qs, 500).Error
}

// DeleteQuestion 删除题目（软删除），返回影响行数供 service 判存在
func (Repo) DeleteQuestion(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ExamQuestion{}, id)
	return res.RowsAffected, res.Error
}

// QuestionsAll 题库全量（组卷抽题池，规则筛选在 service 内存进行）
func (Repo) QuestionsAll(db *gorm.DB) ([]model.ExamQuestion, error) {
	var list []model.ExamQuestion
	err := db.Find(&list).Error
	return list, err
}

// FindQuestionsByIds 按 ID 集取题（拉题/判分取含答案版本）
func (Repo) FindQuestionsByIds(db *gorm.DB, ids []uint) ([]model.ExamQuestion, error) {
	var list []model.ExamQuestion
	err := db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// FindQuestionsByIdsAndTypes 按 ID 集与题型白名单取题（相似度抽查主观题集）
func (Repo) FindQuestionsByIdsAndTypes(db *gorm.DB, ids []uint, types []string) ([]model.ExamQuestion, error) {
	var list []model.ExamQuestion
	err := db.Where("id IN ? AND type IN ?", ids, types).Find(&list).Error
	return list, err
}

// QuestionScoreSum 题目分值合计（试卷满分）
func (Repo) QuestionScoreSum(db *gorm.DB, ids []uint) (float64, error) {
	var total float64
	err := db.Model(&model.ExamQuestion{}).Where("id IN ?", ids).
		Select("COALESCE(SUM(score),0)").Scan(&total).Error
	return total, err
}

// ---------- 试卷 ----------

// PaperList 试卷列表（id 倒序）
func (Repo) PaperList(db *gorm.DB) ([]model.ExamPaper, error) {
	var list []model.ExamPaper
	err := db.Order("id desc").Find(&list).Error
	return list, err
}

// CountPapersByName 同名试卷计数（excludeId>0 排除自身，编辑保存防自撞）
func (Repo) CountPapersByName(db *gorm.DB, name string, excludeId uint) (int64, error) {
	q := db.Model(&model.ExamPaper{}).Where("name = ?", name)
	if excludeId > 0 {
		q = q.Where("id <> ?", excludeId)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// FindPaper 取试卷（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindPaper(db *gorm.DB, id uint) (*model.ExamPaper, error) {
	var row model.ExamPaper
	err := db.First(&row, id).Error
	return &row, err
}

// UpdatePaperColumns 更新试卷指定列
func (Repo) UpdatePaperColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamPaper{}).Where("id = ?", id).Updates(cols).Error
}

// CreatePaper 落库试卷
func (Repo) CreatePaper(db *gorm.DB, p *model.ExamPaper) error {
	return db.Create(p).Error
}

// DeletePaper 删除试卷（软删除），返回影响行数
func (Repo) DeletePaper(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ExamPaper{}, id)
	return res.RowsAffected, res.Error
}

// ---------- 考试场次 ----------

// ExamList 考试列表（状态过滤，开考时间倒序）
func (Repo) ExamList(db *gorm.DB, status string) ([]model.ExamExam, error) {
	q := db.Model(&model.ExamExam{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var list []model.ExamExam
	err := q.Order("start_time desc").Find(&list).Error
	return list, err
}

// ExamsOngoing 进行中考试（终端可考候选集，时间窗过滤在 service）
func (Repo) ExamsOngoing(db *gorm.DB) ([]model.ExamExam, error) {
	var list []model.ExamExam
	err := db.Where("status = 'ongoing'").Find(&list).Error
	return list, err
}

// CountExamsByName 同名考试计数（excludeId>0 排除自身）
func (Repo) CountExamsByName(db *gorm.DB, name string, excludeId uint) (int64, error) {
	q := db.Model(&model.ExamExam{}).Where("name = ?", name)
	if excludeId > 0 {
		q = q.Where("id <> ?", excludeId)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// FindExam 取考试（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindExam(db *gorm.DB, id uint) (*model.ExamExam, error) {
	var row model.ExamExam
	err := db.First(&row, id).Error
	return &row, err
}

// UpdateExamColumns 更新考试指定列
func (Repo) UpdateExamColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamExam{}).Where("id = ?", id).Updates(cols).Error
}

// CreateExam 落库考试
func (Repo) CreateExam(db *gorm.DB, e *model.ExamExam) error {
	return db.Create(e).Error
}

// DeleteExam 删除考试，返回影响行数
func (Repo) DeleteExam(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ExamExam{}, id)
	return res.RowsAffected, res.Error
}

// ---------- 考生 ----------

// FindCandidateByTicket 按券码取考生（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindCandidateByTicket(db *gorm.DB, ticket string) (*model.ExamCandidate, error) {
	var row model.ExamCandidate
	err := db.Where("ticket_code = ?", ticket).First(&row).Error
	return &row, err
}

// FindCandidateByExamUser 按（考试, 用户标识）取考生；withLock 时加行锁（领券防并发）
func (Repo) FindCandidateByExamUser(db *gorm.DB, examId uint, userKey string, withLock bool) (*model.ExamCandidate, error) {
	q := db.Where("exam_id = ? AND user_key = ?", examId, userKey)
	if withLock {
		q = data.WithRowLock(q)
	}
	var row model.ExamCandidate
	err := q.First(&row).Error
	return &row, err
}

// CandidatesByExam 考试考生全量（名单去重 / 成绩导出取名）
func (Repo) CandidatesByExam(db *gorm.DB, examId uint) ([]model.ExamCandidate, error) {
	var list []model.ExamCandidate
	err := db.Where("exam_id = ?", examId).Find(&list).Error
	return list, err
}

// CandidateList 考生列表（examId 过滤，id 升序）。分页口径同 QuestionList（service 归一后
// 与 data.Page 语义等价）。
func (Repo) CandidateList(db *gorm.DB, examId uint, page, pageSize int) ([]model.ExamCandidate, int64, error) {
	q := db.Model(&model.ExamCandidate{})
	if examId > 0 {
		q = q.Where("exam_id = ?", examId)
	}
	var list []model.ExamCandidate
	total, err := data.Page(q.Order("id"), page, pageSize, &list)
	return list, total, err
}

// UpdateCandidateColumns 更新考生指定列
func (Repo) UpdateCandidateColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamCandidate{}).Where("id = ?", id).Updates(cols).Error
}

// CreateCandidate 落库考生（发券）
func (Repo) CreateCandidate(db *gorm.DB, c *model.ExamCandidate) error {
	return db.Create(c).Error
}

// CreateCandidatesBatch 批量落库考生（名单导入，500/批）
func (Repo) CreateCandidatesBatch(db *gorm.DB, cs []model.ExamCandidate) error {
	return db.CreateInBatches(cs, 500).Error
}

// DeleteCandidate 删除考生，返回影响行数
func (Repo) DeleteCandidate(db *gorm.DB, id uint) (int64, error) {
	res := db.Delete(&model.ExamCandidate{}, id)
	return res.RowsAffected, res.Error
}

// ---------- 答卷 ----------

// FindSheet 取答卷（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindSheet(db *gorm.DB, id uint) (*model.ExamAnswerSheet, error) {
	var row model.ExamAnswerSheet
	err := db.First(&row, id).Error
	return &row, err
}

// FindSheetLock 行锁取答卷（保存/交卷状态机防并发）
func (Repo) FindSheetLock(db *gorm.DB, id uint) (*model.ExamAnswerSheet, error) {
	var row model.ExamAnswerSheet
	err := data.WithRowLock(db).First(&row, id).Error
	return &row, err
}

// FindSheetByExamCand 按（考试, 考生）取答卷（券码登录幂等复用）
func (Repo) FindSheetByExamCand(db *gorm.DB, examId, candidateId uint) (*model.ExamAnswerSheet, error) {
	var row model.ExamAnswerSheet
	err := db.Where("exam_id = ? AND candidate_id = ?", examId, candidateId).First(&row).Error
	return &row, err
}

// SheetsSubmittedByExam 考试已交卷答卷（判分发布 / 相似度抽查）
func (Repo) SheetsSubmittedByExam(db *gorm.DB, examId uint) ([]model.ExamAnswerSheet, error) {
	var list []model.ExamAnswerSheet
	err := db.Where("exam_id = ? AND status = 'submitted'", examId).Find(&list).Error
	return list, err
}

// CreateSheet 落库答卷
func (Repo) CreateSheet(db *gorm.DB, s *model.ExamAnswerSheet) error {
	return db.Create(s).Error
}

// UpdateSheetColumns 更新答卷指定列
func (Repo) UpdateSheetColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamAnswerSheet{}).Where("id = ?", id).Updates(cols).Error
}

// ---------- 判分明细 ----------

// DeleteGradesBySheet 清空答卷判分明细（重判幂等先清）
func (Repo) DeleteGradesBySheet(db *gorm.DB, sheetId uint) error {
	return db.Where("sheet_id = ?", sheetId).Delete(&model.ExamGrade{}).Error
}

// CreateGradesBatch 批量落库判分明细（200/批）
func (Repo) CreateGradesBatch(db *gorm.DB, grades []model.ExamGrade) error {
	return db.CreateInBatches(grades, 200).Error
}

// GradeListBySheet 判分明细（question_id 升序）
func (Repo) GradeListBySheet(db *gorm.DB, sheetId uint) ([]model.ExamGrade, error) {
	var list []model.ExamGrade
	err := db.Where("sheet_id = ?", sheetId).Order("question_id").Find(&list).Error
	return list, err
}

// ---------- 成绩与等级规则 ----------

// LevelRulesByExam 等级分数线（examId=0 即全局默认档）
func (Repo) LevelRulesByExam(db *gorm.DB, examId uint) ([]model.ExamLevelRule, error) {
	var list []model.ExamLevelRule
	err := db.Where("exam_id = ?", examId).Find(&list).Error
	return list, err
}

// FindScoreBySheet 按答卷取成绩（幂等 upsert；记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindScoreBySheet(db *gorm.DB, sheetId uint) (*model.ExamScore, error) {
	var row model.ExamScore
	err := db.Where("sheet_id = ?", sheetId).First(&row).Error
	return &row, err
}

// CreateScore 落库成绩
func (Repo) CreateScore(db *gorm.DB, sc *model.ExamScore) error {
	return db.Create(sc).Error
}

// UpdateScoreColumns 更新成绩指定列
func (Repo) UpdateScoreColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamScore{}).Where("id = ?", id).Updates(cols).Error
}

// ScoreList 成绩列表（examId 过滤，总分倒序）。分页口径同 QuestionList。
func (Repo) ScoreList(db *gorm.DB, examId uint, page, pageSize int) ([]model.ExamScore, int64, error) {
	q := db.Model(&model.ExamScore{})
	if examId > 0 {
		q = q.Where("exam_id = ?", examId)
	}
	var list []model.ExamScore
	total, err := data.Page(q.Order("total desc"), page, pageSize, &list)
	return list, total, err
}

// ScoresByExam 考试成绩全量（CSV 导出，总分倒序）
func (Repo) ScoresByExam(db *gorm.DB, examId uint) ([]model.ExamScore, error) {
	var list []model.ExamScore
	err := db.Where("exam_id = ?", examId).Order("total desc").Find(&list).Error
	return list, err
}

// ---------- 防作弊 ----------

// CreateBehavior 落库行为事件
func (Repo) CreateBehavior(db *gorm.DB, b *model.ExamBehavior) error {
	return db.Create(b).Error
}

// BehaviorCountRow 行为事件计数行
type BehaviorCountRow struct {
	Event string
	Cnt   int64
}

// BehaviorStats 行为计数（按事件分组聚合）
func (Repo) BehaviorStats(db *gorm.DB, sheetId uint) ([]BehaviorCountRow, error) {
	var rows []BehaviorCountRow
	err := db.Model(&model.ExamBehavior{}).
		Select("event, COUNT(*) as cnt").Where("sheet_id = ?", sheetId).
		Group("event").Find(&rows).Error
	return rows, err
}

// ---------- 节点与分片 ----------

// FindNodeByKey 按机器码取节点（记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindNodeByKey(db *gorm.DB, nodeKey string) (*model.ExamNode, error) {
	var row model.ExamNode
	err := db.Where("node_key = ?", nodeKey).First(&row).Error
	return &row, err
}

// CreateNode 落库节点
func (Repo) CreateNode(db *gorm.DB, n *model.ExamNode) error {
	return db.Create(n).Error
}

// UpdateNodeColumns 更新节点指定列
func (Repo) UpdateNodeColumns(db *gorm.DB, id uint, cols map[string]any) error {
	return db.Model(&model.ExamNode{}).Where("id = ?", id).Updates(cols).Error
}

// NodeList 节点列表（region, load 序）
func (Repo) NodeList(db *gorm.DB) ([]model.ExamNode, error) {
	var list []model.ExamNode
	err := db.Order("region, load").Find(&list).Error
	return list, err
}

// NodesMarkStaleOffline 心跳超时节点批量离线，返回影响行数
func (Repo) NodesMarkStaleOffline(db *gorm.DB, threshold time.Time) (int64, error) {
	res := db.Model(&model.ExamNode{}).
		Where("status = 'online' AND last_heartbeat < ?", threshold).Update("status", "offline")
	return res.RowsAffected, res.Error
}

// FindShardBySheetLock 行锁取分片（sheetId 唯一索引幂等，记录不存在返回 gorm.ErrRecordNotFound）
func (Repo) FindShardBySheetLock(db *gorm.DB, sheetId uint) (*model.ExamShard, error) {
	var row model.ExamShard
	err := data.WithRowLock(db).Where("sheet_id = ?", sheetId).First(&row).Error
	return &row, err
}

// FindNodesOnlineLock 行锁取在线节点（分配防并发同择）
func (Repo) FindNodesOnlineLock(db *gorm.DB) ([]model.ExamNode, error) {
	var list []model.ExamNode
	err := data.WithRowLock(db).Where("status = 'online'").Find(&list).Error
	return list, err
}

// CreateShard 落库分片
func (Repo) CreateShard(db *gorm.DB, s *model.ExamShard) error {
	return db.Create(s).Error
}

// BumpNodeLoad 节点负载 +1（分配计数）
func (Repo) BumpNodeLoad(db *gorm.DB, nodeKey string) error {
	return db.Model(&model.ExamNode{}).Where("node_key = ?", nodeKey).
		UpdateColumn("load", gorm.Expr("load + 1")).Error
}

// ---------- 实操留痕 ----------

// CreatePracticeLog 落库实操留痕
func (Repo) CreatePracticeLog(db *gorm.DB, l *model.ExamPracticeLog) error {
	return db.Create(l).Error
}

// PracticeLogList 实操日志（sheetId 必带，questionId 可选，id 倒序 20 条）
func (Repo) PracticeLogList(db *gorm.DB, sheetId, questionId uint) ([]model.ExamPracticeLog, error) {
	q := db.Where("sheet_id = ?", sheetId)
	if questionId > 0 {
		q = q.Where("question_id = ?", questionId)
	}
	var list []model.ExamPracticeLog
	err := q.Order("id desc").Limit(20).Find(&list).Error
	return list, err
}
