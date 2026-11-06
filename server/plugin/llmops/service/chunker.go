// Package service llmops：文本切块服务（要求(2) 知识入库基础件）
// 口径：纯 Go 无依赖——段落(\n\n)优先聚块 → 段落超限按句边界(。！？!?；;\n)切 → 单句超限硬切；
// 相邻块可携带重叠（取上块尾部 Overlap rune，且保证不超 MaxLen）；长度一律按 rune 计（中文友好）。
package service

import "strings"

// 默认切块参数
const (
	chunkDefaultMaxLen  = 512
	chunkDefaultOverlap = 64
)

// Chunk 切块结果
type Chunk struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// ChunkOptions 切块参数（0 值取默认；Overlap≥MaxLen 时收敛为 MaxLen/2）
type ChunkOptions struct {
	MaxLen  int
	Overlap int
}

// sentenceEnders 句边界符
const sentenceEnders = "。！？!?；;\n"

// splitSentences 段落→句子（保留结尾符）
func splitSentences(para string) []string {
	var out []string
	var buf strings.Builder
	for _, r := range para {
		buf.WriteRune(r)
		if strings.ContainsRune(sentenceEnders, r) {
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

// hardCut 单句超限硬切（按 rune，无重叠）
func hardCut(s string, maxLen int) []string {
	r := []rune(s)
	if len(r) <= maxLen {
		return []string{s}
	}
	var out []string
	for i := 0; i < len(r); i += maxLen {
		end := i + maxLen
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[i:end]))
	}
	return out
}

// packUnits 贪心聚块：单块超限的单元先硬切，再用 sep 连接不超 maxLen
func packUnits(units []string, sep string, maxLen int) []string {
	var chunks []string
	var cur []string
	curLen := 0
	flush := func() {
		if len(cur) > 0 {
			chunks = append(chunks, strings.Join(cur, sep))
			cur = nil
			curLen = 0
		}
	}
	for _, u := range units {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		for _, piece := range hardCut(u, maxLen) {
			need := len([]rune(piece))
			if need > maxLen { // 理论不可达（hardCut 已保证）
				continue
			}
			add := need
			if len(cur) > 0 {
				add += len(sep)
			}
			if curLen+add > maxLen {
				flush()
			}
			cur = append(cur, piece)
			curLen += add
		}
	}
	flush()
	return chunks
}

// ChunkText 文本切块主入口（纯函数，测试锚点）
func ChunkText(text string, opts ChunkOptions) []Chunk {
	maxLen := opts.MaxLen
	if maxLen <= 0 {
		maxLen = chunkDefaultMaxLen
	}
	overlap := opts.Overlap
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= maxLen {
		overlap = maxLen / 2
	}
	paragraphs := strings.Split(text, "\n\n")
	var units []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len([]rune(p)) <= maxLen {
			units = append(units, p) // 段落整体作为聚块单元
		} else {
			units = append(units, splitSentences(p)...) // 段落超限 → 句子单元
		}
	}
	raw := packUnits(units, "\n\n", maxLen)

	var out []Chunk
	for i, c := range raw {
		if i > 0 && overlap > 0 {
			prev := raw[i-1]
			pr := []rune(prev)
			if len(pr) > overlap {
				pr = pr[len(pr)-overlap:]
			}
			tail := string(pr)
			if len([]rune(tail))+len([]rune(c)) <= maxLen {
				c = tail + c
			}
		}
		out = append(out, Chunk{Index: i, Text: c})
	}
	return out
}
