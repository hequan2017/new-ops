// Package service 白泽终端插件：WebSSH 桥接服务
// 桥接设计：WS 二进制帧 ⇄ SSH PTY 流；文本帧为控制消息（resize/ping/closed）。
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	cssh "golang.org/x/crypto/ssh"

	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
)

// TermService 终端服务
var TermService = new(termService)

type termService struct{}

// TermStartParams WebSSH 启动参数
type TermStartParams struct {
	HostID       uint
	CredentialID uint
	Cols         int
	Rows         int
	Operator     string
}

// StartWebSSH 建立 WS ⇄ SSH 桥接（阻塞直到任一端断开）
func (t *termService) StartWebSSH(ws *websocket.Conn, p TermStartParams) error {
	assetHostSvc := assetSvc.Service.AssetHostService
	host, err := assetHostSvc.GetAssetHost(p.HostID)
	if err != nil {
		return fmt.Errorf("主机不存在: %w", err)
	}
	if host.IP == "" {
		return fmt.Errorf("主机缺少内网IP")
	}
	credSvc := assetSvc.Service.CredCredentialService
	secret, credType, username, err := credSvc.GetPlaintext(p.CredentialID)
	if err != nil {
		return fmt.Errorf("凭据读取失败: %w", err)
	}
	if credType != assetModel.CredTypeSSHPassword && credType != assetModel.CredTypeSSHKey {
		return fmt.Errorf("WebSSH 仅支持 SSH 密码/私钥凭据")
	}

	client, err := assetSvc.DialSSH(host.IP, assetSvc.SSHAuth{
		Username:   username,
		Password:   secret,
		PrivateKey: secret,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	cols, rows := p.Cols, p.Rows
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}
	if err := session.RequestPty("xterm-256color", rows, cols, ssh_TerminalModes()); err != nil {
		return fmt.Errorf("申请 PTY 失败: %w", err)
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	session.Stderr = session.Stdout // 合并 stderr
	if err := session.Shell(); err != nil {
		return fmt.Errorf("启动 shell 失败: %w", err)
	}

	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			_ = session.Close()
			_ = client.Close()
			_ = ws.Close()
		})
	}
	defer closeAll()

	// SSH 输出 → WS（二进制帧）
	go func() {
		buf := make([]byte, 8192)
		for {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				deadline := time.Now().Add(5 * time.Second)
				if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					break
				}
				_ = ws.SetWriteDeadline(deadline)
			}
			if readErr != nil {
				if readErr != io.EOF {
					_ = ws.WriteControl(websocket.CloseMessage,
						websocket.FormatCloseMessage(4000, "SSH 连接已断开"), time.Now().Add(5*time.Second))
				}
				break
			}
		}
	}()

	// WS 消息 → SSH（文本帧=控制/输入，二进制帧=输入）
	ws.SetReadLimit(1 << 20)
	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			return nil // 客户端断开属正常退出
		}
		switch msgType {
		case websocket.TextMessage:
			var ctl struct {
				Type string `json:"type"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			if jsonErr := json.Unmarshal(data, &ctl); jsonErr == nil && ctl.Type != "" {
				switch ctl.Type {
				case "resize":
					_ = session.WindowChange(ctl.Rows, ctl.Cols)
				case "ping":
					_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
				case "close":
					return nil
				}
				continue
			}
			_, _ = stdin.Write(data) // 非控制 JSON 的文本按终端输入处理
		case websocket.BinaryMessage:
			_, _ = stdin.Write(data)
		}
	}
}

// ssh_TerminalModes PTY 终端模式
func ssh_TerminalModes() cssh.TerminalModes {
	return cssh.TerminalModes{
		cssh.ECHO:          1,
		cssh.ICANON:        1,
		cssh.ISIG:          1,
		cssh.ICRNL:         1,
		cssh.TTY_OP_ISPEED: 14400,
		cssh.TTY_OP_OSPEED: 14400,
	}
}
