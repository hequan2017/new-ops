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
		{Path: "/term/logtail", Description: "远程日志 tail WebSocket", ApiGroup: "终端作业", Method: "GET"},
		{Path: "/term/session/list", Description: "会话分页列表", ApiGroup: "终端作业", Method: "POST"},
		{Path: "/term/session/streams", Description: "会话流镜像（回放）", ApiGroup: "终端作业", Method: "GET"},
		{Path: "/term/session/commands", Description: "会话命令列表", ApiGroup: "终端作业", Method: "GET"},
		{Path: "/term/sftp/list", Description: "SFTP 文件列表", ApiGroup: "终端作业", Method: "GET"},
		{Path: "/term/sftp/mkdir", Description: "SFTP 新建目录", ApiGroup: "终端作业", Method: "POST"},
		{Path: "/term/sftp/delete", Description: "SFTP 删除", ApiGroup: "终端作业", Method: "POST"},
		{Path: "/term/sftp/rename", Description: "SFTP 重命名", ApiGroup: "终端作业", Method: "POST"},
		{Path: "/term/sftp/upload", Description: "SFTP 上传", ApiGroup: "终端作业", Method: "POST"},
		{Path: "/term/sftp/download", Description: "SFTP 下载", ApiGroup: "终端作业", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
