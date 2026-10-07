package service

import (
	"strings"
	"testing"
)

const osReleaseSample = `NAME="Ubuntu"
VERSION="22.04.5 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 22.04.5 LTS"
VERSION_ID="22.04"`

func Test_parseOSRelease(t *testing.T) {
	id, pretty := parseOSRelease(osReleaseSample)
	if id != "ubuntu" {
		t.Fatalf("id = %s, want ubuntu", id)
	}
	if pretty != "Ubuntu 22.04.5 LTS" {
		t.Fatalf("pretty = %s", pretty)
	}
	// 兜底：仅 NAME 无 ID/PRETTY_NAME
	id2, pretty2 := parseOSRelease("NAME=\"CentOS Linux\"\n")
	if id2 != "centos" || pretty2 != "" {
		t.Fatalf("兜底解析 = %s / %s", id2, pretty2)
	}
}

func Test_parseMemKB(t *testing.T) {
	out := "MemTotal:       65493212 kB\nMemFree:      1024 kB"
	if got := parseMemKB(out); got != 65493212 {
		t.Fatalf("parseMemKB = %d", got)
	}
	if got := parseMemKB("garbage"); got != 0 {
		t.Fatalf("无 MemTotal 应为 0, got %d", got)
	}
}

func Test_parseDFGB(t *testing.T) {
	out := "Filesystem     1G-blocks  Used Available Use% Mounted on\n/dev/sda4           229G   57G      172G  25% /"
	if got := parseDFGB(out); got != 229 {
		t.Fatalf("parseDFGB = %d, want 229", got)
	}
	if got := parseDFGB("Filesystem 1G-blocks\n"); got != 0 {
		t.Fatalf("无数据行应为 0, got %d", got)
	}
}

func Test_parseNproc(t *testing.T) {
	if got := parseNproc("16\n"); got != 16 {
		t.Fatalf("parseNproc = %d", got)
	}
	if got := parseNproc(""); got != 0 {
		t.Fatalf("空输出应为 0, got %d", got)
	}
	if got := parseNproc(strings.TrimSpace(" 8 ")); got != 8 {
		t.Fatalf("trim 后应为 8, got %d", got)
	}
}
