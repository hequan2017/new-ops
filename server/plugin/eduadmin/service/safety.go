// Package service eduadmin：校园安防-环境监测（需求 3.6-23）
// 口径：阈值告警三档——超阈值为高、≥0.8×阈值为中、其余正常；健康建议模板按
// 类型（空气=通风/净化，水质=停用送检）与等级输出；中/高自动落告警表。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduEnvParam    = 2823
	ErrCodeEduEnvNotFound = 2824
)

// envAlertLevel 告警等级（纯函数，测试锚点）：返回 空=正常/中/高
func envAlertLevel(value, threshold float64) string {
	if threshold <= 0 {
		return ""
	}
	switch {
	case value > threshold:
		return "高"
	case value >= threshold*0.8:
		return "中"
	default:
		return ""
	}
}

// envAdvice 健康建议模板（纯函数，测试锚点）
func envAdvice(envType, metric, level string) string {
	if level == "高" {
		switch envType {
		case "空气":
			return fmt.Sprintf("【空气-高】%s 超标：立即开窗强制通风或启动新风净化，疏散敏感人群并复测。", metric)
		case "水质":
			return fmt.Sprintf("【水质-高】%s 超标：立即停用该供水点，留样送检并排查管路污染。", metric)
		}
		return fmt.Sprintf("【%s-高】%s 超标：立即处置并复测。", envType, metric)
	}
	switch envType {
	case "空气":
		return fmt.Sprintf("【空气-中】%s 接近阈值：加强通风与监测频次（每 2 小时一次）。", metric)
	case "水质":
		return fmt.Sprintf("【水质-中】%s 接近阈值：加密采样检测，安排管路冲洗。", metric)
	}
	return fmt.Sprintf("【%s-中】%s 接近阈值：加密监测。", envType, metric)
}

// SaveEduEnvReading 登记环境读数（中/高自动告警落库）
func (s *EduadminService) SaveEduEnvReading(in *model.EduEnvReading) (*model.EduEnvReading, error) {
	if in.Type != "空气" && in.Type != "水质" {
		return nil, fmt.Errorf("[%d]类型须为 空气/水质", ErrCodeEduEnvParam)
	}
	if strings.TrimSpace(in.Location) == "" || strings.TrimSpace(in.Metric) == "" {
		return nil, fmt.Errorf("[%d]监测点与指标必填", ErrCodeEduEnvParam)
	}
	if in.RecordedAt == "" {
		return nil, fmt.Errorf("[%d]采集时间必填", ErrCodeEduEnvParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	if level := envAlertLevel(in.Value, in.Threshold); level != "" {
		alert := model.EduEnvAlert{
			ReadingId: in.ID, Type: in.Type, Location: in.Location, Metric: in.Metric,
			Level: level, Value: in.Value, Threshold: in.Threshold,
			Advice: envAdvice(in.Type, in.Metric, level),
		}
		if err := global.GVA_DB.Create(&alert).Error; err != nil {
			return nil, err
		}
	}
	return in, nil
}

// GetEduEnvReadings 读数列表（类型过滤，倒序分页）
func (s *EduadminService) GetEduEnvReadings(envType string, page, pageSize int) ([]model.EduEnvReading, int64, error) {
	if envType != "" && envType != "空气" && envType != "水质" {
		return nil, 0, fmt.Errorf("[%d]类型须为 空气/水质", ErrCodeEduEnvParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduEnvReading{})
	if envType != "" {
		db = db.Where("type = ?", envType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduEnvReading
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetEduEnvAlerts 告警列表（倒序，近 500 条）
func (s *EduadminService) GetEduEnvAlerts() ([]model.EduEnvAlert, error) {
	var list []model.EduEnvAlert
	err := global.GVA_DB.Order("id DESC").Limit(500).Find(&list).Error
	return list, err
}
