// Package service eduadmin：设施维护（需求 3.6-22）
// 口径：故障预测联动——按 P12 EduMaintPlan 等级自动派发（严重→机电班组，轻微→综合
// 维修班组）；描述关键字路由（水电/漏水/断电→水电班组，优先级高于等级路由）；
// 状态机 待派→处理中→已闭环；工单号 EDU-年月日-序号。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduRepairParam    = 2821
	ErrCodeEduRepairNotFound = 2822
)

// routeTeam 班组路由规则（纯函数，测试锚点）：关键字 > 等级 > 默认综合维修
func routeTeam(level, description string) string {
	d := description
	if strings.Contains(d, "水电") || strings.Contains(d, "漏水") || strings.Contains(d, "断电") || strings.Contains(d, "水管") {
		return "水电班组"
	}
	switch level {
	case "严重":
		return "机电班组"
	case "轻微":
		return "综合维修班组"
	default:
		return "综合维修班组"
	}
}

// genRepairOrderNo 工单号：EDU-yyyyMMdd-序号（当日序号递增）
func (s *EduadminService) genRepairOrderNo() string {
	today := time.Now().Format("20060102")
	var count int64
	global.GVA_DB.Model(&model.EduRepairOrder{}).Where("order_no LIKE ?", "EDU-"+today+"-%").Count(&count)
	return fmt.Sprintf("EDU-%s-%03d", today, count+1)
}

// CreateEduRepairOrder 创建工单（故障预测联动自动派发）
func (s *EduadminService) CreateEduRepairOrder(assetId uint, description string) (*model.EduRepairOrder, error) {
	if strings.TrimSpace(description) == "" {
		return nil, fmt.Errorf("[%d]故障描述必填", ErrCodeEduRepairParam)
	}
	order := &model.EduRepairOrder{
		AssetId: assetId, Description: description, Status: "待派",
	}
	if assetId > 0 {
		var plan model.EduMaintPlan
		if err := global.GVA_DB.Where("asset_id = ?", assetId).First(&plan).Error; err == nil {
			order.Level = plan.Level
		}
	}
	if order.Level == "" {
		order.Level = "常规"
	}
	order.Team = routeTeam(order.Level, description)
	if order.Level != "常规" {
		order.Status = "处理中" // 预测等级联动：跳过人工待派
	}
	order.OrderNo = s.genRepairOrderNo()
	if err := global.GVA_DB.Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

// AssignEduRepairOrder 人工派单（待派→处理中）
func (s *EduadminService) AssignEduRepairOrder(id uint, team string) (*model.EduRepairOrder, error) {
	if strings.TrimSpace(team) == "" {
		return nil, fmt.Errorf("[%d]班组必填", ErrCodeEduRepairParam)
	}
	var order model.EduRepairOrder
	if err := global.GVA_DB.First(&order, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduRepairNotFound)
	}
	if order.Status != "待派" {
		return nil, fmt.Errorf("[%d]仅待派工单可派发，当前 %s", ErrCodeEduRepairParam, order.Status)
	}
	err := global.GVA_DB.Model(&order).Updates(map[string]any{"team": team, "status": "处理中"}).Error
	return &order, err
}

// CloseEduRepairOrder 闭环（处理中→已闭环）
func (s *EduadminService) CloseEduRepairOrder(id uint) (*model.EduRepairOrder, error) {
	var order model.EduRepairOrder
	if err := global.GVA_DB.First(&order, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]工单不存在", ErrCodeEduRepairNotFound)
	}
	if order.Status != "处理中" {
		return nil, fmt.Errorf("[%d]仅处理中工单可闭环，当前 %s", ErrCodeEduRepairParam, order.Status)
	}
	err := global.GVA_DB.Model(&order).Update("status", "已闭环").Error
	return &order, err
}

// GetEduRepairOrders 工单列表（状态过滤，倒序分页）
func (s *EduadminService) GetEduRepairOrders(status string, page, pageSize int) ([]model.EduRepairOrder, int64, error) {
	if status != "" && status != "待派" && status != "处理中" && status != "已闭环" {
		return nil, 0, fmt.Errorf("[%d]状态须为 待派/处理中/已闭环", ErrCodeEduRepairParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduRepairOrder{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduRepairOrder
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
