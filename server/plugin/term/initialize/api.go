// Package initialize 白泽终端插件 API 记录种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Api 注册终端 API 记录（WS 端点记录用于权限界面展示）
func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/term/ws", Description: "WebSSH 终端 WebSocket", ApiGroup: "终端作业", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
