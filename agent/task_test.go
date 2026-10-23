package main

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestShouldRunDedup(t *testing.T) {
	ex := newTaskExecutor(2, 4)
	if !ex.shouldRun("a") {
		t.Fatal("新任务号应放行")
	}
	if ex.shouldRun("a") {
		t.Fatal("同任务号不应二次放行（幂等）")
	}
	if !ex.shouldRun("b") {
		t.Fatal("不同任务号应放行")
	}
	for _, id := range []string{"c", "d"} {
		ex.shouldRun(id)
	}
	if !ex.shouldRun("e") {
		t.Fatal("seen 超限清空后新任务应放行")
	}
}

func TestDispatchDedupDropsDuplicate(t *testing.T) {
	ex := newTaskExecutor(2, 64)
	sent := make(chan frame, 4)
	send := func(f frame) error { sent <- f; return nil }
	ex.dispatch(frame{Type: "task", TaskNo: "t1", Payload: "echo hi", TaskTimeout: 5}, send)
	ex.dispatch(frame{Type: "task", TaskNo: "t1", Payload: "echo hi", TaskTimeout: 5}, send) // 重复号：静默丢弃
	select {
	case f := <-sent:
		if f.Type != "task_result" || f.TaskNo != "t1" {
			t.Fatalf("应只回传一次结果: %+v", f)
		}
		if !strings.Contains(f.Output, "hi") {
			t.Fatalf("结果输出缺失: %q", f.Output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("任务结果未回传")
	}
	select {
	case f := <-sent:
		t.Fatalf("重复任务不应产生第二笔回传: %+v", f)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestDispatchBusyRejects(t *testing.T) {
	ex := newTaskExecutor(1, 64)
	sent := make(chan frame, 4)
	send := func(f frame) error { sent <- f; return nil }
	ex.dispatch(frame{Type: "task", TaskNo: "slow", Payload: "sleep 1", TaskTimeout: 5}, send) // 占满并发
	ex.dispatch(frame{Type: "task", TaskNo: "busy", Payload: "echo x", TaskTimeout: 5}, send)  // 应被即时拒绝
	select {
	case f := <-sent:
		if f.TaskNo != "busy" || !strings.Contains(f.Msg, "占满") {
			t.Fatalf("应回传并发占满拒绝: %+v", f)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("占满拒绝未回传")
	}
}

func TestRunTaskEcho(t *testing.T) {
	code, out, errMsg := runTask(5*time.Second, "echo baize-a2")
	if code != 0 || errMsg != "" {
		t.Fatalf("echo 应成功: code=%d errMsg=%q", code, errMsg)
	}
	if !strings.Contains(out, "baize-a2") {
		t.Fatalf("输出缺失: %q", out)
	}
}

func TestRunTaskTimeout(t *testing.T) {
	code, _, errMsg := runTask(200*time.Millisecond, "sleep 2")
	if errMsg != taskErrTimeout {
		t.Fatalf("应报执行超时: %q", errMsg)
	}
	if code != -1 {
		t.Fatalf("超时退出码应为 -1: %d", code)
	}
}

func TestRunTaskExitCode(t *testing.T) {
	code, _, errMsg := runTask(5*time.Second, "exit 7")
	if code != 7 || errMsg != "" {
		t.Fatalf("退出码应透传: code=%d errMsg=%q", code, errMsg)
	}
}

func TestTruncateBytes(t *testing.T) {
	if got := truncateBytes("hello", 10); got != "hello" {
		t.Fatalf("未超限不应截断: %q", got)
	}
	long := strings.Repeat("好", 40000) // 120KB
	got := truncateBytes(long, 60000)
	if len(got) > 60000 {
		t.Fatalf("截断后仍超限: %d", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatal("截断应保持 UTF-8 合法（字符边界）")
	}
}
