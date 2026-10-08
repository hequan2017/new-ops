//go:build windows

// Package service compose 子进程执行辅助（M8 C2，Windows 版：隐藏控制台窗口）
// 安全口径同 compose_exec.go（!windows）：绝对路径 + 字面量参数 + 调用方超时 Kill。
package service

import (
	"os/exec"
	"path/filepath"
	"syscall"
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
	cmd := &exec.Cmd{
		Path: bin,
		Args: args,
		Env:  env,
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
