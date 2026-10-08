// Package main 白泽轻量 Agent（M9 A1）：注册/心跳/系统指标采集上报。
// 采集解析为纯函数（内容进、结果出），文件读取按平台分离（collect_linux.go 读 /proc 与 statfs）。
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Sample 一轮采集结果（指标名与 monitor 插件约定对齐，落 monitor_metric 后 PerfChart 直接可用）
type Sample struct {
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float64 `json:"memPercent"`
	DiskPercent float64 `json:"diskPercent"`
	Load1       float64 `json:"load1"`
	NetRxKBs    float64 `json:"netRxKBs"`
	NetTxKBs    float64 `json:"netTxKBs"`
}

// Points 转上报点集（首轮 CPU/网络无增量基线，返回负载/内存/磁盘三项）
func (s Sample) Points() []agentPoint {
	pts := make([]agentPoint, 0, 6)
	if s.CPUPercent >= 0 {
		pts = append(pts, agentPoint{Name: "cpu_percent", Value: s.CPUPercent})
	}
	if s.MemPercent >= 0 {
		pts = append(pts, agentPoint{Name: "mem_percent", Value: s.MemPercent})
	}
	if s.DiskPercent >= 0 {
		pts = append(pts, agentPoint{Name: "disk_percent", Value: s.DiskPercent})
	}
	if s.Load1 >= 0 {
		pts = append(pts, agentPoint{Name: "load1", Value: s.Load1})
	}
	if s.NetRxKBs >= 0 {
		pts = append(pts, agentPoint{Name: "net_rx_kbs", Value: s.NetRxKBs})
	}
	if s.NetTxKBs >= 0 {
		pts = append(pts, agentPoint{Name: "net_tx_kbs", Value: s.NetTxKBs})
	}
	return pts
}

// Snapshot 一轮原始读数（平台采集层填充；-1 表示不可得）
type Snapshot struct {
	CPUTotal     uint64
	CPUIdle      uint64
	HaveCPU      bool
	MemTotal     uint64
	MemAvailable uint64
	DiskTotal    uint64
	DiskFree     uint64
	Load1        float64
	NetRx        uint64
	NetTx        uint64
	HaveNet      bool
}

// Collector 采集器（跨轮次维护 CPU/网络增量基线；Interval 为采样间隔）
type Collector struct {
	prevCPUTotal uint64
	prevCPUIdle  uint64
	prevNetRx    uint64
	prevNetTx    uint64
	haveCPU      bool
	haveNet      bool
	Interval     float64 // 秒
}

// Sample 计算本轮指标（纯逻辑）
func (c *Collector) Sample(snap Snapshot) Sample {
	var s Sample
	s.CPUPercent, s.NetRxKBs, s.NetTxKBs = -1, -1, -1
	if snap.HaveCPU {
		if c.haveCPU && snap.CPUTotal > c.prevCPUTotal {
			totalDelta := float64(snap.CPUTotal - c.prevCPUTotal)
			idleDelta := float64(snap.CPUIdle - c.prevCPUIdle)
			if idleDelta < 0 {
				idleDelta = 0
			}
			if totalDelta > 0 {
				s.CPUPercent = round2((1 - idleDelta/totalDelta) * 100)
			}
		}
		c.prevCPUTotal, c.prevCPUIdle, c.haveCPU = snap.CPUTotal, snap.CPUIdle, true
	}
	if snap.MemTotal > 0 {
		s.MemPercent = round2(float64(snap.MemTotal-snap.MemAvailable) / float64(snap.MemTotal) * 100)
	} else {
		s.MemPercent = -1
	}
	if snap.DiskTotal > 0 {
		s.DiskPercent = round2(float64(snap.DiskTotal-snap.DiskFree) / float64(snap.DiskTotal) * 100)
	} else {
		s.DiskPercent = -1
	}
	s.Load1 = snap.Load1
	if snap.HaveNet {
		if c.haveNet && c.Interval > 0 {
			rx, tx := snap.NetRx, snap.NetTx
			if rx >= c.prevNetRx {
				s.NetRxKBs = round2(float64(rx-c.prevNetRx) / c.Interval / 1024)
			}
			if tx >= c.prevNetTx {
				s.NetTxKBs = round2(float64(tx-c.prevNetTx) / c.Interval / 1024)
			}
		}
		c.prevNetRx, c.prevNetTx, c.haveNet = snap.NetRx, snap.NetTx, true
	}
	return s
}

// ParseProcStat 解析 /proc/stat 首行 cpu 汇总：cpu  user nice system idle iowait ...
// total = 除 cpu 标签外全列和；idle = idle + iowait（与内核口径一致）
func ParseProcStat(b []byte) (total, idle uint64, err error) {
	line := firstLine(string(b))
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "cpu" {
		return 0, 0, errors.New("proc stat 格式非法")
	}
	for i, f := range fields[1:] {
		v, perr := strconv.ParseUint(f, 10, 64)
		if perr != nil {
			return 0, 0, fmt.Errorf("proc stat 数值非法: %w", perr)
		}
		total += v
		if i == 3 || i == 4 { // idle, iowait
			idle += v
		}
	}
	return total, idle, nil
}

// ParseMemInfo 解析 /proc/meminfo 的 MemTotal / MemAvailable（kB）
func ParseMemInfo(b []byte) (total, available uint64, err error) {
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, perr := strconv.ParseUint(fields[1], 10, 64)
		if perr != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			total = v
		case "MemAvailable:":
			available = v
		}
	}
	if total == 0 {
		return 0, 0, errors.New("meminfo 缺少 MemTotal")
	}
	return total, available, nil
}

// ParseLoadavg 解析 /proc/loadavg 首字段 load1
func ParseLoadavg(b []byte) (float64, error) {
	fields := strings.Fields(firstLine(string(b)))
	if len(fields) < 1 {
		return 0, errors.New("loadavg 为空")
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("loadavg 数值非法: %w", err)
	}
	return v, nil
}

// ParseNetDev 解析 /proc/net/dev：汇总非 lo 接口收发字节
func ParseNetDev(b []byte) (rx, tx uint64, err error) {
	for _, line := range strings.Split(string(b), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "" || iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rxV, e1 := strconv.ParseUint(fields[0], 10, 64)
		txV, e2 := strconv.ParseUint(fields[8], 10, 64)
		if e1 != nil || e2 != nil {
			continue
		}
		rx += rxV
		tx += txV
	}
	return rx, tx, nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
