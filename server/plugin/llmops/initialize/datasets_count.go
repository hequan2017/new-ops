// Package initialize llmops：种子样本计数（与 service 口径一致，避免跨包循环）
package initialize

import "strings"

// countSeedSamples jsonl 非空行计数
func countSeedSamples(content string) int {
	n := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
