// Package service 白泽容器管理：端口转发规则管理（M8 C2，docker-gpu-manage 模式）
// Docker 端口绑定不可在线修改：应用新规则 = 按当前 inspect 配置 + 新端口规则重建容器
// （保留镜像/命令/环境/挂载/标签/重启策略/资源限制/网络模式/特权位；运行中容器重建后自动拉起）。
package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// ErrCodePortRuleInvalid 端口规则非法（1411）
const ErrCodePortRuleInvalid = 1411

// PortRule 端口转发规则
type PortRule struct {
	HostIP        string `json:"hostIp"`        // 空=0.0.0.0
	HostPort      string `json:"hostPort"`      // "8080"
	ContainerPort int    `json:"containerPort"` // 80
	Proto         string `json:"proto"`         // tcp|udp
}

// validatePortRules 规则校验（纯逻辑，可单测）：端口范围、协议、hostIP、重复检测
func validatePortRules(rules []PortRule) error {
	if len(rules) == 0 {
		return newCtErr(ErrCodePortRuleInvalid, "端口规则不能为空（如需清除全部映射请直接删除容器重建）")
	}
	seen := map[string]bool{}
	for i, r := range rules {
		proto := strings.ToLower(strings.TrimSpace(r.Proto))
		if proto == "" {
			proto = "tcp"
		}
		if proto != "tcp" && proto != "udp" {
			return newCtErr(ErrCodePortRuleInvalid, fmt.Sprintf("规则#%d 协议仅支持 tcp/udp: %s", i+1, r.Proto))
		}
		if r.ContainerPort < 1 || r.ContainerPort > 65535 {
			return newCtErr(ErrCodePortRuleInvalid, fmt.Sprintf("规则#%d 容器端口非法（1-65535）: %d", i+1, r.ContainerPort))
		}
		host := strings.TrimSpace(r.HostPort)
		hp, err := strconv.Atoi(host)
		if err != nil || hp < 1 || hp > 65535 {
			return newCtErr(ErrCodePortRuleInvalid, fmt.Sprintf("规则#%d 宿主端口非法（1-65535）: %s", i+1, r.HostPort))
		}
		if ip := strings.TrimSpace(r.HostIP); ip != "" && net.ParseIP(ip) == nil {
			return newCtErr(ErrCodePortRuleInvalid, fmt.Sprintf("规则#%d hostIP 非法: %s", i+1, r.HostIP))
		}
		key := strings.TrimSpace(r.HostIP) + "|" + host + "|" + proto
		if seen[key] {
			return newCtErr(ErrCodePortRuleInvalid, fmt.Sprintf("规则#%d 宿主端口重复: %s/%s", i+1, r.HostPort, proto))
		}
		seen[key] = true
	}
	return nil
}

// portRulesToBindings 规则 → nat ExposedPorts/PortMap（纯逻辑，可单测）
func portRulesToBindings(rules []PortRule) (nat.PortSet, nat.PortMap) {
	exposed := nat.PortSet{}
	bindings := nat.PortMap{}
	for _, r := range rules {
		proto := strings.ToLower(strings.TrimSpace(r.Proto))
		if proto == "" {
			proto = "tcp"
		}
		p := nat.Port(strconv.Itoa(r.ContainerPort) + "/" + proto)
		exposed[p] = struct{}{}
		bindings[p] = append(bindings[p], nat.PortBinding{HostIP: strings.TrimSpace(r.HostIP), HostPort: strings.TrimSpace(r.HostPort)})
	}
	return exposed, bindings
}

// bindingsToPortRules inspect 的 PortBindings → 规则列表
func bindingsToPortRules(pm nat.PortMap) []PortRule {
	out := []PortRule{}
	for p, binds := range pm {
		cp, _ := strconv.Atoi(strings.Split(string(p), "/")[0])
		proto := "tcp"
		if parts := strings.Split(string(p), "/"); len(parts) == 2 {
			proto = parts[1]
		}
		if len(binds) == 0 {
			out = append(out, PortRule{ContainerPort: cp, Proto: proto})
			continue
		}
		for _, b := range binds {
			out = append(out, PortRule{HostIP: b.HostIP, HostPort: b.HostPort, ContainerPort: cp, Proto: proto})
		}
	}
	return out
}

// ListPortForwards 读取容器当前端口映射
func (s *EndpointService) ListPortForwards(endpointID uint, containerID string) ([]PortRule, error) {
	insp, err := s.InspectContainer(endpointID, containerID)
	if err != nil {
		return nil, err
	}
	if insp.HostConfig == nil {
		return nil, newCtErr(ErrCodeEpNotFound, "容器配置缺失")
	}
	return bindingsToPortRules(insp.HostConfig.PortBindings), nil
}

// SetPortForwards 应用端口规则（重建容器）；返回新容器 ID
func (s *EndpointService) SetPortForwards(endpointID uint, containerID string, rules []PortRule) (string, error) {
	if err := validatePortRules(rules); err != nil {
		return "", err
	}
	if err := s.ensureEndpointAlive(endpointID); err != nil {
		return "", err
	}
	insp, err := s.InspectContainer(endpointID, containerID)
	if err != nil {
		return "", err
	}
	if insp.Config == nil || insp.HostConfig == nil {
		return "", newCtErr(ErrCodeEpNotFound, "容器配置缺失")
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	name := strings.TrimPrefix(insp.Name, "/")
	wasRunning := insp.State != nil && insp.State.Running
	exposed, bindings := portRulesToBindings(rules)

	cfg := *insp.Config
	cfg.ExposedPorts = exposed
	host := *insp.HostConfig
	host.PortBindings = bindings
	// 重建不继承：自动移除的容器链接与旧网络别名
	host.Links = nil

	if err := cl.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return "", fmt.Errorf("旧容器移除失败: %w", err)
	}
	created, err := cl.ContainerCreate(ctx, &cfg, &host, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("新容器创建失败: %w", err)
	}
	if wasRunning {
		if err := cl.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
			return created.ID, fmt.Errorf("容器已按新规则重建但启动失败: %w", err)
		}
	}
	return created.ID, nil
}
