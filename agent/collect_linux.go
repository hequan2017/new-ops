//go:build linux

// 平台采集层：读 /proc 与根分区 statfs（仅 Linux；Agent 目标平台 Linux 优先）
package main

import (
	"os"
	"syscall"
)

// readSnapshot 一轮 /proc 与磁盘读数；单项失败置不可得位
func readSnapshot() Snapshot {
	var s Snapshot
	if b, err := os.ReadFile("/proc/stat"); err == nil {
		if total, idle, err := ParseProcStat(b); err == nil {
			s.CPUTotal, s.CPUIdle, s.HaveCPU = total, idle, true
		}
	}
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		if total, avail, err := ParseMemInfo(b); err == nil {
			s.MemTotal, s.MemAvailable = total, avail
		}
	}
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		if load1, err := ParseLoadavg(b); err == nil {
			s.Load1 = load1
		}
	} else {
		s.Load1 = -1
	}
	if b, err := os.ReadFile("/proc/net/dev"); err == nil {
		if rx, tx, err := ParseNetDev(b); err == nil {
			s.NetRx, s.NetTx, s.HaveNet = rx, tx, true
		}
	}
	if st, err := statfsRoot(); err == nil {
		s.DiskTotal = st.Blocks * uint64(st.Bsize)
		s.DiskFree = st.Bavail * uint64(st.Bsize)
	}
	return s
}

func statfsRoot() (syscall.Statfs_t, error) {
	var st syscall.Statfs_t
	err := syscall.Statfs("/", &st)
	return st, err
}
