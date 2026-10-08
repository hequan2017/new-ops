//go:build !linux

// 非 Linux 平台采集桩（可编译不可用；Agent 目标平台 Linux 优先）
package main

func readSnapshot() Snapshot {
	return Snapshot{Load1: -1}
}
