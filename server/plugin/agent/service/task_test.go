package service

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestResolveTaskResult(t *testing.T) {
	taskNo := "t-resolve-1"
	ch := make(chan agentFrame, 1)
	pendingTasks.Store(taskNo, ch)
	defer pendingTasks.Delete(taskNo)

	if !resolveTaskResult(taskNo, agentFrame{Type: "task_result", TaskNo: taskNo, ExitCode: 0, Output: "ok"}) {
		t.Fatal("等待中的任务号应被解析")
	}
	select {
	case f := <-ch:
		if f.Output != "ok" {
			t.Fatalf("回传输出不符: %q", f.Output)
		}
	case <-time.After(time.Second):
		t.Fatal("结果未送达等待方")
	}
	if resolveTaskResult(taskNo, agentFrame{TaskNo: taskNo}) {
		t.Fatal("已取走的任务号不应二次解析")
	}
	if resolveTaskResult("", agentFrame{TaskNo: ""}) {
		t.Fatal("空任务号应直接拒绝")
	}
	if resolveTaskResult("no-such-task", agentFrame{TaskNo: "no-such-task"}) {
		t.Fatal("无等待方的任务号应丢弃")
	}
}

func TestTaskStatusOf(t *testing.T) {
	cases := []struct {
		code   int
		errMsg string
		want   string
	}{
		{0, "", TaskStatusSuccess},
		{3, "", TaskStatusFailed},
		{-1, "spawn 失败", TaskStatusFailed},
		{-1, "执行超时", TaskStatusTimeout},
	}
	for _, c := range cases {
		if got := taskStatusOf(c.code, c.errMsg); got != c.want {
			t.Fatalf("taskStatusOf(%d,%q)=%q want %q", c.code, c.errMsg, got, c.want)
		}
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
	if got := truncateBytes("你好", 4); got != "你" {
		t.Fatalf("跨字符截断应回退到边界: %q", got)
	}
}
