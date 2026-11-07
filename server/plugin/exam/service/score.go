// Package service exam：成绩与等级（采购件③）——判分发布落档 + 配置化分数线 + CSV 导出
// 口径：等级按得分率匹配（取满足的最高档；无匹配→不及格）；无配置时默认
// 优秀 0.9 / 良好 0.75 / 及格 0.6；发布逐卷判分并按 sheetId 幂等 upsert。
package service

import (
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const ErrCodeScoreBad = 2416

// defaultLevelRules 默认分数线
func defaultLevelRules() []model.ExamLevelRule {
	return []model.ExamLevelRule{
		{Level: "优秀", MinRatio: 0.9},
		{Level: "良好", MinRatio: 0.75},
		{Level: "及格", MinRatio: 0.6},
	}
}

// DetermineLevel 等级判定（纯函数，测试锚点）：取满足的最高档
func DetermineLevel(rules []model.ExamLevelRule, total, full float64) string {
	if full <= 0 {
		return "不及格"
	}
	ratio := total / full
	sorted := make([]model.ExamLevelRule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].MinRatio > sorted[j].MinRatio })
	for _, r := range sorted {
		if ratio >= r.MinRatio {
			return r.Level
		}
	}
	return "不及格"
}

// scoresToCSV 成绩 CSV（纯函数，测试锚点）
func scoresToCSV(scores []model.ExamScore, names map[uint]string) string {
	var sb strings.Builder
	sb.WriteString("考生,总分,满分,等级\n")
	for _, s := range scores {
		sb.WriteString(fmt.Sprintf("%s,%.1f,%.1f,%s\n", names[s.CandidateId], s.Total, s.Full, s.Level))
	}
	return sb.String()
}

// PublishScores 发布成绩：逐卷判分 + 幂等落档
func (s *ExamService) PublishScores(examId uint, judgeModelId uint, multiPartial bool) (int, error) {
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, examId).Error; err != nil {
		return 0, fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
	}
	var paper model.ExamPaper
	if err := global.GVA_DB.First(&paper, exam.PaperId).Error; err != nil {
		return 0, err
	}
	var sheets []model.ExamAnswerSheet
	if err := global.GVA_DB.Where("exam_id = ? AND status = 'submitted'", examId).Find(&sheets).Error; err != nil {
		return 0, err
	}
	if len(sheets) == 0 {
		return 0, fmt.Errorf("[%d]无已交卷答卷", ErrCodeScoreBad)
	}
	// 等级规则：考试级覆盖全局
	var rules []model.ExamLevelRule
	global.GVA_DB.Where("exam_id = ?", examId).Find(&rules)
	if len(rules) == 0 {
		global.GVA_DB.Where("exam_id = ?", 0).Find(&rules)
	}
	if len(rules) == 0 {
		rules = defaultLevelRules()
	}
	published := 0
	for _, sheet := range sheets {
		total, err := s.GradeSheet(sheet.ID, judgeModelId, multiPartial)
		if err != nil {
			return published, fmt.Errorf("答卷 %d 判分失败: %w", sheet.ID, err)
		}
		level := DetermineLevel(rules, total, paperTotalScore(paper))
		var exist model.ExamScore
		err = global.GVA_DB.Where("sheet_id = ?", sheet.ID).First(&exist).Error
		if err == gorm.ErrRecordNotFound {
			row := model.ExamScore{ExamId: examId, SheetId: sheet.ID, CandidateId: sheet.CandidateId,
				Total: total, Full: paperTotalScore(paper), Level: level}
			if err := global.GVA_DB.Create(&row).Error; err != nil {
				return published, err
			}
		} else if err == nil {
			if err := global.GVA_DB.Model(&exist).Updates(map[string]any{
				"total": total, "full": paperTotalScore(paper), "level": level,
			}).Error; err != nil {
				return published, err
			}
		} else {
			return published, err
		}
		published++
	}
	return published, nil
}

// paperTotalScore 试卷满分（锁定题目分值合计）
func paperTotalScore(paper model.ExamPaper) float64 {
	ids := paperQuestionIds(paper.QuestionIds)
	if len(ids) == 0 {
		return 0
	}
	var total float64
	global.GVA_DB.Model(&model.ExamQuestion{}).Where("id IN ?", ids).
		Select("COALESCE(SUM(score),0)").Scan(&total)
	return total
}

// GetScores 成绩单列表
func (s *ExamService) GetScores(examId uint, page, pageSize int) ([]model.ExamScore, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ExamScore{})
	if examId > 0 {
		db = db.Where("exam_id = ?", examId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ExamScore
	err := db.Order("total desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ExportScores 成绩 CSV 导出（按总分降序）
func (s *ExamService) ExportScores(examId uint) (string, error) {
	var list []model.ExamScore
	if err := global.GVA_DB.Where("exam_id = ?", examId).Order("total desc").Find(&list).Error; err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("[%d]暂无成绩", ErrCodeScoreBad)
	}
	names := map[uint]string{}
	var cands []model.ExamCandidate
	global.GVA_DB.Where("exam_id = ?", examId).Find(&cands)
	for _, c := range cands {
		names[c.ID] = c.Name
	}
	return scoresToCSV(list, names), nil
}

// PublishIn 成绩发布请求
type PublishIn struct {
	ExamId       uint `json:"examId" binding:"required"`
	JudgeModelId uint `json:"judgeModelId"`
	MultiPartial bool `json:"multiPartial"`
}
