// Package service ziqiang：法律术语校验与差分隐私脱敏（需求 3.5-4）
// 口径：混淆规则库（错误表述→修正建议，附上下文条件——上下文词未出现不判定，防误报）；
// 数值差分隐私=Laplace 机制（噪声=-sensitivity/ε·sign(u)·ln(1-2|u|)，注入随机源可测）。
package service

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"
)

// 错误码（ziqiang 段 2700-2799）
const ErrCodeTermParam = 2707

// termRule 术语混淆规则
type termRule struct {
	Wrong   string // 错误表述
	Correct string // 修正建议
	Context string // 上下文条件（非空时文本须含该词才判定）
	Hint    string // 说明
}

// termRules 混淆规则库（专利权/著作权/商标权/开源协议 等 11 条）
var termRules = []termRule{
	{"专利著作权", "专利权与著作权", "", "并列概念应分开表述，二者保护客体不同"},
	{"申请专利版权", "申请专利或登记版权", "", "专利需申请授权，版权（著作权）自创作完成即享有、可自愿登记"},
	{"著作权的专利申请", "专利的申请", "专利", "著作权不存在专利申请程序"},
	{"发明专利的有效期50年", "发明专利的有效期20年", "发明", "发明专利自申请日起 20 年"},
	{"发明专利有效期20年", "发明专利的有效期20年", "发明", "表述一致，确认起始日为申请日"},
	{"著作权有效期20年", "著作权保护期一般为作者终生及死后50年", "", "著作权保护期口径"},
	{"商标权著作权", "商标权与著作权", "", "并列概念应分开表述"},
	{"实用新型保护20年", "实用新型专利保护期10年", "", "实用新型自申请日起 10 年"},
	{"外观设计保护10年", "外观设计专利保护期15年", "", "2021 年专利法修改后外观设计为 15 年"},
	{"开源协议无版权", "开源协议仍有版权与许可证约束", "", "开源不等于放弃版权"},
	{"抄袭专利", "专利侵权", "侵权", "「抄袭」属著作权范畴，专利对应「侵权」"},
}

// TermIssue 术语问题
type TermIssue struct {
	Wrong   string `json:"wrong"`
	Correct string `json:"correct"`
	Hint    string `json:"hint"`
	Pos     int    `json:"pos"`
}

// CheckTerms 术语校验（纯函数，测试锚点）：上下文条件词未出现则跳过
func CheckTerms(text string) []TermIssue {
	var issues []TermIssue
	for _, r := range termRules {
		if r.Context != "" && !strings.Contains(text, r.Context) {
			continue
		}
		idx := strings.Index(text, r.Wrong)
		if idx < 0 {
			continue
		}
		issues = append(issues, TermIssue{Wrong: r.Wrong, Correct: r.Correct, Hint: r.Hint, Pos: idx})
	}
	return issues
}

// numPattern 数值模式（整数/小数）
var numPattern = regexp.MustCompile(`\d+(\.\d+)?`)

// laplaceSample Laplace 采样（纯函数，测试锚点）：u∈[0,1) → 噪声
func laplaceSample(u, scale float64) float64 {
	if scale <= 0 {
		return 0
	}
	if u < 0.5 {
		return scale * math.Log(2*u) / 2
	}
	return -scale * math.Log(2*(1-u)) / 2
}

// DpMaskNumbers 数值差分隐私脱敏（纯函数，测试锚点）：
// 文本中数值加 Laplace 噪声（sensitivity/ε 尺度），保留 2 位小数；商业秘密口径。
func DpMaskNumbers(text string, sensitivity, epsilon float64, rng *rand.Rand) (string, error) {
	if epsilon <= 0 || sensitivity <= 0 {
		return "", zqErr(ErrCodeTermParam, "sensitivity/epsilon 须为正数")
	}
	scale := sensitivity / epsilon
	out := numPattern.ReplaceAllStringFunc(text, func(m string) string {
		v := 0.0
		if _, err := fmt.Sscanf(m, "%g", &v); err != nil {
			return m
		}
		u := rng.Float64()
		noisy := v + laplaceSample(u, scale)
		return fmt.Sprintf("%.2f", noisy)
	})
	return out, nil
}
