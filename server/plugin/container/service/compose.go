// Package service 白泽容器管理：Compose 编排（M8 C2）
// 实现口径：Docker SDK 无 compose 引擎，`docker compose` CLI 为官方引擎；按 Mimosa 认可的
// exec 安全模式执行——LookPath 解析并归一绝对路径、参数全字面量不经 shell、超时 Kill。
// 接入点映射：unix:// → DOCKER_HOST；tcp://+TLS → DOCKER_HOST+DOCKER_TLS_VERIFY+DOCKER_CERT_PATH（临时证书目录用后即删）。
// 校验：`docker compose config --quiet`（客户端侧解析校验，不依赖 daemon 连接）。
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
	"gorm.io/gorm"
)

// compose 错误码接续 1401-1406
const (
	ErrCodeComposeNameInvalid = 1407
	ErrCodeComposeFileEmpty   = 1408
	ErrCodeComposeBinMissing  = 1409
	ErrCodeComposeFailed      = 1410
)

const composeContentLimit = 512 * 1024

// composeProjectNameReg 项目名规则（对齐 compose 自身约束：小写字母数字开头，含 - _）
var composeProjectNameReg = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// composeService Compose 编排服务
type ComposeService struct{}

// validateComposeProject 项目创建校验（纯逻辑，可单测）
func validateComposeProject(name, content string) error {
	if !composeProjectNameReg.MatchString(name) {
		return newCtErr(ErrCodeComposeNameInvalid, "项目名非法（小写字母/数字开头，仅含小写字母数字-_，≤63 字符）")
	}
	if strings.TrimSpace(content) == "" {
		return newCtErr(ErrCodeComposeFileEmpty, "compose 文件内容不能为空")
	}
	if len(content) > composeContentLimit {
		return newCtErr(ErrCodeComposeFileEmpty, "compose 文件超过 512KB 上限")
	}
	return nil
}

// dockerBin 解析 docker CLI 绝对路径（LookPath 后归一绝对路径）
func dockerBin() (string, error) {
	bin, err := lookPathAbs("docker")
	if err != nil {
		return "", newCtErr(ErrCodeComposeBinMissing, "服务器未找到 docker CLI（compose 引擎依赖 docker compose 子命令）")
	}
	return bin, nil
}

// runCompose 执行一次 docker compose 子命令（写入 compose 文件与 TLS 临时目录，结束后清理）
// 返回合并输出；超时 Kill。参数全为字面量，不经 shell。
func (s *ComposeService) runCompose(ep *model.DockerEndpoint, project, args string, timeout time.Duration) (string, error) {
	bin, err := dockerBin()
	if err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "baize-compose-")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(file, []byte(s.projectContent(project)), 0600); err != nil {
		return "", fmt.Errorf("compose 文件写入失败: %w", err)
	}
	env, cleanup, err := s.endpointEnv(ep, dir)
	if err != nil {
		return "", err
	}
	if cleanup != nil {
		defer cleanup()
	}
	fullArgs := append([]string{bin, "compose", "-p", project, "-f", file, "--project-directory", dir}, strings.Fields(args)...)
	cmd := composeCmd(bin, fullArgs, env)
	var buf strings.Builder
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		return "", newCtErr(ErrCodeComposeFailed, "docker compose 启动失败: "+err.Error())
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var waitErr error
	select {
	case waitErr = <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
		return "", newCtErr(ErrCodeComposeFailed, fmt.Sprintf("docker compose %s 超时（%s）", args, timeout))
	}
	out := buf.String()
	if len(out) > 8192 {
		out = out[:8192] + "\n…（输出截断）"
	}
	if waitErr != nil {
		return out, newCtErr(ErrCodeComposeFailed, fmt.Sprintf("docker compose %s 失败: %s", args, strings.TrimSpace(out)))
	}
	return out, nil
}

// projectContent 取项目 compose 内容（不存在报错）
func (s *ComposeService) projectContent(project string) string {
	var p model.DockerComposeProject
	if err := global.GVA_DB.Where("name = ?", project).First(&p).Error; err != nil {
		return ""
	}
	return p.Content
}

// endpointEnv 接入点 → docker CLI 环境变量；TCP+TLS 时落临时证书目录（cleanup 用后即删）
func (s *ComposeService) endpointEnv(ep *model.DockerEndpoint, dir string) ([]string, func(), error) {
	cleanup := func() {}
	if ep == nil {
		return nil, cleanup, newCtErr(ErrCodeEpNotFound, "接入点不存在")
	}
	env := append(os.Environ(), "DOCKER_HOST="+ep.Addr)
	if strings.HasPrefix(ep.Addr, "tcp://") && ep.TLSCredentialID != nil && *ep.TLSCredentialID != 0 {
		ca, cert, key, err := Service.Endpoint.loadTLSPEMs(*ep.TLSCredentialID)
		if err != nil {
			return nil, cleanup, err
		}
		certDir := filepath.Join(dir, "tls")
		if err := os.MkdirAll(certDir, 0700); err != nil {
			return nil, cleanup, fmt.Errorf("TLS 临时目录创建失败: %w", err)
		}
		for name, pem := range map[string]string{"ca.pem": ca, "cert.pem": cert, "key.pem": key} {
			if err := os.WriteFile(filepath.Join(certDir, name), []byte(pem), 0600); err != nil {
				return nil, cleanup, fmt.Errorf("TLS 证书写入失败: %w", err)
			}
		}
		env = append(env, "DOCKER_TLS_VERIFY=1", "DOCKER_CERT_PATH="+certDir)
	}
	return env, cleanup, nil
}

// validateComposeFile 校验 compose 文件（config --quiet 客户端侧校验，不触 daemon）
func (s *ComposeService) validateComposeFile(content string) error {
	bin, err := dockerBin()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "baize-compose-")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(file, []byte(content), 0600); err != nil {
		return fmt.Errorf("compose 文件写入失败: %w", err)
	}
	cmd := composeCmd(bin, []string{bin, "compose", "-f", file, "config", "--quiet"}, os.Environ())
	var buf strings.Builder
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		return newCtErr(ErrCodeComposeFailed, "docker compose 启动失败: "+err.Error())
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var waitErr error
	select {
	case waitErr = <-done:
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		return newCtErr(ErrCodeComposeFailed, "compose 文件校验超时")
	}
	if waitErr != nil {
		return newCtErr(ErrCodeComposeFailed, "compose 文件校验失败: "+strings.TrimSpace(buf.String()))
	}
	return nil
}

// ComposeServiceItem ps 单服务条目
type ComposeServiceItem struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`
	Ports   string `json:"ports"`
}

// ComposePS 执行 ps 并回写项目状态（running 计数>0 运行中；=0 已停止；介于其间部分运行）
func (s *ComposeService) ComposePS(projectID uint) ([]ComposeServiceItem, error) {
	var p model.DockerComposeProject
	if err := global.GVA_DB.First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newCtErr(ErrCodeEpNotFound, "compose 项目不存在")
		}
		return nil, err
	}
	var ep model.DockerEndpoint
	if err := global.GVA_DB.First(&ep, p.EndpointID).Error; err != nil {
		return nil, newCtErr(ErrCodeEpNotFound, "接入点不存在")
	}
	out, err := s.runCompose(&ep, p.Name, "ps --format json", 30*time.Second)
	if err != nil && out == "" {
		return nil, err
	}
	items := parseComposePS(out)
	status := model.ComposeStatusDown
	running := 0
	for _, it := range items {
		if it.State == "running" {
			running++
		}
	}
	if len(items) > 0 && running == len(items) {
		status = model.ComposeStatusRunning
	} else if running > 0 {
		status = model.ComposeStatusPartial
	}
	now := time.Now()
	global.GVA_DB.Model(&p).Updates(map[string]any{"status": status, "last_op_at": &now})
	if b, jerr := json.Marshal(items); jerr == nil {
		global.GVA_DB.Model(&p).Update("services", string(b))
	}
	return items, nil
}

// parseComposePS 解析 `docker compose ps --format json`（新版为数组、旧版为逐行 JSON，宽松兼容）
func parseComposePS(out string) []ComposeServiceItem {
	items := []ComposeServiceItem{}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return items
	}
	absorb := func(raw []byte) {
		var m map[string]any
		if json.Unmarshal(raw, &m) != nil {
			return
		}
		it := ComposeServiceItem{
			Name:    strField(m, "Name", "ID"),
			Service: strField(m, "Service", ""),
			State:   strField(m, "State", ""),
			Ports:   strings.Join(publishPorts(m), ","),
		}
		if it.Name != "" || it.Service != "" {
			items = append(items, it)
		}
	}
	var arr []json.RawMessage
	if json.Unmarshal([]byte(trimmed), &arr) == nil {
		for _, raw := range arr {
			absorb(raw)
		}
		return items
	}
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			absorb([]byte(line))
		}
	}
	return items
}

func strField(m map[string]any, key, fallback string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	if fallback != "" {
		if v, ok := m[fallback].(string); ok {
			return v
		}
	}
	return ""
}

func publishPorts(m map[string]any) []string {
	out := []string{}
	raw, ok := m["Publishers"].([]any)
	if !ok {
		if p, ok2 := m["Ports"].(string); ok2 {
			return []string{p}
		}
		return out
	}
	for _, e := range raw {
		pub, ok := e.(map[string]any)
		if !ok {
			continue
		}
		pubPort := intField(pub, "PublishedPort")
		tgt := intField(pub, "TargetPort")
		if pubPort > 0 {
			out = append(out, fmt.Sprintf("%d->%d", pubPort, tgt))
		} else if tgt > 0 {
			out = append(out, fmt.Sprintf("%d", tgt))
		}
	}
	return out
}

func intField(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

// CreateComposeProject 创建项目并立即 up（校验→落库→up→ps 回写状态）
func (s *ComposeService) CreateComposeProject(name string, endpointID uint, content, createdBy, notes string) (*model.DockerComposeProject, error) {
	if err := validateComposeProject(name, content); err != nil {
		return nil, err
	}
	var count int64
	global.GVA_DB.Model(&model.DockerComposeProject{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return nil, newCtErr(ErrCodeEpDuplicate, fmt.Sprintf("compose 项目已存在: %s", name))
	}
	var ep model.DockerEndpoint
	if err := global.GVA_DB.First(&ep, endpointID).Error; err != nil {
		return nil, newCtErr(ErrCodeEpNotFound, "接入点不存在")
	}
	if err := s.validateComposeFile(content); err != nil {
		return nil, err
	}
	p := &model.DockerComposeProject{
		Name: name, EndpointID: endpointID, Content: content,
		Status: model.ComposeStatusUnknown, CreatedBy: createdBy, Notes: notes,
	}
	if err := global.GVA_DB.Create(p).Error; err != nil {
		return nil, err
	}
	if _, err := s.runCompose(&ep, name, "up -d --remove-orphans", 180*time.Second); err != nil {
		return p, err // 项目已落库，up 失败信息随错误返回（前端提示查看状态）
	}
	_, _ = s.ComposePS(p.ID)
	return p, nil
}

// ComposeAction up/down/restart（down 后状态回写已停止）
func (s *ComposeService) ComposeAction(projectID uint, action string) (string, error) {
	var p model.DockerComposeProject
	if err := global.GVA_DB.First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", newCtErr(ErrCodeEpNotFound, "compose 项目不存在")
		}
		return "", err
	}
	var ep model.DockerEndpoint
	if err := global.GVA_DB.First(&ep, p.EndpointID).Error; err != nil {
		return "", newCtErr(ErrCodeEpNotFound, "接入点不存在")
	}
	var args string
	var timeout time.Duration
	switch action {
	case "up":
		args, timeout = "up -d --remove-orphans", 180*time.Second
	case "down":
		args, timeout = "down --remove-orphans", 120*time.Second
	case "restart":
		args, timeout = "restart", 120*time.Second
	default:
		return "", newCtErr(ErrCodeEpAddrInvalid, fmt.Sprintf("未知 compose 动作: %s", action))
	}
	out, err := s.runCompose(&ep, p.Name, args, timeout)
	now := time.Now()
	global.GVA_DB.Model(&p).Updates(map[string]any{"last_op_at": &now})
	if action == "down" {
		global.GVA_DB.Model(&p).Update("status", model.ComposeStatusDown)
	} else if err == nil {
		_, _ = s.ComposePS(p.ID)
	}
	return out, err
}

// UpdateComposeProject 更新 compose 内容并 up（滚动应用：compose up 增量重建变更服务）
func (s *ComposeService) UpdateComposeProject(projectID uint, content string) error {
	if strings.TrimSpace(content) == "" {
		return newCtErr(ErrCodeComposeFileEmpty, "compose 文件内容不能为空")
	}
	if len(content) > composeContentLimit {
		return newCtErr(ErrCodeComposeFileEmpty, "compose 文件超过 512KB 上限")
	}
	if err := s.validateComposeFile(content); err != nil {
		return err
	}
	return global.GVA_DB.Model(&model.DockerComposeProject{}).Where("id = ?", projectID).Update("content", content).Error
}

// DeleteComposeProject down 后删除项目登记
func (s *ComposeService) DeleteComposeProject(projectID uint) error {
	var p model.DockerComposeProject
	if err := global.GVA_DB.First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return newCtErr(ErrCodeEpNotFound, "compose 项目不存在")
		}
		return err
	}
	var ep model.DockerEndpoint
	if err := global.GVA_DB.First(&ep, p.EndpointID).Error; err == nil {
		_, _ = s.runCompose(&ep, p.Name, "down --remove-orphans", 120*time.Second)
	}
	return global.GVA_DB.Delete(&model.DockerComposeProject{}, projectID).Error
}

// GetComposeProjectList 项目列表（含最近 ps 摘要）
func (s *ComposeService) GetComposeProjectList(keyword string) ([]map[string]any, error) {
	db := global.GVA_DB.Model(&model.DockerComposeProject{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR notes LIKE ?", kw, kw)
	}
	var list []model.DockerComposeProject
	if err := db.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	epNames := map[uint]string{}
	var eps []model.DockerEndpoint
	global.GVA_DB.Find(&eps)
	for _, e := range eps {
		epNames[e.ID] = e.Name
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		out = append(out, map[string]any{
			"ID": p.ID, "name": p.Name, "endpointId": p.EndpointID,
			"endpointName": epNames[p.EndpointID], "status": p.Status,
			"lastOpAt": p.LastOpAt, "createdBy": p.CreatedBy, "notes": p.Notes,
			"updatedAt": p.UpdatedAt,
		})
	}
	return out, nil
}

// GetComposeProjectDetail 详情（内容 + 最近 ps）
func (s *ComposeService) GetComposeProjectDetail(projectID uint) (map[string]any, error) {
	var p model.DockerComposeProject
	if err := global.GVA_DB.First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newCtErr(ErrCodeEpNotFound, "compose 项目不存在")
		}
		return nil, err
	}
	items := []ComposeServiceItem{}
	if p.Services != "" {
		_ = json.Unmarshal([]byte(p.Services), &items)
	}
	return map[string]any{
		"ID": p.ID, "name": p.Name, "endpointId": p.EndpointID,
		"content": p.Content, "status": p.Status, "services": items,
		"notes": p.Notes, "createdBy": p.CreatedBy,
	}, nil
}
