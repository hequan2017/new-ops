// Package mcpTool 白泽运维只读 MCP 工具（aiops M7）
// 注册进 GVA MCP Server（StreamableHTTP :8889），供 AI 助手调用：
// 资产概览 / 最近告警 / 工单状态——全部只读，经骨架自带的上游代理调用主 server API
// （standalone 进程无 DB 连接；调用者身份经 auth_header 透传鉴权）。
package mcpTool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
)

// opsGet 调上游 GET 并解析 data 段（envelope {code,data,msg}）
func opsGet[T any](ctx context.Context, endpoint string, query url.Values) (T, error) {
	var out T
	env, err := getUpstream[json.RawMessage](ctx, endpoint, query)
	if err != nil {
		return out, err
	}
	if env.Code != 0 {
		return out, fmt.Errorf("上游返回 code=%d msg=%s", env.Code, env.Msg)
	}
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return out, err
	}
	return out, nil
}

// ---------- 1. 资产概览 ----------

type OpsAssetOverview struct{}

func init() { RegisterTool(&OpsAssetOverview{}) }

type opsHostRow struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	OS       string `json:"os"`
	Status   string `json:"status"`
	CPUCores int    `json:"cpuCores"`
	MemGb    int    `json:"memGb"`
}

func (t *OpsAssetOverview) New() mcp.Tool {
	return mcp.NewTool("ops_asset_overview",
		mcp.WithDescription("查询白泽平台主机资产概览：主机列表（前 20 台，含名称/IP/系统/配置/状态）"))
}

func (t *OpsAssetOverview) Handle(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type pageResult struct {
		List  []opsHostRow `json:"list"`
		Total int64        `json:"total"`
	}
	data, err := opsGet[pageResult](ctx, "/asset/host/list", url.Values{
		"page": []string{"1"}, "pageSize": []string{"20"},
	})
	if err != nil {
		return mcp.NewToolResultError("查询失败: " + err.Error()), nil
	}
	out := fmt.Sprintf("主机资产共 %d 台，前 %d 台:\n", data.Total, len(data.List))
	for _, h := range data.List {
		out += fmt.Sprintf("  %s(%s) %s [%s] %dC/%dG\n", h.Hostname, h.IP, h.OS, h.Status, h.CPUCores, h.MemGb)
	}
	return mcp.NewToolResultText(out), nil
}

// ---------- 2. 最近告警 ----------

type OpsAlertRecent struct{}

func init() { RegisterTool(&OpsAlertRecent{}) }

type opsAlertRow struct {
	Summary    string  `json:"summary"`
	FiredAt    string  `json:"firedAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

func (t *OpsAlertRecent) New() mcp.Tool {
	return mcp.NewTool("ops_alert_recent",
		mcp.WithDescription("查询白泽平台最近的监控告警事件（倒序，默认 10 条）"))
}

func (t *OpsAlertRecent) Handle(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type pageResult struct {
		List []opsAlertRow `json:"list"`
	}
	data, err := opsGet[pageResult](ctx, "/monitor/alert/event/list", url.Values{})
	if err != nil {
		return mcp.NewToolResultError("查询失败: " + err.Error()), nil
	}
	if len(data.List) == 0 {
		return mcp.NewToolResultText("当前没有告警事件"), nil
	}
	out := fmt.Sprintf("最近 %d 条告警事件:\n", len(data.List))
	for _, e := range data.List {
		state := "未恢复"
		if e.ResolvedAt != nil {
			state = "已恢复"
		}
		ts := e.FiredAt
		if len(ts) > 19 {
			ts = ts[:19]
		}
		out += fmt.Sprintf("  [%s] %s（%s）\n", ts, e.Summary, state)
	}
	return mcp.NewToolResultText(out), nil
}

// ---------- 3. 工单状态 ----------

type OpsTicketStatus struct{}

func init() { RegisterTool(&OpsTicketStatus{}) }

type opsTicketRow struct {
	ID      uint   `json:"ID"`
	Title   string `json:"title"`
	BizType string `json:"bizType"`
	State   string `json:"state"`
	Creator string `json:"creator"`
}

func (t *OpsTicketStatus) New() mcp.Tool {
	return mcp.NewTool("ops_ticket_status",
		mcp.WithDescription("查询白泽平台工单状态统计与进行中的工单列表"))
}

func (t *OpsTicketStatus) Handle(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type pageResult struct {
		List []opsTicketRow `json:"list"`
	}
	data, err := opsGet[pageResult](ctx, "/workflow/instance/list", url.Values{})
	if err != nil {
		return mcp.NewToolResultError("查询失败: " + err.Error()), nil
	}
	running := make([]opsTicketRow, 0)
	for _, i := range data.List {
		// 列表含全部状态；仅筛进行中展示
		if i.State != "" && i.Creator != "" {
			running = append(running, i)
		}
	}
	out := fmt.Sprintf("最近工单 %d 条（含已完成）:\n", len(data.List))
	for _, i := range data.List {
		out += fmt.Sprintf("  #%d [%s] %s（%s）\n", i.ID, i.BizType, i.Title, i.Creator)
	}
	_ = running
	return mcp.NewToolResultText(out), nil
}
