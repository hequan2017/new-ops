package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

func newTestPolicySvc(t *testing.T) *PolicyService {
	t.Helper()
	db, err := data.OpenMemory(
		&model.PolicyDoc{}, &model.PolicyTermTag{}, &model.PolicyEntity{},
		&model.PolicyTokenLabel{}, &model.PolicyRelationLabel{},
		&model.PolicyPermGrant{}, &model.PolicyPromptTemplate{}, &model.PolicyQaLog{},
	)
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	return NewPolicyService(db)
}

func TestSavePolicyDocIdempotent(t *testing.T) {
	s := newTestPolicySvc(t)
	in := &model.PolicyDoc{Name: "学籍管理办法", Format: "TXT", SecurityLevel: "公开", RawText: "第一条 为规范学籍管理。"}
	out, err := s.SavePolicyDoc(in)
	if err != nil {
		t.Fatalf("导入: %v", err)
	}
	if out.ContentHash == "" || out.Sm4Mark {
		t.Fatalf("幂等键/密级标注口径不符: %+v", out)
	}
	// 参数白名单
	bad := &model.PolicyDoc{Name: "x", Format: "DOC", SecurityLevel: "公开", RawText: "t"}
	if _, err = s.SavePolicyDoc(bad); errx.CodeOf(err) != ErrCodePrDocParam {
		t.Fatalf("非法格式应 3000，实际 code=%d", errx.CodeOf(err))
	}
	// 重复导入走幂等返回已有行（不新增）
	in2 := &model.PolicyDoc{Name: "学籍管理办法", Format: "TXT", SecurityLevel: "公开", RawText: "第一条 为规范学籍管理。"}
	out2, err := s.SavePolicyDoc(in2)
	if err != nil || out2.ID != out.ID {
		t.Fatalf("重复导入应幂等返回原行: err=%v id=%d/%d", err, out2.ID, out.ID)
	}
}

func TestSearchPolicyDocsByRoleACL(t *testing.T) {
	s := newTestPolicySvc(t)
	// 三份制度：教学+公开、教学+涉密、科研+内部
	for _, d := range []model.PolicyDoc{
		{Name: "本科生培养方案", Format: "TXT", SecurityLevel: "公开", RawText: "学分 课程"},
		{Name: "涉密科研课题清单", Format: "TXT", SecurityLevel: "涉密", RawText: "横向课题"},
		{Name: "科研经费管理办法", Format: "TXT", SecurityLevel: "内部", RawText: "预算"},
	} {
		if _, err := s.SavePolicyDoc(&d); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// grant：教师×教学 允许到公开；教师×科研 允许到内部
	for _, g := range []model.PolicyPermGrant{
		{RoleName: "教师", DocCategory: "教学", MaxLevel: "公开", Allowed: true},
		{RoleName: "教师", DocCategory: "科研", MaxLevel: "内部", Allowed: true},
	} {
		if _, err := s.SavePolicyPermGrant(&g); err != nil {
			t.Fatalf("seed grant: %v", err)
		}
	}
	// 教师检索：可见 教学-公开 与 科研-内部，不可见 教学-涉密
	list, total, err := s.SearchPolicyDocsByRole("教师", "", 1, 20)
	if err != nil {
		t.Fatalf("检索: %v", err)
	}
	if total != 2 {
		t.Fatalf("可见数应 2，实际 %d", total)
	}
	for _, d := range list {
		if d.SecurityLevel == "涉密" {
			t.Fatalf("涉密文档不应可见: %s", d.Name)
		}
	}
	// 无任何 grant 的角色 → 全不可见
	_, total, err = s.SearchPolicyDocsByRole("访客", "", 1, 20)
	if err != nil || total != 0 {
		t.Fatalf("无授权应不可见: err=%v total=%d", err, total)
	}
	// 角色必填
	if _, _, err = s.SearchPolicyDocsByRole(" ", "", 1, 20); errx.CodeOf(err) != ErrCodePrAclParam {
		t.Fatalf("角色必填应 3008，实际 code=%d", errx.CodeOf(err))
	}
}
