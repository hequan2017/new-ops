package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveDockerAddr(t *testing.T) {
	if got := resolveDockerAddr([]string{"/nonexistent.sock", "/definitely/missing.sock"}); got != "unix:///var/run/docker.sock" {
		t.Fatalf("全部未命中应回退默认: %q", got)
	}
	if got := resolveDockerAddr(nil); got == "" {
		t.Fatal("空候选不应返回空串")
	}
}

func TestHandleDockerReqGetPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ping" {
			w.Header().Set("Api-Version", "1.54")
			_, _ = w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	var reply frame
	send := func(f frame) error { reply = f; return nil }
	handleDockerReq(frame{
		Type: "docker_req", TaskNo: "d1", ReqMethod: http.MethodGet, ReqPath: "/_ping",
	}, send, "tcp://"+strings.TrimPrefix(srv.URL, "http://"))

	if reply.Type != "docker_resp" || reply.TaskNo != "d1" {
		t.Fatalf("回执帧不符: %+v", reply)
	}
	if reply.RespStatus != 200 || reply.Msg != "" {
		t.Fatalf("状态/错误不符: %d %q", reply.RespStatus, reply.Msg)
	}
	body, _ := base64.StdEncoding.DecodeString(reply.RespBody)
	if string(body) != "OK" {
		t.Fatalf("响应体不符: %q", body)
	}
	if reply.RespHeader["Api-Version"][0] != "1.54" {
		t.Fatalf("响应头应透传: %v", reply.RespHeader)
	}
}

func TestHandleDockerReqPostBodyAndError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/containers/create" {
			var in map[string]any
			_ = json.NewDecoder(r.Body).Decode(&in)
			if in["Image"] != "alpine" {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"message":"bad image"}`))
				return
			}
			w.WriteHeader(201)
			_, _ = w.Write([]byte(`{"Id":"abc123"}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	addr := "tcp://" + strings.TrimPrefix(srv.URL, "http://")

	// POST 带体：透传方法/路径/体
	var reply frame
	capture := func(f frame) error { reply = f; return nil }
	handleDockerReq(frame{
		Type: "docker_req", TaskNo: "d2", ReqMethod: http.MethodPost, ReqPath: "/containers/create?name=t1",
		ReqBody:   base64.StdEncoding.EncodeToString([]byte(`{"Image":"alpine"}`)),
		ReqHeader: map[string][]string{"Content-Type": {"application/json"}},
	}, capture, addr)
	decoded, _ := base64.StdEncoding.DecodeString(reply.RespBody)
	if reply.RespStatus != 201 || !strings.Contains(string(decoded), `"Id"`) {
		t.Fatalf("POST 创建应 201 且含 Id: %+v body=%s", reply, decoded)
	}

	// daemon 不可达：errMsg 回传
	handleDockerReq(frame{Type: "docker_req", TaskNo: "d3", ReqMethod: "GET", ReqPath: "/_ping"}, capture, "unix:///nonexistent.sock")
	if reply.TaskNo != "d3" || reply.Msg == "" || reply.RespStatus != 0 {
		t.Fatalf("不可达应回执错误: %+v", reply)
	}

	// 缺字段：静默忽略
	handleDockerReq(frame{Type: "docker_req", TaskNo: "d4"}, capture, addr)
	if reply.TaskNo != "d3" {
		t.Fatal("缺字段的 docker_req 应被忽略")
	}
}
