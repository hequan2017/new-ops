// Package service llmops：模型服务一键部署（对接 container 插件创建容器 + 健康探测）
// 口径：部署计划(待部署) → 执行(建容器+探测 {base}/v1/models 兜底 /health，任意 HTTP 响应视为服务存活)
// → running；探测未过保持 deploying(容器已建服务未就绪)；销毁走 ContainerAction remove。
package service

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	containerSvc "github.com/hequan2017/new-ops/server/plugin/container/service"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeDeployNotFound = 2235
	ErrCodeDeployBad      = 2236
	ErrCodeDeployExec     = 2237
	deployProbeTimeout    = 3 * time.Second
)

var reSlugChars = regexp.MustCompile(`[^a-z0-9_-]+`)

// DeployIn 部署计划入参
type DeployIn struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	EndpointId uint   `json:"endpointId"`
	Image      string `json:"image"`
	Ports      string `json:"ports"`
	Envs       string `json:"envs"`
	BaseUrl    string `json:"baseUrl"`
	Notes      string `json:"notes"`
}

// buildPortSpecs 端口映射串→Docker 规格（缺省 proto 补 /tcp；宿主端口数字校验 + 复用 container.ParsePortBinding）
func buildPortSpecs(ports string) ([]string, error) {
	var out []string
	for _, p := range strings.FieldsFunc(ports, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		spec := p
		if !strings.Contains(spec, "/") {
			spec = spec + "/tcp"
		}
		host, _, err := containerSvc.ParsePortBinding(spec)
		if err != nil {
			return nil, fmt.Errorf("[%d]%w", ErrCodeDeployBad, err)
		}
		hp, nerr := strconv.Atoi(host)
		if nerr != nil || hp < 1 || hp > 65535 {
			return nil, fmt.Errorf("[%d]宿主端口非法: %s", ErrCodeDeployBad, host)
		}
		out = append(out, spec)
	}
	return out, nil
}

// buildEnvSpecs 环境变量串→K=V 列表（非法项跳过）
func buildEnvSpecs(envs string) []string {
	var out []string
	for _, e := range strings.Split(envs, ",") {
		e = strings.TrimSpace(e)
		if strings.Contains(e, "=") && !strings.HasPrefix(e, "=") {
			out = append(out, e)
		}
	}
	return out
}

// deployContainerName 部署名→合法容器名（llmops- 前缀 + 小写 slug，≤40 字符）
func deployContainerName(name string) string {
	slug := reSlugChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "model"
	}
	full := "llmops-" + slug
	if len(full) > 40 {
		full = full[:40]
	}
	return full
}

// GetDeploys 部署列表
func (s *LlmopsServiceStruct) GetDeploys() ([]model.LlmopsModelDeploy, error) {
	var list []model.LlmopsModelDeploy
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SaveDeployPlan 保存部署计划（镜像必填；端口/环境格式即检）
func (s *LlmopsServiceStruct) SaveDeployPlan(in *DeployIn) (*model.LlmopsModelDeploy, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Image) == "" {
		return nil, fmt.Errorf("[%d]部署名与镜像必填", ErrCodeDeployBad)
	}
	if _, err := buildPortSpecs(in.Ports); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsModelDeploy{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]部署名已存在", ErrCodeDeployBad)
	}
	if in.ID > 0 {
		var d model.LlmopsModelDeploy
		if err := global.GVA_DB.First(&d, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]部署不存在", ErrCodeDeployNotFound)
		}
		err := global.GVA_DB.Model(&d).Updates(map[string]any{
			"name": in.Name, "endpoint_id": in.EndpointId, "image": in.Image,
			"ports": in.Ports, "envs": in.Envs, "base_url": in.BaseUrl, "notes": in.Notes,
		}).Error
		if err != nil {
			return nil, err
		}
		return &d, nil
	}
	d := &model.LlmopsModelDeploy{
		Name: in.Name, EndpointId: in.EndpointId, Image: in.Image,
		Ports: in.Ports, Envs: in.Envs, BaseUrl: in.BaseUrl, Notes: in.Notes, Status: "pending",
	}
	if err := global.GVA_DB.Create(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// DeleteDeploy 删除部署计划（软删除；运行中容器不级联销毁，需先销毁）
func (s *LlmopsServiceStruct) DeleteDeploy(id uint) error {
	var d model.LlmopsModelDeploy
	if err := global.GVA_DB.First(&d, id).Error; err != nil {
		return fmt.Errorf("[%d]部署不存在", ErrCodeDeployNotFound)
	}
	if d.Status == "running" || d.Status == "deploying" {
		return fmt.Errorf("[%d]容器运行中(状态 %s)，请先销毁", ErrCodeDeployBad, d.Status)
	}
	return global.GVA_DB.Delete(&d).Error
}

// probeModelService 健康探测：/v1/models 兜底 /health，任意 HTTP 响应视为存活
func probeModelService(baseUrl string) error {
	base := strings.TrimRight(strings.TrimSpace(baseUrl), "/")
	if base == "" {
		return fmt.Errorf("未配置服务地址")
	}
	client := &http.Client{Timeout: deployProbeTimeout}
	for _, path := range []string{"/v1/models", "/health"} {
		resp, err := client.Get(base + path)
		if err != nil {
			continue // 拒连/超时试下一个路径
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		return nil // 有 HTTP 响应即存活（404 亦视为服务已起）
	}
	return fmt.Errorf("健康探测未通过(%s)", base)
}

// ExecuteDeploy 执行部署：建容器→探测→状态流转
func (s *LlmopsServiceStruct) ExecuteDeploy(id uint) (*model.LlmopsModelDeploy, error) {
	var d model.LlmopsModelDeploy
	if err := global.GVA_DB.First(&d, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]部署不存在", ErrCodeDeployNotFound)
	}
	if d.EndpointId == 0 {
		return nil, fmt.Errorf("[%d]请先选择容器接入点", ErrCodeDeployBad)
	}
	ports, err := buildPortSpecs(d.Ports)
	if err != nil {
		return nil, err
	}
	global.GVA_DB.Model(&d).Updates(map[string]any{"status": "deploying", "err_msg": ""})

	req := containerSvc.CreateContainerReq{
		EndpointID: d.EndpointId, Name: deployContainerName(d.Name), Image: d.Image,
		Ports: ports, Envs: buildEnvSpecs(d.Envs),
		RestartPolicy: "unless-stopped", StartNow: true,
	}
	containerId, err := containerSvc.Service.Endpoint.CreateContainer(d.EndpointId, req)
	if err != nil {
		global.GVA_DB.Model(&d).Updates(map[string]any{"status": "failed", "err_msg": truncateSample(err.Error())})
		return nil, fmt.Errorf("[%d]容器创建失败: %w", ErrCodeDeployExec, err)
	}
	status := "deploying"
	errMsg := ""
	if probeErr := probeModelService(d.BaseUrl); probeErr != nil {
		errMsg = truncateSample(probeErr.Error())
	} else {
		status = "running"
	}
	updates := map[string]any{"container_id": containerId, "status": status, "err_msg": errMsg}
	if err := global.GVA_DB.Model(&d).Updates(updates).Error; err != nil {
		return nil, err
	}
	d.ContainerId, d.Status, d.ErrMsg = containerId, status, errMsg
	return &d, nil
}

// DestroyDeploy 销毁部署：强删容器并回置 stopped
func (s *LlmopsServiceStruct) DestroyDeploy(id uint) (*model.LlmopsModelDeploy, error) {
	var d model.LlmopsModelDeploy
	if err := global.GVA_DB.First(&d, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]部署不存在", ErrCodeDeployNotFound)
	}
	if d.ContainerId != "" && d.EndpointId > 0 {
		if err := containerSvc.Service.Endpoint.ContainerAction(d.EndpointId, d.ContainerId, "remove", true); err != nil {
			// 容器可能已不存在：记录后仍回置 stopped，允许重新部署
			global.GVA_LOG.Error(fmt.Sprintf("llmops 销毁容器失败 deploy=%d: %v", id, err))
		}
	}
	if err := global.GVA_DB.Model(&d).Updates(map[string]any{"status": "stopped", "err_msg": "容器已销毁"}).Error; err != nil {
		return nil, err
	}
	d.Status = "stopped"
	return &d, nil
}
