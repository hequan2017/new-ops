// Package service 白泽终端插件：WebSSH 桥接服务
// 桥接设计：WS 二进制帧 ⇄ SSH PTY 流；文本帧为控制消息（resize/ping/closed）。
// 链路：目标主机沿 JumpHostID 级联（≤5 跳、禁环，见 jump.go），逐跳指纹校验/TOFU 录入。
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	cssh "golang.org/x/crypto/ssh"

	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/term/model"
	"go.uber.org/zap"
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
	UserID       uint
	ClientIP     string
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

	// 级联拨号：目标主机沿 JumpHostID 收集跳板链（≤5 跳、禁环），
	// 最外层直连、内层逐跳建隧道；逐跳独立认证与指纹校验（TOFU 回填）
	chain, err := buildJumpChain(host, assetHostSvc.GetAssetHost)
	if err != nil {
		return err
	}
	clients, hostFP, err := t.dialChain(chain, p.CredentialID, p.Operator)
	if err != nil {
		return err
	}
	client := clients[len(clients)-1] // 最内层即目标主机
	defer func() {
		for _, c := range clients {
			_ = c.Close() // 中间跳板是内层通道载体，一并关闭
		}
	}()

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
			for _, c := range clients {
				_ = c.Close()
			}
			_ = ws.Close()
		})
	}
	defer closeAll()

	// 会话审计：元数据落库 + 流镜像 + 命令抽取（审计不可用不阻断终端）
	auditSvc := new(TermAuditService)
	sess := &model.TermSession{
		HostID: p.HostID, Hostname: host.Hostname, IP: host.IP,
		UserID: p.UserID, Username: p.Operator, CredentialID: p.CredentialID,
		ClientIP: p.ClientIP, Cols: cols, Rows: rows,
		Status: model.TermSessionActive, StartedAt: time.Now(),
		Fingerprint: hostFP,
	}
	var sessionID uint
	if auditErr := auditSvc.StartSession(sess); auditErr != nil {
		zap.L().Warn("term 会话审计创建失败: " + auditErr.Error())
	} else {
		sessionID = sess.ID
	}
	defer func() {
		if sessionID > 0 {
			auditSvc.EndSession(sessionID)
		}
	}()
	streamSeq := int64(0)
	cmdAcc := newInputAccumulator()
	recordStream := func(direction uint8, payload string) {
		if sessionID == 0 {
			return
		}
		streamSeq++
		_ = auditSvc.AppendStream(sessionID, streamSeq, direction, payload)
	}

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
				if sessionID > 0 {
					recordStream(model.StreamDirDown, string(buf[:n]))
				}
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
			input := string(data)
			if sessionID > 0 {
				recordStream(model.StreamDirUp, input)
				for _, cmd := range cmdAcc.feed(input) {
					_ = auditSvc.AppendCommand(sessionID, streamSeq, cmd)
				}
			}
			_, _ = stdin.Write([]byte(input)) // 非控制 JSON 的文本按终端输入处理
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
