package service

import (
	"encoding/json"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/job/model"
)

// renderForHost 单主机命令解析决策（纯逻辑，可单测）：
// 显式变量组 > 主机所在资产组绑定的变量组（取最新）> 原样。
func renderForHost(base string, explicitVarGroupID *uint, vgByAssetGroup map[uint]model.JobVariableGroup, hostAssetGroupID *uint) (string, error) {
	if explicitVarGroupID != nil && *explicitVarGroupID != 0 {
		if vg, ok := vgByAssetGroup[*explicitVarGroupID]; ok {
			return RenderTemplate(base, vg.Variables)
		}
		return base, nil
	}
	if hostAssetGroupID != nil {
		if vg, ok := vgByAssetGroup[*hostAssetGroupID]; ok {
			return RenderTemplate(base, vg.Variables)
		}
	}
	return base, nil
}

func Test_renderForHost_explicitGroup(t *testing.T) {
	vgMap := map[uint]model.JobVariableGroup{
		7: {Name: "web-vars", Variables: `[{"key":"port","value":"8080"}]`},
	}
	gid := uint(7)
	out, err := renderForHost("echo {{port}}", &gid, vgMap, nil)
	if err != nil {
		t.Fatalf("显式组渲染失败: %v", err)
	}
	if out != "echo 8080" {
		t.Fatalf("out = %s", out)
	}
}

func Test_renderForHost_autoByAssetGroup(t *testing.T) {
	vgMap := map[uint]model.JobVariableGroup{
		7: {Name: "web-vars", Variables: `[{"key":"env","value":"prod"}]`},
	}
	hostGroup := uint(7)
	out, err := renderForHost("echo {{env}}", nil, vgMap, &hostGroup)
	if err != nil {
		t.Fatalf("自动注入失败: %v", err)
	}
	if out != "echo prod" {
		t.Fatalf("out = %s", out)
	}
}

func Test_renderForHost_noMatch(t *testing.T) {
	vgMap := map[uint]model.JobVariableGroup{
		9: {Name: "other", Variables: `[{"key":"x","value":"1"}]`},
	}
	hostGroup := uint(7) // 主机在 7 组，但变量组绑定的是 9
	out, err := renderForHost("echo {{env}}", nil, vgMap, &hostGroup)
	if err != nil {
		t.Fatalf("无匹配不应报错: %v", err)
	}
	if out != "echo {{env}}" {
		t.Fatalf("无匹配应原样保留占位: %s", out)
	}
}

func Test_mapECSGroupByAssetGroupID(t *testing.T) {
	// 变量组按 asset_group_id 建索引的辅助函数（同组多个时取 ID 最大即最新）
	vgs := []model.JobVariableGroup{
		{Name: "a", AssetGroupID: uintPtr(7), Variables: `[{"key":"k","value":"v1"}]`},
		{Name: "c", AssetGroupID: uintPtr(9), Variables: `[{"key":"k","value":"v9"}]`},
	}
	vgs[0].ID = 3
	vgs[1].ID = 5
	// 追加同组但 ID 更小的，不应覆盖
	vgs = append(vgs, model.JobVariableGroup{Name: "old", AssetGroupID: uintPtr(7), Variables: `[{"key":"k","value":"old"}]`})
	vgs[2].ID = 1
	m := variableGroupsByAssetGroupID(vgs)
	if len(m) != 2 {
		t.Fatalf("索引组数错误: %d", len(m))
	}
	if m[7].Name != "a" {
		t.Fatalf("同组应取 ID 最大者: got %s", m[7].Name)
	}
	if m[9].Name != "c" {
		t.Fatalf("9 组映射错误: got %s", m[9].Name)
	}
}

func Test_variablesJSONValidate(t *testing.T) {
	var arr []map[string]string
	if err := json.Unmarshal([]byte(`[{"key":"k","value":"v"}]`), &arr); err != nil {
		t.Fatalf("合法变量 JSON 解析失败: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"key":"v"}`), &arr); err == nil {
		t.Fatalf("对象形式的变量应解析为数组才合法")
	}
}

func uintPtr(v uint) *uint { return &v }
