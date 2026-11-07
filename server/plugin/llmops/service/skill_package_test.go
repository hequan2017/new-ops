// Package service llmops 技能包单测（打包形状/解析校验/导入规划纯函数）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestMarshalSkillPackage(t *testing.T) {
	pkg := marshalSkillPackage([]model.LlmopsSkill{{Name: "技能A", Desc: "d", DefJSON: `{"kind":"prompt","template":"{input}"}`}})
	for _, want := range []string{`"version":1`, `"name":"技能A"`, `"defJSON"`} {
		if !strings.Contains(pkg, want) {
			t.Fatalf("包缺 %q: %s", want, pkg)
		}
	}
}

func TestParseSkillPackage(t *testing.T) {
	skills, err := ParseSkillPackage(`{"version":1,"skills":[{"name":"A","defJSON":"{\"kind\":\"prompt\",\"template\":\"{input}\"}"}]}`)
	if err != nil || len(skills) != 1 || skills[0].Name != "A" {
		t.Fatalf("解析异常: %+v %v", skills, err)
	}
	if _, err := ParseSkillPackage(`{"version":2,"skills":[{"name":"A","defJSON":"{}"}]}`); err == nil {
		t.Fatal("版本不符应拒绝")
	}
	if _, err := ParseSkillPackage(`{"version":1,"skills":[]}`); err == nil {
		t.Fatal("空包应拒绝")
	}
	if _, err := ParseSkillPackage(`{"version":1,"skills":[{"name":"","defJSON":"{}"}]}`); err == nil {
		t.Fatal("空名应拒绝")
	}
	if _, err := ParseSkillPackage(`{"version":1,"skills":[{"name":"A","defJSON":"{\"kind\":\"code\"}"}]}`); err == nil {
		t.Fatal("def 非法应拒绝")
	}
}

func TestImportPlan(t *testing.T) {
	existing := map[string]bool{"旧技能": true, "双写": true}
	pkg := []pkgSkill{
		{Name: "新技能"}, {Name: "旧技能"}, {Name: "双写"},
	}
	// 不覆盖：1 建 2 冲突
	create, update, conflicts := importPlan(existing, pkg, false)
	if len(create) != 1 || len(update) != 0 || len(conflicts) != 2 {
		t.Fatalf("不覆盖规划异常: %d %d %v", len(create), len(update), conflicts)
	}
	if conflicts[0] != "旧技能" || conflicts[1] != "双写" {
		t.Fatalf("冲突顺序异常: %v", conflicts)
	}
	// 覆盖：1 建 2 更新 0 冲突
	create, update, conflicts = importPlan(existing, pkg, true)
	if len(create) != 1 || len(update) != 2 || len(conflicts) != 0 {
		t.Fatalf("覆盖规划异常: %d %d %v", len(create), len(update), conflicts)
	}
}
