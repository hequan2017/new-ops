// Package service 白泽容器管理：容器实时查询与生命周期（M4 C1）
// 容器数据实时查询 Docker API 不落库；写操作走 GVA 操作日志中间件。
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

// ContainerView 容器列表视图（精简自 Docker API）
type ContainerView struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created int64    `json:"created"`
	Ports   string   `json:"ports"`
}

// ListContainers 拉取接入点容器列表（all=true 含已停止）
func (s *EndpointService) ListContainers(endpointID uint, all bool) ([]ContainerView, error) {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	list, err := cl.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("容器列表拉取失败: %w", err)
	}
	out := make([]ContainerView, 0, len(list))
	for _, c := range list {
		ports := make([]string, 0, len(c.Ports))
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}
		out = append(out, ContainerView{
			ID:      shortenID(c.ID),
			Names:   c.Names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Ports:   strings.Join(ports, ", "),
		})
	}
	return out, nil
}

// ContainerAction 容器生命周期动作（start/stop/restart/remove；离线接入点直接拒绝）
func (s *EndpointService) ContainerAction(endpointID uint, containerID, action string, force bool) error {
	if err := s.ensureEndpointAlive(endpointID); err != nil {
		return err
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	switch action {
	case "start":
		return cl.ContainerStart(ctx, containerID, container.StartOptions{})
	case "stop":
		return cl.ContainerStop(ctx, containerID, container.StopOptions{})
	case "restart":
		return cl.ContainerRestart(ctx, containerID, container.StopOptions{})
	case "remove":
		return cl.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
	default:
		return newCtErr(ErrCodeEpAddrInvalid, fmt.Sprintf("未知容器动作: %s", action))
	}
}

// InspectContainer 容器详情回写场景预留（C1 事件订阅场次使用）
func (s *EndpointService) InspectContainer(endpointID uint, containerID string) (*container.InspectResponse, error) {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	insp, err := cl.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}
	return &insp, nil
}

// ensureEndpointAlive 巡检兜底：操作前列状态（离线直接报错）
func (s *EndpointService) ensureEndpointAlive(endpointID uint) error {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return err
	}
	if ep.Status == model.EndpointOffline {
		return newCtErr(ErrCodeEpConnectFailed, fmt.Sprintf("接入点 %s 当前离线，请先巡检", ep.Name))
	}
	return nil
}

// CreateContainerReq 创建容器参数
type CreateContainerReq struct {
	EndpointID    uint     `json:"endpointId" binding:"required"`
	Name          string   `json:"name"`          // 容器名（空则 Docker 自动分配）
	Image         string   `json:"image"`         // 镜像（必填）
	Command       []string `json:"command"`       // 覆盖入口命令
	Ports         []string `json:"ports"`         // "8080:80/tcp" 形式
	Envs          []string `json:"envs"`          // "K=V" 形式
	Mounts        []string `json:"mounts"`        // "/host:/ct:rw" 形式（Binds）
	CPUCores      float64  `json:"cpuCores"`      // CPU 核数上限（0 不限）
	MemoryMB      int64    `json:"memoryMb"`      // 内存上限 MB（0 不限）
	RestartPolicy string   `json:"restartPolicy"` // no/always/unless-stopped/on-failure
	StartNow      bool     `json:"startNow"`
}

// ParsePortBinding 解析 "8080:80/tcp" → (hostPort, containerPort+proto)（纯逻辑，可单测）
func ParsePortBinding(s string) (hostPort, portProto string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", fmt.Errorf("端口映射为空")
	}
	proto := "tcp"
	pair := s
	if i := strings.LastIndex(s, "/"); i >= 0 {
		proto = s[i+1:]
		pair = s[:i]
	}
	parts := strings.SplitN(pair, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("端口映射格式应为 host:ct[/proto]: %s", s)
	}
	return parts[0], parts[1] + "/" + proto, nil
}

// CreateContainer 创建容器（端口/环境/挂载/资源限制/重启策略）
func (s *EndpointService) CreateContainer(endpointID uint, req CreateContainerReq) (string, error) {
	if err := s.ensureEndpointAlive(endpointID); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.Image) == "" {
		return "", newCtErr(ErrCodeEpAddrInvalid, "镜像不能为空")
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return "", err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return "", err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := &container.Config{
		Image: req.Image,
		Env:   req.Envs,
		Cmd:   req.Command,
	}
	hostCfg := &container.HostConfig{
		Binds: req.Mounts,
	}
	// 端口映射：ParsePortBinding 产出 "80/tcp" 形式，nat.Port 直接作键
	exposed := nat.PortSet{}
	bindings := nat.PortMap{}
	for _, p := range req.Ports {
		host, portProto, perr := ParsePortBinding(p)
		if perr != nil {
			return "", newCtErr(ErrCodeEpAddrInvalid, perr.Error())
		}
		natPort := nat.Port(portProto)
		exposed[natPort] = struct{}{}
		bindings[natPort] = append(bindings[natPort], nat.PortBinding{HostPort: host})
	}
	if len(exposed) > 0 {
		config.ExposedPorts = exposed
		hostCfg.PortBindings = bindings
	}
	// 资源限制
	if req.CPUCores > 0 {
		hostCfg.NanoCPUs = int64(req.CPUCores * 1e9)
	}
	if req.MemoryMB > 0 {
		hostCfg.Memory = req.MemoryMB * 1024 * 1024
	}
	// 重启策略
	switch req.RestartPolicy {
	case "", "no":
		hostCfg.RestartPolicy = container.RestartPolicy{Name: "no"}
	case "always", "unless-stopped", "on-failure":
		hostCfg.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(req.RestartPolicy)}
	default:
		return "", newCtErr(ErrCodeEpAddrInvalid, fmt.Sprintf("未知重启策略: %s", req.RestartPolicy))
	}
	created, err := cl.ContainerCreate(ctx, config, hostCfg, nil, nil, req.Name)
	if err != nil {
		return "", fmt.Errorf("创建容器失败: %w", err)
	}
	if req.StartNow {
		if err := cl.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
			return created.ID, fmt.Errorf("容器已创建但启动失败: %w", err)
		}
	}
	return created.ID, nil
}

// shortenID 容器 ID 截断 12 位展示
func shortenID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
