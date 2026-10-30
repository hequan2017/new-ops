// Package api 监控告警规则与事件接口
package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
	"github.com/hequan2017/new-ops/server/plugin/monitor/service"
	"github.com/hequan2017/new-ops/server/utils"
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
	// webhook 查询串含钉钉 access_token：非超管截掉，防拿到即可伪造群推送
	if utils.GetUserAuthorityId(c) != 888 {
		for i := range list {
			if u := list[i].WebhookURL; u != "" {
				list[i].WebhookURL = strings.SplitN(u, "?", 2)[0]
			}
		}
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// eventRangeRequest 告警事件查询条件（时间范围按 fired_at 过滤）
type eventRangeRequest struct {
	request.PageInfo
	StartAt string `json:"startAt"` // 起始时间（YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD，含）
	EndAt   string `json:"endAt"`   // 结束时间（同上格式，日期格式含当日全天）
}

// parseEventRange 解析时间范围（两种格式兼容；纯日期的 end 取当日 23:59:59）；
// 两端均为空返回 nil,nil 不加过滤
func parseEventRange(startAt, endAt string) (start, end *time.Time, err error) {
	const layoutFull, layoutDate = "2006-01-02 15:04:05", "2006-01-02"
	parse := func(s string, isEnd bool) (*time.Time, error) {
		for _, l := range []string{layoutFull, layoutDate} {
			if t, e := time.ParseInLocation(l, s, time.Local); e == nil {
				if isEnd && l == layoutDate {
					t = t.Add(24*time.Hour - time.Second)
				}
				return &t, nil
			}
		}
		return nil, fmt.Errorf("时间格式不合法: %s（应为 YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）", s)
	}
	if startAt != "" {
		if start, err = parse(startAt, false); err != nil {
			return nil, nil, err
		}
	}
	if endAt != "" {
		if end, err = parse(endAt, true); err != nil {
			return nil, nil, err
		}
	}
	return start, end, nil
}

// ListEvents 告警事件分页
// @Tags MonitorAlert
// @Summary 告警事件列表（倒序分页，支持时间范围过滤）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body eventRangeRequest true "页码/页大小/startAt/endAt(按触发时间过滤)"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /monitor/alert/event/list [post]
func (a *alertApi) ListEvents(c *gin.Context) {
	var req eventRangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	start, end, err := parseEventRange(req.StartAt, req.EndAt)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	db := global.GVA_DB.Model(&model.MonitorAlertEvent{})
	if start != nil {
		db = db.Where("fired_at >= ?", *start)
	}
	if end != nil {
		db = db.Where("fired_at <= ?", *end)
	}
	var total int64
	db.Count(&total)
	var list []model.MonitorAlertEvent
	db.Scopes(req.Paginate()).Order("id DESC").Find(&list)
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: req.Page, PageSize: req.PageSize,
	}, "获取成功", c)
}

// ExportEvents 告警事件 CSV 导出
// @Tags MonitorAlert
// @Summary 告警事件导出 CSV（同一时间过滤口径，上限 10000 行）
// @Security ApiKeyAuth
// @Produce text/csv
// @Param startAt query string false "起始时间（YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD）"
// @Param endAt query string false "结束时间（同上格式，日期格式含当日全天）"
// @Success 200 {string} string "text/csv"
// @Router /monitor/alert/event/export [get]
func (a *alertApi) ExportEvents(c *gin.Context) {
	start, end, err := parseEventRange(c.Query("startAt"), c.Query("endAt"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	db := global.GVA_DB.Model(&model.MonitorAlertEvent{})
	if start != nil {
		db = db.Where("fired_at >= ?", *start)
	}
	if end != nil {
		db = db.Where("fired_at <= ?", *end)
	}
	var list []model.MonitorAlertEvent
	if err := db.Order("fired_at DESC").Limit(10000).Find(&list).Error; err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition",
		fmt.Sprintf(`attachment; filename="alert-events-%s.csv"`, time.Now().Format("20060102-150405")))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", alertEventsCSV(list))
}

// alertEventsCSV 事件列表转 CSV（BOM 前缀，Excel 直开中文不乱码）
func alertEventsCSV(list []model.MonitorAlertEvent) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("\uFEFF")
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"触发时间", "恢复时间", "资产ID", "摘要", "触发值", "已通知"})
	for _, e := range list {
		resolved := ""
		if e.ResolvedAt != nil {
			resolved = e.ResolvedAt.Format("2006-01-02 15:04:05")
		}
		_ = w.Write([]string{
			e.FiredAt.Format("2006-01-02 15:04:05"), resolved,
			strconv.FormatUint(uint64(e.AssetID), 10), e.Summary,
			strconv.FormatFloat(e.Value, 'f', -1, 64),
			map[bool]string{true: "是", false: "否"}[e.Notified],
		})
	}
	w.Flush()
	return buf.Bytes()
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
