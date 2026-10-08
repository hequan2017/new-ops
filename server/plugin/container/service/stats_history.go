// Package service 白泽容器管理：容器资源统计历史（M8 C2 接入图表化）
// 5 分钟一轮采样全部在线接入点的运行中容器，7 天留存；查询供前端 ECharts 时间轴。
package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

const (
	statsSampleInterval = 5 * time.Minute
	statsRetentionDays  = 7
	statsHistoryCap     = 2000
)

// StartStatsLoop 启动统计采样循环（进程生命周期）
func (s *EndpointService) StartStatsLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(statsSampleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.CollectStatsAll()
			}
		}
	}()
}

// CollectStatsAll 一轮全量采样：在线接入点 → 运行中容器（并发 5）
func (s *EndpointService) CollectStatsAll() {
	var eps []model.DockerEndpoint
	if err := global.GVA_DB.Find(&eps).Error; err != nil {
		return
	}
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for i := range eps {
		if eps[i].Status == model.EndpointOffline {
			continue
		}
		wg.Add(1)
		go func(ep model.DockerEndpoint) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			s.collectEndpointStats(&ep)
		}(eps[i])
	}
	wg.Wait()
	cutoff := time.Now().AddDate(0, 0, -statsRetentionDays)
	global.GVA_DB.Where("created_at < ?", cutoff).Delete(&model.DockerStatsSample{})
}

// collectEndpointStats 单接入点采样
func (s *EndpointService) collectEndpointStats(ep *model.DockerEndpoint) {
	views, err := s.ListContainers(ep.ID, false)
	if err != nil {
		return
	}
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for _, v := range views {
		if v.State != "running" {
			continue
		}
		wg.Add(1)
		go func(cv ContainerView) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			st, err := s.ContainerStats(ep.ID, cv.ID)
			if err != nil || st == nil {
				return
			}
			global.GVA_DB.Create(&model.DockerStatsSample{
				EndpointID:  ep.ID,
				ContainerID: cv.ID,
				Name:        containerDisplayName(cv.Names),
				CPUPercent:  st.CPUPercent,
				MemUsedMB:   st.MemUsedMB,
				MemLimitMB:  st.MemLimitMB,
				MemPercent:  st.MemPercent,
				NetRxMB:     st.NetRxMB,
				NetTxMB:     st.NetTxMB,
				Pids:        st.Pids,
			})
		}(v)
	}
	wg.Wait()
}

// containerDisplayName 容器名（去首斜杠）
func containerDisplayName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

// GetStatsHistory 容器统计历史（hours 窗口，按时间倒序截 cap 后反转为正序供绘图）
func (s *EndpointService) GetStatsHistory(endpointID uint, containerID string, hours int) ([]model.DockerStatsSample, error) {
	if hours <= 0 || hours > 24*7 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var list []model.DockerStatsSample
	err := global.GVA_DB.
		Where("endpoint_id = ? AND container_id = ? AND created_at > ?", endpointID, containerID, since).
		Order("created_at DESC").Limit(statsHistoryCap).Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}
