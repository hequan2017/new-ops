// Package router SFTP 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/term/api/v1"
)

type SftpRouter struct{}

// InitSftpRouter SFTP 文件浏览器路由（private + casbin）
func (r *SftpRouter) InitSftpRouter(Router *gin.RouterGroup) {
	sftpGroup := Router.Group("term/sftp")
	{
		sftpGroup.GET("list", v1.Api.Sftp.SftpList)
		sftpGroup.POST("mkdir", v1.Api.Sftp.SftpMkdir)
		sftpGroup.POST("delete", v1.Api.Sftp.SftpDelete)
		sftpGroup.POST("rename", v1.Api.Sftp.SftpRename)
		sftpGroup.POST("upload", v1.Api.Sftp.SftpUpload)
		sftpGroup.GET("download", v1.Api.Sftp.SftpDownload)
	}
}
