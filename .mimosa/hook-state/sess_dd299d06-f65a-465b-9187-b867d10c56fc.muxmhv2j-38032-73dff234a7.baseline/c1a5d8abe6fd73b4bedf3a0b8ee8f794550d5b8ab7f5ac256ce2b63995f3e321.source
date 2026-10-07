// Package service 白泽终端插件：ProxyJump 跳板级联拨号（go-webssh 路线）
// 链路模型：目标主机沿 JumpHostID 自引用向上收集跳板链（≤5 层、禁环），
// 拨号从最外层跳板开始逐层向内：最外层真实 TCP 直连，其后每一跳经上一跳
// SSH client 建立隧道（client.Dial + NewClientConn），逐跳独立认证与指纹校验。
package service

import (
	"fmt"

	cssh "golang.org/x/crypto/ssh"

	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
)

// maxJumpDepth 跳板级联最大层数（不含目标主机，go-webssh 约定）
const maxJumpDepth = 5

// term 插件错误码段：1100-1199（DEV_PLAN 3.6）
const (
	ErrCodeJumpTooDeep      = 1101
	ErrCodeJumpLoop         = 1102
	ErrCodeJumpNoCredential = 1103
)

// TermError 带业务错误码的终端错误
type TermError struct {
	Code int
	Msg  string
}

func (e *TermError) Error() string { return e.Msg }

func newTermErr(code int, msg string) *TermError {
	return &TermError{Code: code, Msg: msg}
}

// buildJumpChain 从目标主机沿 JumpHostID 向上收集级联链，返回 [目标, 跳板1, ..., 跳板N]
// （拨号顺序为倒序，链尾是最外层入口）。lookup 注入取主机实现，纯逻辑可单测。
func buildJumpChain(target *assetModel.AssetHost, lookup func(id uint) (*assetModel.AssetHost, error)) ([]*assetModel.AssetHost, error) {
	chain := []*assetModel.AssetHost{target}
	visited := map[uint]bool{target.ID: true}
	cur := target
	for cur.JumpHostID != nil && *cur.JumpHostID != 0 {
		if len(chain)-1 >= maxJumpDepth {
			return nil, newTermErr(ErrCodeJumpTooDeep,
				fmt.Sprintf("跳板级联超过 %d 层上限（当前链：%s）", maxJumpDepth, chainDesc(chain)))
		}
		jump, err := lookup(*cur.JumpHostID)
		if err != nil {
			return nil, fmt.Errorf("跳板机不存在(id=%d): %w", *cur.JumpHostID, err)
		}
		if visited[jump.ID] {
			return nil, newTermErr(ErrCodeJumpLoop,
				fmt.Sprintf("跳板链成环：主机 %s(%s) 指回已遍历主机 %s(%s)", cur.Hostname, cur.IP, jump.Hostname, jump.IP))
		}
		if jump.IP == "" {
			return nil, fmt.Errorf("跳板机 %s(id=%d) 缺少内网IP", jump.Hostname, jump.ID)
		}
		visited[jump.ID] = true
		chain = append(chain, jump)
		cur = jump
	}
	return chain, nil
}

// chainDesc 链路描述（错误信息用）：目标→跳板1→…
func chainDesc(chain []*assetModel.AssetHost) string {
	s := ""
	for i, h := range chain {
		if i > 0 {
			s += "→"
		}
		s += h.IP
	}
	return s
}

// authForHop 装配某一跳认证：目标主机用用户所选凭据，跳板机用自身绑定凭据；
// 每跳都带该主机已录指纹实现强校验（未录则 TOFU 由 dialChain 回填）。
func (t *termService) authForHop(host *assetModel.AssetHost, isTarget bool, targetCredID uint) (assetSvc.SSHAuth, error) {
	credID := targetCredID
	if !isTarget {
		if host.CredentialID == nil || *host.CredentialID == 0 {
			return assetSvc.SSHAuth{}, newTermErr(ErrCodeJumpNoCredential,
				fmt.Sprintf("跳板机 %s(%s) 未绑定SSH凭据，无法用作跳板", host.Hostname, host.IP))
		}
		credID = *host.CredentialID
	}
	secret, credType, username, err := assetSvc.Service.CredCredentialService.GetPlaintext(credID)
	if err != nil {
		return assetSvc.SSHAuth{}, fmt.Errorf("凭据读取失败(id=%d): %w", credID, err)
	}
	if credType != assetModel.CredTypeSSHPassword && credType != assetModel.CredTypeSSHKey {
		return assetSvc.SSHAuth{}, fmt.Errorf("凭据 id=%d 非 SSH 密码/私钥类型，无法用于终端连接", credID)
	}
	return assetSvc.SSHAuth{
		Username:            username,
		Password:           secret,
		PrivateKey:         secret,
		ExpectedFingerprint: host.SSHFP,
	}, nil
}

// dialChain 按链逐层拨号：返回各跳 client（索引 0=最外层跳板 … 末位=目标）与
// 目标主机实际指纹（会话审计快照用）；全部 client 需由调用方在会话结束后关闭
// （中间层是内层通道的载体，不可提前关）。每跳连接成功且未录指纹时 TOFU 回填
// （回填失败不阻断，仅记日志）。
func (t *termService) dialChain(chain []*assetModel.AssetHost, targetCredID uint, operator string) ([]*cssh.Client, string, error) {
	n := len(chain)
	clients := make([]*cssh.Client, 0, n)
	targetFP := ""
	for i := n - 1; i >= 0; i-- {
		host := chain[i]
		auth, err := t.authForHop(host, i == 0, targetCredID)
		if err != nil {
			t.closeClients(clients)
			return nil, targetFP, err
		}
		var client *cssh.Client
		var fp string
		if len(clients) == 0 {
			// 最外层：真实 TCP 直连（含指纹校验）
			client, fp, err = assetSvc.DialSSH(host.IP, auth)
		} else {
			// 内层：经上一跳建立 TCP 隧道后在其上完成 SSH 握手
			addr := assetSvc.SSHDialTarget(host.IP, 0)
			conn, derr := clients[len(clients)-1].Dial("tcp", addr)
			if derr != nil {
				t.closeClients(clients)
				return nil, targetFP, fmt.Errorf("经跳板隧道拨 %s(%s) 失败: %w", host.Hostname, addr, derr)
			}
			var cfg *cssh.ClientConfig
			cfg, err = assetSvc.BuildSSHClientConfig(auth, &fp)
			if err != nil {
				_ = conn.Close()
				t.closeClients(clients)
				return nil, targetFP, err
			}
			ncc, chans, reqs, herr := cssh.NewClientConn(conn, addr, cfg)
			if herr != nil {
				_ = conn.Close()
				t.closeClients(clients)
				return nil, targetFP, herr // 指纹不匹配等握手错误，fp 已由回调写回
			}
			client = cssh.NewClient(ncc, chans, reqs)
		}
		if err != nil {
			t.closeClients(clients)
			return nil, targetFP, err
		}
		clients = append(clients, client)
		if i == 0 {
			targetFP = fp
		}
		if host.SSHFP == "" {
			_ = assetSvc.Service.AssetHostService.RecordHostFingerprint(host.ID, fp, operator)
		}
	}
	return clients, targetFP, nil
}

// closeClients 关闭已建立的链上连接（出错回滚用）
func (t *termService) closeClients(clients []*cssh.Client) {
	for _, c := range clients {
		_ = c.Close()
	}
}
