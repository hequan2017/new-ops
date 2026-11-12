// Package service artsci：参数 schema 校验（口径对齐 oplib：{"字段":{"type":"string|number|bool","required":true,"enum":[...]}}）
package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

// fieldSchema 单字段 schema
type fieldSchema struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Enum     []any  `json:"enum"`
}

// ValidateArgsAgainstSchema 简化 schema 校验（纯函数，测试锚点）
func ValidateArgsAgainstSchema(schemaRaw string, args map[string]any) error {
	if strings.TrimSpace(schemaRaw) == "" {
		return nil
	}
	var schema map[string]fieldSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return artWrapErr(ErrCodeToolSchema, "schema 解析失败", err)
	}
	for field, fs := range schema {
		v, present := args[field]
		if !present || v == nil {
			if fs.Required {
				return artErr(ErrCodeToolArgs, "缺少必填参数: %s", field)
			}
			continue
		}
		switch fs.Type {
		case "string":
			if _, ok := v.(string); !ok {
				return artErr(ErrCodeToolArgs, "参数 %s 须为字符串", field)
			}
		case "number":
			if _, ok := v.(float64); !ok {
				return artErr(ErrCodeToolArgs, "参数 %s 须为数字", field)
			}
		case "bool":
			if _, ok := v.(bool); !ok {
				return artErr(ErrCodeToolArgs, "参数 %s 须为布尔", field)
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
				return artErr(ErrCodeToolArgs, "参数 %s 不在枚举内", field)
			}
		}
	}
	return nil
}

// ValidateSchemaObject 参数 schema 须为 JSON 对象（保存时校验）
func ValidateSchemaObject(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return artWrapErr(ErrCodeToolSchema, "schema 须为 JSON 对象", err)
	}
	return nil
}

// ValidSubject 学科白名单
func ValidSubject(s string) bool {
	switch s {
	case "数学", "文科", "图像":
		return true
	}
	return false
}

// ValidStatus 状态白名单
func ValidStatus(s string) bool {
	return s == "enabled" || s == "disabled"
}
