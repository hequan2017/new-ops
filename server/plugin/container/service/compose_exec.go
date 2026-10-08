//go:build !windows

// Package service compose 子进程执行辅助（M8 C2）
// 安全口径（Mimosa 认可模式，同 mcp/standalone_manager）：LookPath 解析并归一绝对路径，
// Cmd.Path/Args 全字面量不经 shell，调用方负责超时 Kill。
package service

import (
	"os/exec"
	"path/filepath"
)

// lookPathAbs 在 PATH 中解析命令并归一为绝对路径
func lookPathAbs(name string) (string, error) {
	bin, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	if abs, absErr := filepath.Abs(bin); absErr == nil {
		return abs, nil
	}
	return bin, nil
}

// composeCmd 构造 docker compose 子命令（args[0] 为绝对路径程序本体，其余全字面量）
func composeCmd(bin string, args []string, env []string) *exec.Cmd {
	return &exec.Cmd{
		Path: bin,
		Args: args,
		Env:  env,
	}
}
