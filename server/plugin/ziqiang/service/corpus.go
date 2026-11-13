// Package service ziqiang：文本鉴权语料（需求 3.5-1）
// 口径：四类型白名单（专利/开源协议/论文/合同）；JSONL/CSV 导入逐行校验+错误行报告
// （对齐 N02/K02 口径）；对比样本对=人创/AIGC 成对校验。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeZqType     = 2700
	ErrCodeZqNotFound = 2701
	ErrCodeZqImport   = 2702
	ErrCodeZqPairBad  = 2703
)

// TextTypes 语料类型白名单
var TextTypes = []string{"专利", "开源协议", "论文", "合同"}

// ValidTextType 类型白名单
func ValidTextType(t string) bool {
	for _, v := range TextTypes {
		if v == t {
			return true
		}
	}
	return false
}

// ZqRowError 导入错误行
type ZqRowError struct {
	Line   int    `json:"line"`
	Reason string `json:"reason"`
}

// parseTextImport 解析导入（纯函数，测试锚点）：JSONL 或 CSV（列序 type,title,aigcFlag,content）
func parseTextImport(format, content string) ([]model.ZqTextCorpus, []ZqRowError, error) {
	switch format {
	case "jsonl":
		return parseTextJSONL(content)
	case "csv":
		return parseTextCSV(content)
	default:
		return nil, nil, zqErr(ErrCodeZqImport, "导入格式仅支持 jsonl/csv")
	}
}

func corpusFromRow(ztype, title, aigcStr, content string) (*model.ZqTextCorpus, string) {
	if !ValidTextType(ztype) {
		return nil, fmt.Sprintf("类型须在白名单内: %s", ztype)
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return nil, "标题与正文必填"
	}
	aigc := aigcStr == "true" || aigcStr == "1" || aigcStr == "是"
	return &model.ZqTextCorpus{Type: ztype, Title: title, AigcFlag: aigc, Content: content}, ""
}

func parseTextJSONL(content string) ([]model.ZqTextCorpus, []ZqRowError, error) {
	var list []model.ZqTextCorpus
	var errs []ZqRowError
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var in struct {
			Type     string `json:"type"`
			Title    string `json:"title"`
			Content  string `json:"content"`
			AigcFlag bool   `json:"aigcFlag"`
		}
		if err := json.Unmarshal([]byte(line), &in); err != nil {
			errs = append(errs, ZqRowError{Line: i + 1, Reason: "JSON 解析失败: " + err.Error()})
			continue
		}
		c, reason := corpusFromRow(in.Type, in.Title, fmt.Sprintf("%v", in.AigcFlag), in.Content)
		if reason != "" {
			errs = append(errs, ZqRowError{Line: i + 1, Reason: reason})
			continue
		}
		list = append(list, *c)
	}
	return list, errs, nil
}

func parseTextCSV(content string) ([]model.ZqTextCorpus, []ZqRowError, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.FieldsPerRecord = 4
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, zqWrapErr(ErrCodeZqImport, "CSV 解析失败", err)
	}
	var list []model.ZqTextCorpus
	var errs []ZqRowError
	for i, rec := range records {
		if i == 0 && (rec[0] == "type" || rec[0] == "类型") {
			continue
		}
		c, reason := corpusFromRow(rec[0], rec[1], rec[2], rec[3])
		if reason != "" {
			errs = append(errs, ZqRowError{Line: i + 1, Reason: reason})
			continue
		}
		list = append(list, *c)
	}
	return list, errs, nil
}

// ImportTextCorpora 批量导入文本语料
func (s *ZiqiangService) ImportTextCorpora(format, content string) (map[string]any, error) {
	list, rowErrs, err := parseTextImport(format, content)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		if err := s.repo.CreateTexts(s.dbh(), list); err != nil {
			return nil, errx.Wrap(ErrCodeZqStore, "批量落库语料失败", err)
		}
	}
	return map[string]any{"inserted": len(list), "failedRows": rowErrs}, nil
}

// SaveTextCorpus 新建/更新语料（法律状态白名单，缺省有效）
func (s *ZiqiangService) SaveTextCorpus(in *model.ZqTextCorpus) (*model.ZqTextCorpus, error) {
	if _, reason := corpusFromRow(in.Type, in.Title, fmt.Sprintf("%v", in.AigcFlag), in.Content); reason != "" {
		return nil, zqErr(ErrCodeZqType, "%s", reason)
	}
	if !ValidLegalStatus(in.LegalStatus) {
		in.LegalStatus = "有效"
	}
	if in.ID > 0 {
		exist, err := s.repo.FindText(s.dbh(), in.ID)
		if err != nil {
			return nil, zqErr(ErrCodeZqNotFound, "语料不存在")
		}
		if err := s.repo.UpdateTextColumns(s.dbh(), exist.ID, map[string]any{
			"type": in.Type, "title": in.Title, "content": in.Content, "aigc_flag": in.AigcFlag,
			"source": in.Source, "legal_status": in.LegalStatus, "image_corpus_id": in.ImageCorpusId,
		}); err != nil {
			return exist, errx.Wrap(ErrCodeZqStore, "更新语料失败", err)
		}
		exist.LegalStatus = in.LegalStatus
		exist.ImageCorpusId = in.ImageCorpusId
		return exist, nil
	}
	if err := s.repo.CreateText(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库语料失败", err)
	}
	return in, nil
}

// DeleteTextCorpus 删除语料（软删除）
func (s *ZiqiangService) DeleteTextCorpus(id uint) error {
	rows, err := s.repo.DeleteText(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeZqStore, "删除语料失败", err)
	}
	if rows == 0 {
		return zqErr(ErrCodeZqNotFound, "语料不存在")
	}
	return nil
}

// GetTextCorpora 语料列表（类型/AIGC 过滤，倒序分页）
func (s *ZiqiangService) GetTextCorpora(ztype string, aigc string, keyword string, page, pageSize int) ([]model.ZqTextCorpus, int64, error) {
	if ztype != "" && !ValidTextType(ztype) {
		return nil, 0, zqErr(ErrCodeZqType, "不支持的类型: %s", ztype)
	}
	var aigcFlag *bool
	switch aigc {
	case "yes":
		v := true
		aigcFlag = &v
	case "no":
		v := false
		aigcFlag = &v
	}
	list, total, err := s.repo.TextList(s.dbh(), ztype, aigcFlag, strings.TrimSpace(keyword), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeZqStore, "查询语料列表失败", err)
	}
	return list, total, nil
}

// SaveComparePair 保存对比样本对（人创/AIGC 成对校验）
func (s *ZiqiangService) SaveComparePair(in *model.ZqComparePair) (*model.ZqComparePair, error) {
	if in.HumanCorpusId == 0 || in.AigcCorpusId == 0 || in.HumanCorpusId == in.AigcCorpusId {
		return nil, zqErr(ErrCodeZqPairBad, "须为不同的人创/AIGC 语料成对")
	}
	human, err := s.repo.FindText(s.dbh(), in.HumanCorpusId)
	if err != nil {
		return nil, zqErr(ErrCodeZqNotFound, "人创语料不存在")
	}
	aigc, err := s.repo.FindText(s.dbh(), in.AigcCorpusId)
	if err != nil {
		return nil, zqErr(ErrCodeZqNotFound, "AIGC 语料不存在")
	}
	if human.AigcFlag || !aigc.AigcFlag {
		return nil, zqErr(ErrCodeZqPairBad, "样本对须为 人创(非AIGC)+AIGC 组合")
	}
	if err := s.repo.CreateComparePair(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库对比样本对失败", err)
	}
	return in, nil
}

// GetComparePairs 对比样本对列表（倒序分页）
func (s *ZiqiangService) GetComparePairs(page, pageSize int) ([]model.ZqComparePair, int64, error) {
	list, total, err := s.repo.ComparePairList(s.dbh(), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeZqStore, "查询对比样本对列表失败", err)
	}
	return list, total, nil
}
