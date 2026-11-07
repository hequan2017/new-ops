// Package service oplib：算子执行引擎（简化 schema 校验 + 执行 + 留痕）
// 口径：schema 简化协议 {"字段":{"type":"string|number|bool","required":true,"enum":[...]}}；
// builtin 走注册表，script 执行器 H05 接入；成败均落 OplibRunLog（结果 2MB 截断）。
package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
)

// 错误码（oplib 段续用）
const (
	ErrCodeOpArgs   = 2308
	ErrCodeOpExec   = 2309
	runLogMaxBytes  = 2 << 20
)

// fieldSchema 字段约束
type fieldSchema struct {
	Type     string   `json:"type"`   // string/number/bool
	Required bool     `json:"required"`
	Enum     []any    `json:"enum"`
}

// ValidateArgsAgainstSchema 简化 schema 校验（纯函数，测试锚点）
func ValidateArgsAgainstSchema(schemaRaw string, args map[string]any) error {
	if strings.TrimSpace(schemaRaw) == "" {
		return nil
	}
	var schema map[string]fieldSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return fmt.Errorf("[%d]schema 解析失败: %w", ErrCodeOpSchema, err)
	}
	for field, fs := range schema {
		v, present := args[field]
		if !present || v == nil {
			if fs.Required {
				return fmt.Errorf("[%d]缺少必填参数: %s", ErrCodeOpArgs, field)
			}
			continue
		}
		switch fs.Type {
		case "string":
			if _, ok := v.(string); !ok {
				return fmt.Errorf("[%d]参数 %s 须为字符串", ErrCodeOpArgs, field)
			}
		case "number":
			if _, ok := v.(float64); !ok {
				return fmt.Errorf("[%d]参数 %s 须为数字", ErrCodeOpArgs, field)
			}
		case "bool":
			if _, ok := v.(bool); !ok {
				return fmt.Errorf("[%d]参数 %s 须为布尔", ErrCodeOpArgs, field)
			}
		}
		if len(fs.Enum) > 0 {
			ok := false
			for _, e := range fs.Enum {
				if fmt.Sprintf("%v", e) == fmt.Sprintf("%v", v) {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("[%d]参数 %s 不在枚举内", ErrCodeOpArgs, field)
			}
		}
	}
	return nil
}

// runBuiltin 执行内置算子（名称→注册表；argsJSON→map）
func runBuiltin(name, argsJSON string) (any, error) {
	fn, ok := GetBuiltin(name)
	if !ok {
		return nil, fmt.Errorf("[%d]内置算子未注册: %s", ErrCodeOpExec, name)
	}
	args := map[string]any{}
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("[%d]参数须为 JSON 对象: %w", ErrCodeOpArgs, err)
		}
	}
	return fn(args)
}

// ExecuteOperator 执行算子（校验→执行→留痕，成败均落 OplibRunLog）
func (s *OplibService) ExecuteOperator(operatorId uint, argsJSON string) (*model.OplibRunLog, error) {
	var op model.OplibOperator
	if err := global.GVA_DB.First(&op, operatorId).Error; err != nil {
		return nil, fmt.Errorf("[%d]算子不存在", ErrCodeOpNotFound)
	}
	// 参数预解析供 schema 校验
	args := map[string]any{}
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("[%d]参数须为 JSON 对象: %w", ErrCodeOpArgs, err)
		}
	}
	if err := ValidateArgsAgainstSchema(op.InputSchema, args); err != nil {
		return nil, err
	}
	log := &model.OplibRunLog{OperatorId: op.ID, Name: op.Name, ArgsJSON: argsJSON}
	started := time.Now()
	var result any
	var execErr error
	switch op.ImplType {
	case "builtin":
		result, execErr = runBuiltin(op.Name, argsJSON)
	case "script":
		result, execErr = RunScriptOperator(&op, argsJSON)
	default:
		execErr = fmt.Errorf("[%d]未知实现类型: %s", ErrCodeOpExec, op.ImplType)
	}
	log.LatencyMs = time.Since(started).Milliseconds()
	log.Ok = execErr == nil
	if execErr != nil {
		log.ErrMsg = truncateSample(execErr.Error())
	} else {
		b, _ := json.Marshal(result)
		res := string(b)
		if len(res) > runLogMaxBytes {
			res = res[:runLogMaxBytes]
		}
		log.ResultJSON = res
	}
	if err := global.GVA_DB.Create(log).Error; err != nil {
		return nil, err
	}
	if execErr != nil {
		return log, execErr
	}
	return log, nil
}

// GetRunLogs 执行留痕列表（算子过滤，倒序分页）
func (s *OplibService) GetRunLogs(operatorId uint, page, pageSize int) ([]model.OplibRunLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.OplibRunLog{})
	if operatorId > 0 {
		db = db.Where("operator_id = ?", operatorId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.OplibRunLog
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
