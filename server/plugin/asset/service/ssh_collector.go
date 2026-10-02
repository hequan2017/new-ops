// Package service 主机 SSH 采集器（M1：hostname/os/cpu/mem/disk 现场回填）
// 依赖登记：golang.org/x/crypto/ssh（go.mod 既有 x/crypto，无新增依赖）
package service

import (
	"fmt"
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
	Username string
	Password string // CredTypeSSHPassword 时使用
	PrivateKey string // CredTypeSSHKey 时使用
	Port     int
}

// parseOSRelease 解析 /etc/os-release
func parseOSRelease(out string) (id, pretty string) {
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
		}
	}
	if id == "" && pretty != "" {
		// 兜底：取 PRETTY_NAME 首词小写
		id = strings.ToLower(strings.Fields(pretty)[0])
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

// parseDFGB 解析 `df -BG /` 根分区总容量（GB）
func parseDFGB(out string) int {
	for i, line := range strings.Split(out, "\n") {
		if i == 0 || !strings.Contains(line, "/") {
			continue // 跳过表头与挂载点不含 / 的行
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			gb, _ := strconv.Atoi(strings.TrimSuffix(fields[1], "G"))
			return gb
		}
	}
	return 0
}

// parseNproc 解析 CPU 核数
func parseNproc(out string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(out))
	return n
}

// collectViaSSH 拨号并采集（整体超时 20s）
func collectViaSSH(ip string, auth SSHAuth) (*CollectedInfo, error) {
	if auth.Port <= 0 {
		auth.Port = 22
	}
	var authMethods []cssh.AuthMethod
	if auth.PrivateKey != "" {
		signer, err := cssh.ParsePrivateKey([]byte(auth.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("私钥解析失败: %w", err)
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
		return nil, fmt.Errorf("无可用的 SSH 认证方式")
	}

	// TODO(M2 term): 替换为统一主机指纹校验（accept-new），当前 M1 采集器为既有技术债
	cfg := &cssh.ClientConfig{
		User:            auth.Username,
		Auth:            authMethods,
		Timeout:         10 * time.Second,
		HostKeyCallback: cssh.InsecureIgnoreHostKey(),
	}

	client, err := cssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, auth.Port), cfg)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
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
	if out, err := run("df -BG / 2>/dev/null | tail -1", 5*time.Second); err == nil {
		info.DiskGB = parseDFGB(out)
	}
	return info, nil
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
	auth := SSHAuth{Username: username, Password: secret, PrivateKey: secret}
	info, err := collectViaSSH(host.IP, auth)
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
	err = global.GVA_DB.Model(&model.AssetHost{}).Where("id = ?", hostID).Updates(updates).Error
	if err != nil {
		return nil, err
	}
	recordHostHistory(global.GVA_DB, hostID, model.HostHistoryUpdate,
		map[string]any{"action": "SSH采集", "collected": info}, operator)
	return info, nil
}
