package service

import (
	"encoding/json"
	"strconv"
	"testing"

	pipelineModel "github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
)

// fakeTrigger 假流水线触发器（验证发版钩子行为，不依赖 pipeline 插件）
type fakeTrigger struct {
	calls    int
	lastID   uint
	lastOp   string
	failWith error
}

func (f *fakeTrigger) CreateBuild(pipelineID uint, params map[string]string, operator string, userID uint, triggerIP string) (*pipelineModel.PipelineBuild, error) {
	f.calls++
	f.lastID = pipelineID
	f.lastOp = operator
	if f.failWith != nil {
		return nil, f.failWith
	}
	return &pipelineModel.PipelineBuild{PipelineName: "发布流水线", BuildNo: 42}, nil
}

// threeNodeJSON 三节点审批流定义（提交 → 审批 → 归档）
func threeNodeJSON() (states, transitions string) {
	states = `[{"name":"submitted"},{"name":"approving","isApproval":true},{"name":"archived","isFinal":true}]`
	transitions = `[{"from":"submitted","action":"submit","to":"approving"},
	  {"from":"approving","action":"approve","to":"archived"},
	  {"from":"approving","action":"reject","to":"archived"}]`
	return states, transitions
}

func newTestSvcWithTrigger(t *testing.T) (*WorkflowService, *fakeTrigger) {
	t.Helper()
	db, err := data.OpenMemory(&model.WfDefinition{}, &model.WfInstance{}, &model.WfActionLog{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	ft := &fakeTrigger{}
	return NewWorkflowService(db, ft), ft
}

func seedDefinition(t *testing.T, s *WorkflowService) *model.WfDefinition {
	t.Helper()
	states, transitions := threeNodeJSON()
	d := &model.WfDefinition{
		Name: "发版审批", States: states, Transitions: transitions, StartState: "submitted", Enabled: true,
	}
	if err := s.CreateDefinition(d); err != nil {
		t.Fatalf("seed definition: %v", err)
	}
	return d
}

func TestCreateDefinitionDuplicate(t *testing.T) {
	s, _ := newTestSvcWithTrigger(t)
	seedDefinition(t, s)
	states, transitions := threeNodeJSON()
	dup := &model.WfDefinition{Name: "发版审批", States: states, Transitions: transitions, StartState: "submitted"}
	if err := s.CreateDefinition(dup); errx.CodeOf(err) != ErrCodeWfNameRequired {
		t.Fatalf("重名应 1901，实际 code=%d", errx.CodeOf(err))
	}
	// 非法定义
	if err := s.CreateDefinition(&model.WfDefinition{Name: "x", States: `[{"name":"a"},{"name":"a"}]`, Transitions: `[]`, StartState: "a"}); errx.CodeOf(err) != ErrCodeWfStructInvalid {
		t.Fatalf("重名状态应 1902，实际 code=%d", errx.CodeOf(err))
	}
	// 停用后禁止发起
	if err := s.ToggleDefinition(1, false); err != nil {
		t.Fatalf("停用: %v", err)
	}
	if _, err := s.StartInstance(1, "标题", "", "", "op", 1); errx.CodeOf(err) != ErrCodeWfStructInvalid {
		t.Fatalf("停用发起应 1902，实际 code=%d", errx.CodeOf(err))
	}
	// 有进行中实例拒绝删除（先启用再发起）
	if err := s.ToggleDefinition(1, true); err != nil {
		t.Fatalf("启用: %v", err)
	}
	if _, err := s.StartInstance(1, "工单一", "", "", "op", 1); err != nil {
		t.Fatalf("发起: %v", err)
	}
	if err := s.DeleteDefinition(1); errx.CodeOf(err) != ErrCodeWfStructInvalid {
		t.Fatalf("有进行中工单删除应 1902，实际 code=%d", errx.CodeOf(err))
	}
}

func TestLifecycleWithReleaseHook(t *testing.T) {
	s, ft := newTestSvcWithTrigger(t)
	d := seedDefinition(t, s)
	params, _ := json.Marshal(map[string]any{"pipelineId": 7, "params": map[string]string{"env": "prod"}})

	ins, err := s.StartInstance(d.ID, "v1.2 发版", BizTypeRelease, string(params), "op", 9)
	if err != nil {
		t.Fatalf("发起: %v", err)
	}
	if ft.calls != 0 {
		t.Fatal("未到终态不应触发流水线")
	}
	// 非法动作
	if _, err = s.SubmitAction(ins.ID, "approve", "", "op"); errx.CodeOf(err) != ErrCodeWfActionInvalid {
		t.Fatalf("submitted 不接受 approve 应 1904，实际 code=%d", errx.CodeOf(err))
	}
	// 提交 → 审批
	if _, err = s.SubmitAction(ins.ID, "submit", "", "op"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	// 审批节点受限
	if _, err = s.SubmitAction(ins.ID, "abort", "", "op"); errx.CodeOf(err) != ErrCodeWfActionInvalid {
		t.Fatalf("审批节点仅 approve/reject 应 1904，实际 code=%d", errx.CodeOf(err))
	}
	if _, err = s.SubmitAction(ins.ID, "approve", "同意发版", "op"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	// 终态后不可再操作
	if _, err = s.SubmitAction(ins.ID, "cancel", "", "op"); errx.CodeOf(err) != ErrCodeWfActionInvalid {
		t.Fatalf("终态后操作应 1904，实际 code=%d", errx.CodeOf(err))
	}
	// 发版钩子：恰好一次，参数透传，流水线名/构建号落 hook 日志
	if ft.calls != 1 || ft.lastID != 7 || ft.lastOp != "工单#"+strconv.Itoa(int(ins.ID)) {
		t.Fatalf("钩子口径不符: calls=%d id=%d op=%q", ft.calls, ft.lastID, ft.lastOp)
	}
	logs, err := s.GetInstanceLogs(ins.ID, 9, false)
	if err != nil {
		t.Fatalf("logs: %v", err)
	}
	var hookMsg string
	for _, l := range logs {
		if l.Action == "hook" {
			hookMsg = l.Comment
		}
	}
	if hookMsg != "发版钩子已触发流水线「发布流水线」构建 #42" {
		t.Fatalf("hook 日志口径不符: %q", hookMsg)
	}
	// 状态已收档
	got, _ := s.repo.FindInstance(s.dbh(), ins.ID)
	if got.Status != model.InstanceFinished || got.State != "archived" {
		t.Fatalf("终态口径不符: %+v", got)
	}
}

func TestCancelInstance(t *testing.T) {
	s, _ := newTestSvcWithTrigger(t)
	d := seedDefinition(t, s)
	ins, err := s.StartInstance(d.ID, "撤销单", "", "", "op", 5)
	if err != nil {
		t.Fatalf("发起: %v", err)
	}
	if _, err = s.SubmitAction(ins.ID, "cancel", "不要了", "op"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	got, _ := s.repo.FindInstance(s.dbh(), ins.ID)
	if got.Status != model.InstanceCanceled {
		t.Fatalf("撤销状态不符: %s", got.Status)
	}
}

func TestInstanceListPagingAndScope(t *testing.T) {
	s, _ := newTestSvcWithTrigger(t)
	d := seedDefinition(t, s)
	// 用户 5 发 3 单、用户 6 发 2 单
	for _, u := range []uint{5, 5, 5, 6, 6} {
		if _, err := s.StartInstance(d.ID, "单", "", "", "op", u); err != nil {
			t.Fatalf("发起: %v", err)
		}
	}
	// 非超管只见自己的
	list, total, err := s.GetInstanceList(1, 2, "", "", 5, false)
	if err != nil || total != 3 || len(list) != 2 {
		t.Fatalf("属主过滤口径不符: err=%v total=%d len=%d", err, total, len(list))
	}
	// 888 全量
	_, total, err = s.GetInstanceList(1, 0, "", "", 9, true)
	if err != nil || total != 5 {
		t.Fatalf("888 全量口径不符: err=%v total=%d", err, total)
	}
	// 888 分页缺省 pageSize=10
	list, _, err = s.GetInstanceList(0, 0, "", "", 9, true)
	if err != nil || len(list) != 5 {
		t.Fatalf("全量分页口径不符: err=%v len=%d", err, len(list))
	}
}

func TestInstanceLogsOwnership(t *testing.T) {
	s, _ := newTestSvcWithTrigger(t)
	d := seedDefinition(t, s)
	ins, err := s.StartInstance(d.ID, "别人的单", "", "", "op", 7)
	if err != nil {
		t.Fatalf("发起: %v", err)
	}
	// 非属主非超管 → 拒绝
	if _, err = s.GetInstanceLogs(ins.ID, 8, false); errx.CodeOf(err) != ErrCodeWfNotFound {
		t.Fatalf("越权读日志应 1903，实际 code=%d", errx.CodeOf(err))
	}
	// 属主可读
	if _, err = s.GetInstanceLogs(ins.ID, 7, false); err != nil {
		t.Fatalf("属主读日志不应报错: %v", err)
	}
}
