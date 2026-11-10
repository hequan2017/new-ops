// Package service ziqiang：知识产权问答 API（需求 3.5-14）
// 口径：文本+图像混合输入 → 相似专利检索——文本维度=bigram 包含率（复用 O05
// bigramSet）、图像维度=查询 aHash 与专利关联图像语料 HashHex 的汉明相似度
// （复用 O03 HashSimilarity）；融合排序=0.6×文本+0.4×图像（单维度可得时退化为该
// 维度分）；法律状态字段（有效/审查中/失效/届满白名单）随命中输出；风险评估
// 高/中/低三档纯函数，输出报告 JSON（命中+阈值+证据+免责声明）。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const ErrCodeQaParam = 2718

// 融合权重与命中阈值（口径锚点）
const (
	QaTextWeight  = 0.6
	QaImageWeight = 0.4
	QaTextHit     = 0.8  // 文本维度命中阈（对齐 O05 抄袭默认阈值）
	QaImageHit    = 0.85 // 图像维度命中阈
	QaRiskHigh    = 0.85 // 融合分高风险阈
	QaRiskMid     = 0.60 // 融合分中风险阈
)

// LegalStatuses 法律状态白名单
var LegalStatuses = []string{"有效", "审查中", "失效", "届满"}

// ValidLegalStatus 法律状态白名单
func ValidLegalStatus(s string) bool {
	for _, v := range LegalStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// round4 四位小数（对齐 HashSimilarity 口径）
func round4(v float64) float64 {
	return float64(int(v*10000+0.5)) / 10000
}

// textContainment 文本维度分：查询 bigram 在文档中的包含率（纯函数，测试锚点）
func textContainment(query, doc string) float64 {
	qSet := bigramSet(query)
	if len(qSet) == 0 {
		return 0
	}
	dSet := bigramSet(doc)
	hit := 0
	for bg := range qSet {
		if dSet[bg] {
			hit++
		}
	}
	return round4(float64(hit) / float64(len(qSet)))
}

// fuseScores 融合分（纯函数，测试锚点）：双维度可得时 0.6×文本+0.4×图像，否则退化
func fuseScores(textScore, imageScore float64, hasImage bool) float64 {
	if !hasImage {
		return round4(textScore)
	}
	return round4(QaTextWeight*textScore + QaImageWeight*imageScore)
}

// qaMatch 相似专利命中项
type qaMatch struct {
	CorpusId    uint    `json:"corpusId"`
	Title       string  `json:"title"`
	Source      string  `json:"source"`
	LegalStatus string  `json:"legalStatus"`
	TextScore   float64 `json:"textScore"`
	ImageScore  float64 `json:"imageScore"`
	HasImage    bool    `json:"hasImage"`
	FusedScore  float64 `json:"fusedScore"`
}

// rankPatents 相似专利融合排序（纯函数，测试锚点）
// imgHashByCorpus：图像语料ID→aHash(uint64)；queryHash<0 表示未提供查询图像
func rankPatents(patents []model.ZqTextCorpus, query string, queryHash uint64, hasQueryHash bool, imgHashByCorpus map[uint]uint64) []qaMatch {
	matches := make([]qaMatch, 0, len(patents))
	for _, p := range patents {
		m := qaMatch{
			CorpusId:    p.ID,
			Title:       p.Title,
			Source:      p.Source,
			LegalStatus: p.LegalStatus,
			TextScore:   textContainment(query, p.Title+" "+p.Content),
		}
		if hasQueryHash && p.ImageCorpusId > 0 {
			if h, ok := imgHashByCorpus[p.ImageCorpusId]; ok {
				m.ImageScore = HashSimilarity(queryHash, h)
				m.HasImage = true
			}
		}
		m.FusedScore = fuseScores(m.TextScore, m.ImageScore, m.HasImage)
		matches = append(matches, m)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].FusedScore != matches[j].FusedScore {
			return matches[i].FusedScore > matches[j].FusedScore
		}
		return matches[i].CorpusId < matches[j].CorpusId
	})
	return matches
}

// qaReport 侵权风险评估报告
type qaReport struct {
	TextProvided  bool      `json:"textProvided"`
	ImageProvided bool      `json:"imageProvided"`
	Thresholds    map[string]float64 `json:"thresholds"`
	TopMatches    []qaMatch `json:"topMatches"`
	RiskLevel     string    `json:"riskLevel"`
	RiskScore     float64   `json:"riskScore"`
	Verdict       string    `json:"verdict"`
	Evidence      []string  `json:"evidence"`
	Disclaimer    string    `json:"disclaimer"`
}

// assessRisk 侵权风险评估（纯函数，测试锚点）：
// 高=融合分≥0.85 或文本+图像双维同时命中；中=融合分≥0.60 或单维命中；低=其余。
func assessRisk(matches []qaMatch, imageProvided bool) (level string, score float64, verdict string, evidence []string) {
	const (
		high = "高"
		mid  = "中"
		low  = "低"
	)
	if len(matches) == 0 {
		return low, 0, "库内无相似专利，未见明显侵权风险", nil
	}
	top := matches[0]
	imageBest := 0.0
	bestImgId := uint(0)
	for _, m := range matches {
		if m.HasImage && m.ImageScore > imageBest {
			imageBest = m.ImageScore
			bestImgId = m.CorpusId
		}
	}
	score = top.FusedScore
	if imageProvided && imageBest > score {
		score = imageBest
	}
	score = round4(score)

	textHit := top.TextScore >= QaTextHit
	imageHit := imageProvided && imageBest >= QaImageHit
	evidence = make([]string, 0)
	if textHit {
		evidence = append(evidence, fmt.Sprintf("文本维度：与《%s》bigram 包含率 %.2f ≥ %.2f", top.Title, top.TextScore, QaTextHit))
	}
	if imageHit {
		evidence = append(evidence, fmt.Sprintf("图像维度：与语料 #%d 关联图像相似度 %.2f ≥ %.2f", bestImgId, imageBest, QaImageHit))
	}

	switch {
	case top.FusedScore >= QaRiskHigh || (textHit && imageHit):
		level = high
	case top.FusedScore >= QaRiskMid || textHit || imageHit:
		level = mid
	default:
		level = low
	}
	switch level {
	case high:
		verdict = "与在先专利高度相似，建议立即停止疑似侵权使用并咨询执业律师"
	case mid:
		verdict = "存在部分相似要素，建议比对权利要求书并保留创作过程证据"
	default:
		verdict = "未达相似阈值，未发现明显侵权风险"
	}
	return level, score, verdict, evidence
}

// QaPatentSearch 知识产权问答：文本+图像混合输入 → 相似专利检索 + 风险评估报告
func (s *ZiqiangService) QaPatentSearch(text, imageHash string, topK int) (*qaReport, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("[%d]查询文本必填", ErrCodeQaParam)
	}
	if topK < 1 || topK > 20 {
		topK = 5
	}
	hasQueryHash := strings.TrimSpace(imageHash) != ""
	var queryHash uint64
	if hasQueryHash {
		var err error
		if queryHash, err = parseHashHex(imageHash); err != nil {
			return nil, err
		}
	}
	var patents []model.ZqTextCorpus
	if err := global.GVA_DB.Model(&model.ZqTextCorpus{}).
		Where("type = ?", "专利").Order("id ASC").Limit(2000).Find(&patents).Error; err != nil {
		return nil, err
	}
	imgHashByCorpus := map[uint]uint64{}
	if hasQueryHash {
		var imgs []model.ZqImageCorpus
		if err := global.GVA_DB.Model(&model.ZqImageCorpus{}).
			Where("hash_hex <> ''").Order("id ASC").Limit(2000).Find(&imgs).Error; err != nil {
			return nil, err
		}
		for _, im := range imgs {
			if h, err := parseHashHex(im.HashHex); err == nil {
				imgHashByCorpus[im.ID] = h
			}
		}
	}
	matches := rankPatents(patents, text, queryHash, hasQueryHash, imgHashByCorpus)
	if len(matches) > topK {
		matches = matches[:topK]
	}
	level, score, verdict, evidence := assessRisk(matches, hasQueryHash)
	return &qaReport{
		TextProvided:  true,
		ImageProvided: hasQueryHash,
		Thresholds: map[string]float64{
			"textHit": QaTextHit, "imageHit": QaImageHit,
			"riskHigh": QaRiskHigh, "riskMid": QaRiskMid,
		},
		TopMatches: matches,
		RiskLevel:  level,
		RiskScore:  score,
		Verdict:    verdict,
		Evidence:   evidence,
		Disclaimer: "本报告为算法初筛结论，不构成法律意见；重大决策请咨询执业律师",
	}, nil
}
