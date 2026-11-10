// Package service eduadmin：资产故障预测（需求 3.6-18）
// 口径：时序读数线性回归斜率（复用 talent.go linearRegression）+ 最新读数阈值
// 双因子判级——严重=最新≥阈值；轻微=最新≥0.8×阈值 或 斜率>阈值×0.02/步；其余无故障；
// 维护计划模板（严重立即停机/轻微 7 日巡检/无故障月检），纯函数可测。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduAssetParam    = 2816
	ErrCodeEduAssetNotFound = 2817
)

// faultLevel 故障等级判定（纯函数，测试锚点）
func faultLevel(threshold, slope, latest float64) string {
	if threshold <= 0 {
		return "无故障"
	}
	if latest >= threshold {
		return "严重"
	}
	slopeTol := threshold * 0.02
	if latest >= threshold*0.8 || slope > slopeTol {
		return "轻微"
	}
	return "无故障"
}

// maintenancePlan 维护计划模板（纯函数，测试锚点）
func maintenancePlan(level, assetName string) string {
	switch level {
	case "严重":
		return fmt.Sprintf("【%s】已达故障阈值：立即停机检修，通知设备科与使用部门，预计停机 1-2 个工作日。", assetName)
	case "轻微":
		return fmt.Sprintf("【%s】指标接近阈值或呈上升趋势：7 日内安排巡检复测，必要时降载运行。", assetName)
	default:
		return fmt.Sprintf("【%s】运行正常：纳入常规月度保养计划。", assetName)
	}
}

// PredictAssetFault 故障预测（读数+判定+计划落库，按设备幂等覆盖）
func (s *EduadminService) PredictAssetFault(assetId uint) (*model.EduMaintPlan, error) {
	var asset model.EduAsset
	if err := global.GVA_DB.First(&asset, assetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]设备不存在", ErrCodeEduAssetNotFound)
	}
	var readings []model.EduSensorReading
	if err := global.GVA_DB.Where("asset_id = ?", assetId).Order("id ASC").Limit(200).Find(&readings).Error; err != nil {
		return nil, err
	}
	if len(readings) < 2 {
		return nil, fmt.Errorf("[%d]传感器读数须 ≥2 条方可预测", ErrCodeEduAssetParam)
	}
	xs := make([]float64, len(readings))
	ys := make([]float64, len(readings))
	for i, r := range readings {
		xs[i] = float64(i)
		ys[i] = r.MetricValue
	}
	slope, _, err := linearRegression(xs, ys)
	if err != nil {
		return nil, err
	}
	latest := readings[len(readings)-1].MetricValue
	level := faultLevel(asset.Threshold, slope, latest)
	plan := model.EduMaintPlan{
		AssetId: assetId, Level: level, Slope: round2(slope), Latest: latest,
		Plan: maintenancePlan(level, asset.Name),
	}
	var exist model.EduMaintPlan
	if err := global.GVA_DB.Where("asset_id = ?", assetId).First(&exist).Error; err == nil {
		if err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"level": level, "slope": plan.Slope, "latest": latest, "plan": plan.Plan,
		}).Error; err != nil {
			return nil, err
		}
		return &exist, nil
	}
	if err := global.GVA_DB.Create(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// SaveEduAsset 新建/更新设备
func (s *EduadminService) SaveEduAsset(in *model.EduAsset) (*model.EduAsset, error) {
	if strings.TrimSpace(in.Name) == "" || in.Threshold <= 0 {
		return nil, fmt.Errorf("[%d]设备名与阈值必填", ErrCodeEduAssetParam)
	}
	if in.ID > 0 {
		var exist model.EduAsset
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]设备不存在", ErrCodeEduAssetNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"name": in.Name, "location": in.Location, "metric": in.Metric, "threshold": in.Threshold,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduAssets 设备列表
func (s *EduadminService) GetEduAssets() ([]model.EduAsset, error) {
	var list []model.EduAsset
	err := global.GVA_DB.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// AddEduSensorReading 新增读数
func (s *EduadminService) AddEduSensorReading(in *model.EduSensorReading) (*model.EduSensorReading, error) {
	var asset model.EduAsset
	if err := global.GVA_DB.First(&asset, in.AssetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]设备不存在", ErrCodeEduAssetNotFound)
	}
	if in.RecordedAt == "" {
		return nil, fmt.Errorf("[%d]采集时间必填", ErrCodeEduAssetParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduMaintPlans 维护计划列表
func (s *EduadminService) GetEduMaintPlans() ([]model.EduMaintPlan, error) {
	var list []model.EduMaintPlan
	err := global.GVA_DB.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}
