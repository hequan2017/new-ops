package service

import (
	"strings"
	"testing"
)

func TestStepBackend(t *testing.T) {
	cases := []struct {
		in   string
		want string
		fail bool
	}{
		{"", StepBackendSSH, false},
		{"ssh", StepBackendSSH, false},
		{"agent", StepBackendAgent, false},
		{"exec", "", true},
		{"SSH", "", true},
	}
	for _, c := range cases {
		got, err := stepBackend(c.in)
		if c.fail {
			if err == nil {
				t.Fatalf("stepBackend(%q) 应报错", c.in)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Fatalf("stepBackend(%q)=%q,%v want %q", c.in, got, err, c.want)
		}
	}
}

func TestPipelineShellCommand(t *testing.T) {
	content := "echo hi\necho bye\ndate"
	cmd := pipelineShellCommand(workspaceFor(7), content)
	for _, want := range []string{
		`mkdir -p "/tmp/baize-pipeline/build-7"`,
		`cd "/tmp/baize-pipeline/build-7"`,
		"&& (",
		content,
		"\n)",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("包装命令缺少 %q:\n%s", want, cmd)
		}
	}
	if workspaceFor(8) == workspaceFor(7) {
		t.Fatal("不同构建的工作空间应隔离")
	}
	// 内容原样保留（无额外转义/拼接污染）
	if !strings.HasPrefix(strings.SplitN(cmd, "&& (\n", 2)[1], content) {
		t.Fatalf("内容应原样进入子Shell:\n%s", cmd)
	}
}
