// Package v1 eduadmin API：数据清洗（需求 3.6-2）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// CleanEduRecords 批量清洗
// @Tags Eduadmin
// @Summary 数据清洗（乱码/格式过滤+缺失值填充 均值/前值+3σ 异常标记，dry 试运行）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.cleanReq true "清洗请求"
// @Success 200 {object} response.Response{data=map[string]any} "清洗完成"
// @Router /eduadmin/clean [post]
func (a *EduadminApi) CleanEduRecords(c *gin.Context) {
	var in service.CleanReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.CleanEduRecords(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "清洗完成", c)
}
