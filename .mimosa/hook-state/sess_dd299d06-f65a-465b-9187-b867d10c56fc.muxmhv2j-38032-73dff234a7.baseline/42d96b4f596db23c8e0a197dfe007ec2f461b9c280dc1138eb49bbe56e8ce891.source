// Package service 白泽容器管理：网络与卷（C2，实时查询不落库）
package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"

	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

// NetworkView 网络视图
type NetworkView struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Driver  string   `json:"driver"`
	Subnet  string   `json:"subnet"`
	BuiltIn bool     `json:"builtIn"` // bridge/host/none 等内置网络不可删
}

// VolumeView 卷视图
type VolumeView struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
}

// ListNetworks 网络列表
func (s *EndpointService) ListNetworks(endpointID uint) ([]NetworkView, error) {
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
	list, err := cl.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("网络列表拉取失败: %w", err)
	}
	builtIn := map[string]bool{"bridge": true, "host": true, "none": true}
	out := make([]NetworkView, 0, len(list))
	for _, n := range list {
		subnet := ""
		if n.IPAM.Config != nil && len(n.IPAM.Config) > 0 {
			subnet = n.IPAM.Config[0].Subnet
		}
		out = append(out, NetworkView{
			ID:      shortID(n.ID),
			Name:    n.Name,
			Driver:  n.Driver,
			Subnet:  subnet,
			BuiltIn: builtIn[n.Name],
		})
	}
	return out, nil
}

// CreateNetworkReq 创建网络参数
type CreateNetworkReq struct {
	EndpointID uint   `json:"endpointId" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Driver     string `json:"driver"` // 默认 bridge
	Subnet     string `json:"subnet"` // 如 172.30.0.0/16（选配）
}

// CreateNetwork 创建自定义网络
func (s *EndpointService) CreateNetwork(req CreateNetworkReq) error {
	if err := s.ensureEndpointAlive(req.EndpointID); err != nil {
		return err
	}
	if strings.TrimSpace(req.Name) == "" {
		return newCtErr(ErrCodeEpAddrInvalid, "网络名称不能为空")
	}
	driver := req.Driver
	if driver == "" {
		driver = "bridge"
	}
	ep, err := s.GetEndpoint(req.EndpointID)
	if err != nil {
		return err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	opts := network.CreateOptions{
		Driver: driver,
	}
	if req.Subnet != "" {
		if _, ipNet, err := net.ParseCIDR(req.Subnet); err != nil || ipNet == nil {
			return newCtErr(ErrCodeEpAddrInvalid, "子网格式非法（应如 172.30.0.0/16）")
		}
		opts.IPAM = &network.IPAM{Config: []network.IPAMConfig{{
			Subnet: req.Subnet,
		}}}
	}
	if _, err := cl.NetworkCreate(ctx, req.Name, opts); err != nil {
		return fmt.Errorf("创建网络失败: %w", err)
	}
	return nil
}

// RemoveNetwork 删除自定义网络（内置网络拒绝）
func (s *EndpointService) RemoveNetwork(endpointID uint, name string) error {
	if name == "bridge" || name == "host" || name == "none" {
		return newCtErr(ErrCodeEpAddrInvalid, "内置网络不可删除")
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return cl.NetworkRemove(ctx, name)
}

// ListVolumes 卷列表
func (s *EndpointService) ListVolumes(endpointID uint) ([]VolumeView, error) {
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
	res, err := cl.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("卷列表拉取失败: %w", err)
	}
	out := make([]VolumeView, 0, len(res.Volumes))
	for _, v := range res.Volumes {
		out = append(out, VolumeView{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
		})
	}
	return out, nil
}

// RemoveVolume 删除卷（in-use 由 daemon 拒绝并回传错误）
func (s *EndpointService) RemoveVolume(endpointID uint, name string) error {
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return cl.VolumeRemove(ctx, name, true)
}

// 编译期引用（状态常量复用点）
var _ = model.EndpointOnline
