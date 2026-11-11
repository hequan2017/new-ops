// Package service eduadmin：政策数据更新（需求 3.6-35）
// 口径：教育部/上海市教委两源白名单；HTTP GET 拉取 [{title,url,publishedAt}]
// 数组→SHA256(title+url) 幂等键去重入库（重复跳过）；时效性字段=发布时间+
// 抓取时间（UTC+8 格式 yyyy-MM-dd HH:mm）；解析/哈希纯函数可测。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const (
	ErrCodeEduGovParam    = 2834
	ErrCodeEduGovNotFound = 2835
)

// EduGovSources 政策来源白名单
var EduGovSources = []string{"教育部", "上海市教委"}

// ValidEduGovSource 来源白名单
func ValidEduGovSource(s string) bool {
	for _, v := range EduGovSources {
		if v == s {
			return true
		}
	}
	return false
}

// govPolicyRow 拉取负载行
type govPolicyRow struct {
	Title       string `json:"title"`
	Url         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

// parseGovPolicy 解析负载（纯函数，测试锚点）
func parseGovPolicy(body []byte) ([]govPolicyRow, error) {
	var rows []govPolicyRow
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("[%d]负载须为 [{title,url,publishedAt}] 数组: %v", ErrCodeEduGovParam, err)
	}
	return rows, nil
}

// govContentHash 幂等键（纯函数，测试锚点）：SHA256(title|url)
func govContentHash(title, url string) string {
	sum := sha256.Sum256([]byte(title + "|" + url))
	return hex.EncodeToString(sum[:])
}

// govFetchedAt 抓取时间（UTC+8）
func govFetchedAt() string {
	return time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04")
}

// PullEduGovPolicy 拉取政策条目（HTTP GET→解析→哈希幂等入库）
func (s *EduadminService) PullEduGovPolicy(source, endpoint string) (map[string]any, error) {
	if !ValidEduGovSource(source) {
		return nil, fmt.Errorf("[%d]来源须为 教育部/上海市教委", ErrCodeEduGovParam)
	}
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("[%d]端点必填", ErrCodeEduGovParam)
	}
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("[%d]拉取失败: %v", ErrCodeEduGovParam, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%d]拉取响应异常: status=%d", ErrCodeEduGovParam, resp.StatusCode)
	}
	rows, err := parseGovPolicy(body)
	if err != nil {
		return nil, err
	}
	fetchedAt := govFetchedAt()
	inserted, skipped, failed := 0, 0, 0
	for _, r := range rows {
		if strings.TrimSpace(r.Title) == "" || strings.TrimSpace(r.Url) == "" {
			failed++
			continue
		}
		hash := govContentHash(r.Title, r.Url)
		var count int64
		global.GVA_DB.Model(&model.EduGovPolicy{}).Where("content_hash = ?", hash).Count(&count)
		if count > 0 {
			skipped++
			continue
		}
		rec := model.EduGovPolicy{
			Title: r.Title, Source: source, Url: r.Url,
			PublishedAt: r.PublishedAt, FetchedAt: fetchedAt, ContentHash: hash,
		}
		if err := global.GVA_DB.Create(&rec).Error; err != nil {
			failed++
			continue
		}
		inserted++
	}
	return map[string]any{
		"source": source, "pulled": len(rows), "inserted": inserted,
		"skipped": skipped, "failed": failed, "fetchedAt": fetchedAt,
	}, nil
}

// GetEduGovPolicies 政策列表（来源过滤，发布时间倒序分页）
func (s *EduadminService) GetEduGovPolicies(source string, page, pageSize int) ([]model.EduGovPolicy, int64, error) {
	if source != "" && !ValidEduGovSource(source) {
		return nil, 0, fmt.Errorf("[%d]来源须为 教育部/上海市教委", ErrCodeEduGovParam)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.EduGovPolicy{})
	if source != "" {
		db = db.Where("source = ?", source)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.EduGovPolicy
	err := db.Order("published_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
