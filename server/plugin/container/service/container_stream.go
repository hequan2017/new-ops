// Package service 白泽容器管理：日志流与 exec 终端（WS 桥接，M4 C1 收官）
// 通道约定与 term 插件一致：二进制帧=数据流，文本帧=JSON 控制（resize/ping/close）；
// SSE/WS 端点挂 public 组，query token 自验。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/websocket"
)

// wsCtl 控制帧结构
type wsCtl struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// readControl 读一帧：控制 JSON 解析返回 ctl，其余数据帧原样带回（用户输入）
func readControl(ws *websocket.Conn) (ctl wsCtl, data []byte, err error) {
	msgType, payload, err := ws.ReadMessage()
	if err != nil {
		return wsCtl{}, nil, err
	}
	if msgType == websocket.TextMessage && len(payload) > 0 && payload[0] == '{' {
		_ = json.Unmarshal(payload, &ctl)
		if ctl.Type != "" {
			return ctl, payload, nil
		}
	}
	return wsCtl{}, payload, nil
}

// LogsWS 容器日志流桥接（docker logs -f → WS；阻塞直到任一端断开）
func (s *EndpointService) LogsWS(ws *websocket.Conn, endpointID uint, containerID string, tail int) {
	if tail <= 0 {
		tail = 200
	}
	if tail > 5000 {
		tail = 5000
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, err.Error()), time.Now().Add(5*time.Second))
		return
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, err.Error()), time.Now().Add(5*time.Second))
		return
	}
	defer cl.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TTY 容器输出为裸流；非 TTY 为多路复用（8 字节帧头）需 stdcopy 去复用
	inspect, iErr := cl.ContainerInspect(ctx, containerID)
	if iErr != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, "容器不存在: "+iErr.Error()), time.Now().Add(5*time.Second))
		return
	}
	logsR, err := cl.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true, ShowStderr: true, Follow: true,
		Tail: fmt.Sprintf("%d", tail),
	})
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, "日志流建立失败: "+err.Error()), time.Now().Add(5*time.Second))
		return
	}
	defer logsR.Close()

	var streamR io.Reader = logsR
	if !inspect.Config.Tty {
		pr, pw := io.Pipe()
		go func() {
			// stdout/stderr 合并推送（终端场景无需分色）
			_, _ = stdcopy.StdCopy(pw, pw, logsR)
			_ = pw.Close()
		}()
		defer pr.Close()
		streamR = pr
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			n, readErr := streamR.Read(buf)
			if n > 0 {
				if wErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					return
				}
				_ = ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
			}
			if readErr != nil {
				return
			}
		}
	}()
	// 读循环仅做关闭/心跳（日志流无输入侧）
	for {
		ctl, _, err := readControl(ws)
		if err != nil {
			break
		}
		switch ctl.Type {
		case "close":
			cancel()
		case "ping":
			_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		}
	}
	<-done
}

// ExecWS 容器 exec 交互终端桥接（hijack ⇄ WS；阻塞直到任一端断开）
func (s *EndpointService) ExecWS(ws *websocket.Conn, endpointID uint, containerID string, cols, rows int, operator string) {
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}
	ep, err := s.GetEndpoint(endpointID)
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, err.Error()), time.Now().Add(5*time.Second))
		return
	}
	cl, err := s.NewDockerClient(ep)
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, err.Error()), time.Now().Add(5*time.Second))
		return
	}
	defer cl.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	execResp, err := cl.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin: true, AttachStdout: true, AttachStderr: true,
		Tty: true, // TTY 模式单流裸数据，免去 stdcopy 去复用
		Cmd: []string{"/bin/sh"},
		Env: []string{"TERM=xterm-256color"},
	})
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, "exec 创建失败: "+err.Error()), time.Now().Add(5*time.Second))
		return
	}
	attach, err := cl.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, "exec 连接失败: "+err.Error()), time.Now().Add(5*time.Second))
		return
	}
	defer attach.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			n, readErr := attach.Reader.Read(buf)
			if n > 0 {
				if wErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()
	for {
		ctl, data, err := readControl(ws)
		if err != nil {
			break
		}
		switch ctl.Type {
		case "resize":
			if ctl.Cols > 0 && ctl.Rows > 0 {
				cols, rows = ctl.Cols, ctl.Rows
			}
			_ = cl.ContainerExecResize(ctx, execResp.ID, container.ResizeOptions{
				Width: uint(cols), Height: uint(rows),
			})
		case "ping":
			_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
		case "close":
			cancel()
		default:
			// 数据帧：用户输入转发到 exec 连接
			if len(data) > 0 {
				_, _ = attach.Conn.Write(data)
			}
		}
	}
	<-done
}
