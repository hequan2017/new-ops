// Package service exam：考生管理（名单导入 + 券码发放）
// 口径：券码 12 位大写字母数字（剔除易混淆 0/O/1/I）；发券走 DB 行锁（FOR UPDATE）+
// (examId,userKey) 唯一约束双保险，重复领取幂等返回已有券码。
package service

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const (
	ErrCodeCandBad     = 2410
	ticketCodeLen      = 12
	ticketCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// GenerateTicketCodes 批量生成券码（纯函数注入 rand，测试锚点）
func GenerateTicketCodes(n int, rng *rand.Rand) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		b := make([]byte, ticketCodeLen)
		for j := range b {
			b[j] = ticketCodeAlphabet[rng.Intn(len(ticketCodeAlphabet))]
		}
		out = append(out, string(b))
	}
	return out
}

// filterDupCandidates 名单去重（纯函数：跳过空串、批次内重名、库内已存在姓名）
func filterDupCandidates(existingNames map[string]bool, names []string) (valid []string, skipped int) {
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] || existingNames[name] {
			skipped++
			continue
		}
		seen[name] = true
		valid = append(valid, name)
	}
	return valid, skipped
}

// ImportCandidates 名单导入（自动发券；同批次与库内重名跳过）
func (s *ExamService) ImportCandidates(examId uint, names []string) (imported, skipped int, err error) {
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, examId).Error; err != nil {
		return 0, 0, fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
	}
	var existRows []model.ExamCandidate
	if err := global.GVA_DB.Where("exam_id = ?", examId).Find(&existRows).Error; err != nil {
		return 0, 0, err
	}
	existingNames := map[string]bool{}
	for _, c := range existRows {
		existingNames[c.Name] = true
	}
	validNames, skipped := filterDupCandidates(existingNames, names)
	if len(validNames) == 0 {
		return 0, skipped, nil
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	codes := GenerateTicketCodes(len(validNames), rng)
	rows := make([]model.ExamCandidate, len(validNames))
	for i, name := range validNames {
		rows[i] = model.ExamCandidate{ExamId: examId, Name: name, TicketCode: codes[i], Status: "pending"}
	}
	if err := global.GVA_DB.CreateInBatches(rows, 500).Error; err != nil {
		return 0, skipped, err
	}
	return len(rows), skipped, nil
}

// ClaimTicket 发券（自助领取；行锁 + 唯一约束幂等）
func (s *ExamService) ClaimTicket(examId uint, userKey, name string) (*model.ExamCandidate, error) {
	if strings.TrimSpace(userKey) == "" {
		return nil, fmt.Errorf("[%d]考生标识不能为空", ErrCodeCandBad)
	}
	var result *model.ExamCandidate
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var exist model.ExamCandidate
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("exam_id = ? AND user_key = ?", examId, userKey).First(&exist).Error
		if err == nil {
			// 幂等：已领取直接返回
			if name != "" {
				tx.Model(&exist).Update("name", name)
				exist.Name = name
			}
			result = &exist
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// 新发券：随机码 + 唯一索引兜底
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(examId)))
		code := GenerateTicketCodes(1, rng)[0]
		row := &model.ExamCandidate{ExamId: examId, Name: name, UserKey: userKey, TicketCode: code, Status: "pending"}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		result = row
		return nil
	})
	if err != nil {
		// 唯一约束冲突 = 并发重复领取，重查一次幂等返回
		var exist model.ExamCandidate
		if qerr := global.GVA_DB.Where("exam_id = ? AND user_key = ?", examId, userKey).First(&exist).Error; qerr == nil {
			return &exist, nil
		}
		return nil, err
	}
	return result, nil
}

// GetCandidates 考生列表
func (s *ExamService) GetCandidates(examId uint, page, pageSize int) ([]model.ExamCandidate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ExamCandidate{})
	if examId > 0 {
		db = db.Where("exam_id = ?", examId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ExamCandidate
	err := db.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DeleteCandidate 删除考生
func (s *ExamService) DeleteCandidate(id uint) error {
	res := global.GVA_DB.Delete(&model.ExamCandidate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]考生不存在", ErrCodeCandBad)
	}
	return nil
}

// CandImportIn 名单导入请求
type CandImportIn struct {
	ExamId uint     `json:"examId" binding:"required"`
	Names  []string `json:"names" binding:"required,min=1"`
}

// ClaimIn 领券请求
type ClaimIn struct {
	ExamId  uint   `json:"examId" binding:"required"`
	UserKey string `json:"userKey" binding:"required"`
	Name    string `json:"name"`
}
