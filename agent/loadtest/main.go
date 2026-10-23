// 白泽 Agent 通道压测工具（M9 验收：单实例承载 N Agent 并发心跳与指标上报）
// 流程：登录 → 批量创建资产+签发令牌 → 并发 N 虚拟 agent（WS 注册→周期心跳+指标）→
// 汇总注册/心跳统计与在线数 → 清理（吊销实例+删除资产）。
// 用法：go run ./loadtest -server http://127.0.0.1:8888 -user admin -pass xxx -n 100 -duration 60s
package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type frame struct {
	Type   string `json:"type"`
	Token  string `json:"token,omitempty"`
	TS     int64  `json:"ts,omitempty"`
	Msg    string `json:"msg,omitempty"`
	Agent  agentMeta
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

var (
	flagServer   = flag.String("server", "http://127.0.0.1:8888", "server 基地址")
	flagUser     = flag.String("user", "admin", "管理员用户名")
	flagPass     = flag.String("pass", "", "管理员密码")
	flagN        = flag.Int("n", 100, "虚拟 agent 数量")
	flagDuration = flag.Duration("duration", 60*time.Second, "压测持续时长")
	flagInterval = flag.Duration("interval", 3*time.Second, "心跳周期（压测用短周期）")
	flagKeep     = flag.Bool("keep", false, "结束后保留资产与实例（调试用）")
)

func main() {
	flag.Parse()
	if *flagPass == "" || *flagServer == "" {
		flag.Usage()
		os.Exit(1)
	}
	token, err := login(*flagServer, *flagUser, *flagPass)
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	sweepLeftovers(token)
	fmt.Printf("登录成功，开始准备 %d 个虚拟 agent（持续 %s，心跳 %s）\n", *flagN, *flagDuration, *flagInterval)

	assets, err := setupAgents(token, *flagN)
	if err != nil {
		log.Fatalf("准备失败: %v", err)
	}
	fmt.Printf("资产+令牌就绪: %d/%d\n", len(assets), *flagN)

	var regOK, regFail, beats atomic.Int64
	done := make(chan struct{})
	var wg sync.WaitGroup
	start := time.Now()
	for _, a := range assets {
		wg.Add(1)
		go func(assetID uint, tk string) {
			defer wg.Done()
			runVirtualAgent(*flagServer, assetID, tk, *flagInterval, *flagDuration, &regOK, &regFail, &beats, done)
		}(a.assetID, a.token)
	}
	time.Sleep(*flagDuration)
	close(done)
	wg.Wait()
	elapsed := time.Since(start)

	online := queryOnline(*flagServer, token)
	fmt.Println("---- 压测结果 ----")
	fmt.Printf("目标=%d 注册成功=%d 注册失败=%d 心跳发送=%d 用时=%s\n",
		*flagN, regOK.Load(), regFail.Load(), beats.Load(), elapsed.Truncate(time.Millisecond))
	fmt.Printf("压测结束时 server 在线 agent=%d（期望≈%d）\n", online, regOK.Load())
	if !*flagKeep {
		cleanup(*flagServer, token, assets)
		fmt.Println("清理完成（实例吊销+资产删除）")
	}
	if regFail.Load() > 0 || int64(online) < regOK.Load()*9/10 {
		fmt.Println("结论: ⚠️ 未达验收口径（注册失败>0 或在线数掉线>10%）")
		os.Exit(2)
	}
	fmt.Printf("结论: ✅ 单实例承载 %d 并发 agent 心跳/指标上报通过\n", regOK.Load())
}

type agentHandle struct {
	assetID    uint
	instanceID uint
	token      string
}

// setupAgents 批量创建资产+签发令牌（顺序执行，避免 SQLite 写锁竞争放大）
func setupAgents(token string, n int) ([]agentHandle, error) {
	out := make([]agentHandle, 0, n)
	for i := 1; i <= n; i++ {
		body := fmt.Sprintf(`{"hostname":"baize-loadtest-%03d","ip":"10.99.0.%d","status":"运行中"}`, i, i%240+10)
		resp, err := apiJSON(http.MethodPost, "/asset/host/create", token, body)
		if err != nil {
			return out, fmt.Errorf("创建资产 %d 失败: %w", i, err)
		}
		if resp["code"].(float64) != 0 {
			return out, fmt.Errorf("创建资产 %d 被拒: %v", i, resp["msg"])
		}
		// 从列表取 ID（create 响应不带 ID）
		lr, err := apiJSON(http.MethodPost, "/asset/host/list", token, `{"page":1,"pageSize":1,"keyword":"baize-loadtest-`+fmt.Sprintf("%03d", i)+`"}`)
		if err != nil {
			return out, err
		}
		id, err := extractFirstID(lr["data"])
		if err != nil {
			return out, fmt.Errorf("查询资产 %d 失败: %w", i, err)
		}
		tres, err := apiJSON(http.MethodPost, "/agent/instance/token", token, fmt.Sprintf(`{"assetId":%v}`, id))
		if err != nil {
			return out, err
		}
		tdata, _ := tres["data"].(map[string]any)
		tk, _ := tdata["token"].(string)
		instID, _ := tdata["id"].(float64)
		if tk == "" {
			return out, fmt.Errorf("资产 %v 令牌签发失败", id)
		}
		out = append(out, agentHandle{assetID: uint(id.(float64)), instanceID: uint(instID), token: tk})
	}
	return out, nil
}

// runVirtualAgent 单个虚拟 agent：注册→周期心跳+指标；退出靠 done 或致命错误
func runVirtualAgent(server string, assetID uint, token string, interval, duration time.Duration, regOK, regFail, beats *atomic.Int64, done <-chan struct{}) {
	deadline := time.Now().Add(duration + 10*time.Second)
	conn, err := dialWS(toWSURL(server)+"/agent/ws", token)
	if err != nil {
		if regFail.Load() < 5 {
			fmt.Printf("注册失败样例: %v\n", err)
		}
		regFail.Add(1)
		return
	}
	defer conn.Close()
	regOK.Add(1)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// 读侧不消费（server 无业务下行即忽略）；写心跳+指标
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if time.Now().After(deadline) {
				return
			}
			if err := conn.WriteJSON(frame{Type: "heartbeat", TS: time.Now().Unix()}); err != nil {
				return
			}
			beats.Add(1)
			_ = conn.WriteJSON(frame{Type: "metric", TS: time.Now().Unix(), Points: fakePoints(assetID)})
		}
	}
}

// fakePoints 压测假指标（白名单内，按资产号微调避免全同）
func fakePoints(assetID uint) []agentPoint {
	base := float64(assetID%50) / 2
	return []agentPoint{
		{Name: "cpu_percent", Value: base},
		{Name: "mem_percent", Value: 30 + base},
		{Name: "load1", Value: base / 10},
	}
}

// toWSURL server 基地址 http(s) → ws(s)（与 agent 二进制 buildWSURL 同口径）
func toWSURL(server string) string {
	server = strings.TrimRight(strings.TrimSpace(server), "/")
	switch {
	case strings.HasPrefix(server, "https://"):
		return "wss://" + strings.TrimPrefix(server, "https://")
	case strings.HasPrefix(server, "http://"):
		return "ws://" + strings.TrimPrefix(server, "http://")
	}
	return server
}

func dialWS(wsURL, token string) (*websocket.Conn, error) {
	d := websocket.Dialer{HandshakeTimeout: 10 * time.Second, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	conn, _, err := d.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteJSON(frame{
		Type: "register", Token: token, TS: time.Now().Unix(),
		Agent: agentMeta{Version: "loadtest", Os: "virt", Arch: "virt", Hostname: fmt.Sprintf("loadtest-%d", time.Now().UnixNano()%100000)},
	}); err != nil {
		return nil, err
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var ack frame
	if err := conn.ReadJSON(&ack); err != nil {
		return nil, err
	}
	if ack.Type != "register_ok" {
		return nil, fmt.Errorf("注册被拒: %s", ack.Msg)
	}
	_ = conn.SetReadDeadline(time.Time{})
	return conn, nil
}

// queryOnline 当前在线 agent 数
func queryOnline(server, token string) int {
	resp, err := apiJSON(http.MethodGet, "/agent/instance/list", token, "")
	if err != nil {
		return -1
	}
	list, _ := resp["data"].([]any)
	n := 0
	for _, it := range list {
		m, _ := it.(map[string]any)
		if m["status"] == "在线" {
			n++
		}
	}
	return n
}

// cleanup 吊销实例+删除资产（尽力而为；资产删除为 DELETE /asset/host/delete + JSON body）
func cleanup(server, token string, assets []agentHandle) {
	for _, a := range assets {
		_, _ = apiJSON(http.MethodDelete, fmt.Sprintf("/agent/instance?id=%d", a.instanceID), token, "")
		_, _ = apiJSON(http.MethodDelete, "/asset/host/delete", token, fmt.Sprintf(`{"id":%d}`, a.assetID))
	}
}

// sweepLeftovers 清理历史残留的压测资产（关键字 baize-loadtest-，上轮清理失败时自愈）
func sweepLeftovers(token string) {
	lr, err := apiJSON(http.MethodPost, "/asset/host/list", token, `{"page":1,"pageSize":200,"keyword":"baize-loadtest-"}`)
	if err != nil {
		return
	}
	l, _ := lr["data"].(map[string]any)
	list, _ := l["list"].([]any)
	for _, it := range list {
		m, _ := it.(map[string]any)
		if id, _ := m["ID"].(float64); id > 0 {
			_, _ = apiJSON(http.MethodDelete, "/asset/host/delete", token, fmt.Sprintf(`{"id":%v}`, id))
		}
	}
	if n := len(list); n > 0 {
		fmt.Printf("已清理历史残留压测资产 %d 台\n", n)
	}
}

// ---- HTTP 小工具 ----

func login(server, user, pass string) (string, error) {
	body := fmt.Sprintf(`{"username":%q,"password":%q,"captcha":"","captchaId":""}`, user, pass)
	resp, err := apiJSON(http.MethodPost, "/base/login", "", body)
	if err != nil {
		return "", err
	}
	if resp["code"].(float64) != 0 {
		return "", fmt.Errorf("%v", resp["msg"])
	}
	data, _ := resp["data"].(map[string]any)
	tk, _ := data["token"].(string)
	if tk == "" {
		return "", fmt.Errorf("响应无 token")
	}
	return tk, nil
}

func apiJSON(method, path, token, body string) (map[string]any, error) {
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, *flagServer+path, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("x-token", token)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("响应非 JSON: %s", truncate(string(b), 120))
	}
	return out, nil
}

func extractFirstID(data any) (any, error) {
	l, _ := data.(map[string]any)
	list, _ := l["list"].([]any)
	if len(list) == 0 {
		return nil, fmt.Errorf("列表为空")
	}
	m, _ := list[0].(map[string]any)
	return m["ID"], nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
