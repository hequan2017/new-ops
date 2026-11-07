// Package service oplib：内置算子集（采购件② 八个纯 Go 实现）
// 契约：func(map[string]any) (any, error)；全部无副作用、无外部依赖。
package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// BuiltinFunc 内置算子执行签名
type BuiltinFunc func(args map[string]any) (any, error)

// builtinRegistry 内置算子注册表（名称→实现）
var builtinRegistry = map[string]BuiltinFunc{}

// GetBuiltin 按名取内置算子
func GetBuiltin(name string) (BuiltinFunc, bool) {
	f, ok := builtinRegistry[name]
	return f, ok
}

// argString / argInt 参数提取助手
func argString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return def
}

// opTextClean 文本清洗：去控制字符 + 空白归一 + trim
func opTextClean(args map[string]any) (any, error) {
	text := argString(args, "text")
	var sb strings.Builder
	for _, r := range text {
		if unicode.IsControl(r) && r != '\n' {
			continue
		}
		sb.WriteRune(r)
	}
	lines := strings.Split(sb.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.Join(strings.Fields(l), " ")
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

// splitSentencesLocal 句边界切分（保留结尾符）
func splitSentencesLocal(text string) []string {
	var out []string
	var buf strings.Builder
	for _, r := range text {
		buf.WriteRune(r)
		if strings.ContainsRune("。！？!?；;\n", r) {
			if s := strings.TrimSpace(buf.String()); s != "" {
				out = append(out, s)
			}
			buf.Reset()
		}
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		out = append(out, s)
	}
	return out
}

// opSplitSentences 分句
func opSplitSentences(args map[string]any) (any, error) {
	return splitSentencesLocal(argString(args, "text")), nil
}

// opDedupeLines 按行去重（归一空白后全局保序去重）
func opDedupeLines(args map[string]any) (any, error) {
	text := argString(args, "text")
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		key := strings.Join(strings.Fields(line), " ")
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, line)
	}
	return strings.Join(out, "\n"), nil
}

// bigramLocal 词元切分（ASCII 词 + CJK bigram，与 llmops 口径一致）
func bigramLocal(text string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	var terms []string
	for _, tk := range tokens {
		runes := []rune(strings.ToLower(tk))
		hasCJK := false
		for _, r := range runes {
			if unicode.Is(unicode.Han, r) {
				hasCJK = true
				break
			}
		}
		if !hasCJK || len(runes) == 1 {
			terms = append(terms, string(runes))
			continue
		}
		for i := 0; i+1 < len(runes); i++ {
			terms = append(terms, string(runes[i:i+2]))
		}
	}
	return terms
}

// opKeywordExtract 关键词抽取（bigram 频次 topN）
func opKeywordExtract(args map[string]any) (any, error) {
	text := argString(args, "text")
	topN := argInt(args, "topN", 10)
	if topN <= 0 || topN > 100 {
		topN = 10
	}
	freq := map[string]int{}
	for _, t := range bigramLocal(text) {
		freq[t]++
	}
	type pair struct {
		term string
		n    int
	}
	var pairs []pair
	for t, n := range freq {
		pairs = append(pairs, pair{t, n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].n != pairs[j].n {
			return pairs[i].n > pairs[j].n
		}
		return pairs[i].term < pairs[j].term
	})
	if len(pairs) > topN {
		pairs = pairs[:topN]
	}
	out := make([]map[string]any, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, map[string]any{"term": p.term, "count": p.n})
	}
	return out, nil
}

// opCountStats 字数统计（字符/汉字词元/行/句）
func opCountStats(args map[string]any) (any, error) {
	text := argString(args, "text")
	runes := []rune(text)
	return map[string]any{
		"chars":    len(runes),
		"words":    len(bigramLocal(text)),
		"lines":    len(strings.Split(strings.TrimRight(text, "\n"), "\n")),
		"sentences": len(splitSentencesLocal(text)),
	}, nil
}

// opJSON2CSV JSON 对象数组 → CSV
func opJSON2CSV(args map[string]any) (any, error) {
	raw, err := json.Marshal(args["data"])
	if err != nil {
		return nil, fmt.Errorf("[2305]data 序列化失败: %w", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("[2305]data 须为非空对象数组")
	}
	// 表头取首行键序（稳定性优于并集）
	keys := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	w := csv.NewWriter(&sb)
	_ = w.Write(keys)
	for _, row := range rows {
		record := make([]string, len(keys))
		for i, k := range keys {
			record[i] = fmt.Sprintf("%v", row[k])
		}
		_ = w.Write(record)
	}
	w.Flush()
	return sb.String(), nil
}

// opCSV2JSON CSV → 对象数组
func opCSV2JSON(args map[string]any) (any, error) {
	r := csv.NewReader(strings.NewReader(argString(args, "text")))
	rows, err := r.ReadAll()
	if err != nil || len(rows) < 1 {
		return nil, fmt.Errorf("[2305]CSV 解析失败或为空")
	}
	header := rows[0]
	out := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		obj := map[string]string{}
		for i, h := range header {
			if i < len(row) {
				obj[strings.TrimSpace(h)] = row[i]
			}
		}
		out = append(out, obj)
	}
	return out, nil
}

// opRegexExtract 正则提取
func opRegexExtract(args map[string]any) (any, error) {
	pattern := argString(args, "pattern")
	text := argString(args, "text")
	if pattern == "" {
		return nil, fmt.Errorf("[2306]pattern 不能为空")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("[2306]正则编译失败: %w", err)
	}
	matches := re.FindAllString(text, 1000)
	if matches == nil {
		matches = []string{}
	}
	return matches, nil
}

// opLowercase 大小写归一（lower/upper/title）
func opLowercase(args map[string]any) (any, error) {
	text := argString(args, "text")
	mode := argString(args, "mode")
	if mode == "" {
		mode = "lower"
	}
	switch mode {
	case "lower":
		return strings.ToLower(text), nil
	case "upper":
		return strings.ToUpper(text), nil
	case "title":
		return strings.Title(strings.ToLower(text)), nil //nolint:staticcheck // 逐词首字母大写
	default:
		return nil, fmt.Errorf("[2307]mode 仅支持 lower/upper/title")
	}
}

func init() {
	builtinRegistry["text_clean"] = opTextClean
	builtinRegistry["split_sentences"] = opSplitSentences
	builtinRegistry["dedupe_lines"] = opDedupeLines
	builtinRegistry["keyword_extract"] = opKeywordExtract
	builtinRegistry["count_stats"] = opCountStats
	builtinRegistry["json2csv"] = opJSON2CSV
	builtinRegistry["csv2json"] = opCSV2JSON
	builtinRegistry["regex_extract"] = opRegexExtract
	builtinRegistry["case_normalize"] = opLowercase
}
