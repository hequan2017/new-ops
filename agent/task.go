// 白泽轻量 Agent：任务执行端（M9 A2）
// 收到 server 下发的 task 帧 → 任务号去重（幂等）→ 限时执行 /bin/sh -c → 截断输出回传 task_result。
package main

import (
	"context"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"unicode/utf8"
)

// maxTaskOutputBytes 输出回传上限（与 server 留存口径一致）
const maxTaskOutputBytes = 60 * 1024

// taskErrTimeout 自身执行超时时回传的错误文案（server 据此映射"超时"状态）
const taskErrTimeout = "执行超时"

// taskSeen 任务号去重表（幂等）：同号只放行一次；超限整体清空防长驻泄漏
type taskSeen struct {
	mu      sync.Mutex
	seen    map[string]struct{}
	maxSeen int
}

func newTaskSeen(maxSeen int) *taskSeen {
	return &taskSeen{seen: map[string]struct{}{}, maxSeen: maxSeen}
}

func (t *taskSeen) first(taskNo string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.seen[taskNo]; ok {
		return false
	}
	if len(t.seen) >= t.maxSeen {
		t.seen = map[string]struct{}{}
	}
	t.seen[taskNo] = struct{}{}
	return true
}

// taskExecutor 任务执行器：任务号去重 + 并发上限
type taskExecutor struct {
	seen *taskSeen
	sem  chan struct{}
}

func newTaskExecutor(maxConcurrent, maxSeen int) *taskExecutor {
	return &taskExecutor{seen: newTaskSeen(maxSeen), sem: make(chan struct{}, maxConcurrent)}
}

// shouldRun 任务号去重（幂等）：同号只放行一次
func (t *taskExecutor) shouldRun(taskNo string) bool {
	return t.seen.first(taskNo)
}

// dispatch 处理下发的 task 帧：重复任务号静默丢弃（幂等）；并发占满即时回执失败；否则后台限时执行
func (t *taskExecutor) dispatch(f frame, send func(frame) error) {
	if f.TaskNo == "" || f.Payload == "" {
		return
	}
	if !t.shouldRun(f.TaskNo) {
		return
	}
	select {
	case t.sem <- struct{}{}:
	default:
		_ = send(frame{Type: "task_result", TS: time.Now().Unix(), TaskNo: f.TaskNo, ExitCode: -1, Msg: "agent 并发占满，任务被拒绝"})
		return
	}
	go func() {
		defer func() { <-t.sem }()
		code, out, errMsg := runTask(time.Duration(f.TaskTimeout)*time.Second, f.Payload)
		_ = send(frame{Type: "task_result", TS: time.Now().Unix(), TaskNo: f.TaskNo, ExitCode: code, Output: out, Msg: errMsg})
	}()
}

// runTask 限时执行单条命令（/bin/sh -c；windows 不支持）
func runTask(timeout time.Duration, command string) (exitCode int, output string, errMsg string) {
	if runtime.GOOS == "windows" {
		return -1, "", "windows 暂不支持任务执行"
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()
	capped := truncateBytes(string(out), maxTaskOutputBytes)
	if ctx.Err() == context.DeadlineExceeded {
		return -1, capped, taskErrTimeout
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode(), capped, ""
		}
		return -1, capped, err.Error()
	}
	return 0, capped, ""
}

// truncateBytes 按 UTF-8 字符边界截断到 max 字节内（纯逻辑）
func truncateBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	i := max
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i]
}
