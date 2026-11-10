// Package service eduadmin：数据清洗（需求 3.6-2）
// 口径：乱码/格式错误过滤（UTF-8 乱码率>30% 剔除，对齐 N04 口径）、缺失值填充
// （数值字段 均值/前值 两策略可配）、3σ 异常值检测标记（payload 写 _outlier 字段）、
// 批处理逐条报告；解析与统计为纯函数可测，dry 试运行不改库。
package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// 错误码（eduadmin 段 2800-2899）
const ErrCodeEduClean = 2803

// FillStrategies 缺失值填充策略白名单
var FillStrategies = []string{"均值", "前值"}

// CleanReq 清洗请求
type CleanReq struct {
	Domain   string   `json:"domain"`
	Fields   []string `json:"fields"`
	Strategy string   `json:"strategy"`
	DryRun   bool     `json:"dryRun"`
}

// mojibakeRate 乱码率：非法 UTF-8 RuneError 占比（纯函数，测试锚点）
func mojibakeRate(s string) float64 {
	if s == "" {
		return 0
	}
	bad := 0
	for _, r := range s {
		if r == utf8.RuneError {
			bad++
		}
	}
	runes := utf8.RuneCountInString(s)
	if runes == 0 {
		return 0
	}
	return float64(bad) / float64(runes)
}

// numericOf 提取数值字段值（json.Number/float64 可解析；其余视为缺失）
func numericOf(v any) (float64, bool) {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case float64:
		return x, true
	case string:
		var f float64
		if err := json.Unmarshal([]byte(strings.TrimSpace(x)), &f); err == nil {
			return f, true
		}
	}
	return 0, false
}

// fillMissing 缺失值填充（纯函数，测试锚点）：rows 为逐条 payload 对象；
// 返回更新后的 rows 与填充计数。均值=现存数值均值；前值=行序上一数值（无前值则跳过）。
func fillMissing(rows []map[string]any, fields []string, strategy string) (int, error) {
	valid := false
	for _, s := range FillStrategies {
		if s == strategy {
			valid = true
		}
	}
	if !valid {
		return 0, fmt.Errorf("[%d]填充策略须为 均值/前值", ErrCodeEduClean)
	}
	filled := 0
	for _, field := range fields {
		var vals []float64
		present := make([]bool, len(rows))
		for i, r := range rows {
			v, ok := r[field]
			if ok && v != nil {
				if f, isNum := numericOf(v); isNum {
					vals = append(vals, f)
					present[i] = true
				}
			}
		}
		if len(vals) == 0 {
			continue
		}
		switch strategy {
		case "均值":
			mean := 0.0
			for _, v := range vals {
				mean += v
			}
			mean = math.Round(mean/float64(len(vals))*100) / 100
			for i := range rows {
				if !present[i] {
					rows[i][field] = mean
					filled++
				}
			}
		case "前值":
			last := vals[0]
			for i := range rows {
				if present[i] {
					if f, _ := numericOf(rows[i][field]); true {
						last = f
					}
					continue
				}
				rows[i][field] = last
				filled++
			}
		}
	}
	return filled, nil
}

// sigmaOutliers 3σ 异常检测（纯函数，测试锚点）：|x-均值|>3σ 的下标集合；σ=0 视为无异常
func sigmaOutliers(vals []float64) map[int]bool {
	out := map[int]bool{}
	if len(vals) < 2 {
		return out
	}
	mean := 0.0
	for _, v := range vals {
		mean += v
	}
	mean /= float64(len(vals))
	var ss float64
	for _, v := range vals {
		ss += (v - mean) * (v - mean)
	}
	sigma := math.Sqrt(ss / float64(len(vals)))
	if sigma == 0 {
		return out
	}
	for i, v := range vals {
		if math.Abs(v-mean) > 3*sigma {
			out[i] = true
		}
	}
	return out
}

// eduRow 一条待清洗记录（payload 解析后的对象）
type eduRow struct {
	rec model.EduRecord
	row map[string]any
}

// CleanEduRecords 批量清洗（域过滤；dry 试运行不改库）
func (s *EduadminService) CleanEduRecords(in CleanReq) (map[string]any, error) {
	if !ValidEduDomain(in.Domain) {
		return nil, fmt.Errorf("[%d]域须为人事/财务/资产/后勤", ErrCodeEduClean)
	}
	if len(in.Fields) == 0 {
		return nil, fmt.Errorf("[%d]数值字段列表必填", ErrCodeEduClean)
	}
	if in.Strategy == "" {
		in.Strategy = "均值"
	}
	var records []model.EduRecord
	if err := global.GVA_DB.Where("domain = ?", in.Domain).Order("id ASC").Limit(5000).Find(&records).Error; err != nil {
		return nil, err
	}
	rows := make([]eduRow, 0, len(records))
	mojibake := make([]uint, 0)
	formatErrs := make([]uint, 0)
	for _, rec := range records {
		if mojibakeRate(rec.PayloadJSON) > 0.3 || !validPayloadJSON(rec.PayloadJSON) {
			mojibake = append(mojibake, rec.ID)
			continue
		}
		var m map[string]any
		dec := json.NewDecoder(strings.NewReader(rec.PayloadJSON))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			formatErrs = append(formatErrs, rec.ID)
			continue
		}
		rows = append(rows, eduRow{rec: rec, row: m})
	}
	filled, err := fillMissing(rowsToMaps(rows), in.Fields, in.Strategy)
	if err != nil {
		return nil, err
	}
	outlierIds := make([]uint, 0)
	for _, field := range in.Fields {
		vals := make([]float64, len(rows))
		for i := range rows {
			f, ok := numericOf(rows[i].row[field])
			vals[i] = f
			if !ok {
				vals[i] = math.NaN()
			}
		}
		idx := outlierIndexes(vals)
		for _, i := range idx {
			ids := rows[i].rec.ID
			outlierIds = append(outlierIds, ids)
			if !in.DryRun {
				mark := rows[i].row["_outlier"]
				arr, _ := mark.([]any)
				rows[i].row["_outlier"] = append(arr, field)
			}
		}
	}
	if !in.DryRun {
		for i := range rows {
			b, _ := json.Marshal(rows[i].row)
			if err := global.GVA_DB.Model(&model.EduRecord{}).Where("id = ?", rows[i].rec.ID).
				Update("payload_json", string(b)).Error; err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(outlierIds, func(i, j int) bool { return outlierIds[i] < outlierIds[j] })
	return map[string]any{
		"scanned":     len(records),
		"mojibakeIds": mojibake,
		"formatErrIds": formatErrs,
		"filled":      filled,
		"outlierIds":  outlierIds,
		"dryRun":      in.DryRun,
		"strategy":    in.Strategy,
	}, nil
}

func rowsToMaps(rows []eduRow) []map[string]any {
	out := make([]map[string]any, len(rows))
	for i := range rows {
		out[i] = rows[i].row
	}
	return out
}

// outlierIndexes 忽略 NaN 后的 3σ 异常下标（升序）
func outlierIndexes(vals []float64) []int {
	clean := make([]float64, 0, len(vals))
	for _, v := range vals {
		if !math.IsNaN(v) {
			clean = append(clean, v)
		}
	}
	out := sigmaOutliers(clean)
	idxs := make([]int, 0, len(out))
	// clean 下标→原下标映射：NaN 槽位剔除
	j := 0
	for i, v := range vals {
		if math.IsNaN(v) {
			continue
		}
		if out[j] {
			idxs = append(idxs, i)
		}
		j++
	}
	return idxs
}
