// Package service ziqiang：知识产权实体图谱（需求 3.5-6）
// 口径：三类实体白名单（法律/技术/权利主体）+ 关系边；多跳推理=BFS 最短路径
// （start→hop N→路径与终点返回，对齐 llmops C06 口径自实现），纯函数可测。
package service

import (
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeGraphParam    = 2710
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
	NodeIds []uint           `json:"nodeIds"`
	Nodes   []model.ZqEntity `json:"nodes"`
	Edges   []model.ZqEdge   `json:"edges"`
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
		return nil, zqErr(ErrCodeGraphParam, "实体名必填且分类须在白名单内")
	}
	if in.ID > 0 {
		exist, err := s.repo.FindEntity(s.dbh(), in.ID)
		if err != nil {
			return nil, zqErr(ErrCodeGraphNotFound, "实体不存在")
		}
		if err := s.repo.UpdateEntityColumns(s.dbh(), exist.ID, map[string]any{
			"name": in.Name, "category": in.Category, "kind": in.Kind, "desc": in.Desc,
		}); err != nil {
			return exist, errx.Wrap(ErrCodeZqStore, "更新实体失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateEntity(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库实体失败", err)
	}
	return in, nil
}

// DeleteEntity 删除实体（软删除）
func (s *ZiqiangService) DeleteEntity(id uint) error {
	rows, err := s.repo.DeleteEntity(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeZqStore, "删除实体失败", err)
	}
	if rows == 0 {
		return zqErr(ErrCodeGraphNotFound, "实体不存在")
	}
	return nil
}

// GetEntities 实体列表（分类过滤）
func (s *ZiqiangService) GetEntities(category string) ([]model.ZqEntity, error) {
	if category != "" && !ValidEntityCategory2(category) {
		return nil, zqErr(ErrCodeGraphParam, "不支持的分类: %s", category)
	}
	list, err := s.repo.EntityList(s.dbh(), category, 500)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "查询实体失败", err)
	}
	return list, nil
}

// SaveEdge 新建关系边（禁自环）
func (s *ZiqiangService) SaveEdge(in *model.ZqEdge) (*model.ZqEdge, error) {
	if in.FromId == 0 || in.ToId == 0 || in.FromId == in.ToId {
		return nil, zqErr(ErrCodeGraphParam, "关系边两端须为不同实体")
	}
	if strings.TrimSpace(in.Relation) == "" {
		return nil, zqErr(ErrCodeGraphParam, "关系名必填")
	}
	if _, err := s.repo.FindEntity(s.dbh(), in.FromId); err != nil {
		return nil, zqErr(ErrCodeGraphNotFound, "起点实体不存在")
	}
	if _, err := s.repo.FindEntity(s.dbh(), in.ToId); err != nil {
		return nil, zqErr(ErrCodeGraphNotFound, "终点实体不存在")
	}
	if err := s.repo.CreateEdge(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库关系边失败", err)
	}
	return in, nil
}

// DeleteEdge 删除关系边
func (s *ZiqiangService) DeleteEdge(id uint) error {
	rows, err := s.repo.DeleteEdge(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeZqStore, "删除关系边失败", err)
	}
	if rows == 0 {
		return zqErr(ErrCodeGraphNotFound, "关系边不存在")
	}
	return nil
}

// GetGraph 全图（节点+边）
func (s *ZiqiangService) GetGraph() (map[string]any, error) {
	nodes, err := s.repo.EntityList(s.dbh(), "", 1000)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "查询图谱节点失败", err)
	}
	edges, err := s.repo.EdgeList(s.dbh(), 5000)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "查询图谱边失败", err)
	}
	return map[string]any{"nodes": nodes, "edges": edges}, nil
}

// GraphBFS 多跳推理
func (s *ZiqiangService) GraphBFS(startId uint, maxHops int) ([]graphPath, error) {
	if maxHops < 1 || maxHops > 6 {
		return nil, zqErr(ErrCodeGraphParam, "跳数须在 [1,6] 内")
	}
	if _, err := s.repo.FindEntity(s.dbh(), startId); err != nil {
		return nil, zqErr(ErrCodeGraphNotFound, "起点实体不存在")
	}
	nodes, err := s.repo.EntityList(s.dbh(), "", 1000)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "查询图谱节点失败", err)
	}
	nodeMap := map[uint]model.ZqEntity{}
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}
	edges, err := s.repo.EdgeList(s.dbh(), 5000)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "查询图谱边失败", err)
	}
	return bfsPaths(nodeMap, edges, startId, maxHops), nil
}
