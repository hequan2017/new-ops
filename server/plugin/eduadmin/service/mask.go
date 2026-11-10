// Package service eduadmin：数据脱敏与分级（需求 3.6-3）
// 口径：身份证/银行账号 SHA-256 不可逆散列（确定性，键名含 idCard/bankCard/身份证/银行卡）；
// 财务金额小数位隐藏（键名含 amount/salary/金额/工资，截断到整数）；分类分级四级
// 标记（公开/内部/敏感/机密，命中规则取最高级）；转换纯函数无副作用，不落库。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduMask = 2804

// SecurityLevels 分级白名单（低→高）
var SecurityLevels = []string{"公开", "内部", "敏感", "机密"}

// 身份证号（15/18 位）与银行卡号（13-19 位数字）粗匹配
var (
	idCardPattern    = regexp.MustCompile(`\b\d{6}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]\b|\b\d{15}\b`)
	bankCardPattern  = regexp.MustCompile(`\b\d{13,19}\b`)
	sensitivePattern = regexp.MustCompile(`薪酬|工资|绩效|成绩|健康|病历|合同金额`)
	internalPattern  = regexp.MustCompile(`部门|院系|项目|预算|编制|考核`)
)

// HashSensitive SHA-256 不可逆散列（纯函数，测试锚点）
func HashSensitive(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// HideAmountDecimals 金额小数位隐藏：截断到整数（纯函数，测试锚点）
func HideAmountDecimals(v float64) float64 {
	return math.Trunc(v)
}

// sensitiveKeys 敏感键名匹配（小写包含）
func isSensitiveHashKey(key string) bool {
	lk := strings.ToLower(key)
	return strings.Contains(lk, "idcard") || strings.Contains(lk, "bankcard") ||
		strings.Contains(key, "身份证") || strings.Contains(key, "银行卡") || strings.Contains(key, "账号")
}

func isAmountKey(key string) bool {
	lk := strings.ToLower(key)
	return strings.Contains(lk, "amount") || strings.Contains(lk, "salary") ||
		strings.Contains(key, "金额") || strings.Contains(key, "工资") || strings.Contains(key, "费用")
}

// classifyPayload 分类分级（纯函数，测试锚点）：机密>敏感>内部>公开 取最高
func classifyPayload(payload string) string {
	hasSecret := idCardPattern.MatchString(payload) || bankCardPattern.MatchString(payload) ||
		strings.Contains(payload, "机密")
	if hasSecret {
		return "机密"
	}
	if sensitivePattern.MatchString(payload) {
		return "敏感"
	}
	if internalPattern.MatchString(payload) {
		return "内部"
	}
	return "公开"
}

// MaskResult 脱敏结果
type MaskResult struct {
	Level        string   `json:"level"`
	HashedFields []string `json:"hashedFields"`
	AmountFields []string `json:"amountFields"`
	Masked       string   `json:"masked"`
}

// MaskPayload 脱敏转换（纯函数，测试锚点）：身份证/银行卡哈希、金额取整、分级标记
func MaskPayload(payloadJSON string) (*MaskResult, error) {
	var m map[string]any
	dec := json.NewDecoder(strings.NewReader(payloadJSON))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("[%d]payload 须为 JSON 对象", ErrCodeEduMask)
	}
	hashed := make([]string, 0)
	amounts := make([]string, 0)
	level := classifyPayload(payloadJSON) // 分级基于原文（脱敏后敏感值已被替换）
	for k, v := range m {
		if s, ok := v.(string); ok && isSensitiveHashKey(k) {
			m[k] = HashSensitive(s)
			hashed = append(hashed, k)
			continue
		}
		if f, ok := numericOf(v); ok && isAmountKey(k) {
			m[k] = HideAmountDecimals(f)
			amounts = append(amounts, k)
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &MaskResult{
		Level:        level,
		HashedFields: hashed,
		AmountFields: amounts,
		Masked:       string(b),
	}, nil
}

// MaskEduRecordBatch 批量预览：对域内记录做脱敏预览（dry，不落库）
func (s *EduadminService) MaskEduRecordBatch(domain string) ([]map[string]any, error) {
	if !ValidEduDomain(domain) {
		return nil, fmt.Errorf("[%d]域须为人事/财务/资产/后勤", ErrCodeEduMask)
	}
	var records []model.EduRecord
	if err := global.GVA_DB.Where("domain = ?", domain).Order("id ASC").Limit(500).Find(&records).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		res, err := MaskPayload(rec.PayloadJSON)
		if err != nil {
			out = append(out, map[string]any{"id": rec.ID, "error": err.Error()})
			continue
		}
		out = append(out, map[string]any{"id": rec.ID, "level": res.Level, "hashedFields": res.HashedFields, "amountFields": res.AmountFields})
	}
	return out, nil
}
