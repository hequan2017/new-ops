// Package router 白泽批量作业路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/job/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册批量作业路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	exec := Router.Group("job/exec")
	{
		exec.POST("", v1.Api.BatchExec.CreateBatchExec)
		exec.POST("cancel", v1.Api.BatchExec.CancelBatchExec)
		exec.POST("list", v1.Api.BatchExec.GetBatchList)
		exec.GET("detail", v1.Api.BatchExec.GetBatchDetail)
	}
	script := Router.Group("job/script")
	{
		script.POST("", v1.Api.JobScript.CreateScript)
		script.PUT("", v1.Api.JobScript.UpdateScript)
		script.DELETE("", v1.Api.JobScript.DeleteScript)
		script.GET("list", v1.Api.JobScript.GetScriptList)
		script.GET("versions", v1.Api.JobScript.GetScriptVersions)
	}
	vargroup := Router.Group("job/vargroup")
	{
		vargroup.POST("", v1.Api.JobScript.CreateVariableGroup)
		vargroup.PUT("", v1.Api.JobScript.UpdateVariableGroup)
		vargroup.DELETE("", v1.Api.JobScript.DeleteVariableGroup)
		vargroup.GET("list", v1.Api.JobScript.GetVariableGroupList)
	}
}
