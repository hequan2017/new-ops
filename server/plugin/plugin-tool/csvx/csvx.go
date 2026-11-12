// Package csvx 白泽插件共享 CSV 导出构造器（固化 I07 口径）。
//
// 口径冻结：字段内半角逗号转全角「，」防串列、换行折空格；首行表头 + 数据行，每行 \n 结尾。
// 收敛此前 fanhua/llmops/exam 各自手写的 4 份 CSV 拼接。
package csvx

import (
	"fmt"
	"strings"
)

// Builder CSV 行构造器（非并发安全，按单次导出使用）。
type Builder struct {
	sb strings.Builder
}

// New 写入表头行。
func New(header ...string) *Builder {
	b := &Builder{}
	b.sb.WriteString(strings.Join(header, ","))
	b.sb.WriteString("\n")
	return b
}

// Row 追加一行：各字段经 cleanField 清洗后以半角逗号连接。
func (b *Builder) Row(fields ...any) *Builder {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = cleanField(f)
	}
	b.sb.WriteString(strings.Join(parts, ","))
	b.sb.WriteString("\n")
	return b
}

// Rows 批量追加，行为同多次 Row。
func (b *Builder) Rows(rows [][]any) *Builder {
	for _, r := range rows {
		b.Row(r...)
	}
	return b
}

// String 输出全文。
func (b *Builder) String() string { return b.sb.String() }

// cleanField 单字段清洗：半角逗号→全角防串列、换行→空格（I07 口径）。
func cleanField(field any) string {
	s := fmt.Sprint(field)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, ",", "，")
	return s
}
