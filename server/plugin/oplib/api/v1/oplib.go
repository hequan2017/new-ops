// Package v1 oplib API：算子库接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
	"github.com/hequan2017/new-ops/server/plugin/oplib/service"
)

var Api = new(oplibApi)

type oplibApi struct{}

// GetOperators 算子列表
// @Tags Oplib
// @Summary 算子列表（分类/实现过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "数据处理/特征/检索/后处理/评估"
// @Param implType query string false "builtin/script"
// @Success 200 {object} response.Response{data=[]model.OplibOperator} "获取成功"
// @Router /oplib/operator/list [get]
func (a *oplibApi) GetOperators(c *gin.Context) {
	list, err := service.OplibApp.GetOperators(c.Query("category"), c.Query("implType"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveOperator 保存算子
// @Tags Oplib
// @Summary 新建/更新算子（schema 对象即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.OplibOperator true "算子"
// @Success 200 {object} response.Response{data=model.OplibOperator} "保存成功"
// @Router /oplib/operator [post]
func (a *oplibApi) SaveOperator(c *gin.Context) {
	var in model.OplibOperator
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	op, err := service.OplibApp.SaveOperator(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(op, "保存成功", c)
}

// DeleteOperator 删除算子
// @Tags Oplib
// @Summary 删除算子（软删除）
// @Security ApiKeyAuth
// @Param id query int true "算子ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /oplib/operator [delete]
func (a *oplibApi) DeleteOperator(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.OplibApp.DeleteOperator(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
