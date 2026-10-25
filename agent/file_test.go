package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fileCapture struct{ frames []frame }

func (c *fileCapture) send(f frame) error { c.frames = append(c.frames, f); return nil }

func (c *fileCapture) last() frame { return c.frames[len(c.frames)-1] }

func TestFileReceiveHappyPath(t *testing.T) {
	r := newFileReceiver(2, 64)
	cap := &fileCapture{}
	content := []byte("baize file push 测试内容 0123456789")
	target := filepath.Join(t.TempDir(), "sub", "dir", "a2.txt")
	sum := sha256.Sum256(content)

	r.handle(frame{Type: "file_push", TaskNo: "f1", Payload: target, FileMode: "0644",
		FileSize: int64(len(content)), FileSHA256: hex.EncodeToString(sum[:])}, cap.send)
	r.handle(frame{Type: "file_chunk", TaskNo: "f1", FileData: base64.StdEncoding.EncodeToString(content[:5])}, cap.send)
	r.handle(frame{Type: "file_chunk", TaskNo: "f1", FileData: base64.StdEncoding.EncodeToString(content[5:])}, cap.send)
	r.handle(frame{Type: "file_end", TaskNo: "f1"}, cap.send)

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("目标文件应存在: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("内容不符: %q", got)
	}
	if fi, _ := os.Stat(target); fi.Mode().Perm() != 0644 {
		t.Fatalf("权限应为 0644: %v", fi.Mode().Perm())
	}
	if last := cap.last(); last.ExitCode != 0 || last.TaskNo != "f1" || last.Msg != "" {
		t.Fatalf("应回执成功: %+v", last)
	}
}

func TestFileReceiveSHAMismatch(t *testing.T) {
	r := newFileReceiver(2, 64)
	cap := &fileCapture{}
	target := filepath.Join(t.TempDir(), "bad.txt")
	r.handle(frame{Type: "file_push", TaskNo: "f2", Payload: target, FileSize: 3, FileSHA256: "deadbeef"}, cap.send)
	r.handle(frame{Type: "file_chunk", TaskNo: "f2", FileData: base64.StdEncoding.EncodeToString([]byte("abc"))}, cap.send)
	r.handle(frame{Type: "file_end", TaskNo: "f2"}, cap.send)
	if last := cap.last(); last.ExitCode != -1 || last.Msg == "" {
		t.Fatalf("哈希不符应回执失败: %+v", last)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("校验失败不应落盘目标文件")
	}
}

func TestFileReceiveRejectsRelativeAndOverflow(t *testing.T) {
	r := newFileReceiver(2, 64)
	cap := &fileCapture{}
	// 相对路径在 start 即拒绝
	r.handle(frame{Type: "file_push", TaskNo: "f3", Payload: "relative/x.txt", FileSize: 1, FileSHA256: "aa"}, cap.send)
	if last := cap.last(); last.ExitCode != -1 || last.TaskNo != "f3" {
		t.Fatalf("相对路径应拒绝: %+v", last)
	}
	// 超声明大小的分块应失败且无残留
	abs := filepath.Join(t.TempDir(), "big.bin")
	r.handle(frame{Type: "file_push", TaskNo: "f4", Payload: abs, FileSize: 1, FileSHA256: "aa"}, cap.send)
	r.handle(frame{Type: "file_chunk", TaskNo: "f4", FileData: base64.StdEncoding.EncodeToString([]byte("12345"))}, cap.send)
	if last := cap.last(); last.ExitCode != -1 || last.TaskNo != "f4" {
		t.Fatalf("超大小分块应拒绝: %+v", last)
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatal("失败任务不应落盘")
	}
	// 未知任务号的分块/收尾应回执错误而非 panic
	r.handle(frame{Type: "file_chunk", TaskNo: "nope", FileData: "aGk="}, cap.send)
	if last := cap.last(); last.ExitCode != -1 || last.TaskNo != "nope" {
		t.Fatalf("未知任务应回执错误: %+v", last)
	}
}

func TestFileReceiveDedupDropsDuplicate(t *testing.T) {
	r := newFileReceiver(2, 64)
	cap := &fileCapture{}
	abs := filepath.Join(t.TempDir(), "dup.txt")
	hdr := frame{Type: "file_push", TaskNo: "f5", Payload: abs, FileSize: 0, FileSHA256: hex.EncodeToString(sha256.New().Sum(nil))}
	r.handle(hdr, cap.send)
	r.handle(hdr, cap.send) // 重复推送：静默丢弃
	if len(cap.frames) != 0 { // 成功的 file_push 本就不回帧，重复头更不应产生回执
		t.Fatalf("重复 file_push 不应产生回执, 实际帧数=%d", len(cap.frames))
	}
	// 正常收尾仍成功
	r.handle(frame{Type: "file_end", TaskNo: "f5"}, cap.send)
	if last := cap.last(); last.ExitCode != 0 {
		t.Fatalf("首次推送应可正常收尾: %+v", last)
	}
}

// 回归：会话断开 cleanupAll 与读循环 fail 对同一任务并发摘除时，
// 并发位（sem）必须且只能释放一次——双重释放会让 cleanupAll 收尾排水永久阻塞。
func TestFileReceiverConcurrentFailAndCleanupNoDoubleRelease(t *testing.T) {
	r := newFileReceiver(1, 64)
	cap := &fileCapture{}
	abs := filepath.Join(t.TempDir(), "race.txt")
	r.handle(frame{Type: "file_push", TaskNo: "f6", Payload: abs, FileSize: 10, FileSHA256: "aa"}, cap.send)

	r.mu.Lock()
	it := r.open["f6"] // 模拟读循环已取到接收状态、即将 fail 的瞬间
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.cleanupAll() // 会话断开：清扫含 f6 在内的未完成任务并排水 1 个并发位
	}()
	r.fail("f6", it, "分块解码失败", cap.send) // 读循环 fail 同一任务：摘除失败（已被 cleanupAll 摘走）→ 不得再排水
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanupAll 被双重释放卡死——sem 排水超过实际占用")
	}
	// 并发位应可再次完整占用（容量 1）
	select {
	case r.sem <- struct{}{}:
		<-r.sem
	default:
		t.Fatal("并发位未正确归还")
	}
}
