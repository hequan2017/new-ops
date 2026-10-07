// Package service 白泽容器管理：docker events 区间拉取（M4 C1 收官）
// 设计取舍：不做常驻订阅 goroutine（接入点增删改的订阅生命周期管理与重连成本高），
// 复用 30s 巡检循环按 [LastEventAt, now] 区间拉取事件落库——离线接入点恢复后自动补拉。
package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

// pullEventsSince 事件拉取窗口上限（防首次接入/长期离线补拉过大）
const eventPullWindow = 1 * time.Hour

// eventRetention 事件留存时长（过期由巡检顺带清理）
const eventRetention = 7 * 24 * time.Hour

// eventPullTimeout 单接入点单轮事件拉取限时
const eventPullTimeout = 8 * time.Second

// PullEvents 拉取接入点 [LastEventAt, now] 的容器事件落库，并推进水位。
// v28 SDK 语义：daemon 发完 until 窗口数据后关流，SDK 把 io.EOF 发到 errCh 且不关 msgCh
// ——EOF 视为正常收尾（落库+推水位）；ctx 超时视为异常（已收事件仍落库但不推水位，下轮重拉容忍重复）。
func (s *EndpointService) PullEvents(ep *model.DockerEndpoint) {
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		return
	}
	defer cl.Close()

	since := time.Now().Add(-eventPullWindow)
	if ep.LastEventAt != nil && ep.LastEventAt.After(since) {
		since = *ep.LastEventAt
	}
	until := time.Now()
	opts := container_events_options(since, until)

	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), eventPullTimeout)
		defer cancel()
		msgCh, errCh := cl.Events(ctx, opts)
		batch := make([]model.DockerEventLog, 0, 32)
		lastAt := since
		normalEnd := false
		for !normalEnd {
			select {
			case ev, ok := <-msgCh:
				if !ok {
					normalEnd = true
					break
				}
				if ev.Type != events.ContainerEventType {
					continue
				}
				occurred := time.Unix(ev.Time, 0)
				batch = append(batch, model.DockerEventLog{
					EndpointID:  ep.ID,
					Action:      string(ev.Action),
					ContainerID: shortID(ev.Actor.ID),
					ContainerNm: ev.Actor.Attributes["name"],
					OccurredAt:  occurred,
				})
				if occurred.After(lastAt) {
					lastAt = occurred
				}
				if len(batch) >= 500 {
					storeEventBatch(ep.ID, batch)
					batch = batch[:0]
				}
			case e := <-errCh:
				if errors.Is(e, io.EOF) {
					normalEnd = true // daemon 关流的正常收尾
					break
				}
				global.GVA_LOG.Error(fmt.Sprintf("events 拉取错误（ep=%d）: %v", ep.ID, e))
				normalEnd = true
			case <-ctx.Done():
				global.GVA_LOG.Warn(fmt.Sprintf("events 拉取超时（ep=%d since=%d）", ep.ID, since.Unix()))
			}
			if ctx.Err() != nil && !normalEnd {
				break
			}
		}
		storeEventBatch(ep.ID, batch)
		if normalEnd {
			// 正常收尾推水位到 until（超时不推，下轮区间重拉容忍少量重复）
			global.GVA_DB.Model(&model.DockerEndpoint{}).Where("id = ?", ep.ID).
				Update("last_event_at", until)
		}
	}()
	<-done
}

// container_events_options 构造 Events 过滤（container 类型 + Since/Until）
// Since/Until 用 unix 秒（daemon 对 RFC3339 纳秒格式回放解析不稳，官方 CLI 同用 unix 秒）
func container_events_options(since, until time.Time) events.ListOptions {
	return events.ListOptions{
		Filters: filters.NewArgs(filters.Arg("type", "container")),
		Since:   fmt.Sprintf("%d", since.Unix()),
		Until:   fmt.Sprintf("%d", until.Unix()),
	}
}

// storeEventBatch 批量落库
func storeEventBatch(endpointID uint, batch []model.DockerEventLog) {
	if len(batch) == 0 {
		return
	}
	if err := global.GVA_DB.Create(&batch).Error; err != nil {
		global.GVA_LOG.Error("容器事件批量落库失败: " + err.Error())
	}
}

// pruneEventLogs 清理过期事件（保留 7 天）
func (s *EndpointService) pruneEventLogs() {
	global.GVA_DB.Where("occurred_at < ?", time.Now().Add(-eventRetention)).
		Delete(&model.DockerEventLog{})
}

// shortID 容器 ID 截断 12 位展示
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return strings.TrimSpace(id)
}

// GetEventList 接入点事件列表（倒序，默认 200 条）
func (s *EndpointService) GetEventList(endpointID uint, action string) (list []model.DockerEventLog, err error) {
	db := global.GVA_DB.Model(&model.DockerEventLog{})
	if endpointID > 0 {
		db = db.Where("endpoint_id = ?", endpointID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	err = db.Order("occurred_at DESC").Limit(200).Find(&list).Error
	return list, err
}
