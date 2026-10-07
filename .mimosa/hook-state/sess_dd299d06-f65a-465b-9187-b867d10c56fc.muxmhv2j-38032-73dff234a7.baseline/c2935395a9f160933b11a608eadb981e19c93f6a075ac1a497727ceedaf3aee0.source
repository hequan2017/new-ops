// Package api 监控指标接口
package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
	"github.com/hequan2017/new-ops/server/plugin/monitor/service"
)

var MonitorApi = new(monitorApi)

type monitorApi struct{}

// GetMetrics 指标序列查询
// @Tags Monitor
// @Summary 主机性能指标序列（cpu/mem/disk/load1）
// @Security ApiKeyAuth
// @Produce application/json
// @Param assetId query int true "主机ID"
// @Param names query string false "指标名逗号分隔（默认全部）"
// @Param hours query int false "回看小时数（默认24）"
// @Success 200 {object} response.Response{data=[]model.MonitorMetric} "获取成功"
// @Router /monitor/metric/list [get]
func (a *monitorApi) GetMetrics(c *gin.Context) {
	assetID, err := strconv.ParseUint(c.Query("assetId"), 10, 64)
	if err != nil || assetID == 0 {
		response.FailWithMessage("assetId 无效", c)
		return
	}
	var names []string
	if ns := c.Query("names"); ns != "" {
		for _, n := range strings.Split(ns, ",") {
			n = strings.TrimSpace(n)
			if n == "" {
				continue
			}
			if !model.ValidMetricNames()[n] {
				response.FailWithMessage("非法指标名: "+n, c)
				return
			}
			names = append(names, n)
		}
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 || hours > 24*30 {
		hours = 24
	}
	list, err := service.Service.Monitor.GetMetrics(uint(assetID), names, time.Now().Add(-time.Duration(hours)*time.Hour))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CollectNow 手动触发全量采集
// @Tags Monitor
// @Summary 手动触发一次全量性能采集（异步）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response "已触发"
// @Router /monitor/metric/collect [post]
func (a *monitorApi) CollectNow(c *gin.Context) {
	svc := new(service.MonitorService)
	go svc.CollectAll()
	response.OkWithMessage("已触发采集", c)
}
