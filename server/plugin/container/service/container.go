// Package service 白泽容器管理：容器实时查询与生命周期（M4 C1）
// 容器数据实时查询 Docker API 不落库；写操作走 GVA 操作日志中间件。
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

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

// shortenID 容器 ID 截断 12 位展示
func shortenID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
