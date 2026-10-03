// Package service 白泽批量作业：并发执行池
// worker 池 + 信号量限流：单任务超时与整体取消均生效（autoops/ai-devops 验证过的路线）。
package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrTaskTimeout 单任务超时（pool 注入，非执行体返回）
var ErrTaskTimeout = errors.New("任务执行超时")

// ErrBatchCanceled 批次被取消
var ErrBatchCanceled = errors.New("批次已取消")

// PoolTask 单主机任务：Run 由调用方注入（SSH 执行/单测桩）
type PoolTask struct {
	HostID uint
	Run    func(ctx context.Context) (output string, err error)
}

// PoolResult 单主机结果
type PoolResult struct {
	HostID    uint
	Output    string
	Err       error
	Elapsed   time.Duration
	Canceled  bool // 批次取消导致
	TimedOut  bool // 单任务超时
}

// runPool 并发执行全部任务：maxConcurrent 并发上限（<=0 取 1），
// perTimeout 单任务超时（<=0 不限），parent 取消时未完成任务立即返回取消结果。
// 返回结果与 tasks 的 HostID 一一对应（按提交顺序）。
func runPool(parent context.Context, tasks []PoolTask, maxConcurrent int, perTimeout time.Duration) []PoolResult {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	sem := make(chan struct{}, maxConcurrent)
	results := make([]PoolResult, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t PoolTask) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-parent.Done():
				results[idx] = PoolResult{HostID: t.HostID, Err: ErrBatchCanceled, Canceled: true}
				return
			}
			ctx := parent
			cancel := func() {}
			if perTimeout > 0 {
				ctx, cancel = context.WithTimeout(parent, perTimeout)
			}
			defer cancel()
			start := time.Now()
			out, err := t.Run(ctx)
			r := PoolResult{HostID: t.HostID, Output: out, Err: err, Elapsed: time.Since(start)}
			switch {
			case errors.Is(ctx.Err(), context.DeadlineExceeded) &&
				(err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)):
				// 执行体透传了 ctx 超时（或未感知）→ 统一记为任务超时
				r.Err, r.TimedOut = ErrTaskTimeout, true
			case errors.Is(ctx.Err(), context.Canceled):
				r.Err, r.Canceled = ErrBatchCanceled, true
			}
			results[idx] = r
		}(i, task)
	}
	wg.Wait()
	return results
}
