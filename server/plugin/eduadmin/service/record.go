// Package service eduadmin：全场景数据接入（需求 3.6-1）
// 口径：四域白名单（人事/财务/资产/后勤）；payloadJSON 须为合法 JSON 对象；
// CSV/JSON 导入逐行校验+错误行报告（对齐 N02/K02 口径，解析纯函数）；
// 唯一键=domain+externalKey，文件内保留首条、与库内重复跳过计数。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduParam   = 2800
	ErrCodeEduImport  = 2801
	ErrCodeEduNotFound = 2802
)

// EduDomains 域白名单
var EduDomains = []string{"人事", "财务", "资产", "后勤"}

// ValidEduDomain 域白名单校验
func ValidEduDomain(d string) bool {
	for _, v := range EduDomains {
		if v == d {
			return true
		}
	}
	return false
}

// EduRowError 导入错误行
type EduRowError struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

// validPayloadJSON 负载须为 JSON 对象（纯函数）
func validPayloadJSON(s string) bool {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") {
		return false
	}
	var m map[string]any
	return json.Unmarshal([]byte(trimmed), &m) == nil
}

// parseEduImport 导入解析（纯函数，测试锚点）：json=对象数组；csv 列序 domain,externalKey,source,payloadJSON
func parseEduImport(format, content string) ([]model.EduRecord, []EduRowError, error) {
	switch format {
	case "json":
		return parseEduJSON(content)
	case "csv":
		return parseEduCSV(content)
	default:
		return nil, nil, fmt.Errorf("[%d]导入格式仅支持 json/csv", ErrCodeEduImport)
	}
}

func parseEduJSON(content string) ([]model.EduRecord, []EduRowError, error) {
	var rows []struct {
		Domain      string          `json:"domain"`
		ExternalKey string          `json:"externalKey"`
		Source      string          `json:"source"`
		Payload     json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(content), &rows); err != nil {
		return nil, nil, fmt.Errorf("[%d]JSON 解析失败: %v", ErrCodeEduImport, err)
	}
	list := make([]model.EduRecord, 0, len(rows))
	errs := make([]EduRowError, 0)
	seen := map[string]bool{}
	for i, r := range rows {
		if !ValidEduDomain(r.Domain) {
			errs = append(errs, EduRowError{Line: i + 1, Reason: fmt.Sprintf("域须在白名单: %s", r.Domain)})
			continue
		}
		if strings.TrimSpace(r.ExternalKey) == "" {
			errs = append(errs, EduRowError{Line: i + 1, Reason: "外部唯一键必填"})
			continue
		}
		payload := strings.TrimSpace(string(r.Payload))
		if !validPayloadJSON(payload) {
			errs = append(errs, EduRowError{Line: i + 1, Reason: "payload 须为 JSON 对象"})
			continue
		}
		key := r.Domain + "|" + r.ExternalKey
		if seen[key] {
			errs = append(errs, EduRowError{Line: i + 1, Reason: "文件内唯一键重复（保留首条）"})
			continue
		}
		seen[key] = true
		list = append(list, model.EduRecord{Domain: r.Domain, ExternalKey: r.ExternalKey, Source: r.Source, PayloadJSON: payload})
	}
	return list, errs, nil
}

func parseEduCSV(content string) ([]model.EduRecord, []EduRowError, error) {
	rd := csv.NewReader(strings.NewReader(content))
	rows, err := rd.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("[%d]CSV 解析失败: %v", ErrCodeEduImport, err)
	}
	list := make([]model.EduRecord, 0, len(rows))
	errs := make([]EduRowError, 0)
	seen := map[string]bool{}
	for i, row := range rows {
		line := i + 1
		if i == 0 && len(row) >= 1 && strings.Contains(row[0], "domain") {
			continue // 表头
		}
		if len(row) < 4 {
			errs = append(errs, EduRowError{Line: line, Reason: "列数不足（domain,externalKey,source,payloadJSON）"})
			continue
		}
		if !ValidEduDomain(row[0]) {
			errs = append(errs, EduRowError{Line: line, Reason: fmt.Sprintf("域须在白名单: %s", row[0])})
			continue
		}
		if strings.TrimSpace(row[1]) == "" {
			errs = append(errs, EduRowError{Line: line, Reason: "外部唯一键必填"})
			continue
		}
		if !validPayloadJSON(row[3]) {
			errs = append(errs, EduRowError{Line: line, Reason: "payload 须为 JSON 对象"})
			continue
		}
		key := row[0] + "|" + row[1]
		if seen[key] {
			errs = append(errs, EduRowError{Line: line, Reason: "文件内唯一键重复（保留首条）"})
			continue
		}
		seen[key] = true
		list = append(list, model.EduRecord{Domain: row[0], ExternalKey: row[1], Source: row[2], PayloadJSON: row[3]})
	}
	return list, errs, nil
}

// ImportEduRecords 批量导入（唯一键去重：库内已存在跳过）
func (s *EduadminService) ImportEduRecords(format, content string) (map[string]any, error) {
	list, rowErrs, err := parseEduImport(format, content)
	if err != nil {
		return nil, err
	}
	inserted, deduped := 0, 0
	for i := range list {
		var exist model.EduRecord
		err := global.GVA_DB.Where("domain = ? AND external_key = ?", list[i].Domain, list[i].ExternalKey).First(&exist).Error
		if err == nil {
			deduped++
			continue
		}
		if err := global.GVA_DB.Create(&list[i]).Error; err != nil {
			return nil, err
		}
		inserted++
	}
	return map[string]any{"inserted": inserted, "deduped": deduped, "failedRows": rowErrs}, nil
}

// SaveEduRecord 新建/更新记录（唯一键冲突拒绝）
func (s *EduadminService) SaveEduRecord(in *model.EduRecord) (*model.EduRecord, error) {
	if !ValidEduDomain(in.Domain) {
		return nil, fmt.Errorf("[%d]域须为人事/财务/资产/后勤", ErrCodeEduParam)
	}
	if strings.TrimSpace(in.ExternalKey) == "" || !validPayloadJSON(in.PayloadJSON) {
		return nil, fmt.Errorf("[%d]外部唯一键必填且 payload 须为 JSON 对象", ErrCodeEduParam)
	}
	if in.ID > 0 {
		var exist model.EduRecord
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]记录不存在", ErrCodeEduNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"domain": in.Domain, "external_key": in.ExternalKey,
			"payload_json": in.PayloadJSON, "source": in.Source,
		}).Error
		return &exist, err
	}
	var count int64
	global.GVA_DB.Model(&model.EduRecord{}).Where("domain = ? AND external_key = ?", in.Domain, in.ExternalKey).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]唯一键已存在: %s/%s", ErrCodeEduParam, in.Domain, in.ExternalKey)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteEduRecord 删除记录（软删除）
func (s *EduadminService) DeleteEduRecord(id uint) error {
	res := global.GVA_DB.Delete(&model.EduRecord{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]记录不存在", ErrCodeEduNotFound)
	}
	return nil
}

// GetEduRecords 记录列表（域/来源过滤，倒序分页）
func (s *EduadminService) GetEduRecords(domain, source string, page, pageSize int) ([]model.EduRecord, int64, error) {
	if domain != "" && !ValidEduDomain(domain) {
		return nil, 0, fmt.Errorf("[%d]域须为人事/财务/资产/后勤", ErrCodeEduParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduRecord{})
	if domain != "" {
		db = db.Where("domain = ?", domain)
	}
	if source != "" {
		db = db.Where("source LIKE ?", "%"+source+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduRecord
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// EduadminApp 服务单例出口（eduadmin 域服务聚合）
var EduadminApp = new(EduadminService)

type EduadminService struct{}
