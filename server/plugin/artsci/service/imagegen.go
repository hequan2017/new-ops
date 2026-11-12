// Package service artsci：文生图教学场景库（需求 3.3-10/11/12/14/15）
// 口径：场景模板渲染 {{text}}/{{scene}} 占位；生成走 ImageGenAdapter 可替换点——
// 默认 NoopImageGenAdapter 报未配置（对齐 llmops A09 ASR 口径），任务状态机 queued/done/failed。
package service

import (
	"context"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（artsci 段 2500-2599）
const (
	ErrCodeImgGenNoop  = 2515
	ErrCodeImgGenParam = 2516
)

// ImageGenAdapter 文生图适配器可替换点
type ImageGenAdapter interface {
	Generate(ctx context.Context, prompt string) (url string, err error)
}

// NoopImageGenAdapter 默认适配器：未配置明确报错
type NoopImageGenAdapter struct{}

func (NoopImageGenAdapter) Generate(context.Context, string) (string, error) {
	return "", artErr(ErrCodeImgGenNoop, "文生图适配器未配置")
}

// ImageGen 适配器单例（后续模型服务接入时替换）
var ImageGen ImageGenAdapter = NoopImageGenAdapter{}

// SceneNames 五类教学场景白名单
var SceneNames = []string{"历史场景再现", "文学作品可视化", "地理地貌展示", "图像作文训练", "跨学科学习"}

// renderPrompt 模板渲染（纯函数，测试锚点）
func renderPrompt(tpl, scene, input string) string {
	out := strings.ReplaceAll(tpl, "{{scene}}", scene)
	return strings.ReplaceAll(out, "{{text}}", input)
}

// GetSceneTemplates 场景模板列表
func (s *ArtsciService) GetSceneTemplates() ([]model.ArtsciSceneTemplate, error) {
	list, err := s.repo.SceneTemplateList(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeArtsciStore, "查询场景模板失败", err)
	}
	return list, nil
}

// SubmitImageTask 提交文生图任务（渲染→适配器生成→状态机落库）
func (s *ArtsciService) SubmitImageTask(scene, inputText string) (*model.ArtsciImageTask, error) {
	if strings.TrimSpace(inputText) == "" {
		return nil, artErr(ErrCodeImgGenParam, "教学输入文本必填")
	}
	db := s.dbh()
	tpl, err := s.repo.FindSceneTemplate(db, scene)
	if err != nil {
		return nil, artErr(ErrCodeImgGenParam, "场景模板不存在: %s", scene)
	}
	task := &model.ArtsciImageTask{
		Scene: scene, InputText: inputText,
		RenderedPrompt: renderPrompt(tpl.PromptTemplate, scene, inputText), Status: "queued",
	}
	if err := s.repo.CreateImageTask(db, task); err != nil {
		return nil, errx.Wrap(ErrCodeArtsciStore, "落库文生图任务失败", err)
	}
	url, err := ImageGen.Generate(context.Background(), task.RenderedPrompt)
	if err != nil {
		task.Status, task.ErrMsg = "failed", truncateRunText(err.Error())
	} else {
		task.Status, task.ImageUrl, task.ErrMsg = "done", url, ""
	}
	if err := s.repo.UpdateImageTaskColumns(db, task.ID, map[string]any{
		"status": task.Status, "image_url": task.ImageUrl, "err_msg": task.ErrMsg,
	}); err != nil {
		return nil, errx.Wrap(ErrCodeArtsciStore, "更新文生图任务状态失败", err)
	}
	return task, nil
}

// GetImageTasks 任务列表（场景过滤，倒序分页）
func (s *ArtsciService) GetImageTasks(scene string, page, pageSize int) ([]model.ArtsciImageTask, int64, error) {
	list, total, err := s.repo.ImageTaskList(s.dbh(), scene, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeArtsciStore, "查询文生图任务列表失败", err)
	}
	return list, total, nil
}
