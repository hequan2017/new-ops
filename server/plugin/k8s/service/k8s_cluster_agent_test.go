package service

import (
	"testing"

	agentSvc "github.com/hequan2017/new-ops/server/plugin/agent/service"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
)

// TestCreateAgentClusterRequiresAsset agent 模式注册缺 assetId 应被拒（在建库查询前即拦截）
func TestCreateAgentClusterRequiresAsset(t *testing.T) {
	s := &K8sClusterService{}
	err := s.CreateCluster(&model.K8sCluster{Name: "agent-cluster", ViaAgent: true}, "")
	if err == nil {
		t.Fatal("agent 模式缺 assetId 应报错")
	}
	if ke, ok := err.(*K8sError); !ok || ke.Code != ErrCodeK8sAgentAssetRequired {
		t.Fatalf("应为 1515 错误: %v", err)
	}
}

// TestAgentRESTConfig agent 模式 RESTConfig：host 固定 + WS 代理 transport 注入
func TestAgentRESTConfig(t *testing.T) {
	cfg := agentRESTConfig(&model.K8sCluster{AssetID: 3})
	if cfg.Host != "https://kubernetes.default.svc" {
		t.Fatalf("host 不符: %s", cfg.Host)
	}
	if cfg.Transport == nil {
		t.Fatal("transport 应注入（AgentDockerTransport k8s_req）")
	}
	if _, ok := cfg.Transport.(*agentSvc.AgentDockerTransport); !ok {
		t.Fatalf("transport 应为 AgentDockerTransport: %T", cfg.Transport)
	}
}
