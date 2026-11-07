// Package service llmops：业务系统连接器（要求(6)）——ERP/CRM/OA/SCM 配置 + 连通测试
// 口径：四类系统 + 四种鉴权（none/bearer/basic/apikey，值 AES-GCM 密文不回显即用即弃）；
// 字段映射 JSON（本地字段↔系统字段）供跨系统工作流编排引用；连通测试带鉴权头探测。
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
)

// 错误码（llmops 段续用）
const (
	ErrCodeConnNotFound = 2289
	ErrCodeConnBad      = 2290
	connProbeTimeout    = 5 * time.Second
)

// ConnectorTypes 系统类型白名单
var ConnectorTypes = []string{"erp", "crm", "oa", "scm"}

// ValidConnectorType 类型白名单
func ValidConnectorType(t string) bool {
	for _, v := range ConnectorTypes {
		if v == t {
			return true
		}
	}
	return false
}

// ValidAuthType 鉴权类型白名单
func ValidAuthType(t string) bool {
	switch t {
	case "none", "bearer", "basic", "apikey":
		return true
	}
	return false
}

// buildAuthHeaders 鉴权头构造（纯函数，测试锚点）
func buildAuthHeaders(authType, authValue string) map[string]string {
	switch authType {
	case "bearer":
		return map[string]string{"Authorization": "Bearer " + authValue}
	case "basic":
		return map[string]string{"Authorization": "Basic " + authValue}
	case "apikey":
		return map[string]string{"X-API-Key": authValue}
	default:
		return map[string]string{}
	}
}

// ParseFieldMap 字段映射解析（纯函数）
func ParseFieldMap(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("[%d]字段映射须为 JSON 对象: %w", ErrCodeConnBad, err)
	}
	return m, nil
}

// probeConnector 连通测试（带鉴权头，任意 HTTP 响应视为可达）
func probeConnector(baseUrl string, headers map[string]string) error {
	base := strings.TrimRight(strings.TrimSpace(baseUrl), "/")
	if base == "" {
		return fmt.Errorf("[%d]未配置系统地址", ErrCodeConnBad)
	}
	req, err := http.NewRequest(http.MethodGet, base, nil)
	if err != nil {
		return fmt.Errorf("[%d]地址非法: %w", ErrCodeConnBad, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: connProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[%d]连通测试失败: %w", ErrCodeConnBad, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return nil
}

// GetConnectors 连接器列表
func (s *LlmopsServiceStruct) GetConnectors() ([]model.LlmopsConnector, error) {
	var list []model.LlmopsConnector
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// ConnectorIn 连接器保存入参（authValue 明文独立传输）
type ConnectorIn struct {
	model.LlmopsConnector
	AuthValue string `json:"authValue"`
}

// SaveConnector 新建/更新连接器（authValue 留空不改）
func (s *LlmopsServiceStruct) SaveConnector(in *ConnectorIn) (*model.LlmopsConnector, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidConnectorType(in.Type) || !ValidAuthType(in.AuthType) {
		return nil, fmt.Errorf("[%d]名称必填且类型须为 erp/crm/oa/scm、鉴权 none/bearer/basic/apikey", ErrCodeConnBad)
	}
	if err := ValidateModelBaseURL(in.BaseUrl); err != nil {
		return nil, fmt.Errorf("[%d]%w", ErrCodeConnBad, err)
	}
	if _, err := ParseFieldMap(in.FieldMap); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsConnector{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]连接器名已存在", ErrCodeConnBad)
	}
	encAuth := ""
	if strings.TrimSpace(in.AuthValue) != "" {
		e, err := crypto.Encrypt(in.AuthValue)
		if err != nil {
			return nil, fmt.Errorf("鉴权值加密失败: %w", err)
		}
		encAuth = e
	}
	if in.ID > 0 {
		var c model.LlmopsConnector
		if err := global.GVA_DB.First(&c, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]连接器不存在", ErrCodeConnNotFound)
		}
		updates := map[string]any{
			"name": in.Name, "type": in.Type, "base_url": in.BaseUrl, "auth_type": in.AuthType,
			"field_map": in.FieldMap, "status": in.Status, "notes": in.Notes,
		}
		if encAuth != "" {
			updates["auth_value"] = encAuth
		}
		if err := global.GVA_DB.Model(&c).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &c, nil
	}
	c := &model.LlmopsConnector{
		Name: in.Name, Type: in.Type, BaseUrl: in.BaseUrl, AuthType: in.AuthType,
		AuthValue: encAuth, FieldMap: in.FieldMap, Status: in.Status, Notes: in.Notes,
	}
	if c.Status == "" {
		c.Status = "enabled"
	}
	if err := global.GVA_DB.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// TestConnector 连通测试（解密鉴权值即用即弃）
func (s *LlmopsServiceStruct) TestConnector(id uint) error {
	var c model.LlmopsConnector
	if err := global.GVA_DB.First(&c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("[%d]连接器不存在", ErrCodeConnNotFound)
		}
		return err
	}
	authValue := ""
	if strings.TrimSpace(c.AuthValue) != "" {
		plain, err := crypto.Decrypt(c.AuthValue)
		if err != nil {
			return fmt.Errorf("鉴权值解密失败: %w", err)
		}
		authValue = plain
	}
	return probeConnector(c.BaseUrl, buildAuthHeaders(c.AuthType, authValue))
}

// DeleteConnector 删除连接器
func (s *LlmopsServiceStruct) DeleteConnector(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsConnector{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]连接器不存在", ErrCodeConnNotFound)
	}
	return nil
}
