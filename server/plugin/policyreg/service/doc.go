// Package service policyreg：制度导入（需求 3.8-1/2）
// 口径：格式白名单 Word/PDF/HTML/TXT/MD——TXT/MD 文本直读，Word/PDF/HTML 二进制
// 仅编目（rawText 留空，口径同 A03）；密级三级（公开/内部/涉密），涉密自动脱敏
// （绝密/机密/秘密 行内内容替换 ★，正则纯函数）；SM4 口径标注随涉密置位；
// SHA256(name|rawText) 幂等键。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const (
	ErrCodePrDocParam    = 3000
	ErrCodePrDocNotFound = 3001
)

// PolicyFormats 格式白名单
var PolicyFormats = []string{"Word", "PDF", "HTML", "TXT", "MD"}

// PolicySecurityLevels 密级白名单
var PolicySecurityLevels = []string{"公开", "内部", "涉密"}

// ValidPolicyFormat 格式白名单
func ValidPolicyFormat(f string) bool {
	for _, v := range PolicyFormats {
		if v == f {
			return true
		}
	}
	return false
}

// ValidPolicyLevel 密级白名单
func ValidPolicyLevel(l string) bool {
	for _, v := range PolicySecurityLevels {
		if v == l {
			return true
		}
	}
	return false
}

// secretLinePattern 涉密条款行（绝密/机密/秘密 开头）
var secretLinePattern = regexp.MustCompile(`(?m)^[ \t]*(绝密|机密|秘密)[★☆]?\s*[：:]?.*$`)

// MaskSecretClauses 涉密条款脱敏（纯函数，测试锚点）：密级行整行替换为标记+★
func MaskSecretClauses(text string) (string, int) {
	count := 0
	masked := secretLinePattern.ReplaceAllStringFunc(text, func(line string) string {
		count++
		level := secretLinePattern.FindStringSubmatch(line)[1]
		return fmt.Sprintf("%s★（涉密条款已自动脱敏 %d）", level, count)
	})
	return masked, count
}

// docContentHash 幂等键（纯函数）
func docContentHash(name, rawText string) string {
	sum := sha256.Sum256([]byte(name + "|" + rawText))
	return hex.EncodeToString(sum[:])
}

// SavePolicyDoc 导入制度（txt/md 直读/二进制编目/涉密脱敏/幂等）
func (s *PolicyService) SavePolicyDoc(in *model.PolicyDoc) (*model.PolicyDoc, error) {
	if strings.TrimSpace(in.Name) == "" || !ValidPolicyFormat(in.Format) || !ValidPolicyLevel(in.SecurityLevel) {
		return nil, fmt.Errorf("[%d]名称必填且格式/密级须在白名单内", ErrCodePrDocParam)
	}
	// 二进制格式仅编目
	if in.Format == "TXT" || in.Format == "MD" {
		if strings.TrimSpace(in.RawText) == "" {
			return nil, fmt.Errorf("[%d]TXT/MD 须提供原文文本", ErrCodePrDocParam)
		}
	} else {
		in.RawText = ""
	}
	in.Sm4Mark = in.SecurityLevel == "涉密"
	in.MaskedText, _ = MaskSecretClauses(in.RawText)
	in.ContentHash = docContentHash(in.Name, in.RawText)
	var exist model.PolicyDoc
	if err := global.GVA_DB.Where("content_hash = ?", in.ContentHash).First(&exist).Error; err == nil {
		return &exist, nil // 幂等：重复导入跳过
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetPolicyDocs 制度列表（格式/密级过滤，倒序分页）
func (s *PolicyService) GetPolicyDocs(format, securityLevel string, page, pageSize int) ([]model.PolicyDoc, int64, error) {
	if format != "" && !ValidPolicyFormat(format) {
		return nil, 0, fmt.Errorf("[%d]格式须为 Word/PDF/HTML/TXT/MD", ErrCodePrDocParam)
	}
	if securityLevel != "" && !ValidPolicyLevel(securityLevel) {
		return nil, 0, fmt.Errorf("[%d]密级须为 公开/内部/涉密", ErrCodePrDocParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.PolicyDoc{})
	if format != "" {
		db = db.Where("format = ?", format)
	}
	if securityLevel != "" {
		db = db.Where("security_level = ?", securityLevel)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PolicyDoc
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetPolicyDoc 制度详情
func (s *PolicyService) GetPolicyDoc(id uint) (*model.PolicyDoc, error) {
	var doc model.PolicyDoc
	if err := global.GVA_DB.First(&doc, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]制度不存在", ErrCodePrDocNotFound)
	}
	return &doc, nil
}

// PolicyApp 服务单例出口（policyreg 域服务聚合）
var PolicyApp = new(PolicyService)

type PolicyService struct{}
