// Package service fanhua：城市科学融合查询（需求 3.4-22）
// 口径：问卷/POI/照片元数据三源联合查询（区域/期间参数域过滤）；社区画像复用 N11
// 指标树加权；治理建议=叶得分阈值规则模板（<60 触发建议），纯函数可测。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const ErrCodeFusionParam = 2630

// govSuggestionRules 治理建议规则（叶指标名→建议模板，得分<60 触发）
var govSuggestionRules = map[string]string{
	"空气质量":  "加强扬尘与机动车排放管控，加密空气质量监测点位",
	"噪声水平":  "开展夜间施工与商业噪声专项治理",
	"设施可达性": "补齐 15 分钟生活圈缺项设施，优化 POI 布局",
	"服务满意度": "开展物业与社区服务满意度专项提升",
	"街景美观度": "推进街景整治与店招统一美化",
	"建筑协调性": "加强历史风貌区建筑外立面协调管控",
	"绿化视觉":  "推进立体绿化与口袋公园建设",
}

// buildGovSuggestions 治理建议生成（纯函数，测试锚点）：低分叶触发建议，全部良好输出保持口径
func buildGovSuggestions(leafScores map[string]float64) []string {
	var out []string
	for _, name := range []string{"空气质量", "噪声水平", "设施可达性", "服务满意度", "街景美观度", "建筑协调性", "绿化视觉"} {
		if score, ok := leafScores[name]; ok && score < 60 {
			if tpl, hit := govSuggestionRules[name]; hit {
				out = append(out, fmt.Sprintf("【%s %.0f 分】%s", name, score, tpl))
			}
		}
	}
	if len(out) == 0 {
		out = append(out, "各维度表现良好，保持现有治理节奏并持续监测。")
	}
	return out
}

// collectLeafScores 收集叶指标得分（树→叶名→分，纯函数）
func collectLeafScores(roots []*indicatorNode) map[string]float64 {
	out := map[string]float64{}
	var walk func(n *indicatorNode)
	walk = func(n *indicatorNode) {
		if len(n.Children) == 0 {
			out[n.Name] = n.Avg
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, r := range roots {
		walk(r)
	}
	return out
}

// SaveGovPhoto 登记社区照片元数据
func (s *FanhuaService) SaveGovPhoto(in *model.FanhuaGovPhoto) (*model.FanhuaGovPhoto, error) {
	if strings.TrimSpace(in.Region) == "" || strings.TrimSpace(in.Path) == "" {
		return nil, errx.New(ErrCodeFusionParam, "区域与照片路径必填")
	}
	if err := s.repo.CreateGovPhoto(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "登记照片失败", err)
	}
	return in, nil
}

// FusionQuery 三源融合查询（数据点分源统计+照片+画像+建议）
func (s *FanhuaService) FusionQuery(region, period string) (map[string]any, error) {
	if strings.TrimSpace(region) == "" {
		return nil, errx.New(ErrCodeFusionParam, "区域必填")
	}
	db := s.dbh()
	points, err := s.repo.GovDataPointsByRegion(db, region, period)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询区域数据点失败", err)
	}
	// 分源统计（questionnaire/poi/sensor 计数与均值）
	type srcStat struct {
		Source string  `json:"source"`
		Count  int     `json:"count"`
		Avg    float64 `json:"avg"`
	}
	stats := map[string]*srcStat{"questionnaire": {Source: "questionnaire"}, "poi": {Source: "poi"}, "sensor": {Source: "sensor"}}
	sum := map[string]float64{}
	for _, p := range points {
		if st, ok := stats[p.Source]; ok {
			st.Count++
			sum[p.Source] += p.Value
		}
	}
	for _, st := range stats {
		if st.Count > 0 {
			st.Avg = float64(int(sum[st.Source]/float64(st.Count)*100+0.5)) / 100
		}
	}
	// 照片源
	photos, err := s.repo.GovPhotosByRegion(db, region, period)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询区域照片失败", err)
	}
	photoTotal, err := s.repo.GovPhotoCountByRegion(db, region)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "统计区域照片失败", err)
	}
	// 画像（复用 N11）
	profile, err := s.GetRegionProfile(region, period)
	if err != nil {
		return nil, err
	}
	// 治理建议
	list, err := s.repo.GovIndicatorsAll(db)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询治理指标失败", err)
	}
	roots := buildIndicatorTree(list)
	for _, r := range roots {
		scoreTree(r, points)
	}
	suggestions := buildGovSuggestions(collectLeafScores(roots))
	return map[string]any{
		"region": region, "period": period,
		"sourceStats": stats, "photos": photos, "photoTotal": photoTotal,
		"profile": profile, "suggestions": suggestions,
	}, nil
}
