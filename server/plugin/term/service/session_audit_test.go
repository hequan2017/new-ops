package service

import (
	"testing"
)

func Test_inputAccumulator_feed(t *testing.T) {
	acc := newInputAccumulator()
	// 单条命令
	cmds := acc.feed("ls -la\r")
	if len(cmds) != 1 || cmds[0] != "ls -la" {
		t.Fatalf("单条命令抽取错误: %v", cmds)
	}
	// 带退格修正：systemct → \b\b 越界忽略 → 输入 l → "systemctl"？此处构造确定场景
	cmds = acc.feed("systemctl sta")
	cmds = acc.feed("tus\x7f\x7f\r") // \x7f\x7f 删除 "us" → "systemctl stat"
	if len(cmds) != 1 || cmds[0] != "systemctl stat" {
		t.Fatalf("退格修正错误: %v", cmds)
	}
}

func Test_inputAccumulator_ansiAndPartial(t *testing.T) {
	acc := newInputAccumulator()
	// ANSI 方向键序列 + 部分命令（无回车）→ 抽出空，残留在缓冲
	cmds := acc.feed("\x1b[A")
	if len(cmds) != 0 {
		t.Fatalf("无回车不应抽出命令: %v", cmds)
	}
	cmds = acc.feed("top\r")
	if len(cmds) != 1 || cmds[0] != "top" {
		t.Fatalf("ANSI 剥离后命令抽取错误: %v", cmds)
	}
	// 残留 flush
	acc.feed("hi")
	if rest := acc.flush(); rest != "hi" {
		t.Fatalf("flush 应返回残留: %q", rest)
	}
	if rest := acc.flush(); rest != "" {
		t.Fatalf("flush 后应清空: %q", rest)
	}
}
