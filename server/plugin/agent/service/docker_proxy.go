// Package service 白泽 Agent 通道：Docker HTTP 代理（M9 A4）
// AgentDockerTransport 实现 http.RoundTripper：把 Docker SDK 的 HTTP 请求经 WS 通道
// 转发到 agent 所在主机的 docker daemon（agent 侧 handleDockerReq），响应按 taskNo 关联回来。
// 用法：client.NewClientWithOpts(
//
//	client.WithHTTPClient(&http.Client{Transport: &agentSvc.AgentDockerTransport{AssetID: id}}),
//	client.WithHost("http://docker.proxy"))
//
// v1 边界：仅管理面（inspect/list/create/start/stop/rm/stats 一次性等）；
// 日志流/exec/事件订阅/镜像拉取等流式端点不支持（仍走直连接入点）。
package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// dockerProxyWait 服务端等待时限（略大于 agent 侧 60s 执行超时）
const dockerProxyWait = 70 * time.Second

// dockerProxyBodyLimit 请求体上限（docker 管理 API 请求体为小 JSON）
const dockerProxyBodyLimit = 4 << 20

// AgentDockerTransport 经 agent WS 通道的 HTTP 代理传输（docker_req 默认；FrameType=k8s_req 时
// 由集群内 agent 以 ServiceAccount 身份代理 K8s API，K4）
type AgentDockerTransport struct {
	AssetID   uint
	FrameType string // 空=docker_req；k8s_req=集群内 agent 代理
}

func (t *AgentDockerTransport) frameType() string {
	if t.FrameType == "" {
		return "docker_req"
	}
	return t.FrameType
}

// RoundTrip 实现 http.RoundTripper（SDK 每次调用即一次 WS 往返）
func (t *AgentDockerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		b, err := io.ReadAll(io.LimitReader(req.Body, dockerProxyBodyLimit+1))
		if err != nil {
			return nil, err
		}
		if len(b) > dockerProxyBodyLimit {
			return nil, fmt.Errorf("docker 代理请求体超限")
		}
		body = b
	}
	taskNo := uuid.NewString()
	ch := make(chan agentFrame, 1)
	pendingTasks.Store(taskNo, ch)
	defer pendingTasks.Delete(taskNo)

	header := map[string][]string{}
	for k, vs := range req.Header {
		if strings.EqualFold(k, "Host") {
			continue
		}
		header[k] = vs
	}
	if err := WriteToAgent(t.AssetID, agentFrame{
		Type: t.frameType(), TS: time.Now().Unix(), TaskNo: taskNo,
		ReqMethod: req.Method, ReqPath: req.URL.RequestURI(), ReqHeader: header,
		ReqBody: base64.StdEncoding.EncodeToString(body),
	}); err != nil {
		return nil, err
	}

	timer := time.NewTimer(dockerProxyWait)
	defer timer.Stop()
	select {
	case f := <-ch:
		if f.Msg != "" {
			return nil, fmt.Errorf("agent docker 代理失败: %s", f.Msg)
		}
		if f.RespStatus == 0 {
			return nil, fmt.Errorf("agent docker 代理响应异常（无状态码）")
		}
		decoded, err := base64.StdEncoding.DecodeString(f.RespBody)
		if err != nil {
			return nil, fmt.Errorf("agent docker 响应解码失败: %w", err)
		}
		hdr := http.Header{}
		for k, vs := range f.RespHeader {
			for _, v := range vs {
				hdr.Add(k, v)
			}
		}
		return &http.Response{
			StatusCode: f.RespStatus,
			Status:     fmt.Sprintf("%d %s", f.RespStatus, http.StatusText(f.RespStatus)),
			Header:     hdr,
			Body:       io.NopCloser(bytes.NewReader(decoded)),
			Request:    req,
		}, nil
	case <-timer.C:
		return nil, fmt.Errorf("agent docker 代理超时（%s）", dockerProxyWait)
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}
}
