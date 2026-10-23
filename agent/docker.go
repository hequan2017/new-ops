// 白泽轻量 Agent：Docker HTTP 代理（M9 A4）
// 收到 server 下发的 docker_req → 对本机 docker daemon HTTP API（unix socket / tcp）发起同构请求 →
// 响应缓冲回传 docker_resp（base64）。v1 边界：仅覆盖管理面（inspect/list/create/start/stop/rm/stats
// 一次性等），不支持日志流/exec/事件订阅/镜像拉取等流式端点；响应上限 8MB。
// k8s_req 同构复用（见 k8s.go，K4 集群 ServiceAccount 通道）。
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// maxDockerRespBytes 单响应缓冲上限
const maxDockerRespBytes = 8 << 20

// dockerProxyTimeout 单请求超时（管理面操作秒级，留足余量）
const dockerProxyTimeout = 60 * time.Second

// dockerSem 并发保护：同时处理的代理请求数上限（docker/k8s 共用）
var dockerSem = make(chan struct{}, 8)

// dockerSockCandidates 默认 socket 探测顺序（linux → macOS Docker Desktop）
var dockerSockCandidates = []string{
	"/var/run/docker.sock",
	"/run/docker.sock",
	os.Getenv("HOME") + "/.docker/run/docker.sock",
}

// resolveDockerAddr 默认 daemon 地址探测：首个存在的 socket；否则回退 linux 常规路径（连接时报错更直观）
func resolveDockerAddr(candidates []string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if fi, err := os.Stat(p); err == nil && fi.Mode()&os.ModeSocket != 0 {
			return "unix://" + p
		}
	}
	return "unix:///var/run/docker.sock"
}

// proxyReply 代理回执（status 0 = 代理层错误，语义见 Msg）
func proxyReply(f frame, send func(frame) error, replyType string, status int, header map[string][]string, body []byte, errMsg string) {
	_ = send(frame{
		Type: replyType, TS: time.Now().Unix(), TaskNo: f.TaskNo,
		RespStatus: status, RespHeader: header, RespBody: base64.StdEncoding.EncodeToString(body),
		Msg: errMsg,
	})
}

// handleDockerReq 处理 docker_req：定位 daemon 地址后走通用 HTTP 代理
func handleDockerReq(f frame, send func(frame) error, addr string) {
	client, err := dockerHTTPClient(addr)
	if err != nil {
		proxyReply(f, send, "docker_resp", 0, nil, nil, "docker 地址不可用: "+err.Error())
		return
	}
	// unix 走假主机名（DialContext 忽略）；tcp 直用真实地址，避免假主机名无法解析/被环境代理劫持
	basePrefix := "http://docker.proxy"
	switch {
	case strings.HasPrefix(addr, "tcp://"):
		basePrefix = "http://" + strings.TrimPrefix(addr, "tcp://")
	case strings.HasPrefix(addr, "http://"):
		basePrefix = "http://" + strings.TrimPrefix(addr, "http://")
	}
	handleHTTPProxy("docker_resp", f, send, basePrefix, client, "")
}

// handleHTTPProxy 通用 HTTP-over-WS 代理：请求帧 → 同构请求 → 缓冲回执（按 taskNo 关联）
func handleHTTPProxy(replyType string, f frame, send func(frame) error, basePrefix string, client *http.Client, bearerFile string) {
	if f.TaskNo == "" || f.ReqMethod == "" || f.ReqPath == "" {
		return
	}
	select {
	case dockerSem <- struct{}{}:
		defer func() { <-dockerSem }()
	default:
		proxyReply(f, send, replyType, 0, nil, nil, "代理并发占满")
		return
	}
	var bodyReader io.Reader
	if f.ReqBody != "" {
		b, derr := base64.StdEncoding.DecodeString(f.ReqBody)
		if derr != nil {
			proxyReply(f, send, replyType, 0, nil, nil, "请求体解码失败: "+derr.Error())
			return
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(context.Background(), f.ReqMethod, basePrefix+f.ReqPath, bodyReader)
	if err != nil {
		proxyReply(f, send, replyType, 0, nil, nil, "构造请求失败: "+err.Error())
		return
	}
	for k, vs := range f.ReqHeader {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if bearerFile != "" {
		if tok, rerr := os.ReadFile(bearerFile); rerr == nil {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(tok)))
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		proxyReply(f, send, replyType, 0, nil, nil, "请求失败: "+err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDockerRespBytes+1))
	if err != nil {
		proxyReply(f, send, replyType, 0, nil, nil, "读取响应失败: "+err.Error())
		return
	}
	if len(body) > maxDockerRespBytes {
		proxyReply(f, send, replyType, 0, nil, nil, "响应超过缓冲上限")
		return
	}
	proxyReply(f, send, replyType, resp.StatusCode, resp.Header, body, "")
}

// dockerHTTPClient 按 addr 构造 daemon HTTP 客户端（unix:// 走 socket，tcp:// 直连）
func dockerHTTPClient(addr string) (*http.Client, error) {
	if strings.HasPrefix(addr, "unix://") {
		sock := strings.TrimPrefix(addr, "unix://")
		return &http.Client{
			Timeout: dockerProxyTimeout,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.DialTimeout("unix", sock, 5*time.Second)
				},
			},
		}, nil
	}
	if strings.HasPrefix(addr, "tcp://") || strings.HasPrefix(addr, "http://") {
		// daemon 直连，禁用环境代理（假主机名会被代理拦成 502）
		return &http.Client{Timeout: dockerProxyTimeout, Transport: &http.Transport{Proxy: nil}}, nil
	}
	return nil, fmt.Errorf("不支持的 daemon 地址: %s", addr)
}
