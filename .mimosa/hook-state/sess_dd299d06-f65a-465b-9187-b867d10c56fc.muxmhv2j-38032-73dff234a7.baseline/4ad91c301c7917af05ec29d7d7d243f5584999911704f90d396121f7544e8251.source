package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"strings"
	"testing"

	cssh "golang.org/x/crypto/ssh"
)

// startTestSSHServer 起一个内存 SSH 服务（ed25519 临时主机密钥、任意口令放行），
// 返回监听地址、主机公钥指纹与停止函数——用于验证 HostKeyCallback 指纹三态。
func startTestSSHServer(t *testing.T) (addr, fingerprint string, stop func()) {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("生成临时主机密钥失败: %v", err)
	}
	signer, err := cssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("构造 signer 失败: %v", err)
	}
	cfg := &cssh.ServerConfig{
		PasswordCallback: func(conn cssh.ConnMetadata, password []byte) (*cssh.Permissions, error) {
			return &cssh.Permissions{}, nil // 测试桩：接受任意口令
		},
	}
	cfg.AddHostKey(signer)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				sconn, chans, reqs, err := cssh.NewServerConn(c, cfg)
				if err != nil {
					return
				}
				go cssh.DiscardRequests(reqs)
				for ch := range chans {
					ch.Reject(cssh.UnknownChannelType, "test-server")
				}
				_ = sconn.Close()
			}(conn)
		}
	}()
	return ln.Addr().String(), cssh.FingerprintSHA256(signer.PublicKey()), func() { _ = ln.Close() }
}

func testAuth(expected string) SSHAuth {
	return SSHAuth{Username: "tester", Password: "any", ExpectedFingerprint: expected}
}

func TestDialSSH_FingerprintTOFU(t *testing.T) {
	addr, wantFP, stop := startTestSSHServer(t)
	defer stop()

	client, actualFP, err := DialSSH(addr, testAuth(""))
	if err != nil {
		t.Fatalf("TOFU 模式应放行: %v", err)
	}
	defer client.Close()
	if actualFP != wantFP {
		t.Fatalf("返回指纹应为服务端主机指纹: want %s got %s", wantFP, actualFP)
	}
}

func TestDialSSH_FingerprintMatch(t *testing.T) {
	addr, wantFP, stop := startTestSSHServer(t)
	defer stop()

	client, actualFP, err := DialSSH(addr, testAuth(wantFP))
	if err != nil {
		t.Fatalf("指纹匹配应放行: %v", err)
	}
	defer client.Close()
	if actualFP != wantFP {
		t.Fatalf("返回指纹应为服务端主机指纹: want %s got %s", wantFP, actualFP)
	}
}

func TestDialSSH_FingerprintMismatch(t *testing.T) {
	addr, wantFP, stop := startTestSSHServer(t)
	defer stop()

	client, actualFP, err := DialSSH(addr, testAuth("SHA256:forgedFingerprintForMismatchTest00000="))
	if err == nil {
		_ = client.Close()
		t.Fatal("指纹不匹配必须拒绝连接")
	}
	if !strings.Contains(err.Error(), "指纹不匹配") {
		t.Fatalf("错误信息应说明指纹不匹配: %v", err)
	}
	var se *ServiceError
	if !errors.As(err, &se) || se.Code != ErrCodeHostFPMismatch {
		t.Fatalf("应解出错误码 %d: %v", ErrCodeHostFPMismatch, err)
	}
	// 拒绝时也应带出实际指纹，便于运维核对
	if actualFP != wantFP {
		t.Fatalf("拒绝时应返回实际指纹: want %s got %s", wantFP, actualFP)
	}
}
