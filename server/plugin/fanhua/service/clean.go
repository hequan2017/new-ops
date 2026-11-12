// Package service fanhua：数据清洗流水线（需求 3.4-7 数据整合与预处理 / 3.4-8 脱敏）
// 口径：零宽字符剔除 + UTF-8 乱码率检测（>30% 判乱码）+ 手机号/身份证正则脱敏 +
// 重复行剔除；批处理支持逐条报告；清洗纯函数可测。
package service

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const ErrCodeCleanParam = 2610

// phoneRe 手机号（11 位，1 开头第二位 3-9）
// idRe 身份证（18 位，末位可 X）
var phoneMasker = masker{re: phoneRegexp, keepHead: 3, keepTail: 4}
var idMasker = masker{re: idRegexp, keepHead: 6, keepTail: 4}

// CleanStats 单条清洗统计
type CleanStats struct {
	RemovedZeroWidth int  `json:"removedZeroWidth"`
	MaskedPhone      int  `json:"maskedPhone"`
	MaskedID         int  `json:"maskedId"`
	RemovedDupLines  int  `json:"removedDupLines"`
	Garbled          bool `json:"garbled"`
}

// CleanReport 逐条处理报告
type CleanReport struct {
	Index int        `json:"index"`
	Id    uint       `json:"id"`
	Text  string     `json:"text"`
	Stats CleanStats `json:"stats"`
}

// CleanText 单文本清洗（纯函数，测试锚点）
func CleanText(text string, dedupe bool) (string, CleanStats) {
	var stats CleanStats
	// 零宽字符剔除
	var sb strings.Builder
	for _, r := range text {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF':
			stats.RemovedZeroWidth++
		default:
			sb.WriteRune(r)
		}
	}
	out := sb.String()
	// 乱码率检测（UTF-8 非法 + 替换符占比 >30%）
	if utf8.RuneCountInString(out) > 0 {
		bad := strings.Count(out, "\uFFFD")
		if !utf8.ValidString(out) || float64(bad)/float64(utf8.RuneCountInString(out)) > 0.3 {
			stats.Garbled = true
		}
	}
	// 脱敏（先身份证后手机号：身份证内嵌手机号样式子串，先脱敏长模式避免误命中）
	out, stats.MaskedID = idMasker.mask(out)
	out, stats.MaskedPhone = phoneMasker.mask(out)
	// 重复行剔除
	if dedupe {
		out, stats.RemovedDupLines = dedupeLines(out)
	}
	return strings.TrimSpace(out), stats
}

// dedupeLines 重复行剔除（保留首次出现，纯函数）
func dedupeLines(text string) (string, int) {
	seen := map[string]bool{}
	var out []string
	removed := 0
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			out = append(out, line)
			continue
		}
		if seen[trimmed] {
			removed++
			continue
		}
		seen[trimmed] = true
		out = append(out, line)
	}
	return strings.Join(out, "\n"), removed
}

// BatchCleanTexts 批量清洗（纯文本口径，逐条报告）
func BatchCleanTexts(texts []string, dedupe bool) []CleanReport {
	reports := make([]CleanReport, 0, len(texts))
	for i, t := range texts {
		clean, stats := CleanText(t, dedupe)
		reports = append(reports, CleanReport{Index: i, Text: clean, Stats: stats})
	}
	return reports
}

// CleanCorpora 语料库批量清洗（按 ID 清洗 content 并回写）
func (s *FanhuaService) CleanCorpora(ids []uint, dedupe bool) ([]CleanReport, error) {
	if len(ids) == 0 {
		return nil, errx.New(ErrCodeCleanParam, "语料 ID 列表不能为空")
	}
	db := s.dbh()
	reports := make([]CleanReport, 0, len(ids))
	for _, id := range ids {
		corpus, err := s.repo.FindCorpus(db, id)
		if err != nil {
			return nil, errx.New(ErrCodeCleanParam, fmt.Sprintf("语料 %d 不存在", id))
		}
		clean, stats := CleanText(corpus.Content, dedupe)
		if err := s.repo.UpdateCorpusContent(db, corpus, clean); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "回写清洗语料失败", err)
		}
		reports = append(reports, CleanReport{Index: len(reports), Id: corpus.ID, Text: clean, Stats: stats})
	}
	return reports, nil
}
