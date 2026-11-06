// Package service llmops 媒体域单测（媒体类别映射、时间范围解析、ASR 适配器默认态）
package service

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMediaKindOf(t *testing.T) {
	cases := map[string]string{
		".png": "image", ".JPG": "image", ".svg": "image",
		".mp3": "audio", ".WAV": "audio", ".flac": "audio",
		".mp4": "video", ".mkv": "video", ".webm": "video",
		".txt": "", ".pdf": "", "": "", ".exe": "",
	}
	for in, want := range cases {
		if got := MediaKindOf(in); got != want {
			t.Fatalf("MediaKindOf(%q)=%q want %q", in, got, want)
		}
	}
}

// TestParseTimeRange 三布局 + 区间合法性
func TestParseTimeRange(t *testing.T) {
	// 空入参
	bt, et, err := ParseTimeRange("", "")
	if err != nil || !bt.IsZero() || !et.IsZero() {
		t.Fatalf("空入参应为零值: %v %v %v", bt, et, err)
	}
	// 日期布局
	if _, _, err := ParseTimeRange("2026-11-01", "2026-11-06"); err != nil {
		t.Fatalf("日期布局失败: %v", err)
	}
	// RFC3339
	if _, _, err := ParseTimeRange("2026-11-01T00:00:00Z", "2026-11-06T23:59:59+08:00"); err != nil {
		t.Fatalf("RFC3339 失败: %v", err)
	}
	// datetime 布局
	if _, _, err := ParseTimeRange("2026-11-01 08:00:00", "2026-11-06 09:00:00"); err != nil {
		t.Fatalf("datetime 布局失败: %v", err)
	}
	// begin > end
	if _, _, err := ParseTimeRange("2026-11-06", "2026-11-01"); err == nil || !strings.Contains(err.Error(), "晚于") {
		t.Fatalf("begin>end 应报错: %v", err)
	}
	// 非法格式
	if _, _, err := ParseTimeRange("November 1", ""); err == nil || !strings.Contains(err.Error(), "2226") {
		t.Fatalf("非法格式应报 2226: %v", err)
	}
	// 仅 begin / 仅 end
	if bt, _, err := ParseTimeRange("2026-11-01", ""); err != nil || bt.IsZero() {
		t.Fatalf("仅 begin 失败: %v %v", bt, err)
	}
	if _, et, err := ParseTimeRange("", "2026-11-06"); err != nil || et.IsZero() {
		t.Fatalf("仅 end 失败: %v %v", et, err)
	}
}

// TestNoopAsr 默认适配器契约：未配置明确报 2223（替换点由后续模型接入实现）
func TestNoopAsr(t *testing.T) {
	if _, ok := Asr.(*NoopAsrAdapter); !ok {
		t.Fatalf("默认适配器应为 NoopAsrAdapter")
	}
	_, err := Asr.Transcribe(context.Background(), nil, "a.mp3")
	if err == nil || !strings.Contains(err.Error(), "2223") {
		t.Fatalf("未配置应报 2223: %v", err)
	}
}

// TestAsrAdapterInterface 编译期接口契约
func TestAsrAdapterInterface(t *testing.T) {
	var _ AsrAdapter = &NoopAsrAdapter{}
	fake := fakeAsr{}
	var _ AsrAdapter = &fake
	out, err := fake.Transcribe(context.Background(), nil, "x.wav")
	if err != nil || out != "fake-text" {
		t.Fatalf("替换点未生效: %q %v", out, err)
	}
	_ = time.Now() // 保持 time 引用
}

type fakeAsr struct{}

func (fakeAsr) Transcribe(context.Context, []byte, string) (string, error) { return "fake-text", nil }
