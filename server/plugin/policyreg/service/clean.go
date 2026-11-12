// Package service policyreg：制度清洗（需求 3.8-3）
// 口径：编码统一（UTF-8 乱码率>30% 判异常，对齐 N04/P03 口径）；格式冲突识别
// （同段全角/半角句号混用计数）；条款一致性=重复条号（第X条 正则重号）+冲突
// 关键词对（同条内 应当×不得/必须×禁止/允许×禁止 共现）；纯函数可测。
package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（policyreg 段 3000-3099）
const ErrCodePrCleanParam = 3002

// cleanIssue 清洗问题项
type cleanIssue struct {
	Type     string `json:"type"` // 编码/格式冲突/重复条号/条款冲突
	Location string `json:"location"`
	Detail   string `json:"detail"`
}

// clauseNoPattern 条号（中文数字或阿拉伯数字）
var clauseNoPattern = regexp.MustCompile(`第([一二三四五六七八九十百零0-9]+)条`)

// mojibakeRatePr 乱码率（纯函数）
func mojibakeRatePr(s string) float64 {
	if s == "" {
		return 0
	}
	bad := 0
	for _, r := range s {
		if r == utf8.RuneError {
			bad++
		}
	}
	runes := utf8.RuneCountInString(s)
	if runes == 0 {
		return 0
	}
	return float64(bad) / float64(runes)
}

// splitClauses 按条切分（纯函数辅助）：返回 [(条号, 条文本)]
func splitClauses(text string) [][2]string {
	matches := clauseNoPattern.FindAllStringSubmatchIndex(text, -1)
	out := make([][2]string, 0, len(matches))
	for i, m := range matches {
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		out = append(out, [2]string{text[m[2]:m[3]], strings.TrimSpace(text[m[0]:end])})
	}
	return out
}

// checkClauseConsistency 条款一致性（纯函数，测试锚点）
func checkClauseConsistency(text string) []cleanIssue {
	issues := make([]cleanIssue, 0)
	seen := map[string]int{}
	for _, c := range splitClauses(text) {
		no, body := c[0], c[1]
		if prev, dup := seen[no]; dup {
			issues = append(issues, cleanIssue{
				Type: "重复条号", Location: "第" + no + "条",
				Detail: fmt.Sprintf("与第 %d 处出现的第%s条重复", prev+1, no),
			})
		} else {
			seen[no] = len(issues)
		}
		seen[no]++
		// 冲突关键词对（同条共现）
		pairs := [][2]string{{"应当", "不得"}, {"必须", "禁止"}, {"允许", "禁止"}}
		for _, p := range pairs {
			if strings.Contains(body, p[0]) && strings.Contains(body, p[1]) {
				issues = append(issues, cleanIssue{
					Type: "条款冲突", Location: "第" + no + "条",
					Detail: fmt.Sprintf("同条内「%s」与「%s」共现，须人工复核", p[0], p[1]),
				})
			}
		}
	}
	return issues
}

// checkFormatConflict 格式冲突（纯函数）：全角句号与「CJK 字符后半角句号」混用计数
func checkFormatConflict(text string) []cleanIssue {
	full := strings.Count(text, "。")
	half := 0
	runes := []rune(text)
	for i := 1; i < len(runes); i++ {
		if runes[i] == '.' && runes[i-1] > 127 {
			half++ // CJK 后的半角句号视为句读
		}
	}
	if full > 0 && half > 0 {
		return []cleanIssue{{
			Type: "格式冲突", Location: "全文",
			Detail: fmt.Sprintf("全角句号 %d 处与半角句号 %d 处混用", full, half),
		}}
	}
	return nil
}

// CleanPolicyDoc 制度清洗（按 docId，dry 报告）
func (s *PolicyService) CleanPolicyDoc(docId uint) (map[string]any, error) {
	doc, err := s.repo.FindDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.New(ErrCodePrCleanParam, "制度不存在")
	}
	text := doc.RawText
	if strings.TrimSpace(text) == "" {
		return nil, errx.New(ErrCodePrCleanParam, "二进制编目文档无可清洗文本")
	}
	issues := make([]cleanIssue, 0)
	rate := mojibakeRatePr(text)
	if rate > 0.3 {
		issues = append(issues, cleanIssue{
			Type: "编码", Location: "全文",
			Detail: fmt.Sprintf("UTF-8 乱码率 %.0f%% 超 30%%，须重新转码后再导入", rate*100),
		})
	}
	issues = append(issues, checkFormatConflict(text)...)
	issues = append(issues, checkClauseConsistency(text)...)
	return map[string]any{
		"docId": docId, "name": doc.Name,
		"clauses": len(splitClauses(text)), "issues": issues,
		"ok": len(issues) == 0,
	}, nil
}
