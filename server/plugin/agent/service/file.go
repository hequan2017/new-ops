// Package service 白泽 Agent 通道：文件分发（M9 A2）
// server 读取本地文件，经 WS 长连接分块（base64，60KB/块）下发，agent 端临时文件落盘，
// 结束帧后按"声明大小 + SHA256"校验并原子改名到目标路径，结果回传 task_result
// （复用任务号幂等与 pendingTasks 等待机制）。
package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// fileChunkSize 单块原始字节数（base64 后约 80KB）
const fileChunkSize = 60 * 1024

// filePushTimeout 文件分发等待时限：基础 60s + 每 256KB 加 1s（钳制 ≤30min）
func filePushTimeout(size int64) time.Duration {
	t := 60*time.Second + time.Duration(size/(256*1024))*time.Second
	if t > 30*time.Minute {
		t = 30 * time.Minute
	}
	return t
}

// validateFileMode 目标文件权限校验（空=0644；八进制字符串，纯逻辑）
func validateFileMode(mode string) (os.FileMode, error) {
	if mode == "" {
		return 0644, nil
	}
	v, err := strconv.ParseUint(mode, 8, 32)
	if err != nil || v == 0 || v > 0o7777 {
		return 0, newAgentErr(ErrCodeAgentParamInvalid, "mode 须为八进制权限串（如 0644）")
	}
	return os.FileMode(v), nil
}

// PushFile 向资产上的在线 agent 分发本地文件（同步等待落盘校验结果；写操作仅 888）
func (s *AgentService) PushFile(operator string, assetID uint, localPath, remotePath, mode string) (*TaskResult, error) {
	if assetID == 0 || strings.TrimSpace(localPath) == "" || strings.TrimSpace(remotePath) == "" {
		return nil, newAgentErr(ErrCodeAgentParamInvalid, "assetId/localPath/remotePath 必填")
	}
	if !filepath.IsAbs(remotePath) {
		return nil, newAgentErr(ErrCodeAgentParamInvalid, "remotePath 必须为绝对路径")
	}
	fm, err := validateFileMode(mode)
	if err != nil {
		return nil, err
	}
	src, err := os.Open(localPath)
	if err != nil {
		return nil, newAgentErr(ErrCodeAgentParamInvalid, "本地文件不可读: "+err.Error())
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil || fi.IsDir() {
		return nil, newAgentErr(ErrCodeAgentParamInvalid, "本地路径须为普通文件")
	}

	// 先流式计算 SHA256，再回头分块读
	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	sum := hex.EncodeToString(hasher.Sum(nil))
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("文件重读失败: %w", err)
	}

	taskNo := uuid.NewString()
	ch := make(chan agentFrame, 1)
	pendingTasks.Store(taskNo, ch)
	defer pendingTasks.Delete(taskNo)

	start := time.Now()
	summary := fmt.Sprintf("%s → %s（%d 字节, mode %s）", localPath, remotePath, fi.Size(), fm)
	send := func(f agentFrame) error { return WriteToAgent(assetID, f) }

	if err := send(agentFrame{Type: "file_push", TS: time.Now().Unix(), TaskNo: taskNo,
		Payload: remotePath, FileName: filepath.Base(remotePath), FileSize: fi.Size(),
		FileSHA256: sum, FileMode: fmt.Sprintf("%04o", fm)}); err != nil {
		return nil, err
	}
	buf := make([]byte, fileChunkSize)
	for {
		n, rerr := io.ReadFull(src, buf)
		if n > 0 {
			if err := send(agentFrame{Type: "file_chunk", TS: time.Now().Unix(), TaskNo: taskNo,
				FileData: base64.StdEncoding.EncodeToString(buf[:n])}); err != nil {
				return nil, fmt.Errorf("分块下发失败: %w", err)
			}
		}
		if rerr == io.EOF || rerr == io.ErrUnexpectedEOF {
			break
		}
		if rerr != nil {
			return nil, fmt.Errorf("读取文件失败: %w", rerr)
		}
	}
	if err := send(agentFrame{Type: "file_end", TS: time.Now().Unix(), TaskNo: taskNo}); err != nil {
		return nil, fmt.Errorf("收尾帧下发失败: %w", err)
	}

	timer := time.NewTimer(filePushTimeout(fi.Size()))
	defer timer.Stop()
	select {
	case f := <-ch:
		r := &TaskResult{
			TaskNo: taskNo, ExitCode: f.ExitCode, Output: f.Output, ErrMsg: f.Msg,
			ElapsedMs: time.Since(start).Milliseconds(),
		}
		s.persistTask(operator, assetID, "file", summary, taskStatusOf(r.ExitCode, r.ErrMsg), r)
		return r, nil
	case <-timer.C:
		r := &TaskResult{TaskNo: taskNo, ErrMsg: "agent 文件分发超时", ElapsedMs: time.Since(start).Milliseconds()}
		s.persistTask(operator, assetID, "file", summary, TaskStatusTimeout, r)
		return nil, newAgentErr(ErrCodeAgentTaskTimeout, "agent 文件分发超时（结果未在限时内回传）")
	}
}
