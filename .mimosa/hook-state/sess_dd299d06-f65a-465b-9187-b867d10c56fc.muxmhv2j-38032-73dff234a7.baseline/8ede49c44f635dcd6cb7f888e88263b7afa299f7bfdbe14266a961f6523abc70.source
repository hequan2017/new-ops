// Package service 白泽容器管理：容器资源统计（docker stats 单次采样）
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

// decodeJSONBody 流式 JSON 解析（stats one-shot 响应）
func decodeJSONBody(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// StatsView 单容器即时统计
type StatsView struct {
	CPUPercent   float64 `json:"cpuPercent"`
	MemUsedMB    float64 `json:"memUsedMb"`
	MemLimitMB   float64 `json:"memLimitMb"`
	MemPercent   float64 `json:"memPercent"`
	NetRxMB      float64 `json:"netRxMb"`
	NetTxMB      float64 `json:"netTxMb"`
	BlockReadMB  float64 `json:"blockReadMb"`
	BlockWriteMB float64 `json:"blockWriteMb"`
	Pids         int     `json:"pids"`
}

// ContainerStats 容器即时统计（one-shot）
func (s *EndpointService) ContainerStats(endpointID uint, containerID string) (*StatsView, error) {
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	resp, err := cl.ContainerStats(ctx, containerID, false) // one-shot
	if err != nil {
		return nil, fmt.Errorf("stats 拉取失败: %w", err)
	}
	defer resp.Body.Close()

	var st container.StatsResponse
	if err := decodeJSONBody(resp.Body, &st); err != nil {
		return nil, err
	}
	v := &StatsView{Pids: int(st.PidsStats.Current)}

	// CPU：cpu_delta / system_delta * online_cpus
	cpuDelta := float64(st.CPUStats.CPUUsage.TotalUsage) - float64(st.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(st.CPUStats.SystemUsage) - float64(st.PreCPUStats.SystemUsage)
	online := float64(st.CPUStats.OnlineCPUs)
	if online == 0 {
		online = float64(len(st.CPUStats.CPUUsage.PercpuUsage))
	}
	if sysDelta > 0 && cpuDelta > 0 {
		v.CPUPercent = ctRound2((cpuDelta / sysDelta) * online * 100)
	}
	// 内存
	v.MemUsedMB = ctRound2(float64(st.MemoryStats.Usage-st.MemoryStats.Stats["inactive_file"]) / 1024 / 1024)
	v.MemLimitMB = ctRound2(float64(st.MemoryStats.Limit) / 1024 / 1024)
	if v.MemLimitMB > 0 {
		v.MemPercent = ctRound2(float64(st.MemoryStats.Usage) / float64(st.MemoryStats.Limit) * 100)
	}
	// 网络
	for _, n := range st.Networks {
		v.NetRxMB += ctRound2(float64(n.RxBytes) / 1024 / 1024)
		v.NetTxMB += ctRound2(float64(n.TxBytes) / 1024 / 1024)
	}
	// 块 IO
	for _, b := range st.BlkioStats.IoServiceBytesRecursive {
		op := strings.ToLower(b.Op)
		if op == "read" {
			v.BlockReadMB += float64(b.Value) / 1024 / 1024
		} else if op == "write" {
			v.BlockWriteMB += float64(b.Value) / 1024 / 1024
		}
	}
	v.BlockReadMB = ctRound2(v.BlockReadMB)
	v.BlockWriteMB = ctRound2(v.BlockWriteMB)
	return v, nil
}

// 编译期引用（状态常量复用点）
var _ = model.EndpointOnline

// ctRound2 保留两位小数
func ctRound2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
