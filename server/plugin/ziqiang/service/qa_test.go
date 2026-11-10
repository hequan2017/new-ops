// Package service ziqiang：知识产权问答单测（纯函数，无 DB）
package service

import (
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/global"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// gvaModelOf 构造带 ID 的测试模型
func gvaModelOf(id uint) global.GVA_MODEL {
	return global.GVA_MODEL{ID: id, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

// TestTextContainment bigram 包含率：全命中 1、部分命中比例、无关 0
func TestTextContainment(t *testing.T) {
	if v := textContainment("神经网络训练方法", "一种神经网络训练方法及装置"); v != 1 {
		t.Fatalf("全包含应为 1: %v", v)
	}
	v := textContainment("神经网络训练方法", "卷积神经网络结构研究")
	if v >= 0.5 || v <= 0 {
		t.Fatalf("部分命中应落 (0,0.5): %v", v)
	}
	if v := textContainment("量子加密通信", "食品加工配方专利"); v != 0 {
		t.Fatalf("无关应为 0: %v", v)
	}
	if v := textContainment("", "任意文档"); v != 0 {
		t.Fatalf("空查询应为 0: %v", v)
	}
}

// TestFuseScores 融合分：双维 0.6/0.4 加权、单维退化
func TestFuseScores(t *testing.T) {
	if v := fuseScores(0.9, 0.7, true); v != 0.82 {
		t.Fatalf("0.6*0.9+0.4*0.7 应为 0.82: %v", v)
	}
	if v := fuseScores(0.7, 0.99, false); v != 0.7 {
		t.Fatalf("无图像维度应退化为文本分: %v", v)
	}
	if v := fuseScores(0.5, 1.0, true); v != 0.7 {
		t.Fatalf("0.6*0.5+0.4*1.0 应为 0.7: %v", v)
	}
}

// imgTestHash 构造仅低 k 位为 1 的测试哈希
func imgTestHash(ones int) uint64 {
	var h uint64
	for i := 0; i < ones; i++ {
		h |= 1 << uint(i)
	}
	return h
}

// TestRankPatentsFusion 融合排序：图文双维命中 > 仅文本命中 > 低分
func TestRankPatentsFusion(t *testing.T) {
	patents := []model.ZqTextCorpus{
		{GVA_MODEL: gvaModelOf(1), Title: "一种神经网络训练方法", Content: "涉及神经网络训练方法与装置，采用反向传播与梯度下降。", LegalStatus: "有效", ImageCorpusId: 100},
		{GVA_MODEL: gvaModelOf(2), Title: "神经网络芯片架构", Content: "神经网络加速芯片，涉及训练方法电路。", LegalStatus: "审查中"},
		{GVA_MODEL: gvaModelOf(3), Title: "食品加工工艺", Content: "一种糕点烘焙配方与加工流程。", LegalStatus: "失效"},
	}
	imgs := map[uint]uint64{100: imgTestHash(4)} // 与查询哈希仅 4 位差异→高相似
	matches := rankPatents(patents, "神经网络训练方法", imgTestHash(0), true, imgs)
	if len(matches) != 3 {
		t.Fatalf("应返回全部候选: %d", len(matches))
	}
	if matches[0].CorpusId != 1 {
		t.Fatalf("图文双维命中应居首: %+v", matches[0])
	}
	if !matches[0].HasImage || matches[0].ImageScore <= 0.9 {
		t.Fatalf("首位应带高图像相似度: %+v", matches[0])
	}
	if matches[0].FusedScore <= matches[1].FusedScore {
		t.Fatalf("融合分应严格降序: %+v", matches)
	}
	if matches[2].Title != "食品加工工艺" {
		t.Fatalf("无关专利应垫底: %+v", matches[2])
	}
	// 无查询图像时：退化为文本分排序，HasImage 全 false
	textOnly := rankPatents(patents, "神经网络训练方法", 0, false, imgs)
	if textOnly[0].CorpusId != 1 || textOnly[0].HasImage {
		t.Fatalf("纯文本口径首位应为语料 1 且无图像维度: %+v", textOnly[0])
	}
}

// TestAssessRiskLevels 三档风险判定
func TestAssessRiskLevels(t *testing.T) {
	// 高：融合分 ≥0.85
	highMatches := []qaMatch{{CorpusId: 1, Title: "专利A", TextScore: 0.9, HasImage: true, ImageScore: 0.9, FusedScore: fuseScores(0.9, 0.9, true)}}
	level, score, _, _ := assessRisk(highMatches, true)
	if level != "高" || score < 0.85 {
		t.Fatalf("双高分应判高: %s %v", level, score)
	}
	// 高：文本+图像各自越阈（融合分未达 0.85 也可）
	mid := fuseScores(0.85, 0.9, true) // 0.87? = 0.51+0.36=0.87 → 高；改用低融合组合
	dualHit := []qaMatch{{CorpusId: 2, Title: "专利B", TextScore: 0.85, HasImage: true, ImageScore: 0.9, FusedScore: mid}}
	level, _, _, evidence := assessRisk(dualHit, true)
	if level != "高" || len(evidence) != 2 {
		t.Fatalf("双维命中应判高且两条证据: %s %v", level, evidence)
	}
	// 中：单文本命中
	midMatches := []qaMatch{{CorpusId: 3, Title: "专利C", TextScore: 0.82, FusedScore: 0.82}}
	if level, _, _, ev := assessRisk(midMatches, false); level != "中" || len(ev) != 1 {
		t.Fatalf("单文本命中应判中: %s %v", level, ev)
	}
	// 低：无命中
	if level, score, _, _ := assessRisk([]qaMatch{{CorpusId: 4, Title: "专利D", TextScore: 0.1, FusedScore: 0.1}}, false); level != "低" || score != 0.1 {
		t.Fatalf("低分应判低: %s %v", level, score)
	}
	// 空：库内无相似
	if level, score, _, _ := assessRisk(nil, false); level != "低" || score != 0 {
		t.Fatalf("空命中应判低 0 分: %s %v", level, score)
	}
}

// TestValidLegalStatus 法律状态白名单
func TestValidLegalStatus(t *testing.T) {
	for _, s := range LegalStatuses {
		if !ValidLegalStatus(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if ValidLegalStatus("已撤销") {
		t.Fatal("白名单外应拒绝")
	}
}
