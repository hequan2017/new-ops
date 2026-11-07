// Package service oplib：算子编排（采购件② DAG 管道）
// 口径：def = {"nodes":[{id,operator,args}],"edges":[[from,to]]}；Kahn 拓扑排序校验环与边引用；
// 简化约束——每节点至多一个父（多入边拒绝）；执行按拓扑序，节点 args 模板 {input} 占位父输出，
// 无 args 时默认 {"text": 父输出}；全程节点级 trace。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
)

// 错误码（oplib 段续用）
const ErrCodePipe = 2310

// PipeNode 管道节点
type PipeNode struct {
	ID       string         `json:"id"`
	Operator string         `json:"operator"`
	Args     map[string]any `json:"args"`
}

// PipeDef 管道定义
type PipeDef struct {
	Nodes []PipeNode  `json:"nodes"`
	Edges [][2]string `json:"edges"`
}

// ParsePipelineDef 定义解析 + 拓扑校验（纯函数，测试锚点）：返回拓扑序
func ParsePipelineDef(raw string) (*PipeDef, []string, error) {
	var def PipeDef
	if err := json.Unmarshal([]byte(raw), &def); err != nil {
		return nil, nil, fmt.Errorf("[%d]管道定义解析失败: %w", ErrCodePipe, err)
	}
	if len(def.Nodes) == 0 {
		return nil, nil, fmt.Errorf("[%d]管道至少一个节点", ErrCodePipe)
	}
	idx := make(map[string]int, len(def.Nodes))
	for i := range def.Nodes {
		n := &def.Nodes[i]
		if n.ID == "" {
			return nil, nil, fmt.Errorf("[%d]节点 ID 为空", ErrCodePipe)
		}
		if _, dup := idx[n.ID]; dup {
			return nil, nil, fmt.Errorf("[%d]节点 ID 重复: %s", ErrCodePipe, n.ID)
		}
		if strings.TrimSpace(n.Operator) == "" {
			return nil, nil, fmt.Errorf("[%d]节点 %s 缺算子名", ErrCodePipe, n.ID)
		}
		idx[n.ID] = i
	}
	indegree := make(map[string]int, len(def.Nodes))
	parentOf := make(map[string]string, len(def.Nodes))
	for _, e := range def.Edges {
		from, to := e[0], e[1]
		if _, ok := idx[from]; !ok {
			return nil, nil, fmt.Errorf("[%d]边引用不存在节点: %s", ErrCodePipe, from)
		}
		if _, ok := idx[to]; !ok {
			return nil, nil, fmt.Errorf("[%d]边引用不存在节点: %s", ErrCodePipe, to)
		}
		if _, dup := parentOf[to]; dup {
			return nil, nil, fmt.Errorf("[%d]节点 %s 存在多个父节点（简化约束：至多一父）", ErrCodePipe, to)
		}
		parentOf[to] = from
		indegree[to]++
	}
	// Kahn 拓扑排序
	var order []string
	queue := make([]string, 0, len(def.Nodes))
	for _, n := range def.Nodes {
		if indegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)
		for _, e := range def.Edges {
			if e[0] == cur {
				indegree[e[1]]--
				if indegree[e[1]] == 0 {
					queue = append(queue, e[1])
				}
			}
		}
	}
	if len(order) != len(def.Nodes) {
		return nil, nil, fmt.Errorf("[%d]管道存在环", ErrCodePipe)
	}
	return &def, order, nil
}

// PipeStepTrace 节点执行留痕
type PipeStepTrace struct {
	NodeId string `json:"nodeId"`
	In     string `json:"in"`
	Out    string `json:"out"`
}

// PipeResult 管道执行结果
type PipeResult struct {
	Status string          `json:"status"`
	Output string          `json:"output"`
	Steps  []PipeStepTrace `json:"steps"`
}

// pipeExecutor 节点执行注入点（算子名+参数模板渲染后 JSON → 输出）
type pipeExecutor func(operator string, argsJSON string) (any, error)

// RunPipelineDef 管道执行核心（executor 注入，测试锚点）
func RunPipelineDef(def *PipeDef, order []string, executor pipeExecutor, input string) (*PipeResult, error) {
	res := &PipeResult{Status: "done"}
	byId := make(map[string]*PipeNode, len(def.Nodes))
	for i := range def.Nodes {
		byId[def.Nodes[i].ID] = &def.Nodes[i]
	}
	nodeOut := make(map[string]string, len(def.Nodes))
	parentOf := map[string]string{}
	for _, e := range def.Edges {
		parentOf[e[1]] = e[0]
	}
	for _, nodeId := range order {
		n := byId[nodeId]
		parentInput, hasParent := nodeOut[parentOf[nodeId]]
		var callArgs map[string]any
		if len(n.Args) > 0 {
			callArgs = make(map[string]any, len(n.Args))
			for k, v := range n.Args {
				if s, ok := v.(string); ok && strings.Contains(s, "{input}") {
					prev := input
					if hasParent {
						prev = parentInput
					}
					callArgs[k] = strings.ReplaceAll(s, "{input}", prev)
				} else {
					callArgs[k] = v
				}
			}
		} else {
			prev := input
			if hasParent {
				prev = parentInput
			}
			callArgs = map[string]any{"text": prev}
		}
		b, _ := json.Marshal(callArgs)
		out, err := executor(n.Operator, string(b))
		if err != nil {
			res.Status = "failed"
			res.Steps = append(res.Steps, PipeStepTrace{NodeId: nodeId, In: string(b), Out: err.Error()})
			return res, fmt.Errorf("[%d]节点 %s 执行失败: %w", ErrCodePipe, nodeId, err)
		}
		outStr := fmt.Sprintf("%v", out)
		nodeOut[nodeId] = outStr
		res.Steps = append(res.Steps, PipeStepTrace{NodeId: nodeId, In: string(b), Out: outStr})
		res.Output = outStr
	}
	return res, nil
}

// GetPipelines 管道列表
func (s *OplibService) GetPipelines() ([]model.OplibPipeline, error) {
	var list []model.OplibPipeline
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SavePipeline 保存管道（定义即解析校验）
func (s *OplibService) SavePipeline(in *model.OplibPipeline) (*model.OplibPipeline, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("[%d]管道名必填", ErrCodePipe)
	}
	if _, _, err := ParsePipelineDef(in.DefJSON); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.OplibPipeline{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]管道名已存在", ErrCodePipe)
	}
	if in.ID > 0 {
		var p model.OplibPipeline
		if err := global.GVA_DB.First(&p, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]管道不存在", ErrCodePipe)
		}
		err := global.GVA_DB.Model(&p).Updates(map[string]any{
			"name": in.Name, "desc": in.Desc, "def_json": in.DefJSON, "status": in.Status,
		}).Error
		if err != nil {
			return nil, err
		}
		return &p, nil
	}
	if in.Status == "" {
		in.Status = "draft"
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// RunPipeline 执行管道（节点走 schema 校验+内置执行，整体留痕）
func (s *OplibService) RunPipeline(pipelineId uint, input string) (*PipeResult, error) {
	var p model.OplibPipeline
	if err := global.GVA_DB.First(&p, pipelineId).Error; err != nil {
		return nil, fmt.Errorf("[%d]管道不存在", ErrCodePipe)
	}
	def, order, err := ParsePipelineDef(p.DefJSON)
	if err != nil {
		return nil, err
	}
	// 节点执行器：算子存在性 + schema 校验 + 内置执行
	opCache := map[string]model.OplibOperator{}
	executor := func(operator, argsJSON string) (any, error) {
		op, ok := opCache[operator]
		if !ok {
			if err := global.GVA_DB.Where("name = ?", operator).First(&op).Error; err != nil {
				return nil, fmt.Errorf("算子 %s 不存在", operator)
			}
			opCache[operator] = op
		}
		args := map[string]any{}
		if strings.TrimSpace(argsJSON) != "" {
			if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
				return nil, fmt.Errorf("节点参数非法: %w", err)
			}
		}
		if err := ValidateArgsAgainstSchema(op.InputSchema, args); err != nil {
			return nil, err
		}
		return runBuiltin(op.Name, argsJSON)
	}
	result, err := RunPipelineDef(def, order, executor, input)
	// 整体留痕（成败均记）
	b, _ := json.Marshal(result)
	global.GVA_DB.Create(&model.OplibRunLog{
		OperatorId: 0, Name: "pipeline:" + p.Name, ArgsJSON: input,
		ResultJSON: string(b), Ok: err == nil,
		LatencyMs: 0, ErrMsg: errSummaryLocal(err),
	})
	return result, err
}

// DeletePipeline 删除管道
func (s *OplibService) DeletePipeline(id uint) error {
	res := global.GVA_DB.Delete(&model.OplibPipeline{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]管道不存在", ErrCodePipe)
	}
	return nil
}

func errSummaryLocal(err error) string {
	if err == nil {
		return ""
	}
	return truncateSample(err.Error())
}
