// Package service ziqiang：法律分词器与隐性抄袭识别（需求 3.5-5）
// 口径：复合术语词典（在先使用权/等同侵权…50 词）最大正向匹配，OOV 单字切分；
// 隐性抄袭=可疑文本滑窗 bigram 对源文本 bigram 集合的包含率（阈值可配），纯函数可测。
package service

import (
	"unicode"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/text"
)

// 错误码（ziqiang 段 2700-2799）
const ErrCodeLegalParam = 2709

// legalDictionary 法律复合术语词典（50 词）
var legalDictionary = map[string]bool{
	"在先使用权": true, "等同侵权": true, "专利权": true, "著作权": true, "商标权": true,
	"邻接权": true, "人身权": true, "财产权": true, "优先权": true, "新颖性": true,
	"创造性": true, "实用性": true, "公开使用": true, "现有技术": true, "抵触申请": true,
	"禁止反悔": true, "捐献原则": true, "独立权利要求": true, "从属权利要求": true,
	"权利要求书": true, "说明书附图": true, "请求书": true, "授权公告": true, "驳回": true,
	"复审": true, "无效宣告": true, "行政诉讼": true, "民事诉讼": true, "举证责任": true,
	"诉讼时效": true, "诉讼保全": true, "善意第三人": true, "恶意串通": true, "显失公平": true,
	"重大误解": true, "欺诈": true, "胁迫": true, "乘人之危": true, "无权代理": true,
	"表见代理": true, "违约金": true, "定金": true, "损害赔偿": true, "惩罚性赔偿": true,
	"法定赔偿": true, "合理使用": true, "法定许可": true, "邻接权人": true, "职务发明": true,
	"委托开发": true,
}

// maxLegalTermLen 词典最长词长
var maxLegalTermLen = 1

func init() {
	for w := range legalDictionary {
		if l := len([]rune(w)); l > maxLegalTermLen {
			maxLegalTermLen = l
		}
	}
}

// LegalToken 法律分词结果项
type LegalToken struct {
	Word   string `json:"word"`
	IsDict bool   `json:"isDict"`
}

// LegalSegment 法律文本分词（最大正向匹配 + OOV 单字；纯函数，测试锚点）
func LegalSegment(text string) []LegalToken {
	runes := []rune(text)
	tokens := make([]LegalToken, 0, len(runes)/2)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case r >= 0x4E00 && r <= 0x9FFF:
			matched := ""
			for l := maxLegalTermLen; l >= 1; l-- {
				if i+l <= len(runes) {
					w := string(runes[i : i+l])
					if legalDictionary[w] {
						matched = w
						break
					}
				}
			}
			if matched != "" {
				tokens = append(tokens, LegalToken{Word: matched, IsDict: true})
				i += len([]rune(matched))
			} else {
				tokens = append(tokens, LegalToken{Word: string(r), IsDict: false})
				i++
			}
		case unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			i++
		default:
			j := i
			for j < len(runes) && !(runes[j] >= 0x4E00 && runes[j] <= 0x9FFF) &&
				!unicode.IsSpace(runes[j]) && !unicode.IsPunct(runes[j]) && !unicode.IsSymbol(runes[j]) {
				j++
			}
			// 专利号等标识符内嵌「.」：后继为字母/数字时并入当前 run
			for j < len(runes) && runes[j] == '.' && j+1 < len(runes) && isAlnumRune(runes[j+1]) {
				j++
				for j < len(runes) && !(runes[j] >= 0x4E00 && runes[j] <= 0x9FFF) &&
					!unicode.IsSpace(runes[j]) && !unicode.IsPunct(runes[j]) && !unicode.IsSymbol(runes[j]) {
					j++
				}
			}
			tokens = append(tokens, LegalToken{Word: string(runes[i:j]), IsDict: false})
			i = j
		}
	}
	return tokens
}

// isAlnumRune 字母或数字
func isAlnumRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// PlagiarismHit 抄袭嫌疑片段
type PlagiarismHit struct {
	Position    int     `json:"position"`
	Excerpt     string  `json:"excerpt"`
	Containment float64 `json:"containment"`
}

// SlidingPlagiarism 隐性抄袭识别（纯函数，测试锚点）：
// 可疑文本按 window 滑窗（步长 window/2），窗口 bigram 对源 bigram 集合包含率 ≥ threshold 即命中。
// bigram 集合复用 text.Bigrams 共享口径（rune 级相邻二元组）。
func SlidingPlagiarism(src, susp string, window int, threshold float64) ([]PlagiarismHit, error) {
	if window < 10 || window > 200 {
		return nil, zqErr(ErrCodeLegalParam, "窗口须在 [10,200] 内")
	}
	if threshold < 0.3 || threshold > 1 {
		return nil, zqErr(ErrCodeLegalParam, "阈值须在 [0.3,1] 内")
	}
	srcSet := text.Bigrams(src)
	runes := []rune(susp)
	step := window / 2
	var hits []PlagiarismHit
	for start := 0; start < len(runes); start += step {
		end := start + window
		if end > len(runes) {
			end = len(runes)
		}
		win := string(runes[start:end])
		if len([]rune(win)) < 10 {
			break
		}
		winSet := text.Bigrams(win)
		contained := 0
		for bg := range winSet {
			if _, ok := srcSet[bg]; ok {
				contained++
			}
		}
		ratio := 0.0
		if len(winSet) > 0 {
			ratio = float64(contained) / float64(len(winSet))
		}
		if ratio >= threshold {
			hits = append(hits, PlagiarismHit{Position: start, Excerpt: win, Containment: float64(int(ratio*10000+0.5)) / 10000})
		}
		if end == len(runes) {
			break
		}
	}
	return hits, nil
}
