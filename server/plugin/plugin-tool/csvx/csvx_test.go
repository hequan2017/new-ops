package csvx

import "testing"

func TestNewHeader(t *testing.T) {
	if got := New("来源", "标题", "语种", "内容").String(); got != "来源,标题,语种,内容\n" {
		t.Fatalf("表头口径不符: %q", got)
	}
	if got := New().String(); got != "\n" {
		t.Fatalf("空表头应为空行: %q", got)
	}
}

func TestRowCleaning(t *testing.T) {
	cases := []struct {
		name string
		in   []any
		want string
	}{
		{"半角逗号转全角", []any{"沪语", "普,通话"}, "H\n沪语,普，通话\n"},
		{"换行折空格", []any{"a\nb", "c"}, "H\na b,c\n"},
		{"逗号与换行先后", []any{"a,b\nc,d"}, "H\na，b c，d\n"},
		{"整数原样", []any{1, 23}, "H\n1,23\n"},
		{"无清洗需要的字段不动", []any{"白泽", "GPU"}, "H\n白泽,GPU\n"},
		{"空字段占位", []any{""}, "H\n\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := New("H").Row(tc.in...).String(); got != tc.want {
				t.Fatalf("got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestFullExport(t *testing.T) {
	// 对齐 fanhua parallelsToCSV 口径：沪/普/英/来源 表头 + 全量清洗
	b := New("沪语", "普通话", "英语", "来源").
		Row("侬好", "你好", "hello", "教材").
		Row("谢谢侬", "谢,谢你", "thanks\na lot", "词,书")
	want := "沪语,普通话,英语,来源\n" +
		"侬好,你好,hello,教材\n" +
		"谢谢侬,谢，谢你,thanks a lot,词，书\n"
	if got := b.String(); got != want {
		t.Fatalf("全文口径不符:\ngot=%q\nwant=%q", got, want)
	}
}

func TestRowsBatch(t *testing.T) {
	b := New("A", "B").Rows([][]any{{1, "x"}, {2, "y, z"}})
	want := "A,B\n1,x\n2,y， z\n"
	if got := b.String(); got != want {
		t.Fatalf("批量行口径不符: %q", got)
	}
}
