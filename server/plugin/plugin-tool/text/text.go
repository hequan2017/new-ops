// Package text 白泽插件共享文本助手：rune 安全截断、bigram 集合与检索打分排序。
//
// 收敛此前 10+ 份 truncate 变体与 6 个插件各自实现的 bigram+topK 检索。
// 口径冻结：截断超长以「…」收尾（同 exam/oplib truncateSample）；
// bigram 为 rune 级相邻二元组集合（同 eduadmin/ziqiang）。
package text

import "sort"

// Truncate 按 rune 截断到 n 个字符，超出以「…」收尾；n<=0 不截断。
func Truncate(s string, n int) string {
	if n <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// Bigrams rune 级 bigram 集合；单字符文本退化为空集。
func Bigrams(text string) map[string]struct{} {
	runes := []rune(text)
	set := make(map[string]struct{}, len(runes))
	for i := 0; i+1 < len(runes); i++ {
		set[string(runes[i:i+2])] = struct{}{}
	}
	return set
}

// Score 查询 bigram 集与目标文本的交集计数（检索打分公共口径）。
func Score(querySet map[string]struct{}, target string) int {
	targetSet := Bigrams(target)
	score := 0
	for bg := range querySet {
		if _, ok := targetSet[bg]; ok {
			score++
		}
	}
	return score
}

// Hit 检索命中项：原始条目 + bigram 命中分。
type Hit[T any] struct {
	Item  T
	Score int
}

// RankBigrams bigram 命中检索：以 querySet 对每条 textOf(item) 计分，
// 仅保留 score>0，按分数降序稳定排序（同分保持原序），topK<=0 不限量。
// 需要额外同分决胜（如按 ID 升序）的插件可基于 Bigrams+Score 自组比较器。
func RankBigrams[T any](list []T, textOf func(T) string, query string, topK int) []Hit[T] {
	qSet := Bigrams(query)
	hits := make([]Hit[T], 0, len(list))
	for _, item := range list {
		if score := Score(qSet, textOf(item)); score > 0 {
			hits = append(hits, Hit[T]{Item: item, Score: score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if topK > 0 && len(hits) > topK {
		hits = hits[:topK]
	}
	return hits
}
