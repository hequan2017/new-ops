// Package service 白泽 Agent 通道：网关（M9 A1）
// 职责：接入令牌签发/校验、WS 注册-心跳-指标上报接入、连接注册表（A2 任务下发预留）、
// 在线状态按心跳时效计算（心跳 30s，90s 无心跳判离线，不落库）。
// 指标直接落 monitor_metric（复用 30 天留存与前端 PerfChart 链路）。
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hequan2017/new-ops/server/global"
	model "github.com/hequan2017/new-ops/server/plugin/agent/model"
	monitorModel "github.com/hequan2017/new-ops/server/plugin/monitor/model"
	"gorm.io/gorm"
)

// agent 错误码段：2000-2199（DEV_PLAN 3.6，agent/org/aiops 共用段）
const (
	ErrCodeAgentParamInvalid = 2001
	ErrCodeAgentAssetMissing = 2002
	ErrCodeAgentTokenInvalid = 2003
	ErrCodeAgentUnauthorized = 2004
)

// Agent 在线判定：心跳间隔 30s，3 个周期无心跳判离线
const (
	AgentHeartbeatInterval = 30 * time.Second
	AgentOnlineTTL         = 90 * time.Second
	registerTimeout        = 5 * time.Second
)

var agentUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// AgentService Agent 网关服务
type AgentService struct{}

var AgentSvc = new(AgentService)

// agentConn 已注册连接（A2 任务下发预留路由位）
type agentConn struct {
	AssetID uint
	Conn    *websocket.Conn
	WriteMu sync.Mutex
}

// registry 连接注册表
var registry = struct {
	sync.Mutex
	conns map[uint]*agentConn
}{conns: map[uint]*agentConn{}}

// ---- 消息信封（agent ⇄ server JSON 帧） ----

type agentFrame struct {
	Type   string       `json:"type"` // register / heartbeat / metric / register_ok / register_fail
	Token  string       `json:"token,omitempty"`
	TS     int64        `json:"ts,omitempty"`
	Msg    string       `json:"msg,omitempty"`
	Agent  agentMeta    `json:"agent,omitempty"`
	Points []agentPoint `json:"points,omitempty"`
}

type agentMeta struct {
	Version  string `json:"version"`
	Os       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
}

type agentPoint struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// ---- 纯逻辑（可单测） ----

// HashAgentToken 接入令牌 SHA256（存储口径）
func HashAgentToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// AgentStatusByHeartbeat 按最近心跳计算状态（纯逻辑）
func AgentStatusByHeartbeat(last *time.Time, now time.Time) string {
	if last == nil || now.Sub(*last) > AgentOnlineTTL {
		return "离线"
	}
	return "在线"
}

// agentMetricWhitelist 指标白名单：只接收与 monitor 约定的指标名
var agentMetricWhitelist = map[string]bool{
	monitorModel.MetricCPUPercent:  true,
	monitorModel.MetricMemPercent:  true,
	monitorModel.MetricDiskPercent: true,
	monitorModel.MetricLoad1:       true,
	monitorModel.MetricNetRxKBs:    true,
	monitorModel.MetricNetTxKBs:    true,
}

// sanitizeMetricPoints 指标白名单过滤 + 数值有限性校验（纯逻辑）
func sanitizeMetricPoints(points []agentPoint) []agentPoint {
	out := make([]agentPoint, 0, len(points))
	for _, p := range points {
		if !agentMetricWhitelist[strings.ToLower(strings.TrimSpace(p.Name))] {
			continue
		}
		if p.Value != p.Value || p.Value > 1e12 || p.Value < -1e12 { // NaN / 越界防护
			continue
		}
		out = append(out, agentPoint{Name: strings.ToLower(strings.TrimSpace(p.Name)), Value: p.Value})
	}
	return out
}

// GenerateAgentToken 生成 32 字节随机接入令牌
func GenerateAgentToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// ---- 业务接口 ----

// IssueAgentToken 为资产签发/重置接入令牌（明文仅本次返回，落库仅哈希）
func (s *AgentService) IssueAgentToken(assetID uint, labels string) (*model.AgentInstance, string, error) {
	if assetID == 0 {
		return nil, "", newAgentErr(ErrCodeAgentParamInvalid, "assetId 必填")
	}
	var inst model.AgentInstance
	err := global.GVA_DB.Where("asset_id = ?", assetID).First(&inst).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, "", err
	}
	token, gerr := GenerateAgentToken()
	if gerr != nil {
		return nil, "", gerr
	}
	if err == gorm.ErrRecordNotFound {
		inst = model.AgentInstance{AssetID: assetID, Labels: labels}
	} else if labels != "" {
		inst.Labels = labels
	}
	inst.TokenHash = HashAgentToken(token)
	inst.TokenHint = token[len(token)-4:]
	inst.Version = ""
	inst.LastHeartbeat = nil
	if err := global.GVA_DB.Save(&inst).Error; err != nil {
		return nil, "", err
	}
	return &inst, token, nil
}

// RevokeAgentToken 吊销注册（删除实例登记，Agent 下次注册即被拒）
func (s *AgentService) RevokeAgentToken(id uint) error {
	return global.GVA_DB.Delete(&model.AgentInstance{}, id).Error
}

// AgentInstanceView 实例视图（状态按心跳计算）
type AgentInstanceView struct {
	model.AgentInstance
	Status        string `json:"status"`
	AssetHostname string `json:"assetHostname"`
	AssetIP       string `json:"assetIp"`
}

// ListAgentInstances 实例列表（联资产主机名/IP，状态计算）
func (s *AgentService) ListAgentInstances() ([]AgentInstanceView, error) {
	var list []model.AgentInstance
	if err := global.GVA_DB.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	type hostRow struct {
		ID       uint
		Hostname string
		IP       string
	}
	var hosts []hostRow
	global.GVA_DB.Table("asset_hosts").Select("id, hostname, ip").Find(&hosts)
	hostMap := map[uint]hostRow{}
	for _, h := range hosts {
		hostMap[h.ID] = h
	}
	now := time.Now()
	out := make([]AgentInstanceView, 0, len(list))
	for _, a := range list {
		h := hostMap[a.AssetID]
		out = append(out, AgentInstanceView{
			AgentInstance: a,
			Status:        AgentStatusByHeartbeat(a.LastHeartbeat, now),
			AssetHostname: h.Hostname,
			AssetIP:       h.IP,
		})
	}
	return out, nil
}

// ---- WS 网关 ----

// HandleWS Agent WebSocket 接入（鉴权=首帧 register 携带接入令牌，5s 内完成）
func (s *AgentService) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := agentUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(registerTimeout))
	var first agentFrame
	if err := conn.ReadJSON(&first); err != nil {
		return
	}
	if first.Type != "register" || first.Token == "" {
		_ = conn.WriteJSON(agentFrame{Type: "register_fail", TS: time.Now().Unix(), Msg: "首帧必须为携带令牌的 register"})
		return
	}
	var inst model.AgentInstance
	if err := global.GVA_DB.Where("token_hash = ?", HashAgentToken(first.Token)).First(&inst).Error; err != nil {
		_ = conn.WriteJSON(agentFrame{Type: "register_fail", TS: time.Now().Unix(), Msg: "令牌无效或已吊销"})
		return
	}
	now := time.Now()
	global.GVA_DB.Model(&inst).Updates(map[string]any{
		"version": first.Agent.Version, "os_arch": first.Agent.Os + "/" + first.Agent.Arch,
		"hostname": first.Agent.Hostname, "last_heartbeat": &now,
	})
	if err := conn.WriteJSON(agentFrame{Type: "register_ok", TS: now.Unix()}); err != nil {
		return
	}

	ac := &agentConn{AssetID: inst.AssetID, Conn: conn}
	registry.Lock()
	registry.conns[inst.AssetID] = ac
	registry.Unlock()
	defer func() {
		registry.Lock()
		delete(registry.conns, inst.AssetID)
		registry.Unlock()
	}()

	// 服务端周期 ping 保活（客户端自动 pong）；读超时=2 个心跳周期
	pingDone := make(chan struct{})
	defer close(pingDone)
	go pingLoop(ac, pingDone)
	_ = conn.SetReadDeadline(time.Now().Add(2 * AgentHeartbeatInterval))
	for {
		var f agentFrame
		if err := conn.ReadJSON(&f); err != nil {
			return
		}
		now := time.Now()
		switch f.Type {
		case "heartbeat":
			global.GVA_DB.Model(&inst).Update("last_heartbeat", &now)
		case "metric":
			global.GVA_DB.Model(&inst).Update("last_heartbeat", &now)
			s.storeMetrics(inst.AssetID, f.Points)
		default:
			// 未知类型忽略（向前兼容）
		}
		_ = conn.SetReadDeadline(time.Now().Add(2 * AgentHeartbeatInterval))
	}
}

// pingLoop 周期 ping 保活（经 WriteMu 与业务帧互斥）
func pingLoop(ac *agentConn, done <-chan struct{}) {
	ticker := time.NewTicker(AgentHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ac.WriteMu.Lock()
			_ = ac.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_ = ac.Conn.WriteMessage(websocket.PingMessage, nil)
			ac.WriteMu.Unlock()
		}
	}
}

// storeMetrics 指标批量落 monitor_metric（白名单过滤；本批时间统一，排序仅用于稳定写入）
func (s *AgentService) storeMetrics(assetID uint, points []agentPoint) {
	clean := sanitizeMetricPoints(points)
	if len(clean) == 0 {
		return
	}
	sort.Slice(clean, func(i, j int) bool { return clean[i].Name < clean[j].Name })
	now := time.Now()
	rows := make([]monitorModel.MonitorMetric, 0, len(clean))
	for _, p := range clean {
		rows = append(rows, monitorModel.MonitorMetric{AssetID: assetID, Name: p.Name, Value: p.Value, TS: now})
	}
	global.GVA_DB.Create(&rows)
}

// WriteToAgent 向已注册连接下发 JSON 帧（A2 任务下发预留出口；A1 仅心跳链路无服务端推送）
func WriteToAgent(assetID uint, v any) error {
	registry.Lock()
	ac := registry.conns[assetID]
	registry.Unlock()
	if ac == nil {
		return fmt.Errorf("agent 未在线: asset %d", assetID)
	}
	ac.WriteMu.Lock()
	defer ac.WriteMu.Unlock()
	_ = ac.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return ac.Conn.WriteJSON(v)
}

func newAgentErr(code int, msg string) error { return &agentError{Code: code, Msg: msg} }

type agentError struct {
	Code int
	Msg  string
}

func (e *agentError) Error() string { return e.Msg }
