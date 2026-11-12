// Package service artsci：DB 路径单测（data.OpenMemory 内存库注入，repo 层 GORM 真测）。
// 既有纯函数测试不动，本文件只覆盖关键 DB 路径：重名守卫/留痕落库/任务状态机。
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// newTestArtsciSvc 内存库注入服务（repo 层 GORM 路径真测）
func newTestArtsciSvc(t *testing.T) *ArtsciService {
	t.Helper()
	db, err := data.OpenMemory(
		&model.ArtsciTool{}, &model.ArtsciRunLog{},
		&model.ArtsciQuestion{}, &model.ArtsciPaper{},
		&model.ArtsciSceneTemplate{}, &model.ArtsciImageTask{},
	)
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	return NewArtsciService(db)
}

// TestSaveToolDuplicateGuard 重名守卫：新建撞名拒绝、更新排除自身；停用态白名单兜底
func TestSaveToolDuplicateGuard(t *testing.T) {
	s := newTestArtsciSvc(t)
	in := &model.ArtsciTool{Name: "Allen-Cahn 相场演化", Subject: "数学", Module: "allen-cahn", Status: "enabled"}
	out, err := s.SaveTool(in)
	if err != nil || out.ID == 0 {
		t.Fatalf("首次保存应成功: err=%v id=%d", err, out.ID)
	}
	// 同名新建 → 2503
	if _, err = s.SaveTool(&model.ArtsciTool{Name: in.Name, Subject: "数学", Module: "allen-cahn"}); errx.CodeOf(err) != ErrCodeToolNameDup {
		t.Fatalf("重名应 2503，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	// 自身更新不撞名（excludeId 生效）
	out.Desc = "更新描述"
	if _, err = s.SaveTool(out); err != nil {
		t.Fatalf("更新自身不应报重名: %v", err)
	}
	rows, err := s.repo.ToolCountByName(s.dbh(), in.Name, 0)
	if err != nil || rows != 1 {
		t.Fatalf("重名守卫后应仍只 1 行: rows=%d err=%v", rows, err)
	}
}

// TestRunToolPersistsLogBothOutcomes 执行留痕成败均落库 + 停用守卫
func TestRunToolPersistsLogBothOutcomes(t *testing.T) {
	s := newTestArtsciSvc(t)
	RegisterModule("dbtest-mod", func(args map[string]any) (any, error) {
		return map[string]any{"echo": args["text"]}, nil
	})
	t.Cleanup(func() { delete(builtinModules, "dbtest-mod") })

	tool, err := s.SaveTool(&model.ArtsciTool{
		Name: "回显工具", Subject: "数学", Module: "dbtest-mod", Status: "enabled",
		ParamsSchema: `{"text":{"type":"string","required":true}}`,
	})
	if err != nil {
		t.Fatalf("seed 工具: %v", err)
	}
	// 成功：留痕 Ok=true 且有出参
	log, err := s.RunTool(tool.ID, `{"text":"hi"}`)
	if err != nil || !log.Ok || log.ResultJSON == "" || log.LatencyMs < 0 {
		t.Fatalf("成功执行口径不符: err=%v log=%+v", err, log)
	}
	// 失败（schema 必填缺失）：留痕 Ok=false 带错误消息，返回 2504
	log2, err := s.RunTool(tool.ID, `{}`)
	if errx.CodeOf(err) != ErrCodeToolArgs {
		t.Fatalf("缺必填应 2504，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	if log2 == nil || log2.Ok || log2.ErrMsg == "" {
		t.Fatalf("失败执行应落留痕: %+v", log2)
	}
	// 停用守卫：2506 且不留痕
	if err = s.repo.UpdateToolColumns(s.dbh(), tool.ID, map[string]any{"status": "disabled"}); err != nil {
		t.Fatalf("停用: %v", err)
	}
	if _, err = s.RunTool(tool.ID, `{"text":"x"}`); errx.CodeOf(err) != ErrCodeToolDisabled {
		t.Fatalf("停用应 2506，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	list, total, err := s.GetRunLogs(tool.ID, 1, 20)
	if err != nil || total != 2 || len(list) != 2 {
		t.Fatalf("留痕应 2 条（成败均落库）: total=%d err=%v", total, err)
	}
}

// TestSubmitImageTaskStateMachine 任务状态机：queued→done（假适配器）/failed（未配置适配器），模板缺失 2516
func TestSubmitImageTaskStateMachine(t *testing.T) {
	s := newTestArtsciSvc(t)
	if err := s.dbh().Create(&model.ArtsciSceneTemplate{
		Scene: "历史场景再现", Status: "enabled", PromptTemplate: "场景 {{scene}}：{{text}}",
	}).Error; err != nil {
		t.Fatalf("seed 模板: %v", err)
	}
	// 假适配器 → done
	old := ImageGen
	ImageGen = fakeImageGen{}
	t.Cleanup(func() { ImageGen = old })

	task, err := s.SubmitImageTask("历史场景再现", "淞沪抗战")
	if err != nil || task.Status != "done" || task.ImageUrl == "" {
		t.Fatalf("假适配器应 done: err=%v task=%+v", err, task)
	}
	if !strings.Contains(task.RenderedPrompt, "淞沪抗战") {
		t.Fatalf("渲染提示词应含输入文本: %q", task.RenderedPrompt)
	}
	// 未配置适配器 → failed（任务返回但不报错，错误进留痕）
	ImageGen = NoopImageGenAdapter{}
	task2, err := s.SubmitImageTask("历史场景再现", "第二次提交")
	if err != nil || task2.Status != "failed" || !strings.Contains(task2.ErrMsg, "2515") {
		t.Fatalf("Noop 适配器应 failed: err=%v task=%+v", err, task2)
	}
	// 模板缺失 → 2516，不落任务
	if _, err = s.SubmitImageTask("不存在场景", "文本"); errx.CodeOf(err) != ErrCodeImgGenParam {
		t.Fatalf("模板缺失应 2516，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	_, total, err := s.GetImageTasks("历史场景再现", 1, 20)
	if err != nil || total != 2 {
		t.Fatalf("任务落库应 2 条: total=%d err=%v", total, err)
	}
}
