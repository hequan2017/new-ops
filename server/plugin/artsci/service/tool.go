// Package service artsci：工具注册表（CRUD + module 分发执行 + 留痕；内置实现 M03+ 逐条注册）
package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// 错误码（artsci 段 2500-2599）
const (
	ErrCodeToolNotFound  = 2500
	ErrCodeToolSubject   = 2501
	ErrCodeToolSchema    = 2502
	ErrCodeToolNameDup   = 2503
	ErrCodeToolArgs      = 2504
	ErrCodeToolNoModule  = 2505
	ErrCodeToolDisabled  = 2506
	ErrCodeToolArgsParse = 2507
)

// runLogMaxBytes 出参留痕截断上限（2MB）
const runLogMaxBytes = 2 << 20

// builtinModules 内置实现注册表（module 标识 → 执行函数；M03+ 求解器逐条注册）
var builtinModules = map[string]func(map[string]any) (any, error){}

// RegisterModule 注册内置实现模块（init 期调用，幂等覆盖）
func RegisterModule(module string, fn func(map[string]any) (any, error)) {
	builtinModules[module] = fn
}

// ArtsciApp 服务单例出口（工具域服务聚合）
var ArtsciApp = new(ArtsciService)

type ArtsciService struct{}

// GetTools 工具列表（学科/关键字过滤）
func (s *ArtsciService) GetTools(subject, keyword string) ([]model.ArtsciTool, error) {
	db := global.GVA_DB.Model(&model.ArtsciTool{})
	if subject != "" {
		if !ValidSubject(subject) {
			return nil, fmt.Errorf("[%d]不支持的学科: %s", ErrCodeToolSubject, subject)
		}
		db = db.Where("subject = ?", subject)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("name LIKE ? OR desc LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	var list []model.ArtsciTool
	err := db.Order("subject, id").Find(&list).Error
	return list, err
}

// SaveTool 新建/更新工具
func (s *ArtsciService) SaveTool(in *model.ArtsciTool) (*model.ArtsciTool, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Module) == "" || !ValidSubject(in.Subject) {
		return nil, fmt.Errorf("[%d]名称/模块标识必填且学科须在白名单内", ErrCodeToolSubject)
	}
	if !ValidStatus(in.Status) {
		in.Status = "enabled"
	}
	if err := ValidateSchemaObject(in.ParamsSchema); err != nil {
		return nil, err
	}
	var count int64
	q := global.GVA_DB.Model(&model.ArtsciTool{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]工具名已存在", ErrCodeToolNameDup)
	}
	if in.ID > 0 {
		var tool model.ArtsciTool
		if err := global.GVA_DB.First(&tool, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]工具不存在", ErrCodeToolNotFound)
		}
		err := global.GVA_DB.Model(&tool).Updates(map[string]any{
			"name": in.Name, "subject": in.Subject, "module": in.Module,
			"params_schema": in.ParamsSchema, "status": in.Status, "desc": in.Desc,
		}).Error
		if err != nil {
			return nil, err
		}
		return &tool, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteTool 删除工具（软删除）
func (s *ArtsciService) DeleteTool(id uint) error {
	res := global.GVA_DB.Delete(&model.ArtsciTool{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]工具不存在", ErrCodeToolNotFound)
	}
	return nil
}

// dispatch 按 module 分发内置实现（schema 校验→注册表查找→执行；无 DB 依赖，测试锚点）
func (s *ArtsciService) dispatch(module, schemaRaw, argsJSON string) (any, error) {
	args := map[string]any{}
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("[%d]参数须为 JSON 对象: %w", ErrCodeToolArgsParse, err)
		}
	}
	if err := ValidateArgsAgainstSchema(schemaRaw, args); err != nil {
		return nil, err
	}
	fn, ok := builtinModules[module]
	if !ok {
		return nil, fmt.Errorf("[%d]内置实现未注册: %s", ErrCodeToolNoModule, module)
	}
	return fn(args)
}

// RunTool 执行工具（校验→分发→留痕，成败均落 ArtsciRunLog）
func (s *ArtsciService) RunTool(toolId uint, argsJSON string) (*model.ArtsciRunLog, error) {
	var tool model.ArtsciTool
	if err := global.GVA_DB.First(&tool, toolId).Error; err != nil {
		return nil, fmt.Errorf("[%d]工具不存在", ErrCodeToolNotFound)
	}
	if tool.Status != "enabled" {
		return nil, fmt.Errorf("[%d]工具已停用", ErrCodeToolDisabled)
	}
	log := &model.ArtsciRunLog{ToolId: tool.ID, Name: tool.Name, ArgsJSON: argsJSON}
	started := time.Now()
	result, execErr := s.dispatch(tool.Module, tool.ParamsSchema, argsJSON)
	log.LatencyMs = time.Since(started).Milliseconds()
	log.Ok = execErr == nil
	if execErr != nil {
		log.ErrMsg = truncateRunText(execErr.Error())
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

// GetRunLogs 执行留痕（工具过滤，倒序分页）
func (s *ArtsciService) GetRunLogs(toolId uint, page, pageSize int) ([]model.ArtsciRunLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ArtsciRunLog{})
	if toolId > 0 {
		db = db.Where("tool_id = ?", toolId)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ArtsciRunLog
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// truncateRunText 错误留痕截断（500 字节口径，防超列宽）
func truncateRunText(s string) string {
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
