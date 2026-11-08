// Package service ziqiang：术语校验与差分隐私单测（纯函数，无 DB）
package service

import (
	"math"
	"math/rand"
	"regexp"
	"testing"
)

// TestCheckTerms 上下文条件与命中
func TestCheckTerms(t *testing.T) {
	// 带上下文规则：上下文出现才判定
	issues := CheckTerms("该发明专利的有效期50年，自申请日起算")
	if len(issues) != 1 || issues[0].Wrong != "发明专利的有效期50年" {
		t.Fatalf("应命中发明专利年限规则: %+v", issues)
	}
	// 同样表述但无「发明」上下文 → 不判定
	if issues := CheckTerms("论文写道：本成果的有效期50年"); len(issues) != 0 {
		t.Fatalf("无上下文不应误报: %+v", issues)
	}
	// 无上下文规则直接命中
	issues2 := CheckTerms("此开源协议无版权，可随意商用")
	if len(issues2) != 1 || issues2[0].Correct == "" {
		t.Fatalf("应命中开源协议规则: %+v", issues2)
	}
	// 多条命中
	issues3 := CheckTerms("专利著作权与商标权著作权常被混淆")
	if len(issues3) != 2 {
		t.Fatalf("应命中 2 条: %+v", issues3)
	}
}

// TestLaplaceSample 采样函数数学性质
func TestLaplaceSample(t *testing.T) {
	scale := 2.0
	if laplaceSample(0.5, scale) != 0 {
		t.Fatal("u=0.5 应零噪声")
	}
	half := laplaceSample(0.25, scale) // scale*ln(0.5)/2 = -scale*ln2/2
	if math.Abs(half-(-scale*math.Log(2)/2)) > 1e-9 {
		t.Fatalf("u=0.25 应为 -scale·ln2/2: %v", half)
	}
	threeQ := laplaceSample(0.75, scale)
	if math.Abs(threeQ-scale*math.Log(2)/2) > 1e-9 {
		t.Fatalf("u=0.75 应为 +scale·ln2/2: %v", threeQ)
	}
	if laplaceSample(0.5, 0) != 0 {
		t.Fatal("scale=0 应零噪声")
	}
}

// TestDpMaskNumbers 数值扰动：有噪声、保留两位、非数值原样
func TestDpMaskNumbers(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	text := "本专利许可费 100000 元，市场占比 12.5%，共 3 家竞品。"
	masked, err := DpMaskNumbers(text, 1000, 1.0, rng)
	if err != nil {
		t.Fatalf("脱敏不应报错: %v", err)
	}
	nums := regexp.MustCompile(`\d+(\.\d+)?`).FindAllString(masked, -1)
	if len(nums) != 3 {
		t.Fatalf("仍应 3 个数值: %v", nums)
	}
	for _, n := range nums {
		if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(n) {
			t.Fatalf("数值应保留两位小数: %s", n)
		}
	}
	if masked == text {
		t.Fatal("带噪声后文本应有变化")
	}
	// 参数非法
	if _, err := DpMaskNumbers(text, 0, 0, rng); err == nil {
		t.Fatal("非法参数应报错")
	}
}
