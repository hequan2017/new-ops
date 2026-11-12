// Package service policyreg：制度分词与属性（需求 3.8-4）
// 口径：制度术语词典（30 词）最大正向匹配（对齐 N05/O05 口径），OOV 汉字单字；
// 属性标注=分类标签（五类关键词多者胜）+生效日期正则抽取（自/自发布之日起 +
// yyyy年M月d日 优先首处）。纯函数可测。
package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// 错误码（policyreg 段 3000-3099）
const ErrCodePrTermParam = 3003

// policyDictionary 制度术语词典
var policyDictionary = []string{
	"学位评定委员会", "学术委员会", "教职工代表大会", "校长办公会", "党委常委会",
	"招生简章", "培养方案", "学分绩点", "学分互认", "双学位", "辅修专业", "转专业",
	"师德师风", "职称评审", "岗位聘用", "绩效工资", "编制核定",
	"科研经费", "横向课题", "纵向课题", "成果转化", "知识产权",
	"预算决算", "政府采购", "招标投标", "国有资产", "审计监督",
	"学生申诉", "学籍管理", "奖助学金", "双一流", "产教融合",
}

// prMaxLen 词典最长词长（init 计算）
var prMaxLen = 0

func init() {
	for _, w := range policyDictionary {
		if l := len([]rune(w)); l > prMaxLen {
			prMaxLen = l
		}
	}
}

// PolicySegment 制度分词（纯函数，测试锚点）：词典最大正向匹配，OOV 汉字单字，
// 拉丁/数字连串成词，空白标点跳过。
func PolicySegment(text string) []string {
	runes := []rune(text)
	out := make([]string, 0, len(runes)/2)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			i++
		case isPrPunct(r):
			i++
		case isPrLatin(r):
			j := i
			for j < len(runes) && (isPrLatin(runes[j]) || (runes[j] == '/' || runes[j] == '.') && j+1 < len(runes) && isPrLatin(runes[j+1])) {
				j++
			}
			out = append(out, string(runes[i:j]))
			i = j
		default:
			matched := ""
			max := prMaxLen
			if len(runes)-i < max {
				max = len(runes) - i
			}
			for l := max; l >= 2; l-- {
				cand := string(runes[i : i+l])
				for _, w := range policyDictionary {
					if w == cand {
						matched = cand
						break
					}
				}
				if matched != "" {
					break
				}
			}
			if matched != "" {
				out = append(out, matched)
				i += len([]rune(matched))
			} else {
				out = append(out, string(r))
				i++
			}
		}
	}
	return out
}

func isPrPunct(r rune) bool {
	return strings.ContainsRune("，。、；：？！「」『』（）《》〈〉【】…—·,. ;:!?", r)
}

func isPrLatin(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
}

// effectiveDatePattern 生效日期正则
var effectiveDatePattern = regexp.MustCompile(`(?:自|自发布之日起|自公布之日起)?[ \t]*([0-9]{4})年([0-9]{1,2})月([0-9]{1,2})日`)

// ExtractEffectiveDate 生效日期抽取（纯函数，测试锚点）：返回 yyyy-MM-dd 或空
func ExtractEffectiveDate(text string) string {
	m := effectiveDatePattern.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%s-%02s-%02s", m[1], m[2], m[3])
}

// policyLabelKeywords 分类标签关键词
var policyLabelKeywords = map[string][]string{
	"教学": {"招生", "培养", "学分", "学籍", "专业", "课程"},
	"科研": {"科研", "课题", "成果转化", "知识产权", "实验"},
	"人事": {"职称", "聘用", "绩效", "师德", "编制", "人事"},
	"财务": {"预算", "采购", "招标", "财务", "审计"},
	"资产": {"国有资产", "资产", "设备"},
}

// ClassifyPolicyLabel 分类标签（纯函数，测试锚点）：多者胜平局按白名单序
func ClassifyPolicyLabel(text string) string {
	best, bestCnt := "综合", 0
	for _, label := range []string{"教学", "科研", "人事", "财务", "资产"} {
		cnt := 0
		for _, kw := range policyLabelKeywords[label] {
			if strings.Contains(text, kw) {
				cnt++
			}
		}
		if cnt > bestCnt {
			best, bestCnt = label, cnt
		}
	}
	return best
}

// TagPolicyTerms 分词+属性标注（按 docId 幂等覆盖）
func (s *PolicyService) TagPolicyTerms(docId uint) (*model.PolicyTermTag, error) {
	doc, err := s.repo.FindDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.New(ErrCodePrTermParam, "制度不存在")
	}
	text := doc.RawText
	if strings.TrimSpace(text) == "" {
		return nil, errx.New(ErrCodePrTermParam, "二进制编目文档无可标注文本")
	}
	tokens := PolicySegment(text)
	b, _ := json.Marshal(tokens)
	tag := &model.PolicyTermTag{
		DocId:         docId,
		Tokens:        string(b),
		CategoryLabel: ClassifyPolicyLabel(doc.Name + " " + text),
		EffectiveDate: ExtractEffectiveDate(text),
	}
	if exist, err := s.repo.FindTermTagByDoc(s.dbh(), docId); err == nil {
		if err := s.repo.UpdateTermTagColumns(s.dbh(), exist.ID, map[string]any{
			"tokens": tag.Tokens, "category_label": tag.CategoryLabel, "effective_date": tag.EffectiveDate,
		}); err != nil {
			return nil, errx.Wrap(ErrCodePrStore, "更新标注失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateTermTag(s.dbh(), tag); err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "创建标注失败", err)
	}
	return tag, nil
}

// GetPolicyTermTags 标注列表
func (s *PolicyService) GetPolicyTermTags(docId uint) ([]model.PolicyTermTag, error) {
	list, err := s.repo.TermTagListByDoc(s.dbh(), docId)
	if err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "查询标注列表失败", err)
	}
	return list, nil
}

// labelCount 分类计数
type labelCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// termLabelStats 标签聚合（纯函数，测试锚点）：六类固定序，零计数保留
func termLabelStats(tags []model.PolicyTermTag) []labelCount {
	counts := map[string]int64{}
	for _, t := range tags {
		if t.CategoryLabel != "" {
			counts[t.CategoryLabel]++
		}
	}
	out := make([]labelCount, 0, len(PolicyCategories))
	for _, label := range PolicyCategories {
		out = append(out, labelCount{Label: label, Count: counts[label]})
	}
	return out
}

// GetPolicyTermStats 分类标签统计
func (s *PolicyService) GetPolicyTermStats() ([]labelCount, error) {
	tags, err := s.repo.TermTagsAll(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodePrStore, "查询标注统计失败", err)
	}
	return termLabelStats(tags), nil
}
