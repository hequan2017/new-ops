package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectK8sInClusterDir(t *testing.T) {
	if detectK8sInClusterDir("", "443", t.TempDir()) != nil {
		t.Fatal("无 host env 应 nil")
	}
	if detectK8sInClusterDir("10.0.0.1", "", t.TempDir()) != nil {
		t.Fatal("无 port env 应 nil")
	}
	dir := t.TempDir()
	if detectK8sInClusterDir("10.0.0.1", "443", dir) != nil {
		t.Fatal("缺 SA 文件应 nil")
	}
	if err := os.WriteFile(filepath.Join(dir, "ca.crt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if detectK8sInClusterDir("10.0.0.1", "443", dir) != nil {
		t.Fatal("缺 token 应 nil")
	}
	if err := os.WriteFile(filepath.Join(dir, "token"), []byte("tok"), 0644); err != nil {
		t.Fatal(err)
	}
	ic := detectK8sInClusterDir("10.0.0.1", "443", dir)
	if ic == nil || ic.APIAddr != "https://10.0.0.1:443" {
		t.Fatalf("in-cluster 探测不符: %+v", ic)
	}
}

func TestHandleK8sReqNotInCluster(t *testing.T) {
	var got frame
	handleK8sReq(frame{Type: "k8s_req", TaskNo: "k1", ReqMethod: http.MethodGet, ReqPath: "/version"},
		func(f frame) error { got = f; return nil }, nil)
	if got.Type != "k8s_resp" || got.RespStatus != 0 || got.Msg == "" || got.TaskNo != "k1" {
		t.Fatalf("非集群环境应回执错误: %+v", got)
	}
}

func TestK8sHTTPClientBadCA(t *testing.T) {
	if _, err := k8sHTTPClient("/nonexistent/ca.crt"); err == nil {
		t.Fatal("CA 文件不存在应报错")
	}
	p := filepath.Join(t.TempDir(), "bad.crt")
	if err := os.WriteFile(p, []byte("not-a-cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := k8sHTTPClient(p); err == nil {
		t.Fatal("非法 CA 应报错")
	}
}
