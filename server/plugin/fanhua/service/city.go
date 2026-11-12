// Package service fanhua：城市科学知识库（需求 3.4-17）
// 口径：治理指标父子树（含视觉评价维度），数据点按区域×指标×期间聚合；
// 得分=叶子均值×权重逐层加权归一（权重和为 0 时按 1 处理），纯函数可测。
package service

import (
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeGovParam    = 2621
	ErrCodeGovNotFound = 2622
)

// indicatorNode 树节点
type indicatorNode struct {
	model.FanhuaGovIndicator
	Children []*indicatorNode `json:"children"`
	Avg      float64          `json:"avg"`
}

// buildIndicatorTree 平铺→树（纯函数，测试锚点）
func buildIndicatorTree(list []model.FanhuaGovIndicator) []*indicatorNode {
	nodes := make(map[uint]*indicatorNode, len(list))
	for i := range list {
		nodes[list[i].ID] = &indicatorNode{FanhuaGovIndicator: list[i], Children: []*indicatorNode{}}
	}
	var roots []*indicatorNode
	for i := range list {
		n := nodes[list[i].ID]
		if p, ok := nodes[list[i].ParentId]; ok && list[i].ParentId != 0 {
			p.Children = append(p.Children, n)
		} else {
			roots = append(roots, n)
		}
	}
	return roots
}

// scoreTree 递归加权（纯函数，测试锚点）：叶子=数据点均值，父=子加权归一
func scoreTree(n *indicatorNode, points []model.FanhuaGovDataPoint) float64 {
	if len(n.Children) == 0 {
		sum, cnt := 0.0, 0
		for _, p := range points {
			if p.IndicatorId == n.ID {
				sum += p.Value
				cnt++
			}
		}
		if cnt > 0 {
			n.Avg = sum / float64(cnt)
		}
		return n.Avg
	}
	weighted, total := 0.0, 0.0
	for _, c := range n.Children {
		scoreTree(c, points)
		w := float64(c.Weight)
		if w <= 0 {
			w = 1
		}
		weighted += c.Avg * w
		total += w
	}
	if total > 0 {
		n.Avg = weighted / total
	}
	return n.Avg
}

// GetIndicatorTree 指标树
func (s *FanhuaService) GetIndicatorTree() ([]*indicatorNode, error) {
	list, err := s.repo.GovIndicatorsAll(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询治理指标失败", err)
	}
	return buildIndicatorTree(list), nil
}

// SaveIndicator 新建/更新指标
func (s *FanhuaService) SaveIndicator(in *model.FanhuaGovIndicator) (*model.FanhuaGovIndicator, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errx.New(ErrCodeGovParam, "指标名必填")
	}
	if in.Weight < 0 {
		in.Weight = 1
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindGovIndicator(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeGovNotFound, "指标不存在")
		}
		if err := s.repo.UpdateGovIndicatorColumns(db, exist.ID, map[string]any{
			"name": in.Name, "parent_id": in.ParentId, "weight": in.Weight,
			"visual": in.Visual, "desc": in.Desc,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新指标失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateGovIndicator(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建指标失败", err)
	}
	return in, nil
}

// DeleteIndicator 删除指标（软删除）
func (s *FanhuaService) DeleteIndicator(id uint) error {
	rows, err := s.repo.DeleteGovIndicator(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除指标失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeGovNotFound, "指标不存在")
	}
	return nil
}

// SaveDataPoint 写数据点
func (s *FanhuaService) SaveDataPoint(in *model.FanhuaGovDataPoint) (*model.FanhuaGovDataPoint, error) {
	if strings.TrimSpace(in.Region) == "" || strings.TrimSpace(in.Period) == "" {
		return nil, errx.New(ErrCodeGovParam, "区域与期间必填")
	}
	db := s.dbh()
	if _, err := s.repo.FindGovIndicator(db, in.IndicatorId); err != nil {
		return nil, errx.New(ErrCodeGovNotFound, "指标不存在")
	}
	if in.Value < 0 || in.Value > 100 {
		return nil, errx.New(ErrCodeGovParam, "指标值须在 [0,100] 内")
	}
	if err := s.repo.CreateGovDataPoint(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "写入数据点失败", err)
	}
	return in, nil
}

// GetDataPoints 数据点列表（区域/期间过滤，倒序分页）
func (s *FanhuaService) GetDataPoints(region, period string, page, pageSize int) ([]model.FanhuaGovDataPoint, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	list, total, err := s.repo.GovDataPointList(s.dbh(), region, period, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询数据点失败", err)
	}
	return list, total, nil
}

// GetRegionProfile 社区画像（指标树加权得分）
func (s *FanhuaService) GetRegionProfile(region, period string) (map[string]any, error) {
	if strings.TrimSpace(region) == "" {
		return nil, errx.New(ErrCodeGovParam, "区域必填")
	}
	db := s.dbh()
	points, err := s.repo.GovDataPointsByRegion(db, region, period)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询区域数据点失败", err)
	}
	list, err := s.repo.GovIndicatorsAll(db)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询治理指标失败", err)
	}
	roots := buildIndicatorTree(list)
	overall := 0.0
	total := 0.0
	for _, r := range roots {
		scoreTree(r, points)
		w := float64(r.Weight)
		if w <= 0 {
			w = 1
		}
		overall += r.Avg * w
		total += w
	}
	if total > 0 {
		overall /= total
	}
	return map[string]any{"region": region, "period": period, "overall": roundScore(overall), "tree": roots, "dataPoints": len(points)}, nil
}

// roundScore 两位小数
func roundScore(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
