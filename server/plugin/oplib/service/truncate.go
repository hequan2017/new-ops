// Package service oplib：截断助手（错误摘要/长文本留痕）
package service

// truncateSample 截断到 64 rune + 省略号
func truncateSample(s string) string {
	r := []rune(s)
	if len(r) <= 64 {
		return s
	}
	return string(r[:64]) + "…"
}
