// Package service llmops：问答对采集（JSON/CSV 解析→同题去重→批量落库）
// 解析口径：JSON 为对象数组 [{question,answer}]（兼容 q/a 别名）；CSV 首行表头须含 question(或 q)/answer(或 a) 列；
// 格式自动探测（首字符 [/{ → JSON，否则 CSV）；问题/答案任一为空的行跳过并计数。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeQaParse    = 2208
	ErrCodeQaEmpty    = 2209
	ErrCodeQaTooMany  = 2210
	qaMaxPairs        = 10000
	qaMaxQuestionLen  = 512  // 与表 varchar(512) 对齐
	qaMaxAnswerRune   = 10000
)

// QaItem 解析出的单条问答对（Line 从 1 起，JSON 数组下标或 CSV 行号）
type QaItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Line     int    `json:"line"`
}

// DetectQAFormat 格式自动探测：首非空字符 [/{ → json，否则 csv
func DetectQAFormat(content []byte) string {
	for _, b := range content {
		switch b {
		case ' ', '\t', '\r', '\n', 0xef, 0xbb, 0xbf: // 跳过空白与 UTF-8 BOM
			continue
		case '[', '{':
			return "json"
		default:
			return "csv"
		}
	}
	return "csv"
}

// ParseQaJSON 解析 JSON 问答对数组（question/q 与 answer/a 别名兼容）
func ParseQaJSON(content []byte) ([]QaItem, error) {
	var raw []map[string]any
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("[%d]JSON 解析失败: %w", ErrCodeQaParse, err)
	}
	items := make([]QaItem, 0, len(raw))
	for i, m := range raw {
		q := anyToStr(m, "question", "q")
		a := anyToStr(m, "answer", "a")
		items = append(items, QaItem{Question: q, Answer: a, Line: i + 1})
	}
	return items, nil
}

func anyToStr(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// ParseQaCSV 解析 CSV 问答对（表头映射 question/q、answer/a；返回跳过行数）
func ParseQaCSV(content []byte) ([]QaItem, int, error) {
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(content), "\xef\xbb\xbf")))
	reader.FieldsPerRecord = -1 // 允许行内列数不齐，逐行校验
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("[%d]CSV 解析失败: %w", ErrCodeQaParse, err)
	}
	if len(rows) == 0 {
		return nil, 0, fmt.Errorf("[%d]CSV 内容为空", ErrCodeQaParse)
	}
	qIdx, aIdx := -1, -1
	for i, h := range rows[0] {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "question", "q":
			qIdx = i
		case "answer", "a":
			aIdx = i
		}
	}
	if qIdx < 0 || aIdx < 0 {
		return nil, 0, fmt.Errorf("[%d]CSV 表头须含 question(或 q) 与 answer(或 a) 列", ErrCodeQaParse)
	}
	items := make([]QaItem, 0, len(rows)-1)
	skipped := 0
	for i, row := range rows[1:] {
		line := i + 2 // 表头占第 1 行
		if qIdx >= len(row) || aIdx >= len(row) {
			skipped++
			continue
		}
		items = append(items, QaItem{Question: row[qIdx], Answer: row[aIdx], Line: line})
	}
	return items, skipped, nil
}

// NormalizeQaKey 问题归一键（去首尾空白+内部空白归一+ASCII 小写），同题去重依据
func NormalizeQaKey(q string) string {
	return strings.ToLower(strings.Join(strings.Fields(q), " "))
}

// ValidateDedupeQa 逐条校验（空跳过/限长）+ 同题去重；返回有效条目、跳过数与重复数
func ValidateDedupeQa(items []QaItem) (kept []QaItem, skipped, dup int, err error) {
	kept = make([]QaItem, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, it := range items {
		q := strings.TrimSpace(it.Question)
		a := strings.TrimSpace(it.Answer)
		if q == "" || a == "" {
			skipped++
			continue
		}
		if len([]rune(q)) > qaMaxQuestionLen || len([]rune(a)) > qaMaxAnswerRune {
			return nil, 0, 0, fmt.Errorf("[%d]第 %d 行问题超 %d 字或答案超 %d 字", ErrCodeQaParse, it.Line, qaMaxQuestionLen, qaMaxAnswerRune)
		}
		key := NormalizeQaKey(q)
		if seen[key] {
			dup++
			continue
		}
		seen[key] = true
		kept = append(kept, QaItem{Question: q, Answer: a, Line: it.Line})
	}
	if len(kept) == 0 {
		return nil, skipped, dup, fmt.Errorf("[%d]无有效问答对（共 %d 条）", ErrCodeQaEmpty, len(items))
	}
	if len(kept) > qaMaxPairs {
		return nil, 0, 0, fmt.Errorf("[%d]有效问答对 %d 超上限 %d", ErrCodeQaTooMany, len(kept), qaMaxPairs)
	}
	return kept, skipped, dup, nil
}

// ImportQaPairs 问答对采集落库：资产(qa/import/parsed) + 去重后批量明细，事务保证
func (s *LlmopsServiceStruct) ImportQaPairs(providedName string, content []byte, tags string) (asset *model.LlmopsDataAsset, imported, skipped, dup int, err error) {
	var items []QaItem
	var skipRows int
	switch DetectQAFormat(content) {
	case "json":
		items, err = ParseQaJSON(content)
		if err != nil {
			return nil, 0, 0, 0, err
		}
	default:
		items, skipRows, err = ParseQaCSV(content)
		if err != nil {
			return nil, 0, 0, 0, err
		}
	}
	kept, skipInvalid, dup, err := ValidateDedupeQa(items)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	skipped = skipRows + skipInvalid

	meta, _ := json.Marshal(map[string]any{
		"parseMode": "qa-" + DetectQAFormat(content),
		"imported":  len(kept), "skipped": skipped, "dup": dup,
	})
	asset = &model.LlmopsDataAsset{
		Name:     BuildDocName(providedName, "qa.json"),
		DataType: "qa",
		Source:   "import",
		Status:   "parsed",
		MetaJSON: string(meta),
		Tags:     strings.TrimSpace(tags),
	}
	pairs := make([]model.LlmopsQaPair, len(kept))
	for i, it := range kept {
		pairs[i] = model.LlmopsQaPair{Question: it.Question, Answer: it.Answer}
	}
	txErr := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(asset).Error; err != nil {
			return err
		}
		for i := range pairs {
			pairs[i].AssetId = asset.ID
		}
		return tx.CreateInBatches(pairs, 500).Error
	})
	if txErr != nil {
		return nil, 0, 0, 0, fmt.Errorf("[%d]问答对入库失败: %w", ErrCodeImportFailed, txErr)
	}
	return asset, len(kept), skipped, dup, nil
}
