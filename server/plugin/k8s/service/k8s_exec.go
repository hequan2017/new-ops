// Package service K2 收官：Pod WebShell（client-go SPDY remotecommand ⇄ WS 桥接）
// 通道约定与 term/container 插件一致：二进制帧=数据流，文本帧=JSON 控制（resize/ping/close）。
package service

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// k8sWsCtl 控制帧结构
type k8sWsCtl struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// k8sWsIO exec 输出 → WS（写互斥：stdout 泵与 pong 应答可能并发写 WS）
type k8sWsIO struct {
	ws *websocket.Conn
	mu sync.Mutex
}

// Write remotecommand stdout 回写 WS（二进制帧）
func (w *k8sWsIO) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := w.ws.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// writeJSON 控制应答（pong 等）
func (w *k8sWsIO) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ws.WriteMessage(websocket.TextMessage, b)
}

// k8sWsStdin WS → exec stdin（Read 内联消化控制帧：resize/ping/close）
type k8sWsStdin struct {
	wio    *k8sWsIO
	resize chan remotecommand.TerminalSize
}

// Read remotecommand stdin 泵拉取：数据帧返回，控制帧就地处理
func (s *k8sWsStdin) Read(p []byte) (int, error) {
	for {
		msgType, payload, err := s.wio.ws.ReadMessage()
		if err != nil {
			return 0, err
		}
		if msgType == websocket.TextMessage && len(payload) > 0 && payload[0] == '{' {
			var ctl k8sWsCtl
			_ = json.Unmarshal(payload, &ctl)
			switch ctl.Type {
			case "resize":
				if ctl.Cols > 0 && ctl.Rows > 0 {
					select {
					case s.resize <- remotecommand.TerminalSize{Width: uint16(ctl.Cols), Height: uint16(ctl.Rows)}:
					default: // 队列满丢弃过期尺寸，保留最新由后续帧补充
					}
				}
				continue
			case "ping":
				_ = s.wio.writeJSON(map[string]string{"type": "pong"})
				continue
			case "close":
				return 0, io.EOF
			}
		}
		if len(payload) == 0 {
			continue
		}
		n := copy(p, payload)
		return n, nil
	}
}

// k8sSizeQueue remotecommand 终端尺寸队列
type k8sSizeQueue struct {
	ch chan remotecommand.TerminalSize
}

// Next 阻塞等待下一个尺寸（通道关闭返回 nil 结束监听）
func (q *k8sSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.ch
	if !ok {
		return nil
	}
	return &size
}

// PodExecWS Pod 交互终端桥接（TTY 模式单流，/bin/sh；阻塞直到任一端断开）
func (s *K8sClusterService) PodExecWS(ws *websocket.Conn, clusterID uint, namespace, pod, container string, cols, rows int) {
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}
	fail := func(msg string) {
		_ = ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4000, msg), time.Now().Add(5*time.Second))
	}
	cfg, _, err := restConfigForCluster(clusterID)
	if err != nil {
		fail(err.Error())
		return
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fail("clientset 构造失败: " + err.Error())
		return
	}
	ctx := context.Background()
	// 容器未指定：取 Pod 首个容器（多容器场景前端提供选择）
	if container == "" {
		p, gerr := cs.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
		if gerr != nil {
			fail("Pod 不存在: " + gerr.Error())
			return
		}
		if len(p.Spec.Containers) == 0 {
			fail("Pod 无容器")
			return
		}
		container = p.Spec.Containers[0].Name
	}
	req := cs.CoreV1().RESTClient().Post().
		Resource("pods").Namespace(namespace).Name(pod).SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"/bin/sh"},
			Container: container,
			Stdin:     true, Stdout: true, Stderr: true,
			TTY: true, // TTY 单流裸数据，stderr 并入 stdout
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		fail("exec 构造失败: " + err.Error())
		return
	}

	sizeCh := make(chan remotecommand.TerminalSize, 4)
	sizeCh <- remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	wio := &k8sWsIO{ws: ws}
	stdin := &k8sWsStdin{wio: wio, resize: sizeCh}

	// StreamWithContext 阻塞：WS 断开 → Read 错误 → 流终止；exec 进程退出（exit）→ 返回
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            wio,
		Stderr:            wio,
		Tty:               true,
		TerminalSizeQueue: &k8sSizeQueue{ch: sizeCh},
	})
	if err != nil {
		fail("exec 已结束: " + err.Error())
		return
	}
	_ = ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(1000, "session closed"), time.Now().Add(3*time.Second))
}
