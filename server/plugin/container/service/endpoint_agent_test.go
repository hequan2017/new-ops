package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/container/model"
)

func TestValidateEndpointAgentMode(t *testing.T) {
	// agent 模式必须关联资产
	ep := &model.DockerEndpoint{Name: "agent-ep", Addr: "tcp://ignored:2375", ViaAgent: true}
	if err := validateEndpoint(ep); err == nil {
		t.Fatal("agent 模式缺 assetId 应报错")
	}
	ep.AssetID = 7
	if err := validateEndpoint(ep); err != nil {
		t.Fatalf("agent 模式带 assetId 应通过: %v", err)
	}
	// 直连模式不受影响
	if err := validateEndpoint(&model.DockerEndpoint{Name: "d", Addr: "unix:///var/run/docker.sock"}); err != nil {
		t.Fatalf("直连模式应通过: %v", err)
	}
}

// TestNewDockerClientAgentMode agent 模式仅构造客户端（transport 换成 WS 通道），不拨号、不依赖 DB
func TestNewDockerClientAgentMode(t *testing.T) {
	svc := &EndpointService{}
	cl, err := svc.NewDockerClient(&model.DockerEndpoint{Name: "a", Addr: "tcp://ignored:2375", ViaAgent: true, AssetID: 9})
	if err != nil || cl == nil {
		t.Fatalf("agent 模式构造应成功: %v", err)
	}
	_ = cl.Close()
}

// TestRequireDirect 守卫：agent 模式拒绝流式/CLI，直连放行
func TestRequireDirect(t *testing.T) {
	if err := requireDirect(&model.DockerEndpoint{ViaAgent: true}); err == nil {
		t.Fatal("agent 模式应被守卫拒绝")
	}
	if err := requireDirect(&model.DockerEndpoint{}); err != nil {
		t.Fatalf("直连应放行: %v", err)
	}
}
