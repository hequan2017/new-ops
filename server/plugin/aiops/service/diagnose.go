// Package service aiops：Pod AI 诊断
// 采集（k8s 服务复用：Pod 详情含容器状态/条件/事件 + 日志尾部）→ buildPodPrompt 固定模板 → LLM 网关 → 留档
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	k8sSvc "github.com/hequan2017/new-ops/server/plugin/k8s/service"
	"github.com/hequan2017/new-ops/server/plugin/aiops/model"
)

// 诊断 prompt 模板：system 定位角色与输出口径，user 为采集快照
const (
	diagSystemPrompt = `你是资深 Kubernetes 运维专家。基于给出的 Pod 状态、容器信息、事件与日志片段，输出诊断结论。
要求：
1. 先一句话给出根因判断；
2. 列出关键证据（引用事件/日志原文片段）；
3. 给出可执行的修复建议（按优先级，含具体命令或 YAML 修改点）；
4. 如信息不足，明确说明还需采集什么。用中文回答，简洁专业。`
	diagLogTailLines = 50
)

// buildPodPrompt 把采集结果压成快照文本（纯函数，单测覆盖）
func buildPodPrompt(detail *k8sSvc.PodDetail, logs map[string]string) string {
	var buf bytes.Buffer
	dj, _ := json.MarshalIndent(detail, "", "  ")
	buf.WriteString("## Pod 状态（JSON）\n")
	buf.Write(dj)
	buf.WriteString("\n\n## 日志尾部\n")
	if len(logs) == 0 {
		buf.WriteString("（无可用日志）\n")
	}
	for c, l := range logs {
		buf.WriteString("### 容器 " + c + "（tail " + fmt.Sprint(diagLogTailLines) + " 行）\n")
		if strings.TrimSpace(l) == "" {
			buf.WriteString("（空）\n")
		} else {
			buf.WriteString(l)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// DiagnosePod 诊断一个 Pod：采集 → prompt → LLM → 留档
func (s *AiopsServiceStruct) DiagnosePod(ctx context.Context, clusterID uint, cluster, namespace, pod string) (*model.AiopsDiagnosis, error) {
	detail, err := k8sSvc.Service.K8sClusterService.GetPodDetail(clusterID, namespace, pod)
	if err != nil {
		return nil, fmt.Errorf("[%d]Pod 采集失败: %w", ErrCodeDiagnosisFailed, err)
	}
	logs := map[string]string{}
	for _, c := range detail.Containers {
		l, lerr := k8sSvc.Service.K8sClusterService.GetPodLogs(clusterID, namespace, pod, c.Name, diagLogTailLines)
		if lerr == nil {
			logs[c.Name] = l
		}
	}
	snapshot := buildPodPrompt(detail, logs)

	analysis, providerName, modelName, err := s.Chat(ctx, 0, diagSystemPrompt, snapshot)
	if err != nil {
		return nil, err
	}
	rec := &model.AiopsDiagnosis{
		ClusterID: clusterID, Cluster: cluster, Namespace: namespace, Pod: pod,
		Snapshot: snapshot, Analysis: analysis, Provider: providerName, Model: modelName,
	}
	if err := global.GVA_DB.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// GetDiagnosisList 诊断历史（快照大字段列表不回传）
func (s *AiopsServiceStruct) GetDiagnosisList(page, pageSize int) ([]model.AiopsDiagnosis, int64, error) {
	var list []model.AiopsDiagnosis
	var total int64
	db := global.GVA_DB.Model(&model.AiopsDiagnosis{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Select("id, created_at, updated_at, cluster_id, cluster, namespace, pod, analysis, provider, model").
		Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetDiagnosis 诊断详情（含快照与结论）
func (s *AiopsServiceStruct) GetDiagnosis(id uint) (*model.AiopsDiagnosis, error) {
	var rec model.AiopsDiagnosis
	if err := global.GVA_DB.First(&rec, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]诊断记录不存在", ErrCodeDiagnosisNotFound)
	}
	return &rec, nil
}
