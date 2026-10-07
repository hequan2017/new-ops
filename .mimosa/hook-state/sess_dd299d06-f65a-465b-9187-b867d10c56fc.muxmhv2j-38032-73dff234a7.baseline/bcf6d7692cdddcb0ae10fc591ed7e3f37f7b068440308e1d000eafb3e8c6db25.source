// Package service SFTP 文件浏览服务（M2 场21）
// 依赖登记：github.com/pkg/sftp（DEV_PLAN 3.5 已更新），经保险库凭据 DialSSH 后建立 SFTP 会话。
package service

import (
	"fmt"
	"io"
	"path"
	"sort"
	"time"

	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"github.com/pkg/sftp"
)

// sftpClientFor 为主机建立 SFTP 客户端（凭据取用→DialSSH→sftp.NewClient）
func sftpClientFor(host *model.AssetHost, credentialID uint) (*sftp.Client, func(), error) {
	secret, credType, username, err := assetSvc.Service.CredCredentialService.GetPlaintext(credentialID)
	if err != nil {
		return nil, nil, fmt.Errorf("凭据读取失败: %w", err)
	}
	if credType != model.CredTypeSSHPassword && credType != model.CredTypeSSHKey {
		return nil, nil, fmt.Errorf("SFTP 仅支持 SSH 密码/私钥凭据")
	}
	client, _, err := assetSvc.DialSSH(host.IP, assetSvc.SSHAuth{Username: username, Password: secret, PrivateKey: secret})
	if err != nil {
		return nil, nil, err
	}
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		_ = client.Close()
		return nil, nil, fmt.Errorf("建立 SFTP 会话失败: %w", err)
	}
	cleanup := func() {
		_ = sftpClient.Close()
		_ = client.Close()
	}
	return sftpClient, cleanup, nil
}

// SftpEntry 目录条目
type SftpEntry struct {
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"modTime"`
}

// normalizeSftpPath 规范化路径（默认 /root，禁止越级逃逸）
func normalizeSftpPath(p string) string {
	if p == "" {
		return "/root"
	}
	clean := path.Clean("/" + p)
	return clean
}

// SftpService SFTP 文件服务
type SftpService struct{}

// SftpList 列目录（目录优先，按名排序）
func (s *SftpService) SftpList(host *model.AssetHost, credentialID uint, remotePath string) ([]SftpEntry, error) {
	dir := normalizeSftpPath(remotePath)
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	infos, err := cl.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}
	entries := make([]SftpEntry, 0, len(infos))
	for _, fi := range infos {
		entries = append(entries, SftpEntry{
			Name: fi.Name(), IsDir: fi.IsDir(), Size: fi.Size(),
			Mode: fi.Mode().String(), ModTime: fi.ModTime(),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

// SftpMkdir 新建目录
func (s *SftpService) SftpMkdir(host *model.AssetHost, credentialID uint, remotePath string) error {
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return err
	}
	defer cleanup()
	return cl.MkdirAll(normalizeSftpPath(remotePath))
}

// SftpDelete 删除文件或空目录（目录递归删除）
func (s *SftpService) SftpDelete(host *model.AssetHost, credentialID uint, remotePath string) error {
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return err
	}
	defer cleanup()
	target := normalizeSftpPath(remotePath)
	fi, err := cl.Stat(target)
	if err != nil {
		return fmt.Errorf("目标不存在: %w", err)
	}
	if fi.IsDir() {
		return s.removeDir(cl, target)
	}
	return cl.Remove(target)
}

func (s *SftpService) removeDir(cl *sftp.Client, dir string) error {
	entries, err := cl.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		sub := dir + "/" + e.Name()
		if e.IsDir() {
			if err := s.removeDir(cl, sub); err != nil {
				return err
			}
		} else if err := cl.Remove(sub); err != nil {
			return err
		}
	}
	return cl.RemoveDirectory(dir)
}

// SftpRename 重命名/移动
func (s *SftpService) SftpRename(host *model.AssetHost, credentialID uint, from, to string) error {
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return err
	}
	defer cleanup()
	return cl.PosixRename(normalizeSftpPath(from), normalizeSftpPath(to))
}

// SftpDownload 打开远端文件读流（调用方负责关闭）
func (s *SftpService) SftpDownload(host *model.AssetHost, credentialID uint, remotePath string) (io.ReadCloser, string, int64, func(), error) {
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return nil, "", 0, nil, err
	}
	target := normalizeSftpPath(remotePath)
	fi, err := cl.Stat(target)
	if err != nil {
		cleanup()
		return nil, "", 0, nil, fmt.Errorf("文件不存在: %w", err)
	}
	if fi.IsDir() {
		cleanup()
		return nil, "", 0, nil, fmt.Errorf("目标为目录，不能下载")
	}
	src, err := cl.Open(target)
	if err != nil {
		cleanup()
		return nil, "", 0, nil, err
	}
	name := path.Base(target)
	return src, name, fi.Size(), cleanup, nil
}

// SftpUpload 写入远端文件（覆盖）
func (s *SftpService) SftpUpload(host *model.AssetHost, credentialID uint, remotePath string, src io.Reader, size int64) error {
	cl, cleanup, err := sftpClientFor(host, credentialID)
	if err != nil {
		return err
	}
	defer cleanup()
	target := normalizeSftpPath(remotePath)
	dst, err := cl.Create(target)
	if err != nil {
		return fmt.Errorf("创建远端文件失败: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}
	return nil
}
