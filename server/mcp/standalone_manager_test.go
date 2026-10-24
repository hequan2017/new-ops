// Package mcpTool 托管进程归属校验单测
package mcpTool

import (
	"testing"
)

func TestManagedProcessAliveWith(t *testing.T) {
	meta := &managedProcessMeta{PID: 4242, Command: "/srv/new-ops/.tmp/mcp/ops-mcp-standalone"}
	alive := func(pid int) bool { return pid == 4242 }

	cases := []struct {
		name    string
		meta    *managedProcessMeta
		alive   func(int) bool
		cmdline func(int) (string, bool)
		want    bool
	}{
		{
			name:    "PID 存活且命令行匹配",
			meta:    meta,
			alive:   alive,
			cmdline: func(int) (string, bool) { return "/srv/new-ops/.tmp/mcp/ops-mcp-standalone -config x.yaml", true },
			want:    true,
		},
		{
			name:  "PID 被复用为无关进程",
			meta:  meta,
			alive: alive,
			cmdline: func(int) (string, bool) {
				return "/Applications/Editor.app/Contents/MacOS/Editor", true
			},
			want: false,
		},
		{
			name:    "命令行不可得退回存活判断",
			meta:    meta,
			alive:   alive,
			cmdline: func(int) (string, bool) { return "", false },
			want:    true,
		},
		{
			name:    "PID 已退出",
			meta:    meta,
			alive:   func(int) bool { return false },
			cmdline: func(int) (string, bool) { return "whatever", true },
			want:    false,
		},
		{
			name:    "meta 为空或 PID 非法",
			meta:    &managedProcessMeta{PID: 0, Command: "/x/ops-mcp-standalone"},
			alive:   func(int) bool { return true },
			cmdline: func(int) (string, bool) { return "/x/ops-mcp-standalone", true },
			want:    false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := managedProcessAliveWith(c.meta, c.alive, c.cmdline); got != c.want {
				t.Fatalf("managedProcessAliveWith = %v, want %v", got, c.want)
			}
		})
	}
}
