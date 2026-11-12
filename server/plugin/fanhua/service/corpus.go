// Package service fanhua：沪语语料库（CRUD + 批量导入 + 检索；需求 3.4-1~6）
// 口径：导入 JSONL/CSV 逐行校验（source 白名单+标题/正文必填），错误行报告不中断，
// 合法行批量入库（对齐 llmops K02 口径）；解析为纯函数可测。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/repo"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/csvx"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeCorpusNotFound = 2600
	ErrCodeCorpusSource   = 2601
	ErrCodeImportFormat   = 2602
	ErrCodeCorpusField    = 2603
)

// ErrCodeFanhuaStore 存储层统一包装码（段内未用码 2606）：裸 gorm 错误经 errx.Wrap 携带分段码
const ErrCodeFanhuaStore = 2606

// FanhuaService fanhua 服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB）
type FanhuaService struct {
	db   *gorm.DB
	repo repo.Repo
}

// NewFanhuaService 构造：显式传 db 供单测内存库注入，生产传 nil
func NewFanhuaService(db *gorm.DB) *FanhuaService {
	return &FanhuaService{db: db}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *FanhuaService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// FanhuaApp 服务单例出口（繁花域服务聚合）
var FanhuaApp = NewFanhuaService(nil)

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
		return nil, nil, errx.New(ErrCodeImportFormat, "导入格式仅支持 jsonl/csv")
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
		return nil, nil, errx.Wrap(ErrCodeImportFormat, "CSV 解析失败", err)
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

// ImportCorpora 批量导入（解析→合法行批量入库→错误行报告）
func (s *FanhuaService) ImportCorpora(format, content string) (map[string]any, error) {
	list, rowErrs, err := parseCorpusImport(format, content)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		if err := s.repo.CreateCorporaBatch(s.dbh(), list, 200); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "批量写库语料失败", err)
		}
	}
	return map[string]any{"inserted": len(list), "failedRows": rowErrs}, nil
}

// SaveCorpus 新建/更新语料
func (s *FanhuaService) SaveCorpus(in *model.FanhuaCorpus) (*model.FanhuaCorpus, error) {
	if _, reason := corpusFromFields(in.Source, in.Title, in.Lang, in.Content, 0); reason != "" {
		return nil, errx.New(ErrCodeCorpusSource, reason)
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindCorpus(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeCorpusNotFound, "语料不存在")
		}
		if err := s.repo.UpdateCorpusColumns(db, exist.ID, map[string]any{
			"source": in.Source, "title": in.Title, "content": in.Content, "lang": in.Lang, "meta_json": in.MetaJSON,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新语料失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateCorpus(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建语料失败", err)
	}
	return in, nil
}

// DeleteCorpus 删除语料（软删除）
func (s *FanhuaService) DeleteCorpus(id uint) error {
	rows, err := s.repo.DeleteCorpus(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除语料失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeCorpusNotFound, "语料不存在")
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
	if source != "" && !ValidSource(source) {
		return nil, 0, errx.New(ErrCodeCorpusSource, "不支持的来源: "+source)
	}
	kw := strings.TrimSpace(keyword)
	list, total, err := s.repo.CorpusList(s.dbh(), source, lang, kw, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询语料失败", err)
	}
	return list, total, nil
}

// corporaToCSV 语料 CSV（纯函数，测试锚点，I07 口径）：复用 csvx.Builder 全字段清洗
// （换行折空格先于逗号转全角）。fanhua 原实现仅清洗 title/content，csvx 为全字段清洗超集，
// 既有测试未覆盖差异字段（source/lang 无逗号/换行样本），口径兼容。
func corporaToCSV(list []model.FanhuaCorpus) string {
	b := csvx.New("来源", "标题", "语种", "内容")
	for _, c := range list {
		b.Row(c.Source, c.Title, c.Lang, c.Content)
	}
	return b.String()
}

// ExportCorpora 语料 CSV 导出（来源/语种过滤，倒序）
func (s *FanhuaService) ExportCorpora(source, lang string) (string, error) {
	list, err := s.repo.CorporaForExport(s.dbh(), source, lang)
	if err != nil {
		return "", errx.Wrap(ErrCodeFanhuaStore, "扫描语料导出失败", err)
	}
	return corporaToCSV(list), nil
}
