// Package service eduadmin：跨校身份认证接口（需求 3.6-34）
// 口径：SAML2.0/OAuth2.0 适配器元数据配置（端点须 http/https）；互认范围逗号分隔
// 白名单校验（教学资源/图书馆/算力/实验平台/电子书库）；跳转 URL 组装纯函数
// （SAML→SSO 端点带 RelayState，OAuth2→authorize 端点带 client_id/state/
// response_type）；12 所共建高校种子；用户同步任务记录（频率 ≤5 分钟口径）。
package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduIdpParam    = 2832
	ErrCodeEduIdpNotFound = 2833
)

// EduIdpProtocols 协议白名单
var EduIdpProtocols = []string{"SAML2.0", "OAuth2.0"}

// EduMutualScopes 互认范围白名单
var EduMutualScopes = []string{"教学资源", "图书馆", "算力", "实验平台", "电子书库"}

// ValidEduMutualScope 互认范围逐项校验（逗号分隔）
func ValidEduMutualScope(scope string) bool {
	if strings.TrimSpace(scope) == "" {
		return false
	}
	for _, part := range strings.Split(scope, ",") {
		p := strings.TrimSpace(part)
		ok := false
		for _, v := range EduMutualScopes {
			if v == p {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// buildIdpRedirect 跳转 URL 组装（纯函数，测试锚点）
func buildIdpRedirect(adapter *model.EduIdpAdapter, state string) (string, error) {
	if adapter.Protocol != "SAML2.0" && adapter.Protocol != "OAuth2.0" {
		return "", fmt.Errorf("[%d]协议须为 SAML2.0/OAuth2.0", ErrCodeEduIdpParam)
	}
	base, err := url.Parse(adapter.MetadataEndpoint)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("[%d]元数据端点须为合法 http(s) URL", ErrCodeEduIdpParam)
	}
	q := base.Query()
	switch adapter.Protocol {
	case "SAML2.0":
		q.Set("RelayState", state)
	case "OAuth2.0":
		if strings.TrimSpace(adapter.ClientId) == "" {
			return "", fmt.Errorf("[%d]OAuth2.0 适配器须配置 ClientId", ErrCodeEduIdpParam)
		}
		q.Set("client_id", adapter.ClientId)
		q.Set("response_type", "code")
		q.Set("state", state)
		q.Set("scope", adapter.MutualScope)
	}
	base.RawQuery = q.Encode()
	return base.String(), nil
}

// BuildEduIdpRedirect 认证跳转（适配器查询+URL 组装）
func (s *EduadminService) BuildEduIdpRedirect(school, state string) (string, error) {
	var adapter model.EduIdpAdapter
	if err := global.GVA_DB.Where("school = ?", school).First(&adapter).Error; err != nil {
		return "", fmt.Errorf("[%d]共建高校适配器不存在: %s", ErrCodeEduIdpNotFound, school)
	}
	if adapter.Status != "启用" {
		return "", fmt.Errorf("[%d]适配器已停用: %s", ErrCodeEduIdpParam, school)
	}
	if strings.TrimSpace(state) == "" {
		return "", fmt.Errorf("[%d]state 必填（防 CSRF）", ErrCodeEduIdpParam)
	}
	return buildIdpRedirect(&adapter, state)
}

// SaveEduIdpAdapter 新建/更新适配器
func (s *EduadminService) SaveEduIdpAdapter(in *model.EduIdpAdapter) (*model.EduIdpAdapter, error) {
	if strings.TrimSpace(in.School) == "" {
		return nil, fmt.Errorf("[%d]高校必填", ErrCodeEduIdpParam)
	}
	validProto := false
	for _, p := range EduIdpProtocols {
		if p == in.Protocol {
			validProto = true
		}
	}
	if !validProto {
		return nil, fmt.Errorf("[%d]协议须为 SAML2.0/OAuth2.0", ErrCodeEduIdpParam)
	}
	if !ValidEduMutualScope(in.MutualScope) {
		return nil, fmt.Errorf("[%d]互认范围须在白名单: %s", ErrCodeEduIdpParam, strings.Join(EduMutualScopes, "/"))
	}
	u, err := url.Parse(in.MetadataEndpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("[%d]元数据端点须为合法 http(s) URL", ErrCodeEduIdpParam)
	}
	if in.Status == "" {
		in.Status = "启用"
	}
	var exist model.EduIdpAdapter
	if err := global.GVA_DB.Where("school = ?", in.School).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"protocol": in.Protocol, "metadata_endpoint": in.MetadataEndpoint,
			"client_id": in.ClientId, "mutual_scope": in.MutualScope, "status": in.Status,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduIdpAdapters 适配器列表
func (s *EduadminService) GetEduIdpAdapters() ([]model.EduIdpAdapter, error) {
	var list []model.EduIdpAdapter
	err := global.GVA_DB.Order("id ASC").Limit(50).Find(&list).Error
	return list, err
}

// RunEduIdpUserSync 用户同步任务记录（接口口径留痕，频率 ≤5 分钟）
func (s *EduadminService) RunEduIdpUserSync(school string) (*model.EduIdpUserSync, error) {
	var adapter model.EduIdpAdapter
	if err := global.GVA_DB.Where("school = ?", school).First(&adapter).Error; err != nil {
		return nil, fmt.Errorf("[%d]共建高校适配器不存在: %s", ErrCodeEduIdpNotFound, school)
	}
	task := &model.EduIdpUserSync{
		School: school, SyncFreqMinutes: 5, OK: adapter.Status == "启用",
		Message: fmt.Sprintf("适配器 %s/%s 状态 %s，按 ≤5 分钟频率由调度器执行增量同步", adapter.Protocol, adapter.MetadataEndpoint, adapter.Status),
	}
	if err := global.GVA_DB.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}
