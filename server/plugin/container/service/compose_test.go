// Package service Compose 编排纯函数单测（项目名校验 / ps 输出双格式解析）
package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateComposeProject(t *testing.T) {
	cases := []struct {
		name    string
		project string
		content string
		wantErr bool
	}{
		{"正常项目名", "my-app", "services:\n  web:\n    image: nginx", false},
		{"数字开头", "1st-app", "services: {}", false},
		{"大写字母非法", "MyApp", "services: {}", true},
		{"空内容", "my-app", "  ", true},
		{"特殊字符", "my app", "services: {}", true},
		{"点号非法", "my.app", "services: {}", true},
		{"超长 64", strings.Repeat("a", 64), "services: {}", true},
		{"63 字符合法", strings.Repeat("a", 63), "services: {}", false},
	}
	for _, tc := range cases {
		err := validateComposeProject(tc.project, tc.content)
		if (err != nil) != tc.wantErr {
			t.Fatalf("%s: wantErr=%v got=%v", tc.name, tc.wantErr, err)
		}
	}
}

func TestParseComposePS(t *testing.T) {
	// 新版 docker compose：JSON 数组
	arr := `[
	  {"Name":"baize-web-1","Service":"web","State":"running","Publishers":[{"PublishedPort":8081,"TargetPort":80}]},
	  {"Name":"baize-db-1","Service":"db","State":"exited"}
	]`
	items := parseComposePS(arr)
	if len(items) != 2 {
		t.Fatalf("数组格式应解析出 2 项，got %d", len(items))
	}
	if items[0].Name != "baize-web-1" || items[0].State != "running" {
		t.Fatalf("首项字段错误: %+v", items[0])
	}
	if items[0].Ports != "8081->80" {
		t.Fatalf("端口映射错误: %s", items[0].Ports)
	}
	if items[1].State != "exited" {
		t.Fatalf("次项状态错误: %+v", items[1])
	}

	// 旧版：逐行 JSON
	lines := `{"Name":"a-1","Service":"a","State":"running"}
{"Name":"b-1","Service":"b","State":"running","Ports":"0.0.0.0:53->53/tcp"}`
	items = parseComposePS(lines)
	if len(items) != 2 {
		t.Fatalf("逐行格式应解析出 2 项，got %d", len(items))
	}
	if items[1].Ports != "0.0.0.0:53->53/tcp" {
		t.Fatalf("Ports 字符串字段回退失败: %s", items[1].Ports)
	}

	// 空输出与非法内容
	if got := parseComposePS(""); len(got) != 0 {
		t.Fatalf("空输出应为 0 项，got %d", len(got))
	}
	if got := parseComposePS("not json at all"); len(got) != 0 {
		t.Fatalf("非法输出应为 0 项，got %d", len(got))
	}
}

func TestComposeStatusMarshal(t *testing.T) {
	// Services 列 JSON 序列化回读闭环（落库口径）
	items := []ComposeServiceItem{{Name: "n", Service: "s", State: "running", Ports: "80->80"}}
	b, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal 失败: %v", err)
	}
	var back []ComposeServiceItem
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal 失败: %v", err)
	}
	if len(back) != 1 || back[0].Ports != "80->80" {
		t.Fatalf("回读不一致: %+v", back)
	}
}
