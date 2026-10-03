// Package service 主机 SSH 采集器（M1：hostname/os/cpu/mem/disk 现场回填）
// 依赖登记：golang.org/x/crypto/ssh（go.mod 既有 x/crypto，无新增依赖）
package service

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	cssh "golang.org/x/crypto/ssh"
)

// CollectedInfo SSH 采集结果
type CollectedInfo struct {
	Hostname  string
	OS        string // 发行版 ID（如 ubuntu / centos）
	OSVersion string // PRETTY_NAME 或版本号
	CPUCores  int
	MemGB     int
	DiskGB    int
}

// SSHAuth SSH 认证参数（凭据保险库解密后传入，用后即弃）
type SSHAuth struct {
	Username           string
	Password           string // CredTypeSSHPassword 时使用
	PrivateKey         string // CredTypeSSHKey 时使用
	Port               int
	ExpectedFingerprint string // 主机公钥指纹期望值（空=TOFU 首次信任）
}

// parseOSRelease 解析 /etc/os-release
func parseOSRelease(out string) (id, pretty string) {
	var name string
	for _, line := range strings.Split(out, "\n") {
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k, v := strings.TrimSpace(kv[0]), strings.Trim(strings.TrimSpace(kv[1]), `"`)
		switch k {
		case "ID":
			id = v
		case "PRETTY_NAME":
			pretty = v
		case "NAME":
			name = v
		}
	}
	if id == "" && name != "" {
		// 兜底：取 NAME 首词小写（如 NAME="CentOS Linux" → centos）
		id = strings.ToLower(strings.Fields(name)[0])
	}
	return id, pretty
}

// parseMemKB 解析 /proc/meminfo 的 MemTotal（KB）
func parseMemKB(out string) int {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.Atoi(fields[1])
				return kb
			}
		}
	}
	return 0
}

// parseDFGB 解析 `df -BG / --output=size` 尾行（纯 "NNNG"）；兼容表格格式取 size 列
func parseDFGB(out string) int {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Filesystem") || strings.HasPrefix(line, "1G-blocks") {
			continue
		}
		fields := strings.Fields(line)
		// 单列输出（--output=size）取最后一个字段；表格输出取 size 列（fields[1]）
		candidates := []string{}
		if len(fields) >= 2 {
			candidates = append(candidates, fields[1])
		}
		candidates = append(candidates, fields[len(fields)-1])
		for _, f := range candidates {
			if g, err := strconv.Atoi(strings.TrimSuffix(f, "G")); err == nil && g > 0 {
				return g
			}
		}
	}
	return 0
}

// parseNproc 解析 CPU 核数
func parseNproc(out string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(out))
	return n
}

// DialSSH 建立带主机指纹校验的 SSH 连接（导出供 term 插件 WebSSH/跳板级联复用），
// 返回实际见到的主机公钥指纹（SHA256:xxx）。
// 指纹策略 TOFU：auth.ExpectedFingerprint 为空时信任首次指纹（调用方负责落库），
// 非空时强校验，不匹配拒绝握手（错误码 ErrCodeHostFPMismatch）。
func DialSSH(ip string, auth SSHAuth) (*cssh.Client, string, error) {
	var authMethods []cssh.AuthMethod
	if auth.PrivateKey != "" {
		signer, err := cssh.ParsePrivateKey([]byte(auth.PrivateKey))
		if err != nil {
			return nil, "", fmt.Errorf("私钥解析失败: %w", err)
		}
		authMethods = append(authMethods, cssh.PublicKeys(signer))
	}
	if auth.Password != "" {
		authMethods = append(authMethods, cssh.Password(auth.Password))
		authMethods = append(authMethods, cssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range questions {
				answers[i] = auth.Password
			}
			return answers, nil
		}))
	}
	if len(authMethods) == 0 {
		return nil, "", fmt.Errorf("无可用的 SSH 认证方式")
	}
	var actualFP string
	cfg := &cssh.ClientConfig{
		User:            auth.Username,
		Auth:            authMethods,
		Timeout:         10 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key cssh.PublicKey) error {
			actualFP = cssh.FingerprintSHA256(key)
			if auth.ExpectedFingerprint == "" {
				return nil // TOFU：首次信任，由调用方回填资产
			}
			if actualFP != auth.ExpectedFingerprint {
				return newServiceErr(ErrCodeHostFPMismatch, fmt.Sprintf(
					"主机指纹不匹配：期望 %s 实际 %s（主机可能被重装或存在中间人风险，确认安全后可清空指纹重新录入）",
					auth.ExpectedFingerprint, actualFP))
			}
			return nil
		},
	}
	client, err := cssh.Dial("tcp", SSHDialTarget(ip, auth.Port), cfg)
	return client, actualFP, err
}

// SSHDialTarget 计算拨号地址：纯 IP + 可选端口 → host:port；入参已带端口时直接归一化使用
// （跳板级联在隧道内拨下一跳时复用）
func SSHDialTarget(ip string, port int) string {
	host, sport, err := net.SplitHostPort(ip)
	if err == nil {
		if port <= 0 {
			return net.JoinHostPort(host, sport)
		}
		return net.JoinHostPort(host, strconv.Itoa(port)) // 显式端口优先
	}
	if port <= 0 {
		port = 22
	}
	return net.JoinHostPort(ip, strconv.Itoa(port))
}

// collectViaSSH 拨号并采集（整体超时 20s）；auth.ExpectedFingerprint 由调用方带入实现强校验
func collectViaSSH(ip string, auth SSHAuth) (*CollectedInfo, string, error) {
	client, actualFP, err := DialSSH(ip, auth)
	if err != nil {
		return nil, actualFP, err
	}
	defer client.Close()

	run := func(cmd string, timeout time.Duration) (string, error) {
		session, err := client.NewSession()
		if err != nil {
			return "", err
		}
		defer session.Close()
		type result struct {
			out string
			err error
		}
		ch := make(chan result, 1)
		go func() {
			out, e := session.CombinedOutput(cmd)
			ch <- result{string(out), e}
		}()
		select {
		case r := <-ch:
			return r.out, r.err
		case <-time.After(timeout):
			return "", fmt.Errorf("命令超时(%s): %s", timeout, cmd)
		}
	}

	info := &CollectedInfo{}
	if out, err := run("hostname", 5*time.Second); err == nil {
		info.Hostname = strings.TrimSpace(out)
	}
	if out, err := run("cat /etc/os-release 2>/dev/null || cat /etc/redhat-release 2>/dev/null", 5*time.Second); err == nil {
		if strings.Contains(out, "=") {
			info.OS, info.OSVersion = parseOSRelease(out)
		} else {
			info.OS = "centos"
			info.OSVersion = strings.TrimSpace(out)
		}
	}
	if out, err := run("nproc", 5*time.Second); err == nil {
		info.CPUCores = parseNproc(out)
	}
	if out, err := run("cat /proc/meminfo 2>/dev/null | head -1", 5*time.Second); err == nil {
		info.MemGB = parseMemKB(out) / 1024 / 1024
	}
	if out, err := run("df -BG / --output=size 2>/dev/null | tail -1", 5*time.Second); err == nil {
		info.DiskGB = parseDFGB(out)
	}
	return info, actualFP, nil
}

// CollectHostFromCredential 用绑定的凭据采集并回填主机资产
func (s *AssetHostService) CollectHostFromCredential(hostID, credentialID uint, operator string) (*CollectedInfo, error) {
	var host model.AssetHost
	if err := global.GVA_DB.First(&host, hostID).Error; err != nil {
		return nil, err
	}
	if host.IP == "" {
		return nil, newServiceErr(ErrCodeIPInvalid, "主机缺少内网IP，无法采集")
	}
	credSvc := new(CredCredentialService)
	secret, credType, username, err := credSvc.GetPlaintext(credentialID)
	if err != nil {
		return nil, fmt.Errorf("凭据读取失败: %w", err)
	}
	if credType != model.CredTypeSSHPassword && credType != model.CredTypeSSHKey {
		return nil, newServiceErr(ErrCodeCredTypeInvalid, "采集仅支持 SSH 密码/私钥凭据")
	}
	auth := SSHAuth{Username: username, Password: secret, PrivateKey: secret, ExpectedFingerprint: host.SSHFP}
	info, actualFP, err := collectViaSSH(host.IP, auth)
	if err != nil {
		return nil, err
	}

	// 回填（仅更新采集到的非空字段）
	updates := map[string]any{"last_collect_at": time.Now()}
	if info.Hostname != "" && host.Hostname == "" {
		updates["hostname"] = info.Hostname // 主机名为空时才回填，不覆盖人工命名
	}
	if info.OS != "" {
		updates["os"] = info.OS
		updates["os_version"] = info.OSVersion
	}
	if info.CPUCores > 0 {
		updates["cpu_cores"] = info.CPUCores
	}
	if info.MemGB > 0 {
		updates["mem_gb"] = info.MemGB
	}
	if info.DiskGB > 0 {
		updates["disk_gb"] = info.DiskGB
	}
	if credentialID > 0 {
		updates["credential_id"] = credentialID
	}
	if host.SSHFP == "" && actualFP != "" {
		updates["ssh_fp"] = actualFP // TOFU 首次录入指纹；已有指纹时 DialSSH 已强校验
	}
	err = global.GVA_DB.Model(&model.AssetHost{}).Where("id = ?", hostID).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	recordHostHistory(global.GVA_DB, hostID, model.HostHistoryUpdate,
		map[string]any{"action": "SSH采集", "collected": info, "fingerprint": actualFP}, operator)
	return info, nil
}
