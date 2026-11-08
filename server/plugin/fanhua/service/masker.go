// Package service fanhua：正则脱敏器（手机号/身份证）
package service

import (
	"regexp"
	"strings"
)

var phoneRegexp = regexp.MustCompile(`1[3-9]\d{9}`)
var idRegexp = regexp.MustCompile(`\d{17}[\dXx]`)

// masker 正则脱敏：保留前 keepHead 后 keepTail，中间以 * 填充
type masker struct {
	re        *regexp.Regexp
	keepHead  int
	keepTail  int
}

// mask 返回脱敏文本与命中次数
func (m masker) mask(s string) (string, int) {
	count := 0
	out := m.re.ReplaceAllStringFunc(s, func(match string) string {
		count++
		r := []rune(match)
		head := string(r[:m.keepHead])
		tail := ""
		if m.keepTail > 0 {
			tail = string(r[len(r)-m.keepTail:])
		}
		return head + strings.Repeat("*", len(r)-m.keepHead-m.keepTail) + tail
	})
	return out, count
}
