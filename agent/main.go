// 白泽轻量 Agent（M9 A1）：单文件二进制，出站 WebSocket 反向连接 server。
// 注册（接入令牌绑定资产）→ 心跳 + 系统指标上报（30s 周期）→ 断线退避重连（1s→10s）。
// 用法：baize-agent -server http://192.168.112.138:8888 -token <接入令牌> [-interval 30s] [-hostname xx]
package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const agentVersion = "0.1.0"

type agentPoint struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type frame struct {
	Type   string       `json:"type"`
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

var (
	flagServer   = flag.String("server", "", "server 基地址（http://host:8888）")
	flagToken    = flag.String("token", "", "接入令牌（资产页 Agent 接入签发）")
	flagInterval = flag.Duration("interval", 30*time.Second, "心跳与指标上报周期")
	flagHostname = flag.String("hostname", "", "上报主机名（默认取 os.Hostname）")
)

func main() {
	flag.Parse()
	if *flagServer == "" || *flagToken == "" {
		flag.Usage()
		os.Exit(1)
	}
	hostname := *flagHostname
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		}
	}
	meta := agentMeta{Version: agentVersion, Os: runtime.GOOS, Arch: runtime.GOARCH, Hostname: hostname}

	wsURL := buildWSURL(*flagServer)
	log.Printf("baize-agent %s → %s (interval %s)", agentVersion, wsURL, *flagInterval)

	backoff := time.Second
	for {
		err := runSession(wsURL, *flagToken, meta, *flagInterval)
		if rr, ok := err.(*registerRejectedError); ok {
			// 令牌被拒不会自愈：打印后退出（等人工换令牌再启动）
			log.Fatalf("注册被拒（%s），退出——请在资产页重新签发令牌", rr.reason)
		}
		if err == nil {
			backoff = time.Second
			continue
		}
		log.Printf("会话断开: %v，%s 后重连", err, backoff)
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 10*time.Second {
			backoff = 10 * time.Second
		}
	}
}

// buildWSURL server 基地址 → WS 地址（http→ws / https→wss，拼 /agent/ws）
func buildWSURL(server string) string {
	server = strings.TrimRight(strings.TrimSpace(server), "/")
	switch {
	case strings.HasPrefix(server, "https://"):
		server = "wss://" + strings.TrimPrefix(server, "https://")
	case strings.HasPrefix(server, "http://"):
		server = "ws://" + strings.TrimPrefix(server, "http://")
	}
	return server + "/agent/ws"
}

// runSession 一次完整会话：拨号 → 注册 → 心跳/指标循环；连接断开返回 error
func runSession(wsURL, token string, meta agentMeta, interval time.Duration) error {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	var wmu sync.Mutex
	send := func(f frame) error {
		wmu.Lock()
		defer wmu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteJSON(f)
	}

	// 注册
	if err := send(frame{Type: "register", Token: token, TS: time.Now().Unix(), Agent: meta}); err != nil {
		return err
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var ack frame
	if err := conn.ReadJSON(&ack); err != nil {
		return err
	}
	if ack.Type != "register_ok" {
		return &registerRejectedError{reason: ack.Msg}
	}
	log.Printf("注册成功: %s@%s", meta.Hostname, meta.Os+"/"+meta.Arch)

	// 读循环：消化 server ping（默认 pong 应答），读超时即判死链
	readErr := make(chan error, 1)
	go func() {
		for {
			_ = conn.SetReadDeadline(time.Now().Add(2 * interval))
			if _, _, err := conn.ReadMessage(); err != nil {
				readErr <- err
				return
			}
		}
	}()

	collector := &Collector{Interval: interval.Seconds()}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case err := <-readErr:
			return err
		case <-ticker.C:
			if err := send(frame{Type: "heartbeat", TS: time.Now().Unix()}); err != nil {
				return err
			}
			sample := collector.Sample(readSnapshot())
			if err := send(frame{Type: "metric", TS: time.Now().Unix(), Points: sample.Points()}); err != nil {
				return err
			}
		}
	}
}

type registerRejectedError struct{ reason string }

func (e *registerRejectedError) Error() string { return "注册被拒: " + e.reason }
