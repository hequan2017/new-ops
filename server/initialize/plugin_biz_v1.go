package initialize

import (
	"github.com/gin-gonic/gin"
)

// bizPluginV1 旧式插件挂载占位：email 已迁移至 v2 接口化注册（plugin/register.go 空导入），
// 仅保留 router_biz.go 的 holder 以维持底座 router_biz 兼容面
func bizPluginV1(group ...*gin.RouterGroup) {
	private := group[0]
	public := group[1]
	holder(public, private)
}
