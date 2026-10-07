// Package api SFTP 文件浏览器接口
package api

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	termSvc "github.com/hequan2017/new-ops/server/plugin/term/service"
)

type sftpApi struct{}

// SftpList 列目录
// @Tags Sftp
// @Summary 列目录
// @Security ApiKeyAuth
// @Produce application/json
// @Param hostId query int true "主机ID"
// @Param credentialId query int true "SSH凭据ID"
// @Param path query string false "远端路径（默认 /root）"
// @Success 200 {object} response.Response{data=[]service.SftpEntry} "获取成功"
// @Router /term/sftp/list [get]
func (s *sftpApi) SftpList(c *gin.Context) {
	hostID, _ := parseUintQuery(c.Query("hostId"))
	credID, _ := parseUintQuery(c.Query("credentialId"))
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(hostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	entries, err := termSvc.Service.SftpService.SftpList(host, credID, c.DefaultQuery("path", "/root"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(entries, c)
}

// SftpMkdir 新建目录
// @Tags Sftp
// @Summary 新建目录
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "hostId/credentialId/path"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /term/sftp/mkdir [post]
func (s *sftpApi) SftpMkdir(c *gin.Context) {
	var req struct {
		HostID       uint   `json:"hostId" binding:"required"`
		CredentialID uint   `json:"credentialId" binding:"required"`
		Path         string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(req.HostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	if err := termSvc.Service.SftpService.SftpMkdir(host, req.CredentialID, req.Path); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// SftpDelete 删除文件/目录
// @Tags Sftp
// @Summary 删除文件或目录（目录递归）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "hostId/credentialId/path"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /term/sftp/delete [post]
func (s *sftpApi) SftpDelete(c *gin.Context) {
	var req struct {
		HostID       uint   `json:"hostId" binding:"required"`
		CredentialID uint   `json:"credentialId" binding:"required"`
		Path         string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(req.HostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	if err := termSvc.Service.SftpService.SftpDelete(host, req.CredentialID, req.Path); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// SftpRename 重命名/移动
// @Tags Sftp
// @Summary 重命名/移动
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "hostId/credentialId/from/to"
// @Success 200 {object} response.Response{msg=string} "重命名成功"
// @Router /term/sftp/rename [post]
func (s *sftpApi) SftpRename(c *gin.Context) {
	var req struct {
		HostID       uint   `json:"hostId" binding:"required"`
		CredentialID uint   `json:"credentialId" binding:"required"`
		From         string `json:"from" binding:"required"`
		To           string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(req.HostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	if err := termSvc.Service.SftpService.SftpRename(host, req.CredentialID, req.From, req.To); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("重命名成功", c)
}

// SftpDownload 下载远端文件
// @Tags Sftp
// @Summary 下载远端文件
// @Security ApiKeyAuth
// @Produce application/octet-stream
// @Param hostId query int true "主机ID"
// @Param credentialId query int true "SSH凭据ID"
// @Param path query string true "远端文件路径"
// @Success 200 {file} file "文件流"
// @Router /term/sftp/download [get]
func (s *sftpApi) SftpDownload(c *gin.Context) {
	hostID, _ := parseUintQuery(c.Query("hostId"))
	credID, _ := parseUintQuery(c.Query("credentialId"))
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(hostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	src, name, size, cleanup, err := termSvc.Service.SftpService.SftpDownload(host, credID, c.Query("path"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer cleanup()
	c.Header("Content-Disposition", "attachment; filename=\""+path.Base(name)+"\"")
	c.Header("Content-Length", fmt.Sprint(size))
	c.Data(http.StatusOK, "application/octet-stream", streamToBytes(src))
}

// SftpUpload 上传文件（覆盖远端同名）
// @Tags Sftp
// @Summary 上传文件
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param hostId formData int true "主机ID"
// @Param credentialId formData int true "SSH凭据ID"
// @Param path formData string true "远端目录"
// @Param file formData file true "文件"
// @Success 200 {object} response.Response{msg=string} "上传成功"
// @Router /term/sftp/upload [post]
func (s *sftpApi) SftpUpload(c *gin.Context) {
	hostID, _ := parseUintQuery(c.PostForm("hostId"))
	credID, _ := parseUintQuery(c.PostForm("credentialId"))
	dir := c.PostForm("path")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择上传文件", c)
		return
	}
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(hostID)
	if err != nil {
		response.FailWithMessage("主机不存在", c)
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer src.Close()
	target := dir + "/" + fileHeader.Filename
	if err := termSvc.Service.SftpService.SftpUpload(host, credID, target, src, fileHeader.Size); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("上传成功", c)
}

// streamToBytes 读取全部内容（SFTP 文件有限大小场景）
func streamToBytes(r interface{ Read([]byte) (int, error) }) []byte {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := r.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf
}
