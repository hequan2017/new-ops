// Package service eduadmin：数据同步接口（需求 3.6-29~33）
// 口径：适配器配置（source/protocol/endpoint/凭据 AES-GCM 密文，对齐 asset/crypto
// 口径不回显明文，同步频率 ≤5 分钟）；同步任务=HTTP GET 拉取 JSON 数组→域白名单/
// 唯一键校验→落库幂等（唯一键去重）→执行记录；人事/财务/资产/后勤/教学科研 五类种子。
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
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduSyncParam    = 2830
	ErrCodeEduSyncNotFound = 2831
)

// EduSyncSources 五类来源白名单
var EduSyncSources = []string{"人事", "财务", "资产", "后勤", "教学科研"}

// ValidEduSyncSource 来源白名单
func ValidEduSyncSource(s string) bool {
	for _, v := range EduSyncSources {
		if v == s {
			return true
		}
	}
	return false
}

// pullRows 拉取负载行
type pullRows []struct {
	ExternalKey string          `json:"externalKey"`
	Payload     json.RawMessage `json:"payload"`
}

// parsePull 解析拉取负载（纯函数，测试锚点）
func parsePull(body []byte) (pullRows, error) {
	var rows pullRows
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("[%d]拉取负载须为 [{externalKey,payload}] 数组: %v", ErrCodeEduSyncParam, err)
	}
	return rows, nil
}

// httpClientSync 同步用 HTTP 客户端（测试可替换）
var httpClientSync = &http.Client{Timeout: 30 * time.Second}

// RunEduSyncTask 执行同步任务（拉取-校验-幂等落库-执行记录）
func (s *EduadminService) RunEduSyncTask(adapterId uint) (*model.EduSyncTask, error) {
	var adapter model.EduSyncAdapter
	if err := global.GVA_DB.First(&adapter, adapterId).Error; err != nil {
		return nil, fmt.Errorf("[%d]适配器不存在", ErrCodeEduSyncNotFound)
	}
	if !ValidEduSyncSource(adapter.Source) {
		return nil, fmt.Errorf("[%d]来源须为人事/财务/资产/后勤/教学科研", ErrCodeEduSyncParam)
	}
	task := &model.EduSyncTask{AdapterId: adapterId}
	if adapter.Status != "启用" {
		task.Message = "适配器已停用，跳过执行"
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	// 凭据解密注入（密文校验，不回显明文）
	token := ""
	if adapter.CredentialEnc != "" {
		plain, err := crypto.Decrypt(adapter.CredentialEnc)
		if err != nil {
			task.Message = "凭据解密失败: " + err.Error()
			if err := global.GVA_DB.Create(task).Error; err != nil {
				return nil, err
			}
			return task, nil
		}
		token = plain
	}
	req, err := http.NewRequest(http.MethodGet, adapter.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("[%d]端点非法: %v", ErrCodeEduSyncParam, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClientSync.Do(req)
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
	rows, err := parsePull(body)
	if err != nil {
		task.Message = err.Error()
		if err := global.GVA_DB.Create(task).Error; err != nil {
			return nil, err
		}
		return task, nil
	}
	task.Pulled = len(rows)
	for _, r := range rows {
		payload := strings.TrimSpace(string(r.Payload))
		if strings.TrimSpace(r.ExternalKey) == "" || !validPayloadJSON(payload) {
			task.Failed++
			continue
		}
		var count int64
		global.GVA_DB.Model(&model.EduRecord{}).
			Where("domain = ? AND external_key = ?", adapter.Source, r.ExternalKey).Count(&count)
		if count > 0 {
			task.Skipped++
			continue
		}
		rec := model.EduRecord{
			Domain: adapter.Source, ExternalKey: r.ExternalKey,
			PayloadJSON: payload, Source: adapter.Source + "同步",
		}
		if err := global.GVA_DB.Create(&rec).Error; err != nil {
			task.Failed++
			continue
		}
		task.Inserted++
	}
	task.OK = task.Failed == 0
	task.Message = fmt.Sprintf("拉取 %d，新入库 %d，去重跳过 %d，失败 %d", task.Pulled, task.Inserted, task.Skipped, task.Failed)
	if err := global.GVA_DB.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

// SaveEduSyncAdapter 新建/更新适配器（凭据明文入口→密文落库）
func (s *EduadminService) SaveEduSyncAdapter(in *model.EduSyncAdapter, credentialPlain string) (*model.EduSyncAdapter, error) {
	if !ValidEduSyncSource(in.Source) {
		return nil, fmt.Errorf("[%d]来源须为人事/财务/资产/后勤/教学科研", ErrCodeEduSyncParam)
	}
	if strings.TrimSpace(in.Endpoint) == "" {
		return nil, fmt.Errorf("[%d]端点必填", ErrCodeEduSyncParam)
	}
	if in.SyncFreqMinutes < 1 || in.SyncFreqMinutes > 5 {
		in.SyncFreqMinutes = 5 // 需求口径：同步频率 ≤5 分钟
	}
	if in.Status == "" {
		in.Status = "启用"
	}
	if credentialPlain != "" {
		enc, err := crypto.Encrypt(credentialPlain)
		if err != nil {
			return nil, fmt.Errorf("[%d]凭据加密失败: %v", ErrCodeEduSyncParam, err)
		}
		in.CredentialEnc = enc
	}
	if in.ID > 0 {
		var exist model.EduSyncAdapter
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]适配器不存在", ErrCodeEduSyncNotFound)
		}
		updates := map[string]any{
			"source": in.Source, "protocol": in.Protocol, "endpoint": in.Endpoint,
			"sync_freq_minutes": in.SyncFreqMinutes, "status": in.Status,
		}
		if in.CredentialEnc != "" {
			updates["credential_enc"] = in.CredentialEnc
		}
		err := global.GVA_DB.Model(&exist).Updates(updates).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetEduSyncAdapters 适配器列表（凭据不回显）
func (s *EduadminService) GetEduSyncAdapters() ([]model.EduSyncAdapter, error) {
	var list []model.EduSyncAdapter
	err := global.GVA_DB.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// GetEduSyncTasks 任务执行记录（适配器过滤，倒序分页）
func (s *EduadminService) GetEduSyncTasks(adapterId uint, page, pageSize int) ([]model.EduSyncTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduSyncTask{})
	if adapterId > 0 {
		db = db.Where("adapter_id = ?", adapterId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduSyncTask
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
