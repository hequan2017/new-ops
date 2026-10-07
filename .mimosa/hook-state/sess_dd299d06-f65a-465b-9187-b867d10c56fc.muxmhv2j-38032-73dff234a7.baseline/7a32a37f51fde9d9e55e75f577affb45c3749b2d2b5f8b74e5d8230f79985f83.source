// Package api 监控告警规则与事件接口
package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
	"github.com/hequan2017/new-ops/server/plugin/monitor/service"
)

type alertApi struct{}

// AlertApi 告警接口实例
var AlertApi = new(alertApi)

// CreateRule 创建告警规则
// @Tags MonitorAlert
// @Summary 创建告警规则（指标阈值/端口探活，可配钉钉 webhook）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MonitorAlertRule true "规则"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /monitor/alert/rule [post]
func (a *alertApi) CreateRule(c *gin.Context) {
	var r model.MonitorAlertRule
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := validateRule(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.GVA_DB.Create(&r).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateRule 更新告警规则
// @Tags MonitorAlert
// @Summary 更新告警规则
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.MonitorAlertRule true "含ID"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /monitor/alert/rule [put]
func (a *alertApi) UpdateRule(c *gin.Context) {
	var r model.MonitorAlertRule
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := validateRule(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := global.GVA_DB.Model(&model.MonitorAlertRule{}).Where("id = ?", r.ID).Updates(map[string]any{
		"asset_id": r.AssetID, "type": r.Type, "metric_name": r.MetricName,
		"operator": r.Operator, "threshold": r.Threshold, "port": r.Port,
		"duration": r.Duration, "silence_min": r.SilenceMin,
		"webhook_url": r.WebhookURL, "enabled": r.Enabled, "notes": r.Notes,
	}).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteRule 删除告警规则
// @Tags MonitorAlert
// @Summary 删除告警规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "规则ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /monitor/alert/rule [delete]
func (a *alertApi) DeleteRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	global.GVA_DB.Delete(&model.MonitorAlertRule{}, id)
	response.OkWithMessage("删除成功", c)
}

// ListRules 规则列表
// @Tags MonitorAlert
// @Summary 告警规则列表（全量）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.MonitorAlertRule} "获取成功"
// @Router /monitor/alert/rule/list [get]
func (a *alertApi) ListRules(c *gin.Context) {
	var list []model.MonitorAlertRule
	global.GVA_DB.Order("id DESC").Limit(500).Find(&list)
	response.OkWithDetailed(list, "获取成功", c)
}

// ListEvents 告警事件分页
// @Tags MonitorAlert
// @Summary 告警事件列表（倒序分页）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /monitor/alert/event/list [post]
func (a *alertApi) ListEvents(c *gin.Context) {
	var info request.PageInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var total int64
	db := global.GVA_DB.Model(&model.MonitorAlertEvent{})
	db.Count(&total)
	var list []model.MonitorAlertEvent
	db.Scopes(info.Paginate()).Order("id DESC").Find(&list)
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// validateRule 规则校验（类型/字段/http(s) webhook）
func validateRule(r *model.MonitorAlertRule) error {
	if r.AssetID == 0 {
		return errInvalidRule("必须关联资产")
	}
	switch r.Type {
	case model.RuleTypeMetric:
		if !model.ValidMetricNames()[r.MetricName] {
			return errInvalidRule("非法指标名: " + r.MetricName)
		}
		if r.Operator != model.OpGT && r.Operator != model.OpLT {
			return errInvalidRule("比较符仅支持 > 或 <")
		}
	case model.RuleTypePort:
		if r.Port <= 0 || r.Port > 65535 {
			return errInvalidRule("端口非法")
		}
	default:
		return errInvalidRule("类型必须为 metric 或 port")
	}
	if r.WebhookURL != "" && !strings.HasPrefix(r.WebhookURL, "http://") && !strings.HasPrefix(r.WebhookURL, "https://") {
		return errInvalidRule("webhook 必须为 http(s) 地址")
	}
	return nil
}

func errInvalidRule(msg string) error { return &ruleErr{msg} }

type ruleErr struct{ msg string }

func (e *ruleErr) Error() string { return e.msg }

var _ = service.Service // 引用保持（服务在 alert_engine 中直接经 global.DB 工作）
