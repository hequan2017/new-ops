package service

import (
	"strings"
	"testing"
)

// 无 NET 块的旧格式（向后兼容）
const sampleOutput = "0.52 0.58 0.60 1/512 12345\n" +
	"cpu  100 0 50 850 0 0 0 0 0 0\n" +
	"cpu  110 0 60 830 0 0 0 0 0 0\n" +
	"MemTotal:       16000000 kB\n" +
	"MemAvailable:    4000000 kB\n" +
	"/dev/sda1 500000000 300000000 200000000 61% /"

// 含 NET 块的新格式：1s 窗口内 eth0 rx 增 20480B（20KB/s）、tx 增 10240B（10KB/s）
const sampleWithNet = "0.10 0.20 0.30 1/8 999\n" +
	"cpu  100 0 50 850 0 0 0 0 0 0\n" +
	"cpu  200 0 50 850 0 0 0 0 0 0\n" +
	"MemTotal:       8000000 kB\n" +
	"MemAvailable:   4000000 kB\n" +
	"/dev/sda1 100 50 50 50% /\n" +
	"---NET1---\n" +
	"  eth0: 1000000 1000 0 0 0 0 0 0 500000 500 0 0 0 0 0 0\n" +
	"  eth1: 2000000 2000 0 0 0 0 0 0 300000 300 0 0 0 0 0 0\n" +
	"---NET2---\n" +
	"  eth0: 1020480 1000 0 0 0 0 0 0 510240 500 0 0 0 0 0 0\n" +
	"  eth1: 2020480 2000 0 0 0 0 0 0 310240 300 0 0 0 0 0 0"

func TestParsePerfOutput_Legacy(t *testing.T) {
	s, err := ParsePerfOutput(sampleOutput)
	if err != nil {
		t.Fatalf("解析不应报错: %v", err)
	}
	if s.Load1 != 0.52 {
		t.Fatalf("load1 应为 0.52: %v", s.Load1)
	}
	// delta_total=0 → CPU 保持 0（除零保护）
	if s.CPUPercent != 0 {
		t.Fatalf("delta_total=0 时 CPU 应为 0: %v", s.CPUPercent)
	}
	if s.MemPercent != 75 {
		t.Fatalf("mem 应为 75: %v", s.MemPercent)
	}
	if s.DiskPercent != 61 {
		t.Fatalf("disk 应为 61: %v", s.DiskPercent)
	}
	// 旧格式无 NET 块 → 网络指标为 0
	if s.NetRxKBs != 0 || s.NetTxKBs != 0 {
		t.Fatalf("无 NET 块时网络指标应为 0: %v/%v", s.NetRxKBs, s.NetTxKBs)
	}
}

func TestParsePerfOutput_WithNet(t *testing.T) {
	s, err := ParsePerfOutput(sampleWithNet)
	if err != nil {
		t.Fatalf("解析不应报错: %v", err)
	}
	if s.CPUPercent != 100 {
		t.Fatalf("CPU 应为 100: %v", s.CPUPercent)
	}
	// rx: (1020480+2020480)-(1000000+2000000)=40960B → 40KB/s；tx: (510240+310240)-(500000+300000)=20480B → 20KB/s
	if s.NetRxKBs != 40 {
		t.Fatalf("net rx 应为 40KB/s: %v", s.NetRxKBs)
	}
	if s.NetTxKBs != 20 {
		t.Fatalf("net tx 应为 20KB/s: %v", s.NetTxKBs)
	}
	// lo 网卡已在命令侧排除；计数器回绕保护由 netBlockRate 处理
}

func TestNetBlockRate_CounterWrap(t *testing.T) {
	b1 := []string{"  eth0: 5000 0 0 0 0 0 0 0 3000 0 0 0 0 0 0 0"}
	b2 := []string{"  eth0: 100 0 0 0 0 0 0 0 50 0 0 0 0 0 0 0"}
	rx, tx := netBlockRate(b1, b2, 1_000_000_000)
	if rx != 0 || tx != 0 {
		t.Fatalf("计数器回绕应按 0 处理: %v/%v", rx, tx)
	}
}

func TestParsePerfOutput_TooFewLines(t *testing.T) {
	if _, err := ParsePerfOutput(strings.Repeat("x\n", 3)); err == nil {
		t.Fatal("行数不足应报错")
	}
}
