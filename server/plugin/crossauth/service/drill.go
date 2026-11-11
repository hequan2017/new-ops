// Package service crossauth：应急演练与预案（需求 3.7-16）
// 口径：预案 10 类事件种子幂等（响应流程/责任分工/处置措施），更新即版本+1；
// 演练计划年度 2 次口径字段（达成判定纯函数）。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaDrillParam    = 2925
	ErrCodeCaDrillNotFound = 2926
)

// drillTargetMet 年度演练达成（纯函数，测试锚点）：执行数 ≥ 年度目标(2)
func drillTargetMet(executed, yearlyTarget int) bool {
	if yearlyTarget < 1 {
		yearlyTarget = 2
	}
	return executed >= yearlyTarget
}

// SaveCaPlaybook 新建/更新预案（更新版本+1）
func (s *CaService) SaveCaPlaybook(in *model.CaPlaybook) (*model.CaPlaybook, error) {
	if strings.TrimSpace(in.EventName) == "" || strings.TrimSpace(in.ResponseFlow) == "" {
		return nil, fmt.Errorf("[%d]事件类型与响应流程必填", ErrCodeCaDrillParam)
	}
	if in.Status == "" {
		in.Status = "启用"
	}
	var exist model.CaPlaybook
	if err := global.GVA_DB.Where("event_name = ?", in.EventName).First(&exist).Error; err == nil {
		// 版本管理：内容更新版本+1
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"response_flow": in.ResponseFlow, "duties": in.Duties, "measures": in.Measures,
			"version": exist.Version + 1, "status": in.Status,
		}).Error
		if err == nil {
			exist.Version++
		}
		return &exist, err
	}
	in.Version = 1
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaPlaybooks 预案列表
func (s *CaService) GetCaPlaybooks() ([]model.CaPlaybook, error) {
	var list []model.CaPlaybook
	err := global.GVA_DB.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// SaveCaDrillPlan 新建/更新演练计划
func (s *CaService) SaveCaDrillPlan(in *model.CaDrillPlan) (*model.CaDrillPlan, error) {
	if strings.TrimSpace(in.PlaybookName) == "" || in.Year < 2000 {
		return nil, fmt.Errorf("[%d]预案事件与年度必填", ErrCodeCaDrillParam)
	}
	var exist model.CaPlaybook
	if err := global.GVA_DB.Where("event_name = ?", in.PlaybookName).First(&exist).Error; err != nil {
		return nil, fmt.Errorf("[%d]预案不存在: %s", ErrCodeCaDrillNotFound, in.PlaybookName)
	}
	if in.YearlyTarget < 1 {
		in.YearlyTarget = 2
	}
	var count int64
	global.GVA_DB.Model(&model.CaDrillPlan{}).Where("playbook_name = ? AND year = ?", in.PlaybookName, in.Year).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]该预案本年度已有演练计划（年度 %d 次口径）", ErrCodeCaDrillParam, in.YearlyTarget)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// FinishCaDrillPlan 演练复盘归档
func (s *CaService) FinishCaDrillPlan(id uint, report, grade string) (*model.CaDrillPlan, error) {
	if strings.TrimSpace(report) == "" {
		return nil, fmt.Errorf("[%d]复盘报告必填", ErrCodeCaDrillParam)
	}
	if grade != "优" && grade != "良" && grade != "需改进" {
		return nil, fmt.Errorf("[%d]评定须为 优/良/需改进", ErrCodeCaDrillParam)
	}
	var plan model.CaDrillPlan
	if err := global.GVA_DB.First(&plan, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]演练计划不存在", ErrCodeCaDrillNotFound)
	}
	now := time.Now().In(caCst).Format("2006-01-02 15:04")
	if err := global.GVA_DB.Model(&plan).Updates(map[string]any{
		"report": report, "grade": grade, "executed_at": now,
	}).Error; err != nil {
		return nil, err
	}
	plan.Report = report
	plan.Grade = grade
	plan.ExecutedAt = now
	return &plan, nil
}

// GetCaDrillPlans 演练计划列表（年度过滤）
func (s *CaService) GetCaDrillPlans(year int) ([]model.CaDrillPlan, error) {
	db := global.GVA_DB.Model(&model.CaDrillPlan{})
	if year > 0 {
		db = db.Where("year = ?", year)
	}
	var list []model.CaDrillPlan
	err := db.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}
