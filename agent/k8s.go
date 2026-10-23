// 白泽轻量 Agent：K8s in-cluster 代理（M9 A4/K4）
// 集群内以 Pod 运行时，探测 ServiceAccount（CA + Bearer token），把 server 下发的
// k8s_req 代理到本 Pod 可达的 API server，响应按 taskNo 关联回传 k8s_resp。
// v1 边界：管理面 HTTP 可代理；WebShell（SPDY 升级）不支持。
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
)

const k8sSAPath = "/var/run/secrets/kubernetes.io/serviceaccount"

// k8sInCluster in-cluster 访问要素
type k8sInCluster struct {
	APIAddr   string // https://host:port
	CAFile    string
	TokenFile string
}

// detectK8sInCluster 探测 in-cluster 环境（KUBERNETES_* env 与 SA 文件齐备才算）
func detectK8sInCluster() *k8sInCluster {
	return detectK8sInClusterDir(os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT"), k8sSAPath)
}

// detectK8sInClusterDir 按给定要素探测（纯逻辑，可单测；dir 为 ServiceAccount 目录）
func detectK8sInClusterDir(host, port, dir string) *k8sInCluster {
	if host == "" || port == "" || dir == "" {
		return nil
	}
	caFile, tokenFile := dir+"/ca.crt", dir+"/token"
	for _, p := range []string{caFile, tokenFile} {
		if fi, err := os.Stat(p); err != nil || fi.IsDir() {
			return nil
		}
	}
	return &k8sInCluster{APIAddr: "https://" + net.JoinHostPort(host, port), CAFile: caFile, TokenFile: tokenFile}
}

// k8sHTTPClient 集群 API 客户端（in-cluster CA 校验、直连不走环境代理、HTTP2）
func k8sHTTPClient(caFile string) (*http.Client, error) {
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("集群 CA 证书解析失败")
	}
	return &http.Client{
		Timeout: dockerProxyTimeout,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
			Proxy:             nil,
			ForceAttemptHTTP2: true,
		},
	}, nil
}

// handleK8sReq 处理 k8s_req：in-cluster ServiceAccount 通道（https + CA 校验 + Bearer 注入）
func handleK8sReq(f frame, send func(frame) error, ic *k8sInCluster) {
	if ic == nil {
		proxyReply(f, send, "k8s_resp", 0, nil, nil, "agent 不在 K8s 集群内（未检测到 ServiceAccount）")
		return
	}
	client, err := k8sHTTPClient(ic.CAFile)
	if err != nil {
		proxyReply(f, send, "k8s_resp", 0, nil, nil, "集群 CA 不可用: "+err.Error())
		return
	}
	handleHTTPProxy("k8s_resp", f, send, ic.APIAddr, client, ic.TokenFile)
}
