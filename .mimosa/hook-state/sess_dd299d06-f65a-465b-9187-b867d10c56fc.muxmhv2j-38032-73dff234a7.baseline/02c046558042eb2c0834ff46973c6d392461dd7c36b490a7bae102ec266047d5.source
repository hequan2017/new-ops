// Package service 白泽终端插件：远程日志 tail
// 复用 WebSSH 的级联拨号能力（跳板+指纹校验/TOFU），在远端执行
// `tail -n N -f -- <path>` 并把输出流推给 WebSocket（二进制帧）。
package service

import (
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"

	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
)

// LogTailParams 日志 tail 启动参数
type LogTailParams struct {
	HostID       uint
	CredentialID uint
	Path         string
	Lines        int // 初始回看行数（1-2000，默认 200）
	Operator     string
}

// StartLogTail 建立 WS ⇄ `tail -f` 桥接（阻塞直到任一端断开）
func (t *termService) StartLogTail(ws *websocket.Conn, p LogTailParams) error {
	assetHostSvc := assetSvc.Service.AssetHostService
	host, err := assetHostSvc.GetAssetHost(p.HostID)
	if err != nil {
		return fmt.Errorf("主机不存在: %w", err)
	}
	if host.IP == "" {
		return fmt.Errorf("主机缺少内网IP")
	}
	if p.Path == "" || p.Path[0] != '/' {
		return fmt.Errorf("日志路径必须为绝对路径（以 / 开头）")
	}
	lines := p.Lines
	if lines <= 0 {
		lines = 200
	}
	if lines > 2000 {
		lines = 2000
	}

	chain, err := buildJumpChain(host, assetHostSvc.GetAssetHost)
	if err != nil {
		return err
	}
	clients, _, err := t.dialChain(chain, p.CredentialID, p.Operator)
	if err != nil {
		return err
	}
	client := clients[len(clients)-1]
	closeAll := func() {
		for _, c := range clients {
			_ = c.Close()
		}
		_ = ws.Close()
	}
	defer closeAll()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	// -- 终止选项解析，防路径以 - 开头注入；文件不存在时 tail 持续等待与 -f 语义一致
	cmd := fmt.Sprintf("tail -n %d -F -- %q", lines, p.Path)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("启动 tail 失败: %w", err)
	}

	// SSH 输出 → WS（二进制帧）；客户端 close 文本帧 → 退出
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
				_ = ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
			}
			if readErr != nil {
				if readErr != io.EOF {
					_ = ws.WriteControl(websocket.CloseMessage,
						websocket.FormatCloseMessage(4000, "tail 已断开"), time.Now().Add(5*time.Second))
				}
				return
			}
		}
	}()

	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			break // 客户端断开
		}
		if msgType == websocket.TextMessage && len(data) >= 16 && string(data[:7]) == `{"type"` {
			if string(data) == `{"type":"close"}` {
				break
			}
		}
	}
	_ = session.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = session.Signal("KILL")
		_ = session.Close()
		<-done
	}
	return nil
}
