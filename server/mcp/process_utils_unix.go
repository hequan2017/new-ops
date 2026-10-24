//go:build !windows

package mcpTool

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func prepareDetachedProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// processCommandLine 取 PID 当前命令行（linux 读 /proc/<pid>/cmdline；darwin 回退 /bin/ps）。
// 用于托管进程 PID 归属校验——托管进程退出后 PID 被系统复用时，仅凭存活判断会把
// 无关进程误判为"我们的"，停用动作会误杀它
func processCommandLine(pid int) (string, bool) {
	if pid <= 0 {
		return "", false
	}
	if raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline"); err == nil {
		return strings.ReplaceAll(string(bytes.TrimRight(raw, "\x00")), "\x00", " "), true
	}
	// darwin 无 /proc：/bin/ps 为系统固定位（不经 PATH 查找）
	out, err := exec.Command("/bin/ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

func terminateProcess(pid int) error {
	if pid <= 0 {
		return nil
	}

	err := syscall.Kill(pid, syscall.SIGTERM)
	if err == nil || errors.Is(err, syscall.ESRCH) {
		return nil
	}

	return err
}
