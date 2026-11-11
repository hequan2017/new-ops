// Package service eduadmin：跨校身份认证单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestBuildIdpRedirect SAML/OAuth2.0 跳转组装
func TestBuildIdpRedirect(t *testing.T) {
	saml := &model.EduIdpAdapter{
		School: "上海大学", Protocol: "SAML2.0",
		MetadataEndpoint: "https://idp.shu.edu/saml/sso", MutualScope: "教学资源,图书馆",
	}
	u, err := buildIdpRedirect(saml, "state-xyz")
	if err != nil {
		t.Fatalf("SAML 组装失败: %v", err)
	}
	if !strings.HasPrefix(u, "https://idp.shu.edu/saml/sso?") || !strings.Contains(u, "RelayState=state-xyz") {
		t.Fatalf("SAML 跳转应带 RelayState: %s", u)
	}

	oauth := &model.EduIdpAdapter{
		School: "华东理工", Protocol: "OAuth2.0",
		MetadataEndpoint: "https://oauth.ecust.edu.cn/authorize",
		ClientId: "eduadmin-gateway", MutualScope: "教学资源,算力",
	}
	u2, err := buildIdpRedirect(oauth, "state-abc")
	if err != nil {
		t.Fatalf("OAuth 组装失败: %v", err)
	}
	for _, want := range []string{"client_id=eduadmin-gateway", "response_type=code", "state=state-abc"} {
		if !strings.Contains(u2, want) {
			t.Fatalf("OAuth 跳转缺 %s: %s", want, u2)
		}
	}
	// OAuth 缺 ClientId
	oauth.ClientId = ""
	if _, err := buildIdpRedirect(oauth, "s"); err == nil {
		t.Fatal("OAuth 缺 ClientId 应报错")
	}
	// 端点非法
	bad := &model.EduIdpAdapter{Protocol: "SAML2.0", MetadataEndpoint: "not-a-url"}
	if _, err := buildIdpRedirect(bad, "s"); err == nil {
		t.Fatal("非法端点应报错")
	}
	// 协议白名单
	bad2 := &model.EduIdpAdapter{Protocol: "CAS", MetadataEndpoint: "https://a.b/c"}
	if _, err := buildIdpRedirect(bad2, "s"); err == nil {
		t.Fatal("协议白名单外应报错")
	}
}

// TestValidEduMutualScope 互认范围白名单
func TestValidEduMutualScope(t *testing.T) {
	if !ValidEduMutualScope("教学资源, 算力") {
		t.Fatal("含空格白名单内应通过")
	}
	if ValidEduMutualScope("教学资源,宿舍") {
		t.Fatal("白名单外应拒绝")
	}
	if ValidEduMutualScope("") {
		t.Fatal("空范围应拒绝")
	}
}
