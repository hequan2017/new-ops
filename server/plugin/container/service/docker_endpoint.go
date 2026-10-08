// Package service 白泽容器管理：Docker 接入点（M4 C1）
// 连接：unix socket 直连 / TCP+TLS（证书组合走凭据保险库 docker_tls，secret 为 JSON{ca,cert,key}）；
// 巡检：30s 合并巡检（单 goroutine 遍历全部接入点，tianqi 模式），状态/版本回写。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/hequan2017/new-ops/server/global"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
	"gorm.io/gorm"
)

// container 插件错误码段：1400-1499（DEV_PLAN 3.6）
const (
	ErrCodeEpNameRequired  = 1401
	ErrCodeEpDuplicate     = 1402
	ErrCodeEpAddrInvalid   = 1403
	ErrCodeEpNotFound      = 1404
	ErrCodeEpTLSInvalid    = 1405
	ErrCodeEpConnectFailed = 1406
)

// ContainerService 容器管理服务
type ContainerService struct{}

// EndpointService 接入点服务
type EndpointService struct{}

// Service 服务聚合出口
var Service = new(containerSvc)

type containerSvc struct {
	Endpoint EndpointService
	Compose  ComposeService
}

func newCtErr(code int, msg string) error { return &containerError{Code: code, Msg: msg} }

type containerError struct {
	Code int
	Msg  string
}

func (e *containerError) Error() string { return e.Msg }

// validateEndpoint 接入点校验（纯逻辑，可单测）
func validateEndpoint(ep *model.DockerEndpoint) error {
	if ep == nil || strings.TrimSpace(ep.Name) == "" {
		return newCtErr(ErrCodeEpNameRequired, "接入点名称不能为空")
	}
	if !strings.HasPrefix(ep.Addr, "unix://") && !strings.HasPrefix(ep.Addr, "tcp://") {
		return newCtErr(ErrCodeEpAddrInvalid, fmt.Sprintf("地址必须以 unix:// 或 tcp:// 开头: %s", ep.Addr))
	}
	if strings.HasPrefix(ep.Addr, "tcp://") && ep.TLSCredentialID != nil && *ep.TLSCredentialID == 0 {
		ep.TLSCredentialID = nil
	}
	return nil
}

// NewDockerClient 按接入点构造 Docker 客户端（TCP+TLS 时从凭据保险库取证书）
func (s *EndpointService) NewDockerClient(ep *model.DockerEndpoint) (*client.Client, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}
	if strings.HasPrefix(ep.Addr, "unix://") {
		opts = append(opts, client.WithHost(ep.Addr))
	} else {
		opts = append(opts, client.WithHost(ep.Addr))
		if ep.TLSCredentialID != nil && *ep.TLSCredentialID != 0 {
			caPEM, certPEM, keyPEM, err := s.loadTLSPEMs(*ep.TLSCredentialID)
			if err != nil {
				return nil, err
			}
			// docker SDK v28：直接接收 PEM 三件套
			opts = append(opts, client.WithTLSClientConfig(caPEM, certPEM, keyPEM))
		}
	}
	return client.NewClientWithOpts(opts...)
}

// loadTLSPEMs 从凭据保险库解密 docker_tls 组合（secret=JSON{ca,cert,key}）返回 PEM 三件套
func (s *EndpointService) loadTLSPEMs(credentialID uint) (ca, cert, key string, err error) {
	secret, credType, _, err := assetSvc.Service.CredCredentialService.GetPlaintext(credentialID)
	if err != nil {
		return "", "", "", fmt.Errorf("TLS凭据读取失败(id=%d): %w", credentialID, err)
	}
	if credType != assetModel.CredTypeDockerTLS {
		return "", "", "", newCtErr(ErrCodeEpTLSInvalid, fmt.Sprintf("凭据 id=%d 非 docker_tls 类型", credentialID))
	}
	var bundle struct {
		CA   string `json:"ca"`
		Cert string `json:"cert"`
		Key  string `json:"key"`
	}
	if err := json.Unmarshal([]byte(secret), &bundle); err != nil {
		return "", "", "", newCtErr(ErrCodeEpTLSInvalid, "TLS凭据内容不合法（应为 {ca,cert,key} JSON）")
	}
	if bundle.CA == "" || bundle.Cert == "" || bundle.Key == "" {
		return "", "", "", newCtErr(ErrCodeEpTLSInvalid, "TLS凭据缺少 ca/cert/key 之一")
	}
	return bundle.CA, bundle.Cert, bundle.Key, nil
}

// PingEndpoint 连通巡检：Ping 并回写状态/版本/时间
func (s *EndpointService) PingEndpoint(ep *model.DockerEndpoint) (bool, string, error) {
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		s.markStatus(ep.ID, model.EndpointOffline, "")
		return false, "", err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ping, err := cl.Ping(ctx)
	if err != nil {
		s.markStatus(ep.ID, model.EndpointOffline, "")
		return false, "", err
	}
	version := ping.APIVersion
	s.markStatus(ep.ID, model.EndpointOnline, version)
	return true, version, nil
}

// markStatus 巡检结果回写
func (s *EndpointService) markStatus(id uint, status, version string) {
	updates := map[string]any{"status": status, "last_check_at": time.Now()}
	if version != "" {
		updates["docker_version"] = version
	}
	global.GVA_DB.Model(&model.DockerEndpoint{}).Where("id = ?", id).Updates(updates)
}

// CheckAll 30s 合并巡检入口：并发 5 遍历全部接入点（Ping + 事件区间拉取）
func (s *EndpointService) CheckAll() {
	var list []model.DockerEndpoint
	if err := global.GVA_DB.Find(&list).Error; err != nil {
		return
	}
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for i := range list {
		wg.Add(1)
		go func(ep model.DockerEndpoint) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			_, _, _ = s.PingEndpoint(&ep)
			s.PullEvents(&ep)
		}(list[i])
	}
	wg.Wait()
	s.pruneEventLogs()
}

// StartInspectLoop 启动 30s 合并巡检循环（进程生命周期）
func (s *EndpointService) StartInspectLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.CheckAll()
			}
		}
	}()
}

// CreateEndpoint 创建接入点（校验+重名拦截+连通试测不阻断）
func (s *EndpointService) CreateEndpoint(ep *model.DockerEndpoint) error {
	if err := validateEndpoint(ep); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.DockerEndpoint{}).Where("name = ?", ep.Name).Count(&count)
	if count > 0 {
		return newCtErr(ErrCodeEpDuplicate, fmt.Sprintf("接入点已存在: %s", ep.Name))
	}
	ep.Status = model.EndpointUnknown
	return global.GVA_DB.Create(ep).Error
}

// UpdateEndpoint 更新接入点
func (s *EndpointService) UpdateEndpoint(ep *model.DockerEndpoint) error {
	if ep == nil || ep.ID == 0 {
		return newCtErr(ErrCodeEpNameRequired, "接入点ID不能为空")
	}
	if err := validateEndpoint(ep); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.DockerEndpoint{}).Where("name = ? AND id <> ?", ep.Name, ep.ID).Count(&count)
	if count > 0 {
		return newCtErr(ErrCodeEpDuplicate, fmt.Sprintf("接入点已存在: %s", ep.Name))
	}
	return global.GVA_DB.Model(&model.DockerEndpoint{}).Where("id = ?", ep.ID).Updates(map[string]any{
		"name": ep.Name, "addr": ep.Addr, "tls_credential_id": ep.TLSCredentialID, "notes": ep.Notes,
	}).Error
}

// DeleteEndpoint 删除接入点
func (s *EndpointService) DeleteEndpoint(id uint) error {
	return global.GVA_DB.Delete(&model.DockerEndpoint{}, id).Error
}

// GetEndpointList 接入点列表（keyword/status 过滤）
func (s *EndpointService) GetEndpointList(keyword, status string) (list []*model.DockerEndpoint, err error) {
	db := global.GVA_DB.Model(&model.DockerEndpoint{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR addr LIKE ? OR notes LIKE ?", kw, kw, kw)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// GetEndpoint 详情
func (s *EndpointService) GetEndpoint(id uint) (*model.DockerEndpoint, error) {
	var ep model.DockerEndpoint
	if err := global.GVA_DB.First(&ep, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newCtErr(ErrCodeEpNotFound, "接入点不存在")
		}
		return nil, err
	}
	return &ep, nil
}
