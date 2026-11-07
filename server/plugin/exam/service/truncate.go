// Package service exam：截断助手
package service

// truncateSample 截断到 64 rune + 省略号
func truncateSample(s string) string {
	r := []rune(s)
	if len(r) <= 64 {
		return s
	}
	return string(r[:64]) + "…"
}
