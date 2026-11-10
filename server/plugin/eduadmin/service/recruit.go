// Package service eduadmin：智能招聘（需求 3.6-12）
// 口径：匹配度=关键词重合率×0.7 + 经验达标率×0.3（经验达标率=min(1,年限/max(最低年限,1))），
// 高≥0.7 / 中≥0.45 / 低；筛选=按融合分降序输出三档标记，纯函数可测。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduRecruitParam   = 2807
	ErrCodeEduRecruitNotFound = 2808
)

// splitSkills 逗号分隔技能集合（小写去空白）
func splitSkills(s string) map[string]bool {
	set := map[string]bool{}
	for _, part := range strings.Split(s, ",") {
		p := strings.ToLower(strings.TrimSpace(part))
		if p != "" {
			set[p] = true
		}
	}
	return set
}

// recruitMatch 匹配结果
type recruitMatch struct {
	model.EduResume
	Overlap float64 `json:"overlap"` // 技能重合率
	Score   float64 `json:"score"`   // 融合分
	Level   string  `json:"level"`   // 高/中/低
}

// matchResume 匹配打分（纯函数，测试锚点）
func matchResume(resume model.EduResume, position model.EduPosition) recruitMatch {
	req := splitSkills(position.RequiredSkill)
	own := splitSkills(resume.Skills)
	overlap := 0.0
	if len(req) > 0 {
		hit := 0
		for k := range req {
			if own[k] {
				hit++
			}
		}
		overlap = float64(hit) / float64(len(req))
	}
	minYears := position.MinYears
	if minYears < 1 {
		minYears = 1
	}
	expFactor := float64(resume.Years) / float64(minYears)
	if expFactor > 1 {
		expFactor = 1
	}
	if expFactor < 0 {
		expFactor = 0
	}
	score := round2(0.7*overlap + 0.3*expFactor)
	level := "低"
	switch {
	case score >= 0.7:
		level = "高"
	case score >= 0.45:
		level = "中"
	}
	return recruitMatch{EduResume: resume, Overlap: round2(overlap), Score: score, Level: level}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// ScreenCandidates 候选人筛选（按融合分降序）
func (s *EduadminService) ScreenCandidates(positionId uint) ([]recruitMatch, error) {
	var position model.EduPosition
	if err := global.GVA_DB.First(&position, positionId).Error; err != nil {
		return nil, fmt.Errorf("[%d]岗位不存在", ErrCodeEduRecruitNotFound)
	}
	var resumes []model.EduResume
	if err := global.GVA_DB.Order("id ASC").Limit(500).Find(&resumes).Error; err != nil {
		return nil, err
	}
	matches := make([]recruitMatch, 0, len(resumes))
	for _, r := range resumes {
		matches = append(matches, matchResume(r, position))
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].ID < matches[j].ID
	})
	return matches, nil
}

// SaveEduResume 新建/更新简历
func (s *EduadminService) SaveEduResume(in *model.EduResume) (*model.EduResume, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("[%d]姓名必填", ErrCodeEduRecruitParam)
	}
	if in.ID > 0 {
		var exist model.EduResume
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]简历不存在", ErrCodeEduRecruitNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"name": in.Name, "education": in.Education, "years": in.Years,
			"skills": in.Skills, "resume_text": in.ResumeText,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduResumes 简历列表
func (s *EduadminService) GetEduResumes(page, pageSize int) ([]model.EduResume, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduResume{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduResume
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveEduPosition 新建/更新岗位
func (s *EduadminService) SaveEduPosition(in *model.EduPosition) (*model.EduPosition, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.RequiredSkill) == "" {
		return nil, fmt.Errorf("[%d]岗位名与必备技能必填", ErrCodeEduRecruitParam)
	}
	if in.ID > 0 {
		var exist model.EduPosition
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]岗位不存在", ErrCodeEduRecruitNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"title": in.Title, "dept": in.Dept, "required_skill": in.RequiredSkill,
			"min_years": in.MinYears, "min_education": in.MinEducation, "headcount": in.Headcount,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduPositions 岗位列表
func (s *EduadminService) GetEduPositions(page, pageSize int) ([]model.EduPosition, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduPosition{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduPosition
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
