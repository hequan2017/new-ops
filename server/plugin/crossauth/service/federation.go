// Package service crossauth：跨校身份系统对接（需求 3.7-1）
// 口径：协议白名单 SAML2.0/OAuth2.0；密钥 AES-GCM 密文落库不回显（对齐 asset/crypto）；
// 拉取=HTTP GET [{name,userNo,dept,identityStatus}]，school+userNo 幂等 upsert
// （存在更新身份状态，否则新建）；同步任务记录含 ≤5 分钟频率口径字段；
// 身份状态白名单 在读/在职/离校/离职。
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaFedParam    = 2900
	ErrCodeCaFedNotFound = 2901
)

// CaProtocols 协议白名单
var CaProtocols = []string{"SAML2.0", "OAuth2.0"}

// CaIdentityStatuses 身份状态白名单
var CaIdentityStatuses = []string{"在读", "在职", "离校", "离职"}

// ValidCaIdentityStatus 身份状态白名单
func ValidCaIdentityStatus(s string) bool {
	for _, v := range CaIdentityStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// caPullRow 拉取负载行
type caPullRow struct {
	Name           string `json:"name"`
	UserNo         string `json:"userNo"`
	Dept           string `json:"dept"`
	IdentityStatus string `json:"identityStatus"`
}

// parseCaPull 解析拉取负载（纯函数，测试锚点）
func parseCaPull(body []byte) ([]caPullRow, error) {
	var rows []caPullRow
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("[%d]负载须为 [{name,userNo,dept,identityStatus}] 数组: %v", ErrCodeCaFedParam, err)
	}
	return rows, nil
}

// caHttpClient 同步 HTTP 客户端（测试可替换）
var caHttpClient = &http.Client{Timeout: 30 * time.Second}

// RunCaSyncTask 执行用户同步（拉取-校验-幂等落库-执行记录）
func (s *CaService) RunCaSyncTask(federationId uint) (*model.CaSyncTask, error) {
	var fed model.CaFederation
	if err := global.GVA_DB.First(&fed, federationId).Error; err != nil {
		return nil, fmt.Errorf("[%d]对接配置不存在", ErrCodeCaFedNotFound)
	}
	task := &model.CaSyncTask{FederationId: federationId, FreqMinutes: 5}
	if fed.Status != "启用" {
		task.Message = "对接配置已停用，跳过执行"
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	token := ""
	if fed.SecretEnc != "" {
		plain, err := crypto.Decrypt(fed.SecretEnc)
		if err != nil {
			task.Message = "密钥解密失败: " + err.Error()
			if err := global.GVA_DB.Create(task).Error; err != nil {
				return nil, err
			}
			return task, nil
		}
		token = plain
	}
	req, err := http.NewRequest(http.MethodGet, fed.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("[%d]端点非法: %v", ErrCodeCaFedParam, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := caHttpClient.Do(req)
	if err != nil {
		task.Message = "拉取失败: " + err.Error()
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		task.Message = fmt.Sprintf("拉取响应异常: status=%d", resp.StatusCode)
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	rows, err := parseCaPull(body)
	if err != nil {
		task.Message = err.Error()
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	task.Pulled = len(rows)
	for _, r := range rows {
		if strings.TrimSpace(r.UserNo) == "" || strings.TrimSpace(r.Name) == "" || !ValidCaIdentityStatus(r.IdentityStatus) {
			task.Message = fmt.Sprintf("%s", "存在校验失败行")
			continue
		}
		var exist model.CaSyncUser
		err := global.GVA_DB.Where("school = ? AND user_no = ?", fed.School, r.UserNo).First(&exist).Error
		if err == nil {
			if exist.IdentityStatus != r.IdentityStatus {
				if err := global.GVA_DB.Model(&exist).Updates(map[string]any{
					"name": r.Name, "dept": r.Dept, "identity_status": r.IdentityStatus,
				}).Error; err != nil {
					continue
				}
				task.Updated++
			}
			continue
		}
		rec := model.CaSyncUser{
			Name: r.Name, UserNo: r.UserNo, School: fed.School,
			Dept: r.Dept, IdentityStatus: r.IdentityStatus,
		}
		if err := global.GVA_DB.Create(&rec).Error; err != nil {
			continue
		}
		task.Inserted++
	}
	task.OK = true
	task.Message = fmt.Sprintf("拉取 %d，新入库 %d，状态更新 %d", task.Pulled, task.Inserted, task.Updated)
	if err := global.GVA_DB.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

// SaveCaFederation 新建/更新对接配置（密钥明文入口→密文落库）
func (s *CaService) SaveCaFederation(in *model.CaFederation, secretPlain string) (*model.CaFederation, error) {
	valid := false
	for _, p := range CaProtocols {
		if p == in.Protocol {
			valid = true
		}
	}
	if !valid {
		return nil, fmt.Errorf("[%d]协议须为 SAML2.0/OAuth2.0", ErrCodeCaFedParam)
	}
	if strings.TrimSpace(in.School) == "" || strings.TrimSpace(in.Endpoint) == "" {
		return nil, fmt.Errorf("[%d]高校与端点必填", ErrCodeCaFedParam)
	}
	if in.Status == "" {
		in.Status = "启用"
	}
	if secretPlain != "" {
		enc, err := crypto.Encrypt(secretPlain)
		if err != nil {
			return nil, fmt.Errorf("[%d]密钥加密失败: %v", ErrCodeCaFedParam, err)
		}
		in.SecretEnc = enc
	}
	if in.ID > 0 {
		var exist model.CaFederation
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]对接配置不存在", ErrCodeCaFedNotFound)
		}
		updates := map[string]any{
			"school": in.School, "protocol": in.Protocol, "endpoint": in.Endpoint, "status": in.Status,
		}
		if in.SecretEnc != "" {
			updates["secret_enc"] = in.SecretEnc
		}
		err := global.GVA_DB.Model(&exist).Updates(updates).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaFederations 对接配置列表（密钥不回显）
func (s *CaService) GetCaFederations() ([]model.CaFederation, error) {
	var list []model.CaFederation
	err := global.GVA_DB.Order("id ASC").Limit(50).Find(&list).Error
	return list, err
}

// GetCaSyncUsers 同步用户列表（院校/身份状态过滤，倒序分页）
func (s *CaService) GetCaSyncUsers(school, identityStatus string, page, pageSize int) ([]model.CaSyncUser, int64, error) {
	if identityStatus != "" && !ValidCaIdentityStatus(identityStatus) {
		return nil, 0, fmt.Errorf("[%d]身份状态须为 在读/在职/离校/离职", ErrCodeCaFedParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.CaSyncUser{})
	if school != "" {
		db = db.Where("school = ?", school)
	}
	if identityStatus != "" {
		db = db.Where("identity_status = ?", identityStatus)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CaSyncUser
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetCaSyncTasks 任务记录（对接配置过滤，倒序分页）
func (s *CaService) GetCaSyncTasks(federationId uint, page, pageSize int) ([]model.CaSyncTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.CaSyncTask{})
	if federationId > 0 {
		db = db.Where("federation_id = ?", federationId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.CaSyncTask
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CaApp 服务单例出口（crossauth 域服务聚合）
var CaApp = new(CaService)

type CaService struct{}
