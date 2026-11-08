// Package service fanhua：沪语语料库（CRUD + 批量导入 + 检索；需求 3.4-1~6）
// 口径：导入 JSONL/CSV 逐行校验（source 白名单+标题/正文必填），错误行报告不中断，
// 合法行批量入库（对齐 llmops K02 口径）；解析为纯函数可测。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeCorpusNotFound = 2600
	ErrCodeCorpusSource   = 2601
	ErrCodeImportFormat   = 2602
	ErrCodeCorpusField    = 2603
)

// CorpusSources 语料来源白名单
var CorpusSources = []string{"新闻", "社交", "文学", "对照", "舆情", "城市"}

// ValidSource 来源白名单
func ValidSource(s string) bool {
	for _, v := range CorpusSources {
		if v == s {
			return true
		}
	}
	return false
}

// ValidLang 语种白名单
func ValidLang(l string) bool {
	switch l {
	case "", "沪语", "普通话", "英语", "混合":
		return true
	}
	return false
}

// RowError 导入错误行报告
type RowError struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

// parseCorpusImport 解析导入内容（纯函数，测试锚点）：JSONL 或 CSV（列序 source,title,lang,content）
func parseCorpusImport(format, content string) ([]model.FanhuaCorpus, []RowError, error) {
	switch format {
	case "jsonl":
		return parseCorpusJSONL(content)
	case "csv":
		return parseCorpusCSV(content)
	default:
		return nil, nil, fmt.Errorf("[%d]导入格式仅支持 jsonl/csv", ErrCodeImportFormat)
	}
}

func corpusFromFields(source, title, lang, content string, line int) (*model.FanhuaCorpus, string) {
	if !ValidSource(source) {
		return nil, fmt.Sprintf("来源须在白名单内: %s", source)
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return nil, "标题与正文必填"
	}
	if !ValidLang(lang) {
		return nil, fmt.Sprintf("语种非法: %s", lang)
	}
	return &model.FanhuaCorpus{Source: source, Title: title, Lang: lang, Content: content}, ""
}

func parseCorpusJSONL(content string) ([]model.FanhuaCorpus, []RowError, error) {
	var list []model.FanhuaCorpus
	var errs []RowError
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var in struct {
			Source  string `json:"source"`
			Title   string `json:"title"`
			Content string `json:"content"`
			Lang    string `json:"lang"`
		}
		if err := json.Unmarshal([]byte(line), &in); err != nil {
			errs = append(errs, RowError{Line: i + 1, Reason: "JSON 解析失败: " + err.Error()})
			continue
		}
		c, reason := corpusFromFields(in.Source, in.Title, in.Lang, in.Content, i+1)
		if reason != "" {
			errs = append(errs, RowError{Line: i + 1, Reason: reason})
			continue
		}
		list = append(list, *c)
	}
	return list, errs, nil
}

func parseCorpusCSV(content string) ([]model.FanhuaCorpus, []RowError, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.FieldsPerRecord = 4
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("[%d]CSV 解析失败: %w", ErrCodeImportFormat, err)
	}
	var list []model.FanhuaCorpus
	var errs []RowError
	for i, rec := range records {
		if i == 0 && (rec[0] == "source" || rec[0] == "来源") {
			continue // 表头
		}
		c, reason := corpusFromFields(rec[0], rec[1], rec[2], rec[3], i+1)
		if reason != "" {
			errs = append(errs, RowError{Line: i + 1, Reason: reason})
			continue
		}
		list = append(list, *c)
	}
	return list, errs, nil
}

// FanhuaApp 服务单例出口（繁花域服务聚合）
var FanhuaApp = new(FanhuaService)

type FanhuaService struct{}

// ImportCorpora 批量导入（解析→合法行批量入库→错误行报告）
func (s *FanhuaService) ImportCorpora(format, content string) (map[string]any, error) {
	list, rowErrs, err := parseCorpusImport(format, content)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		if err := global.GVA_DB.CreateInBatches(list, 200).Error; err != nil {
			return nil, err
		}
	}
	return map[string]any{"inserted": len(list), "failedRows": rowErrs}, nil
}

// SaveCorpus 新建/更新语料
func (s *FanhuaService) SaveCorpus(in *model.FanhuaCorpus) (*model.FanhuaCorpus, error) {
	if _, reason := corpusFromFields(in.Source, in.Title, in.Lang, in.Content, 0); reason != "" {
		return nil, fmt.Errorf("[%d]%s", ErrCodeCorpusSource, reason)
	}
	if in.ID > 0 {
		var exist model.FanhuaCorpus
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]语料不存在", ErrCodeCorpusNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"source": in.Source, "title": in.Title, "content": in.Content, "lang": in.Lang, "meta_json": in.MetaJSON,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteCorpus 删除语料（软删除）
func (s *FanhuaService) DeleteCorpus(id uint) error {
	res := global.GVA_DB.Delete(&model.FanhuaCorpus{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]语料不存在", ErrCodeCorpusNotFound)
	}
	return nil
}

// GetCorpora 语料列表（来源/语种/关键字过滤，倒序分页）
func (s *FanhuaService) GetCorpora(source, lang, keyword string, page, pageSize int) ([]model.FanhuaCorpus, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaCorpus{})
	if source != "" {
		if !ValidSource(source) {
			return nil, 0, fmt.Errorf("[%d]不支持的来源: %s", ErrCodeCorpusSource, source)
		}
		db = db.Where("source = ?", source)
	}
	if lang != "" {
		db = db.Where("lang = ?", lang)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaCorpus
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
