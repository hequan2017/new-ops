// 白泽轻量 Agent：文件接收端（M9 A2）
// file_push（头：目标绝对路径/大小/SHA256/权限）→ file_chunk（base64 分块写临时文件）→
// file_end（校验大小与哈希，原子改名到目标路径），任务号幂等，结果回传 task_result。
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type incomingFile struct {
	target  string
	mode    fs.FileMode
	size    int64
	wantSHA string
	written int64
	sha     hash.Hash
	tmp     *os.File
}

// fileReceiver 文件接收器：任务号去重 + 并发上限 + 按任务号维护接收状态（临时文件流式落盘，不占内存）
type fileReceiver struct {
	seen *taskSeen
	mu   sync.Mutex
	open map[string]*incomingFile
	// unknownReplied 未知任务分块已回执过的任务号（每任务只回一次，防大文件逐块回帧风暴）
	unknownReplied map[string]struct{}
	sem            chan struct{}
}

func newFileReceiver(maxConcurrent, maxSeen int) *fileReceiver {
	return &fileReceiver{seen: newTaskSeen(maxSeen), open: map[string]*incomingFile{},
		unknownReplied: map[string]struct{}{}, sem: make(chan struct{}, maxConcurrent)}
}

func (r *fileReceiver) handle(f frame, send func(frame) error) {
	switch f.Type {
	case "file_push":
		r.start(f, send)
	case "file_chunk":
		r.chunk(f, send)
	case "file_end":
		r.finish(f, send)
	}
}

func (r *fileReceiver) reply(taskNo string, code int, errMsg string, send func(frame) error) {
	_ = send(frame{Type: "task_result", TS: time.Now().Unix(), TaskNo: taskNo, ExitCode: code, Msg: errMsg})
}

// release 原子摘除接收状态：仅当 open 表中确实摘除成功时才释放并发位——
// 会话断开 cleanupAll 与读循环 fail/finish 并发摘除同一任务时，sem 释放与摘除
// 一一对应；若两处都无条件 <-r.sem 会多释放，cleanupAll 收尾排水时永久阻塞（agent 挂死不重连）
func (r *fileReceiver) release(taskNo string) {
	r.mu.Lock()
	_, ok := r.open[taskNo]
	delete(r.open, taskNo)
	r.mu.Unlock()
	if ok {
		<-r.sem
	}
}

// fail 清理接收状态（关删临时文件、释放并发位）并回执失败
func (r *fileReceiver) fail(taskNo string, it *incomingFile, errMsg string, send func(frame) error) {
	if it != nil {
		if it.tmp != nil {
			_ = it.tmp.Close()
			_ = os.Remove(it.tmp.Name())
		}
		r.release(taskNo)
	}
	r.reply(taskNo, -1, errMsg, send)
}

func (r *fileReceiver) start(f frame, send func(frame) error) {
	if f.TaskNo == "" || f.Payload == "" {
		return
	}
	if !r.seen.first("file:" + f.TaskNo) {
		return // 重复推送静默丢弃（幂等）
	}
	if !filepath.IsAbs(f.Payload) {
		r.reply(f.TaskNo, -1, "目标路径必须为绝对路径", send)
		return
	}
	mode := fs.FileMode(0644)
	if f.FileMode != "" {
		v, err := strconv.ParseUint(f.FileMode, 8, 32)
		if err != nil || v == 0 || v > 0o7777 {
			r.reply(f.TaskNo, -1, "fileMode 非法", send)
			return
		}
		mode = fs.FileMode(v)
	}
	if f.FileSize < 0 || f.FileSize > 8<<30 {
		r.reply(f.TaskNo, -1, "fileSize 非法（上限 8GB）", send)
		return
	}
	select {
	case r.sem <- struct{}{}:
	default:
		r.reply(f.TaskNo, -1, "agent 文件接收并发占满", send)
		return
	}
	tmp, err := os.CreateTemp("", "baize-file-*")
	if err != nil {
		<-r.sem
		r.reply(f.TaskNo, -1, "创建临时文件失败: "+err.Error(), send)
		return
	}
	r.mu.Lock()
	r.open[f.TaskNo] = &incomingFile{
		target: f.Payload, mode: mode, size: f.FileSize, wantSHA: f.FileSHA256,
		sha: sha256.New(), tmp: tmp,
	}
	r.mu.Unlock()
}

func (r *fileReceiver) chunk(f frame, send func(frame) error) {
	r.mu.Lock()
	it := r.open[f.TaskNo]
	if it == nil {
		if f.TaskNo != "" {
			// 每任务号只回执一次：server 侧有早停但迟到分块仍可能多帧，逐块回执会放大成错误风暴
			if _, done := r.unknownReplied[f.TaskNo]; !done {
				if len(r.unknownReplied) > 1024 { // 防长驻膨胀
					r.unknownReplied = map[string]struct{}{}
				}
				r.unknownReplied[f.TaskNo] = struct{}{}
				r.mu.Unlock()
				r.reply(f.TaskNo, -1, "未知文件任务分块", send)
				return
			}
		}
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()
	data, err := base64.StdEncoding.DecodeString(f.FileData)
	if err != nil {
		r.fail(f.TaskNo, it, "分块解码失败: "+err.Error(), send)
		return
	}
	if it.written+int64(len(data)) > it.size {
		r.fail(f.TaskNo, it, "分块超出声明大小", send)
		return
	}
	if _, err := it.tmp.Write(data); err != nil {
		r.fail(f.TaskNo, it, "写临时文件失败: "+err.Error(), send)
		return
	}
	it.sha.Write(data)
	it.written += int64(len(data))
}

func (r *fileReceiver) finish(f frame, send func(frame) error) {
	r.mu.Lock()
	it := r.open[f.TaskNo]
	r.mu.Unlock()
	if it == nil {
		if f.TaskNo != "" {
			r.reply(f.TaskNo, -1, "未知文件任务收尾", send)
		}
		return
	}
	ok := func() {
		r.release(f.TaskNo)
		r.reply(f.TaskNo, 0, "", send)
	}
	if it.written != it.size {
		r.fail(f.TaskNo, it, "接收字节数与声明不符", send)
		return
	}
	_ = it.tmp.Sync()
	if hex.EncodeToString(it.sha.Sum(nil)) != it.wantSHA {
		r.fail(f.TaskNo, it, "SHA256 校验不一致", send)
		return
	}
	if err := it.tmp.Close(); err != nil {
		r.fail(f.TaskNo, it, "关闭临时文件失败: "+err.Error(), send)
		return
	}
	if err := os.MkdirAll(filepath.Dir(it.target), 0755); err != nil {
		r.fail(f.TaskNo, it, "创建目标目录失败: "+err.Error(), send)
		return
	}
	if err := os.Chmod(it.tmp.Name(), it.mode); err != nil {
		r.fail(f.TaskNo, it, "设置权限失败: "+err.Error(), send)
		return
	}
	if err := os.Rename(it.tmp.Name(), it.target); err != nil {
		r.fail(f.TaskNo, it, "落盘失败: "+err.Error(), send)
		return
	}
	ok()
}

// cleanupAll 会话断开时清空未完成接收（关闭并删除临时文件，释放并发位）
func (r *fileReceiver) cleanupAll() {
	r.mu.Lock()
	n := len(r.open)
	for taskNo, it := range r.open {
		if it.tmp != nil {
			_ = it.tmp.Close()
			_ = os.Remove(it.tmp.Name())
		}
		delete(r.open, taskNo)
	}
	r.mu.Unlock()
	for i := 0; i < n; i++ {
		<-r.sem
	}
}
