package main

import (
	"math"
	"testing"
)

func TestParseProcStat(t *testing.T) {
	content := []byte("cpu  100 0 50 800 100 0 0 0 0 0\ncpu0 1 0 1 9 0 0 0 0 0 0\nintr 1\n")
	total, idle, err := ParseProcStat(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// total = 100+0+50+800+100 = 1050；idle = 800+100 = 900
	if total != 1050 {
		t.Fatalf("total 应为 1050: %d", total)
	}
	if idle != 900 {
		t.Fatalf("idle 应为 900: %d", idle)
	}
	if _, _, err := ParseProcStat([]byte("garbage")); err == nil {
		t.Fatal("非 cpu 行应报错")
	}
}

func TestParseMemInfo(t *testing.T) {
	content := []byte("MemTotal:       32784216 kB\nMemFree:         1024 kB\nMemAvailable:   16384216 kB\nBuffers:         100 kB\n")
	total, avail, err := ParseMemInfo(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if total != 32784216 || avail != 16384216 {
		t.Fatalf("字段错误: total=%d avail=%d", total, avail)
	}
	if _, _, err := ParseMemInfo([]byte("SwapTotal: 1 kB\n")); err == nil {
		t.Fatal("缺 MemTotal 应报错")
	}
}

func TestParseLoadavg(t *testing.T) {
	v, err := ParseLoadavg([]byte("0.52 0.58 0.59 1/1234 5678\n"))
	if err != nil || v != 0.52 {
		t.Fatalf("load1 应为 0.52: %v %v", v, err)
	}
}

func TestParseNetDev(t *testing.T) {
	content := []byte(
		"Inter-|   Receive                                                |  Transmit\n" +
			" face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n" +
			"    lo: 1000      10    0    0    0     0          0         0     1000      10    0    0    0     0       0          0\n" +
			"  eth0: 5000      50    0    0    0     0          0         0     7000      70    0    0    0     0       0          0\n" +
			"  eth1: 300       3    0    0    0     0          0         0       400      4    0    0    0     0       0          0\n")
	rx, tx, err := ParseNetDev(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if rx != 5300 || tx != 7400 {
		t.Fatalf("汇总错误: rx=%d tx=%d（lo 应排除）", rx, tx)
	}
}

func TestCollectorSample(t *testing.T) {
	c := &Collector{Interval: 10}
	// 首轮：建立基线，CPU/网络不可得（-1），负载/内存/磁盘可用
	s1 := c.Sample(Snapshot{
		CPUTotal: 1000, CPUIdle: 800, HaveCPU: true,
		MemTotal: 1000, MemAvailable: 400,
		DiskTotal: 2000, DiskFree: 500,
		Load1: 0.5, NetRx: 10000, NetTx: 20000, HaveNet: true,
	})
	if s1.CPUPercent != -1 || s1.NetRxKBs != -1 {
		t.Fatalf("首轮 CPU/网络应为 -1: %+v", s1)
	}
	if s1.MemPercent != 60 || s1.DiskPercent != 75 || s1.Load1 != 0.5 {
		t.Fatalf("静态指标错误: %+v", s1)
	}
	// 二轮：10s 内 cpu total +500（idle +200 → busy 60%），rx +102400B → 10KB/s
	s2 := c.Sample(Snapshot{
		CPUTotal: 1500, CPUIdle: 1000, HaveCPU: true,
		MemTotal: 1000, MemAvailable: 400,
		DiskTotal: 2000, DiskFree: 500,
		Load1: 0.7, NetRx: 112400, NetTx: 20000, HaveNet: true,
	})
	if s2.CPUPercent != 60 {
		t.Fatalf("CPU 百分比应为 60: %v", s2.CPUPercent)
	}
	if s2.NetRxKBs != 10 {
		t.Fatalf("net_rx 应为 10KB/s: %v", s2.NetTxKBs)
	}
	if s2.NetTxKBs != 0 {
		t.Fatalf("net_tx 无增量应为 0: %v", s2.NetTxKBs)
	}
	pts := s2.Points()
	names := map[string]bool{}
	for _, p := range pts {
		names[p.Name] = true
		if math.IsNaN(p.Value) {
			t.Fatalf("指标含 NaN: %+v", p)
		}
	}
	for _, want := range []string{"cpu_percent", "mem_percent", "disk_percent", "load1", "net_rx_kbs", "net_tx_kbs"} {
		if !names[want] {
			t.Fatalf("缺少指标 %s", want)
		}
	}
}

func TestBuildWSURL(t *testing.T) {
	cases := map[string]string{
		"http://1.2.3.4:8888":    "ws://1.2.3.4:8888/agent/ws",
		"https://ops.example.cn": "wss://ops.example.cn/agent/ws",
		"ws://1.2.3.4:8888/":     "ws://1.2.3.4:8888/agent/ws",
	}
	for in, want := range cases {
		if got := buildWSURL(in); got != want {
			t.Fatalf("%s → %s，期望 %s", in, got, want)
		}
	}
}

// TestSample_PointsContract 指标名协议契约（monitor 图表序列与告警规则 metric_name 依赖这些名字，漂移即图表空白+告警失效）：
// 六指标全量时按 cpu/mem/disk/load/rx/tx 序输出，负值（不可得）跳过。
func TestSample_PointsContract(t *testing.T) {
	full := Sample{CPUPercent: 12.5, MemPercent: 40.1, DiskPercent: 55.0, Load1: 0.8, NetRxKBs: 100, NetTxKBs: 50}
	pts := full.Points()
	wantNames := []string{"cpu_percent", "mem_percent", "disk_percent", "load1", "net_rx_kbs", "net_tx_kbs"}
	if len(pts) != len(wantNames) {
		t.Fatalf("期望 %d 个指标点 got %d", len(wantNames), len(pts))
	}
	for i, w := range wantNames {
		if pts[i].Name != w {
			t.Fatalf("指标序[%d] want %s got %s", i, w, pts[i].Name)
		}
	}

	// 负值=不可得：跳过对应点，其余保留
	partial := Sample{CPUPercent: 10, MemPercent: -1, DiskPercent: -1, Load1: -1, NetRxKBs: -1, NetTxKBs: -1}
	pts = partial.Points()
	got := map[string]float64{}
	for _, p := range pts {
		got[p.Name] = p.Value
	}
	if len(pts) != 1 || got["cpu_percent"] != 10 {
		t.Fatalf("负值跳过口径不符: %+v", pts)
	}
	if _, ok := got["load1"]; ok {
		t.Fatal("load1=-1 应被跳过")
	}
}
