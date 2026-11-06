// Package service llmops：半结构化数据解析（JSON/YAML/XML 顶层字段抽取→编目元数据）
// 口径：JSON/YAML 共用 summarize（sigs.k8s.io/yaml 将 YAML 转为 JSON 形状）；
// XML 走 encoding/xml 顶层元素遍历；字段数上限 200、样本值截断 64 rune；数组按首元素取字段。
package service

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	yamlLib "sigs.k8s.io/yaml"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeSemiParse     = 2216
	ErrCodeSemiNoField   = 2217
	semiMaxFields        = 200
	semiSampleMaxRune    = 64
	semiMaxBytes   int64 = 10 << 20
)

// SemiField 抽取出的顶层字段
type SemiField struct {
	Name   string `json:"name"`
	Type   string `json:"type"` // string/number/bool/array/object/null/element
	Sample string `json:"sample,omitempty"`
}

// jsonTypeOf 值→类型标签
func jsonTypeOf(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case string:
		return "string"
	case float64, int, int64, json.Number:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// truncateSample 样本值截断（64 rune + …）
func truncateSample(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > semiSampleMaxRune {
		return string(r[:semiSampleMaxRune]) + "…"
	}
	return string(r)
}

// summarizeAny 顶层字段抽取（map 直接展开按名排序保编目稳定；数组取首元素字段）
func summarizeAny(v any) []SemiField {
	switch t := v.(type) {
	case map[string]any:
		fields := make([]SemiField, 0, len(t))
		for name, val := range t {
			f := SemiField{Name: name, Type: jsonTypeOf(val), Sample: truncateSample(fmt.Sprintf("%v", val))}
			if val == nil {
				f.Sample = ""
			}
			fields = append(fields, f)
		}
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		if len(fields) > semiMaxFields {
			fields = fields[:semiMaxFields]
		}
		return fields
	case []any:
		if len(t) == 0 {
			return nil
		}
		fields := summarizeAny(t[0]) // 数组口径：取首元素字段
		for i := range fields {
			fields[i].Type = fields[i].Type + "[]" // 标注数组语义
		}
		return fields
	default:
		return nil
	}
}

// DetectSemiFormat 半结构化格式探测：json/yaml/xml/空串(未知)
func DetectSemiFormat(content []byte) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(string(content), "\xef\xbb\xbf"))
	switch {
	case strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["):
		return "json"
	case strings.HasPrefix(trimmed, "<"):
		return "xml"
	case trimmed == "":
		return ""
	default:
		return "yaml"
	}
}

// ParseJSONSchema JSON 顶层字段（对象→字段；数组→首元素字段）
func ParseJSONSchema(content []byte) ([]SemiField, error) {
	var v any
	if err := json.Unmarshal(content, &v); err != nil {
		return nil, fmt.Errorf("[%d]JSON 解析失败: %w", ErrCodeSemiParse, err)
	}
	return summarizeAny(v), nil
}

// ParseYAMLSchema YAML 顶层字段（sigs.k8s.io/yaml 转 JSON 后共用 summarize）
func ParseYAMLSchema(content []byte) ([]SemiField, error) {
	var v any
	if err := yamlLib.Unmarshal(content, &v); err != nil {
		return nil, fmt.Errorf("[%d]YAML 解析失败: %w", ErrCodeSemiParse, err)
	}
	return summarizeAny(v), nil
}

// ParseXMLSchema XML 顶层元素（名称 + 首段文本样本）
func ParseXMLSchema(content []byte) ([]SemiField, error) {
	dec := xml.NewDecoder(bytes.NewReader(content))
	var fields []SemiField
	depth := 0
	var curName string
	var sample strings.Builder
	inTop := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("[%d]XML 解析失败: %w", ErrCodeSemiParse, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 { // 根元素本身不算字段
				depth++
				continue
			}
			if depth == 1 {
				if len(fields) >= semiMaxFields {
					return fields, nil
				}
				curName = t.Name.Local
				sample.Reset()
				inTop = true
			}
			depth++
		case xml.EndElement:
			depth--
			if depth == 1 && inTop {
				fields = append(fields, SemiField{Name: curName, Type: "element", Sample: truncateSample(sample.String())})
				inTop = false
			}
		case xml.CharData:
			if inTop && depth == 2 { // 仅顶层子元素直属文本，嵌套子树文本不混入样本
				sample.Write(t)
			}
		}
	}
	return fields, nil
}

// ImportSemiStruct 半结构化数据编目落库（格式自动探测，字段清单入 metaJSON）
func (s *LlmopsServiceStruct) ImportSemiStruct(providedName string, content []byte, tags string) (*model.LlmopsDataAsset, error) {
	if int64(len(content)) > semiMaxBytes {
		return nil, fmt.Errorf("[%d]内容超过 %dMB 上限", ErrCodeSemiParse, semiMaxBytes>>20)
	}
	format := DetectSemiFormat(content)
	if format == "" {
		return nil, fmt.Errorf("[%d]内容为空", ErrCodeImportEmpty)
	}
	var fields []SemiField
	var err error
	switch format {
	case "json":
		fields, err = ParseJSONSchema(content)
	case "yaml":
		fields, err = ParseYAMLSchema(content)
	default:
		fields, err = ParseXMLSchema(content)
	}
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("[%d]未抽取到顶层字段", ErrCodeSemiNoField)
	}
	meta, _ := json.Marshal(map[string]any{
		"parseMode":  "semi-" + format,
		"fieldCount": len(fields),
		"fields":     fields,
	})
	asset := &model.LlmopsDataAsset{
		Name:     BuildDocName(providedName, "semi.json"),
		DataType: "semistructured",
		Source:   "import",
		Status:   "parsed",
		MetaJSON: string(meta),
		Tags:     strings.TrimSpace(tags),
		Size:     int64(len(content)),
	}
	if err := global.GVA_DB.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("[%d]编目入库失败: %w", ErrCodeImportFailed, err)
	}
	return asset, nil
}
