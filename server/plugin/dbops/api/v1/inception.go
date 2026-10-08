// Package api 白泽数据库工单：审核引擎配置与执行接口（M6 收官）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	dbModel "github.com/hequan2017/new-ops/server/plugin/dbops/model"
	"github.com/hequan2017/new-ops/server/utils"
)

// GetInceptionConfig 读审核引擎配置
// @Tags DbopsInception
// @Summary 读取 goInception 审核引擎配置（密码不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=dbModel.DbopsInceptionConfig} "获取成功"
// @Router /dbops/inception [get]
func (a *dbopsApi) GetInceptionConfig(c *gin.Context) {
	cfg, err := dbSvc.GetInceptionConfig()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(cfg, c)
}

// SaveInceptionConfig 保存审核引擎配置
// @Tags DbopsInception
// @Summary 保存 goInception 审核引擎配置（备份密码留空不改；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body dbModel.DbopsInceptionConfig true "引擎/备份库配置"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /dbops/inception [post]
func (a *dbopsApi) SaveInceptionConfig(c *gin.Context) {
	var cfg dbModel.DbopsInceptionConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	if err := dbSvc.SaveInceptionConfig(&cfg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// TestInception 测试审核引擎连通
// @Tags DbopsInception
// @Summary 测试 goInception 引擎连通性
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=string,msg=string} "测试通过"
// @Router /dbops/inception/test [post]
func (a *dbopsApi) TestInception(c *gin.Context) {
	ok, msg, err := dbSvc.TestInception()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if !ok {
		response.FailWithMessage(msg, c)
		return
	}
	response.OkWithMessage(msg, c)
}

// ExecuteOrder 执行已审核工单
// @Tags DbopsOrder
// @Summary 执行审核通过的 SQL 工单（goInception 执行+备份回滚查询；写操作仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=dbModel.DbopsOrder,msg=string} "执行完成"
// @Router /dbops/order/execute [post]
func (a *dbopsApi) ExecuteOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	order, err := dbSvc.ExecuteOrder(uint(id), utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(order, "执行完成", c)
}
