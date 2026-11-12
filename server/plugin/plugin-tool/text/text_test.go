package text

import "testing"

func TestTruncate(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"短文本不截断", "abc", 64, "abc"},
		{"等长不截断", "abcdef", 6, "abcdef"},
		{"英文截断加省略号", "abcdefghij", 5, "abcde…"},
		{"中文按 rune 截", "白泽通晓万物之情", 2, "白泽…"},
		{"中英混排", "op白泽平台v3", 4, "op白泽…"},
		{"n 为零不截断", "abc", 0, "abc"},
		{"n 为负不截断", "abc", -1, "abc"},
		{"空串", "", 8, ""},
		{"emoji 按 rune 不切半个", "👍👍👍👍", 2, "👍👍…"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Truncate(tc.in, tc.n); got != tc.want {
				t.Fatalf("Truncate(%q,%d)=%q want %q", tc.in, tc.n, got, tc.want)
			}
		})
	}
}

func TestBigrams(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"中文四字", "白泽平台", 3},
		{"单字空集", "白", 0},
		{"空串空集", "", 0},
		{"拉丁含空格对", "ai ops", 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(Bigrams(tc.in)); got != tc.want {
				t.Fatalf("Bigrams(%q) len=%d want %d", tc.in, got, tc.want)
			}
		})
	}
	// 集合语义：重复 bigram 去重（aaaa → aa×3 位置，集合去重为 1？否：位置 0aa,1aa,2aa 均为「aa」→ 集合 1 个）
	if got := len(Bigrams("aaaa")); got != 1 {
		t.Fatalf("重复 bigram 应去重：len=%d want 1", got)
	}
}

func TestScore(t *testing.T) {
	// 「GPU防超卖」bigram 集：GP,PU,U防,防超,超卖
	qSet := Bigrams("GPU防超卖")
	cases := []struct {
		name   string
		target string
		want   int
	}{
		{"全文命中", "GPU防超卖引擎", 5},
		{"部分命中", "节点防超卖", 2},
		{"零命中", "磁盘告警", 0},
		{"空目标", "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Score(qSet, tc.target); got != tc.want {
				t.Fatalf("Score=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestRankBigrams(t *testing.T) {
	type doc struct {
		ID    uint
		Title string
	}
	docs := []doc{
		{1, "GPU 防超卖引擎设计"}, // GP,PU,防超,超卖 = 4
		{2, "GPU 规格与节点台账"}, // GP,PU = 2
		{3, "磁盘监控告警"},      // 0
		{4, "GPU防超卖与节点联动"}, // GP,PU,U防,防超,超卖 = 5
	}
	hits := RankBigrams(docs, func(d doc) string { return d.Title }, "GPU防超卖", 0)
	if len(hits) != 3 {
		t.Fatalf("命中数=%d want 3", len(hits))
	}
	if hits[0].Item.ID != 4 || hits[0].Score != 5 {
		t.Fatalf("首项应为 doc4 分 5，实际 ID=%d 分=%d", hits[0].Item.ID, hits[0].Score)
	}
	if hits[1].Item.ID != 1 || hits[1].Score != 4 {
		t.Fatalf("次项应为 doc1 分 4，实际 ID=%d 分=%d", hits[1].Item.ID, hits[1].Score)
	}
	if hits[2].Item.ID != 2 || hits[2].Score != 2 {
		t.Fatalf("末项应为 doc2 分 2，实际 ID=%d 分=%d", hits[2].Item.ID, hits[2].Score)
	}
	// topK 截断
	top1 := RankBigrams(docs, func(d doc) string { return d.Title }, "GPU", 1)
	if len(top1) != 1 {
		t.Fatalf("topK 截断失效：len=%d", len(top1))
	}
	// 稳定序：同分保持原序
	same := []doc{{9, "GPU"}, {5, "GPU"}}
	tied := RankBigrams(same, func(d doc) string { return d.Title }, "GPU", 0)
	if tied[0].Item.ID != 9 || tied[1].Item.ID != 5 {
		t.Fatal("同分未保持稳定原序")
	}
	// 零分过滤
	none := RankBigrams(docs, func(d doc) string { return d.Title }, "区块链", 0)
	if len(none) != 0 {
		t.Fatalf("零命中应过滤：len=%d", len(none))
	}
}
