// Package service llmops 切块服务单测（段落聚块/句边界/硬切/重叠/默认参数/rune 安全）
package service

import (
	"strings"
	"testing"
)

func runeLen(s string) int { return len([]rune(s)) }

func TestChunkText_SingleParagraph(t *testing.T) {
	chunks := ChunkText("白泽统一运维平台。", ChunkOptions{})
	if len(chunks) != 1 || chunks[0].Index != 0 || chunks[0].Text != "白泽统一运维平台。" {
		t.Fatalf("短文本应单块: %+v", chunks)
	}
}

// TestChunkText_ParagraphPack 段落贪心聚块（含空段跳过）
func TestChunkText_ParagraphPack(t *testing.T) {
	text := "第一段。\n\n第二段。\n\n\n\n第三段。\n\n\n\n\n"
	chunks := ChunkText(text, ChunkOptions{MaxLen: 100})
	if len(chunks) != 1 {
		t.Fatalf("三短段应聚为一块: %+v", chunks)
	}
	for _, want := range []string{"第一段。", "第二段。", "第三段。"} {
		if !strings.Contains(chunks[0].Text, want) {
			t.Fatalf("缺段落 %q: %q", want, chunks[0].Text)
		}
	}
	if strings.Contains(chunks[0].Text, "\n\n\n") {
		t.Fatalf("空段应被清理: %q", chunks[0].Text)
	}
}

// TestChunkText_SentenceSplit 段落超限 → 句边界切
func TestChunkText_SentenceSplit(t *testing.T) {
	sent := strings.Repeat("这是第九句。", 1) // 6 rune
	text := strings.Repeat(sent, 6)       // 段落 36 rune
	chunks := ChunkText(text, ChunkOptions{MaxLen: 15, Overlap: 0})
	if len(chunks) < 3 {
		t.Fatalf("应切多块: %d", len(chunks))
	}
	for _, c := range chunks {
		if runeLen(c.Text) > 15 {
			t.Fatalf("块超限: %d %q", runeLen(c.Text), c.Text)
		}
	}
	// 边界不应切在句中（每块以句号收尾或为最后一块）
	for i, c := range chunks {
		if i < len(chunks)-1 && !strings.HasSuffix(c.Text, "。") {
			t.Fatalf("句边界切分应保留完整句: %q", c.Text)
		}
	}
}

// TestChunkText_HardCut 单句超限硬切
func TestChunkText_HardCut(t *testing.T) {
	long := strings.Repeat("字", 30)
	chunks := ChunkText(long, ChunkOptions{MaxLen: 10})
	if len(chunks) != 3 {
		t.Fatalf("30 rune 单句按 10 应切 3 块: %d", len(chunks))
	}
	if chunks[0].Text != strings.Repeat("字", 10) || chunks[2].Text != strings.Repeat("字", 10) {
		t.Fatalf("硬切内容异常: %+v", chunks)
	}
}

// TestChunkText_Overlap 聚块边界携带上块尾部重叠且不超 MaxLen（尽力而为：放不下时跳过，内容完整性优先）
func TestChunkText_Overlap(t *testing.T) {
	// 四个 8 rune 段落，MaxLen 20：块=p1+p2 / tail+p3+p4 → 重叠可放下
	text := "一二三四五六七八。\n\n一二三四五六七八。\n\n一二三四五六七八。\n\n一二三四五六七八。"
	chunks := ChunkText(text, ChunkOptions{MaxLen: 20, Overlap: 4})
	if len(chunks) < 2 {
		t.Fatalf("应切多块: %+v", chunks)
	}
	prev := chunks[0].Text
	tail := []rune(prev)
	if len(tail) > 4 {
		tail = tail[len(tail)-4:]
	}
	if !strings.HasPrefix(chunks[1].Text, string(tail)) {
		t.Fatalf("块 1 应以上块尾部开头: %q vs %q", chunks[1].Text, string(tail))
	}
	for _, c := range chunks {
		if runeLen(c.Text) > 20 {
			t.Fatalf("重叠后仍不得超限: %d %q", runeLen(c.Text), c.Text)
		}
	}
	// 硬切流场景：重叠放不下时跳过，但内容不丢（各块拼接去重叠后覆盖原文）
	stream := strings.Repeat("甲乙丙丁戊己庚辛壬癸", 3)
	hard := ChunkText(stream, ChunkOptions{MaxLen: 12, Overlap: 4})
	joined := ""
	for _, c := range hard {
		joined += c.Text
		if runeLen(c.Text) > 12 {
			t.Fatalf("硬切块超限: %d", runeLen(c.Text))
		}
	}
	for _, r := range []rune(stream) {
		if !strings.ContainsRune(joined, r) {
			t.Fatalf("硬切丢字符: %q", string(r))
		}
	}
}

// TestChunkText_Defaults 默认参数与重叠收敛
func TestChunkText_Defaults(t *testing.T) {
	// 默认 MaxLen 512：500 rune 单段单块
	if chunks := ChunkText(strings.Repeat("测", 500), ChunkOptions{}); len(chunks) != 1 {
		t.Fatalf("500 rune 应单块: %d", len(chunks))
	}
	// Overlap ≥ MaxLen 收敛为 MaxLen/2（不越限不 panic）
	chunks := ChunkText(strings.Repeat("测", 2000), ChunkOptions{MaxLen: 100, Overlap: 500})
	if len(chunks) == 0 {
		t.Fatal("不应为空")
	}
	for _, c := range chunks {
		if runeLen(c.Text) > 100 {
			t.Fatalf("收敛后仍超限: %d", runeLen(c.Text))
		}
	}
}

// TestSplitSentences 句边界符覆盖中英文
func TestSplitSentences(t *testing.T) {
	sentences := splitSentences("第一句。Second! 第三；last?\n第五句")
	if len(sentences) != 5 {
		t.Fatalf("应切 5 句: %v", sentences)
	}
}
