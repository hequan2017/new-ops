// Package service ziqiang：文本鉴权语料（需求 3.5-1）
// 口径：四类型白名单（专利/开源协议/论文/合同）；JSONL/CSV 导入逐行校验+错误行报告
// （对齐 N02/K02 口径）；对比样本对=人创/AIGC 成对校验。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeZqType      = 2700
	ErrCodeZqNotFound  = 2701
	ErrCodeZqImport    = 2702
	ErrCodeZqPairBad   = 2703
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
		return nil, nil, fmt.Errorf("[%d]导入格式仅支持 jsonl/csv", ErrCodeZqImport)
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
		return nil, nil, fmt.Errorf("[%d]CSV 解析失败: %w", ErrCodeZqImport, err)
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

// ZiqiangApp 服务单例出口（自强域服务聚合）
var ZiqiangApp = new(ZiqiangService)

type ZiqiangService struct{}

// ImportTextCorpora 批量导入文本语料
func (s *ZiqiangService) ImportTextCorpora(format, content string) (map[string]any, error) {
	list, rowErrs, err := parseTextImport(format, content)
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

// SaveTextCorpus 新建/更新语料（法律状态白名单，缺省有效）
func (s *ZiqiangService) SaveTextCorpus(in *model.ZqTextCorpus) (*model.ZqTextCorpus, error) {
	if _, reason := corpusFromRow(in.Type, in.Title, fmt.Sprintf("%v", in.AigcFlag), in.Content); reason != "" {
		return nil, fmt.Errorf("[%d]%s", ErrCodeZqType, reason)
	}
	if !ValidLegalStatus(in.LegalStatus) {
		in.LegalStatus = "有效"
	}
	if in.ID > 0 {
		var exist model.ZqTextCorpus
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]语料不存在", ErrCodeZqNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"type": in.Type, "title": in.Title, "content": in.Content, "aigc_flag": in.AigcFlag,
			"source": in.Source, "legal_status": in.LegalStatus, "image_corpus_id": in.ImageCorpusId,
		}).Error
		if err == nil {
			exist.LegalStatus = in.LegalStatus
			exist.ImageCorpusId = in.ImageCorpusId
		}
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteTextCorpus 删除语料（软删除）
func (s *ZiqiangService) DeleteTextCorpus(id uint) error {
	res := global.GVA_DB.Delete(&model.ZqTextCorpus{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]语料不存在", ErrCodeZqNotFound)
	}
	return nil
}

// GetTextCorpora 语料列表（类型/AIGC 过滤，倒序分页）
func (s *ZiqiangService) GetTextCorpora(ztype string, aigc string, keyword string, page, pageSize int) ([]model.ZqTextCorpus, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ZqTextCorpus{})
	if ztype != "" {
		if !ValidTextType(ztype) {
			return nil, 0, fmt.Errorf("[%d]不支持的类型: %s", ErrCodeZqType, ztype)
		}
		db = db.Where("type = ?", ztype)
	}
	if aigc == "yes" {
		db = db.Where("aigc_flag = ?", true)
	} else if aigc == "no" {
		db = db.Where("aigc_flag = ?", false)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ZqTextCorpus
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveComparePair 保存对比样本对（人创/AIGC 成对校验）
func (s *ZiqiangService) SaveComparePair(in *model.ZqComparePair) (*model.ZqComparePair, error) {
	if in.HumanCorpusId == 0 || in.AigcCorpusId == 0 || in.HumanCorpusId == in.AigcCorpusId {
		return nil, fmt.Errorf("[%d]须为不同的人创/AIGC 语料成对", ErrCodeZqPairBad)
	}
	var human, aigc model.ZqTextCorpus
	if err := global.GVA_DB.First(&human, in.HumanCorpusId).Error; err != nil {
		return nil, fmt.Errorf("[%d]人创语料不存在", ErrCodeZqNotFound)
	}
	if err := global.GVA_DB.First(&aigc, in.AigcCorpusId).Error; err != nil {
		return nil, fmt.Errorf("[%d]AIGC 语料不存在", ErrCodeZqNotFound)
	}
	if human.AigcFlag || !aigc.AigcFlag {
		return nil, fmt.Errorf("[%d]样本对须为 人创(非AIGC)+AIGC 组合", ErrCodeZqPairBad)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetComparePairs 对比样本对列表（倒序分页）
func (s *ZiqiangService) GetComparePairs(page, pageSize int) ([]model.ZqComparePair, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ZqComparePair{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ZqComparePair
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
