// Package service exam：分布式调度（采购件③）——节点注册/心跳 + 领卷分片分配
// 口径：节点按 region 优先 + 负载升序择优；分片以 sheetId 唯一索引幂等（重复领取返回原分配）；
// 心跳超时节点自动离线（阈值可注入）。
package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const (
	ErrCodeNodeBad      = 2417
	nodeStaleDefaultMin = 5
)

// isStale 心跳超时判定（纯函数，测试锚点）
func isStale(last, now time.Time, staleMinutes int) bool {
	return now.Sub(last) > time.Duration(staleMinutes)*time.Minute
}

// pickNode 择优节点（纯函数，测试锚点）：online → region 优先 → 负载升序 → 名称序
func pickNode(nodes []model.ExamNode, preferRegion string) (string, error) {
	var candidates []model.ExamNode
	for _, n := range nodes {
		if n.Status == "online" {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("[%d]无在线考试节点", ErrCodeNodeBad)
	}
	sort.Slice(candidates, func(i, j int) bool {
		ri, rj := candidates[i].Region == preferRegion, candidates[j].Region == preferRegion
		if ri != rj {
			return ri
		}
		if candidates[i].Load != candidates[j].Load {
			return candidates[i].Load < candidates[j].Load
		}
		return candidates[i].NodeKey < candidates[j].NodeKey
	})
	return candidates[0].NodeKey, nil
}

// ExamServiceApp 引用（避免循环说明：调度方法挂 ExamService）

// RegisterNode 注册/更新节点（upsert 幂等）
func (s *ExamService) RegisterNode(nodeKey, name, region string) (*model.ExamNode, error) {
	if strings.TrimSpace(nodeKey) == "" {
		return nil, fmt.Errorf("[%d]节点标识不能为空", ErrCodeNodeBad)
	}
	var node model.ExamNode
	err := global.GVA_DB.Where("node_key = ?", nodeKey).First(&node).Error
	if err == gorm.ErrRecordNotFound {
		node = model.ExamNode{NodeKey: nodeKey, Name: name, Region: region,
			LastHeartbeat: time.Now(), Status: "online"}
		if err := global.GVA_DB.Create(&node).Error; err != nil {
			return nil, err
		}
		return &node, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{"last_heartbeat": time.Now(), "status": "online"}
	if name != "" {
		updates["name"] = name
	}
	if region != "" {
		updates["region"] = region
	}
	if err := global.GVA_DB.Model(&node).Updates(updates).Error; err != nil {
		return nil, err
	}
	node.Status = "online"
	return &node, nil
}

// NodeHeartbeat 心跳（不存在自动注册；负载回传）
func (s *ExamService) NodeHeartbeat(nodeKey string, load int) (*model.ExamNode, error) {
	var node model.ExamNode
	if err := global.GVA_DB.Where("node_key = ?", nodeKey).First(&node).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.RegisterNode(nodeKey, "", "")
		}
		return nil, err
	}
	if err := global.GVA_DB.Model(&node).Updates(map[string]any{
		"last_heartbeat": time.Now(), "load": load, "status": "online",
	}).Error; err != nil {
		return nil, err
	}
	node.Load, node.Status = load, "online"
	return &node, nil
}

// MarkStaleNodesOffline 超时节点批量离线（返回离线数）
func (s *ExamService) MarkStaleNodesOffline(staleMinutes int) (int64, error) {
	if staleMinutes <= 0 {
		staleMinutes = nodeStaleDefaultMin
	}
	threshold := time.Now().Add(-time.Duration(staleMinutes) * time.Minute)
	res := global.GVA_DB.Model(&model.ExamNode{}).
		Where("status = 'online' AND last_heartbeat < ?", threshold).Update("status", "offline")
	return res.RowsAffected, res.Error
}

// GetNodes 节点列表
func (s *ExamService) GetNodes() ([]model.ExamNode, error) {
	var list []model.ExamNode
	err := global.GVA_DB.Order("region, load").Find(&list).Error
	return list, err
}

// AssignShard 领卷分片（行锁事务 + sheetId 唯一幂等）
func (s *ExamService) AssignShard(examId, sheetId uint, preferRegion string) (*model.ExamShard, error) {
	var result *model.ExamShard
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 幂等：已分配直接返回
		var exist model.ExamShard
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sheet_id = ?", sheetId).First(&exist).Error
		if err == nil {
			result = &exist
			return nil
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		// 候选节点（行锁防并发同择）
		var nodes []model.ExamNode
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status = 'online'").Find(&nodes).Error; err != nil {
			return err
		}
		nodeKey, err := pickNode(nodes, preferRegion)
		if err != nil {
			return err
		}
		shard := &model.ExamShard{ExamId: examId, SheetId: sheetId, NodeKey: nodeKey, IssuedAt: time.Now()}
		if err := tx.Create(shard).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ExamNode{}).Where("node_key = ?", nodeKey).
			UpdateColumn("load", gorm.Expr("load + 1")).Error; err != nil {
			return err
		}
		result = shard
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NodeIn 节点注册请求
type NodeIn struct {
	NodeKey string `json:"nodeKey" binding:"required"`
	Name    string `json:"name"`
	Region  string `json:"region"`
}

// NodeHeartbeatIn 心跳请求
type NodeHeartbeatIn struct {
	NodeKey string `json:"nodeKey" binding:"required"`
	Load    int    `json:"load"`
}

// AssignIn 分片分配请求
type AssignIn struct {
	ExamId  uint   `json:"examId" binding:"required"`
	SheetId uint   `json:"sheetId" binding:"required"`
	Region  string `json:"region"`
}
