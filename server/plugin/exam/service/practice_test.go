// Package service exam 实操通道单测（白名单/输出截断）
package service

import "testing"

func TestValidPracticeChannel(t *testing.T) {
	if !ValidPracticeChannel("operator") || !ValidPracticeChannel("agent") {
		t.Fatal("白名单应放行 operator/agent")
	}
	for _, bad := range []string{"", "shell", "Operator", "script"} {
		if ValidPracticeChannel(bad) {
			t.Fatalf("非法通道 %q 应拒绝", bad)
		}
	}
}

func TestTruncatePracticeOutput(t *testing.T) {
	if got := truncatePracticeOutput("  短输出  "); got != "短输出" {
		t.Fatalf("短输出应 trim: %q", got)
	}
	long := ""
	for i := 0; i < practiceOutputMax+100; i++ {
		long += "字"
	}
	got := truncatePracticeOutput(long)
	if want := practiceOutputMax + len([]rune("…(截断)")); len([]rune(got)) != want {
		t.Fatalf("截断口径异常: %d want %d", len([]rune(got)), want)
	}
}
