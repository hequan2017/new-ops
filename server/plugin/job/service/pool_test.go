package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestRunPool_ConcurrencyLimit 并发上限：峰值并发不得超过 maxConcurrent
func TestRunPool_ConcurrencyLimit(t *testing.T) {
	var running, peak int32
	tasks := make([]PoolTask, 12)
	for i := range tasks {
		tasks[i] = PoolTask{
			HostID: uint(i + 1),
			Run: func(ctx context.Context) (string, error) {
				cur := atomic.AddInt32(&running, 1)
				for {
					old := atomic.LoadInt32(&peak)
					if cur <= old || atomic.CompareAndSwapInt32(&peak, old, cur) {
						break
					}
				}
				time.Sleep(60 * time.Millisecond)
				atomic.AddInt32(&running, -1)
				return "ok", nil
			},
		}
	}
	results := runPool(context.Background(), tasks, 3, 0)
	if len(results) != 12 {
		t.Fatalf("结果数应与任务数一致: %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("任务 %d 不应失败: %v", r.HostID, r.Err)
		}
	}
	if atomic.LoadInt32(&peak) > 3 {
		t.Fatalf("峰值并发 %d 超过上限 3", peak)
	}
	if atomic.LoadInt32(&peak) < 2 {
		t.Fatalf("峰值并发 %d 异常偏低，未真正并发", peak)
	}
}

// TestRunPool_TaskTimeout 单任务超时被中断并标记 TimedOut
func TestRunPool_TaskTimeout(t *testing.T) {
	tasks := []PoolTask{
		{HostID: 1, Run: func(ctx context.Context) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(2 * time.Second):
				return "slow", nil
			}
		}},
	}
	start := time.Now()
	results := runPool(context.Background(), tasks, 1, 100*time.Millisecond)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("超时未生效，耗时 %s", elapsed)
	}
	if !results[0].TimedOut || !errors.Is(results[0].Err, ErrTaskTimeout) {
		t.Fatalf("应标记超时: %+v", results[0])
	}
}

// TestRunPool_BatchCancel 整体取消：未完成任务返回取消
func TestRunPool_BatchCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tasks := make([]PoolTask, 6)
	for i := range tasks {
		tasks[i] = PoolTask{
			HostID: uint(i + 1),
			Run: func(c context.Context) (string, error) {
				<-c.Done() // 阻塞直到取消
				return "", c.Err()
			},
		}
	}
	go func() {
		time.Sleep(80 * time.Millisecond)
		cancel()
	}()
	results := runPool(ctx, tasks, 2, 0)
	for _, r := range results {
		if !r.Canceled || !errors.Is(r.Err, ErrBatchCanceled) {
			t.Fatalf("任务 %d 应为取消: %+v", r.HostID, r)
		}
	}
}

// TestRunPool_ResultOrder 结果与任务顺序一一对应
func TestRunPool_ResultOrder(t *testing.T) {
	tasks := []PoolTask{
		{HostID: 10, Run: func(context.Context) (string, error) { return "a", nil }},
		{HostID: 20, Run: func(context.Context) (string, error) { return "b", errors.New("boom") }},
	}
	results := runPool(context.Background(), tasks, 2, 0)
	if results[0].HostID != 10 || results[0].Output != "a" || results[0].Err != nil {
		t.Fatalf("结果错位: %+v", results[0])
	}
	if results[1].HostID != 20 || results[1].Err == nil || results[1].Err.Error() != "boom" {
		t.Fatalf("结果错位: %+v", results[1])
	}
}
