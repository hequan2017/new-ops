// Package service llmops：知识图谱库（要求(2) 四库之一）——节点/边管理 + BFS 多跳推理
// 口径：有向图；路径为 simple path（环内不重复节点）；路径数上限 100 防爆炸； hops 夹取 1-5。
package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeGpKbTypeBad  = 2246
	ErrCodeGpNodeMiss   = 2247
	ErrCodeGpNodeDup    = 2248
	ErrCodeGpEdgeMiss   = 2249
	graphMaxHops        = 5
	graphMaxPaths       = 100
)

// PathStep 推理路径步骤
type PathStep struct {
	FromId   uint   `json:"fromId"`
	FromName string `json:"fromName"`
	Relation string `json:"relation"`
	ToId     uint   `json:"toId"`
	ToName   string `json:"toName"`
}

// GraphEdgeRef BFS 内存引用
type GraphEdgeRef struct {
	To       uint
	Relation string
}

// SaveGraphNode 新建/更新节点（库内同名去重）
func (s *LlmopsServiceStruct) SaveGraphNode(n *model.LlmopsGraphNode) error {
	if n == nil || strings.TrimSpace(n.Name) == "" {
		return fmt.Errorf("[%d]节点名不能为空", ErrCodeGpNodeMiss)
	}
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, n.KbId).Error; err != nil {
		return fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "graph" {
		return fmt.Errorf("[%d]仅 graph 类型知识库支持图谱节点（当前 %s）", ErrCodeGpKbTypeBad, kb.Type)
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsGraphNode{}).Where("kb_id = ? AND name = ?", n.KbId, n.Name)
	if n.ID > 0 {
		q = q.Where("id <> ?", n.ID)
	}
	q.Count(&count)
	if count > 0 {
		return fmt.Errorf("[%d]库内节点名已存在", ErrCodeGpNodeDup)
	}
	if n.ID > 0 {
		return global.GVA_DB.Model(n).Updates(map[string]any{"name": n.Name, "type": n.Type, "props": n.Props}).Error
	}
	return global.GVA_DB.Create(n).Error
}

// GetGraphNodes 库内节点列表
func (s *LlmopsServiceStruct) GetGraphNodes(kbId uint) ([]model.LlmopsGraphNode, error) {
	var list []model.LlmopsGraphNode
	err := global.GVA_DB.Where("kb_id = ?", kbId).Order("id").Find(&list).Error
	return list, err
}

// DeleteGraphNode 删除节点（级联删关联边）
func (s *LlmopsServiceStruct) DeleteGraphNode(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var n model.LlmopsGraphNode
		if err := tx.First(&n, id).Error; err != nil {
			return fmt.Errorf("[%d]节点不存在", ErrCodeGpNodeMiss)
		}
		if err := tx.Where("from_node_id = ? OR to_node_id = ?", id, id).Delete(&model.LlmopsGraphEdge{}).Error; err != nil {
			return err
		}
		return tx.Delete(&n).Error
	})
}

// SaveGraphEdge 新建边（起终点须存在且同库）
func (s *LlmopsServiceStruct) SaveGraphEdge(e *model.LlmopsGraphEdge) error {
	if e == nil || strings.TrimSpace(e.Relation) == "" || e.FromNodeId == 0 || e.ToNodeId == 0 {
		return fmt.Errorf("[%d]关系名与起终点必填", ErrCodeGpEdgeMiss)
	}
	var from, to model.LlmopsGraphNode
	if err := global.GVA_DB.First(&from, e.FromNodeId).Error; err != nil {
		return fmt.Errorf("[%d]起点节点不存在", ErrCodeGpNodeMiss)
	}
	if err := global.GVA_DB.First(&to, e.ToNodeId).Error; err != nil {
		return fmt.Errorf("[%d]终点节点不存在", ErrCodeGpNodeMiss)
	}
	if from.KbId != to.KbId {
		return fmt.Errorf("[%d]起终点须同库", ErrCodeGpEdgeMiss)
	}
	e.KbId = from.KbId
	return global.GVA_DB.Create(e).Error
}

// GetGraphEdges 库内边列表
func (s *LlmopsServiceStruct) GetGraphEdges(kbId uint) ([]model.LlmopsGraphEdge, error) {
	var list []model.LlmopsGraphEdge
	err := global.GVA_DB.Where("kb_id = ?", kbId).Order("id").Find(&list).Error
	return list, err
}

// DeleteGraphEdge 删除边
func (s *LlmopsServiceStruct) DeleteGraphEdge(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsGraphEdge{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]边不存在", ErrCodeGpEdgeMiss)
	}
	return nil
}

// BfsPaths 多跳路径枚举（纯函数，测试锚点）：simple path + hops 上限 + 路径数截断
func BfsPaths(adj map[uint][]GraphEdgeRef, start uint, maxHops int) [][]PathStep {
	if maxHops <= 0 {
		maxHops = 1
	}
	if maxHops > graphMaxHops {
		maxHops = graphMaxHops
	}
	var out [][]PathStep
	var walk func(node uint, hops int, path []PathStep, visited map[uint]bool)
	walk = func(node uint, hops int, path []PathStep, visited map[uint]bool) {
		if len(out) >= graphMaxPaths {
			return
		}
		if hops == 0 {
			return
		}
		for _, e := range adj[node] {
			if visited[e.To] {
				continue
			}
			step := PathStep{FromId: node, Relation: e.Relation, ToId: e.To}
			next := append(append([]PathStep{}, path...), step)
			out = append(out, next)
			if len(out) >= graphMaxPaths {
				return
			}
			visited[e.To] = true
			walk(e.To, hops-1, next, visited)
			visited[e.To] = false
			if len(out) >= graphMaxPaths { // 递归触顶后不再继续兄弟分支
				return
			}
		}
	}
	walk(start, maxHops, nil, map[uint]bool{start: true})
	return out
}

// GraphMultiHopQuery 多跳推理：起点出发枚举路径（relation 可选过滤）
func (s *LlmopsServiceStruct) GraphMultiHopQuery(kbId, startNodeId uint, maxHops int, relation string) ([][]PathStep, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "graph" {
		return nil, fmt.Errorf("[%d]仅 graph 类型知识库支持多跳推理（当前 %s）", ErrCodeGpKbTypeBad, kb.Type)
	}
	var start model.LlmopsGraphNode
	if err := global.GVA_DB.Where("id = ? AND kb_id = ?", startNodeId, kbId).First(&start).Error; err != nil {
		return nil, fmt.Errorf("[%d]起点节点不存在", ErrCodeGpNodeMiss)
	}
	var edges []model.LlmopsGraphEdge
	if err := global.GVA_DB.Where("kb_id = ?", kbId).Find(&edges).Error; err != nil {
		return nil, err
	}
	adj := make(map[uint][]GraphEdgeRef, len(edges))
	for _, e := range edges {
		if relation != "" && e.Relation != relation {
			continue
		}
		adj[e.FromNodeId] = append(adj[e.FromNodeId], GraphEdgeRef{To: e.ToNodeId, Relation: e.Relation})
	}
	paths := BfsPaths(adj, startNodeId, maxHops)
	// 回填节点名
	names := map[uint]string{}
	var nodes []model.LlmopsGraphNode
	if err := global.GVA_DB.Where("kb_id = ?", kbId).Find(&nodes).Error; err == nil {
		for _, n := range nodes {
			names[n.ID] = n.Name
		}
	}
	for i := range paths {
		for j := range paths[i] {
			paths[i][j].FromName = names[paths[i][j].FromId]
			paths[i][j].ToName = names[paths[i][j].ToId]
		}
	}
	return paths, nil
}
