// Package service llmops：模型调用日志 CSV 导出（备用池 P08，最后一条）
// 口径：全量导出（模型/协议过滤），倒序；CSV 列 时间,模型,协议,时延ms,输入,输出,总tokens,结果,错误；
// 时间 RFC3339 带时区（与 I07 成绩导出口径一致用 fmt 拼接）。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// callLogsToCSV 调用日志 CSV（纯函数，测试锚点）
func callLogsToCSV(rows []model.LlmopsModelCallLog) string {
	var sb strings.Builder
	sb.WriteString("时间,模型,协议,时延ms,输入tokens,输出tokens,总tokens,结果,错误\n")
	for _, r := range rows {
		result := "成功"
		if !r.Ok {
			result = "失败"
		}
		errField := r.Error
		errField = strings.ReplaceAll(errField, ",", "，")
		errField = strings.ReplaceAll(errField, "\n", " ")
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d,%d,%d,%s,%s\n",
			r.CreatedAt.Format("2006-01-02 15:04:05"), r.ModelName, r.Protocol,
			r.LatencyMs, r.PromptTokens, r.CompletionTokens, r.TotalTokens, result, errField))
	}
	return sb.String()
}

// ExportModelCallLogs 调用日志导出（过滤口径同列表）
func (s *LlmopsServiceStruct) ExportModelCallLogs(modelId uint, okFilter string) (string, error) {
	db := global.GVA_DB.Model(&model.LlmopsModelCallLog{})
	if modelId > 0 {
		db = db.Where("model_id = ?", modelId)
	}
	if okFilter == "true" || okFilter == "false" {
		db = db.Where("ok = ?", okFilter == "true")
	}
	var rows []model.LlmopsModelCallLog
	if err := db.Order("id desc").Limit(10000).Find(&rows).Error; err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("[%d]暂无调用日志", ErrCodeQuota)
	}
	return callLogsToCSV(rows), nil
}
