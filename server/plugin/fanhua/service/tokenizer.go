// Package service fanhua：沪语分词器（需求 3.4-9 基础数据分词）
// 口径：专用词典（侬/嗲/弄堂/滑稽戏… 100+ 词）最大正向匹配；
// 词典未命中的连续汉字 run 以 bigram 两两成词兜底（奇数尾单字）；非汉字按符号/空白切分、
// 拉丁/数字连串成词。纯函数无 DB，词频统计 API。
package service

import (
	"sort"
	"unicode"
)

// shDictionary 沪语专用词典（含单字与多字词）
var shDictionary = map[string]bool{
	"侬": true, "嗲": true, "搿": true, "覅": true, "呒": true, "汏": true, "伊": true,
	"阿拉": true, "侬好": true, "侬好伐": true, "弄堂": true, "滑稽戏": true, "小囡": true,
	"囡儿": true, "伲子": true, "事体": true, "今朝": true, "明朝": true, "昨日子": true,
	"交关": true, "老灵": true, "灵额": true, "呒没": true, "晓得": true, "适意": true,
	"夜到": true, "早上向": true, "中浪向": true, "夜里向": true, "灶披间": true,
	"亭子间": true, "客堂间": true, "老虎窗": true, "老虎灶": true, "汤婆子": true,
	"石库门": true, "天井": true, "晒台": true, "山墙": true, "洋房": true,
	"迭个": true, "伊个": true, "迭桩": true, "啥物事": true, "为啥": true, "哪能": true,
	"老早": true, "面孔": true, "眼睛": true, "脚踏车": true, "黄鱼车": true,
	"三轮车": true, "电车": true, "马路": true, "兜马路": true, "荡马路": true,
	"白相": true, "白相相": true, "困觉": true, "吃饭": true, "吃茶": true,
	"吃酒": true, "吃烟": true, "买小菜": true, "小菜": true, "烧饭": true,
	"烧菜": true, "汏衣裳": true, "衣裳": true, "揩台子": true, "扫地": true,
	"掼垃圾": true, "垃圾": true, "生病": true, "吃药": true, "打针": true,
	"看医生": true, "挂号": true, "体温": true, "先生": true, "学生": true,
	"同学": true, "老师": true, "朋友": true, "亲眷": true, "爷娘": true,
	"爷叔": true, "姆妈": true, "阿婆": true, "好婆": true, "阿哥": true,
	"兄弟": true, "姐妹": true, "女婿": true, "新妇": true, "小辈": true,
	"长辈": true, "混堂": true, "书场": true, "戏院": true, "电影院": true,
	"南货店": true, "烟纸店": true, "布店": true, "米行": true, "剃头店": true,
	"排车队": true, "排队": true, "摆渡": true, "赶头班": true, "价钿": true,
	"钿": true, "结账": true, "钞票": true, "角子": true, "一分钿": true,
	"闹猛": true, "尴尬": true, "拎勿清": true, "扎劲": true, "煞根": true,
	"魂灵头": true, "牵丝攀藤": true, "笃悠悠": true, "慢悠悠": true, "慢慢交": true,
}

// maxWordLen 词典最长词长（init 期计算）
var maxWordLen = 1

func init() {
	for w := range shDictionary {
		if l := len([]rune(w)); l > maxWordLen {
			maxWordLen = l
		}
	}
}

// Token 分词结果项
type Token struct {
	Word   string `json:"word"`
	IsDict bool   `json:"isDict"`
}

// isCJK 常用汉字区
func isCJK(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

// Segment 沪语分词（最大正向匹配 + OOV bigram 兜底；纯函数，测试锚点）
func Segment(text string) []Token {
	runes := []rune(text)
	tokens := make([]Token, 0, len(runes)/2)
	pending := make([]rune, 0, 8)
	flush := func() { // 连续未登录字 bigram 兜底
		for k := 0; k < len(pending); k += 2 {
			if k+1 < len(pending) {
				tokens = append(tokens, Token{Word: string(pending[k : k+2]), IsDict: false})
			} else {
				tokens = append(tokens, Token{Word: string(pending[k]), IsDict: false})
			}
		}
		pending = pending[:0]
	}
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case isCJK(r):
			matched := ""
			for l := maxWordLen; l >= 1; l-- {
				if i+l <= len(runes) {
					w := string(runes[i : i+l])
					if shDictionary[w] {
						matched = w
						break
					}
				}
			}
			if matched != "" {
				flush()
				tokens = append(tokens, Token{Word: matched, IsDict: true})
				i += len([]rune(matched))
			} else {
				pending = append(pending, r)
				i++
			}
		case unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			flush()
			i++
		default: // 拉丁/数字连串成词
			flush()
			j := i
			for j < len(runes) && !isCJK(runes[j]) && !unicode.IsSpace(runes[j]) &&
				!unicode.IsPunct(runes[j]) && !unicode.IsSymbol(runes[j]) {
				j++
			}
			tokens = append(tokens, Token{Word: string(runes[i:j]), IsDict: false})
			i = j
		}
	}
	flush()
	return tokens
}

// FreqItem 词频项
type FreqItem struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// WordFreq 词频统计（降序）
func WordFreq(text string) []FreqItem {
	counts := map[string]int{}
	for _, t := range Segment(text) {
		counts[t.Word]++
	}
	out := make([]FreqItem, 0, len(counts))
	for w, c := range counts {
		out = append(out, FreqItem{Word: w, Count: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}
