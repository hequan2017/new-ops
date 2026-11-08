// Package service ziqiang：知识产权实体图谱（需求 3.5-6）
// 口径：三类实体白名单（法律/技术/权利主体）+ 关系边；多跳推理=BFS 最短路径
// （start→hop N→路径与终点返回，对齐 llmops C06 口径自实现），纯函数可测。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeGraphParam   = 2710
	ErrCodeGraphNotFound = 2711
)

// EntityCategories 实体分类白名单
var EntityCategories = []string{"法律", "技术", "权利主体"}

// ValidEntityCategory2 分类白名单
func ValidEntityCategory2(c string) bool {
	for _, v := range EntityCategories {
		if v == c {
			return true
		}
	}
	return false
}

// graphPath BFS 路径
type graphPath struct {
	NodeIds []uint          `json:"nodeIds"`
	Nodes   []model.ZqEntity `json:"nodes"`
	Edges   []model.ZqEdge  `json:"edges"`
}

// bfsPaths BFS 多跳（纯函数，测试锚点）：返回 start 可达的最短路径（hop ≤ maxHops）
func bfsPaths(nodes map[uint]model.ZqEntity, edges []model.ZqEdge, startId uint, maxHops int) []graphPath {
	adj := map[uint][]model.ZqEdge{}
	for _, e := range edges {
		adj[e.FromId] = append(adj[e.FromId], e)
	}
	type queueItem struct {
		nodeId uint
		depth  int
		path   []uint
		edges  []model.ZqEdge
	}
	visited := map[uint]bool{startId: true}
	var out []graphPath
	queue := []queueItem{{nodeId: startId, depth: 0, path: []uint{startId}}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.depth > 0 {
			out = append(out, graphPath{NodeIds: cur.path, Nodes: idsToNodes(cur.path, nodes), Edges: cur.edges})
		}
		if cur.depth >= maxHops {
			continue
		}
		for _, e := range adj[cur.nodeId] {
			if visited[e.ToId] {
				continue
			}
			visited[e.ToId] = true
			np := append([]uint{}, cur.path...)
			np = append(np, e.ToId)
			ne := append([]model.ZqEdge{}, cur.edges...)
			ne = append(ne, e)
			queue = append(queue, queueItem{nodeId: e.ToId, depth: cur.depth + 1, path: np, edges: ne})
		}
	}
	return out
}

// idsToNodes id 路径→实体列表
func idsToNodes(ids []uint, nodes map[uint]model.ZqEntity) []model.ZqEntity {
	out := make([]model.ZqEntity, 0, len(ids))
	for _, id := range ids {
		if n, ok := nodes[id]; ok {
			out = append(out, n)
		}
	}
	return out
}

// SaveEntity 新建/更新实体
func (s *ZiqiangService) SaveEntity(in *model.ZqEntity) (*model.ZqEntity, error) {
	if strings.TrimSpace(in.Name) == "" || !ValidEntityCategory2(in.Category) {
		return nil, fmt.Errorf("[%d]实体名必填且分类须在白名单内", ErrCodeGraphParam)
	}
	if in.ID > 0 {
		var exist model.ZqEntity
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]实体不存在", ErrCodeGraphNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"name": in.Name, "category": in.Category, "kind": in.Kind, "desc": in.Desc,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteEntity 删除实体（软删除）
func (s *ZiqiangService) DeleteEntity(id uint) error {
	res := global.GVA_DB.Delete(&model.ZqEntity{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]实体不存在", ErrCodeGraphNotFound)
	}
	return nil
}

// GetEntities 实体列表（分类过滤）
func (s *ZiqiangService) GetEntities(category string) ([]model.ZqEntity, error) {
	db := global.GVA_DB.Model(&model.ZqEntity{})
	if category != "" {
		if !ValidEntityCategory2(category) {
			return nil, fmt.Errorf("[%d]不支持的分类: %s", ErrCodeGraphParam, category)
		}
		db = db.Where("category = ?", category)
	}
	var list []model.ZqEntity
	err := db.Order("id").Limit(500).Find(&list).Error
	return list, err
}

// SaveEdge 新建关系边（禁自环）
func (s *ZiqiangService) SaveEdge(in *model.ZqEdge) (*model.ZqEdge, error) {
	if in.FromId == 0 || in.ToId == 0 || in.FromId == in.ToId {
		return nil, fmt.Errorf("[%d]关系边两端须为不同实体", ErrCodeGraphParam)
	}
	if strings.TrimSpace(in.Relation) == "" {
		return nil, fmt.Errorf("[%d]关系名必填", ErrCodeGraphParam)
	}
	var from, to model.ZqEntity
	if err := global.GVA_DB.First(&from, in.FromId).Error; err != nil {
		return nil, fmt.Errorf("[%d]起点实体不存在", ErrCodeGraphNotFound)
	}
	if err := global.GVA_DB.First(&to, in.ToId).Error; err != nil {
		return nil, fmt.Errorf("[%d]终点实体不存在", ErrCodeGraphNotFound)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteEdge 删除关系边
func (s *ZiqiangService) DeleteEdge(id uint) error {
	res := global.GVA_DB.Delete(&model.ZqEdge{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]关系边不存在", ErrCodeGraphNotFound)
	}
	return nil
}

// GetGraph 全图（节点+边）
func (s *ZiqiangService) GetGraph() (map[string]any, error) {
	var nodes []model.ZqEntity
	if err := global.GVA_DB.Order("id").Limit(1000).Find(&nodes).Error; err != nil {
		return nil, err
	}
	var edges []model.ZqEdge
	if err := global.GVA_DB.Order("id").Limit(5000).Find(&edges).Error; err != nil {
		return nil, err
	}
	return map[string]any{"nodes": nodes, "edges": edges}, nil
}

// GraphBFS 多跳推理
func (s *ZiqiangService) GraphBFS(startId uint, maxHops int) ([]graphPath, error) {
	if maxHops < 1 || maxHops > 6 {
		return nil, fmt.Errorf("[%d]跳数须在 [1,6] 内", ErrCodeGraphParam)
	}
	var start model.ZqEntity
	if err := global.GVA_DB.First(&start, startId).Error; err != nil {
		return nil, fmt.Errorf("[%d]起点实体不存在", ErrCodeGraphNotFound)
	}
	var nodes []model.ZqEntity
	if err := global.GVA_DB.Limit(1000).Find(&nodes).Error; err != nil {
		return nil, err
	}
	nodeMap := map[uint]model.ZqEntity{}
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}
	var edges []model.ZqEdge
	if err := global.GVA_DB.Limit(5000).Find(&edges).Error; err != nil {
		return nil, err
	}
	return bfsPaths(nodeMap, edges, startId, maxHops), nil
}
