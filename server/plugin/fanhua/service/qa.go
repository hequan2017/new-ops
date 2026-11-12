// Package service fanhua：沪普英问答 API（需求 3.4-20）
// 口径：问答前置平行语料注入（N03 对照检索 top3）+ 价值观前置审查（N10 block 词即拒答）；
// ModelId>0 走 chat（复用 gradingChatFn），ModelId=0 规则口径（平行语料组合）；
// 翻译引擎走 TranslateAdapter 可替换点（默认 Noop 未配置口径）；时延落日志（1.2s 阈值标记）；
// elderMode 输出助老语音唤醒兼容口径（短句+唤醒提示）。
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	llmopsService "github.com/hequan2017/new-ops/server/plugin/llmops/service"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeQaParam    = 2625
	ErrCodeQaLLM      = 2626
	ErrCodeTranslateN = 2627
	ErrCodeQaBlocked  = 2628
)

// qaLatencyThresholdMs 响应时延阈值（需求口径 <1.2s）
const qaLatencyThresholdMs = 1200

// TranslateAdapter 翻译引擎可替换点（需求口径：实时翻译 <1.2s）
type TranslateAdapter interface {
	Translate(ctx context.Context, text, targetLang string) (string, error)
}

// NoopTranslateAdapter 默认适配器：未配置明确报错。
// 注：此处保留 "[码]" 前缀 fmt.Errorf 口径——既有单测锚定 err.Error() 含 "[2627]"，不可改测。
type NoopTranslateAdapter struct{}

func (NoopTranslateAdapter) Translate(context.Context, string, string) (string, error) {
	return "", fmt.Errorf("[%d]翻译引擎未配置", ErrCodeTranslateN)
}

// Translate 翻译单例（后续接入实时翻译引擎时替换）
var Translate TranslateAdapter = NoopTranslateAdapter{}

// qaChatFn chat 网关调用点（测试替换锚点；默认 llmops 网关，先例 exam J03）
var qaChatFn = func(ctx context.Context, modelId uint, prompt string) (string, error) {
	content, _, err := llmopsService.LlmopsService.ChatCompletion(ctx, modelId,
		[]llmopsService.ChatMessage{{Role: "user", Content: prompt}}, nil)
	return content, err
}

// buildQaPrompt 问答提示词（注入平行语料对照与助老口径）
func buildQaPrompt(question string, refs []paraHit, elderMode bool) string {
	var sb strings.Builder
	sb.WriteString("你是沪语-普通话-英语三语问答助手。用沪语回答并给出普通话对照，不输出任何不合规内容。\n")
	if len(refs) > 0 {
		sb.WriteString("参考对照语料：\n")
		for _, r := range refs {
			fmt.Fprintf(&sb, "- 沪：%s / 普：%s\n", r.Sh, r.Pu)
		}
	}
	if elderMode {
		sb.WriteString("助老模式：句子短、语气慢，先给结论再解释，结尾提示可用语音再次提问。\n")
	}
	sb.WriteString("问题：" + question)
	return sb.String()
}

// ruleQaAnswer 规则口径回答（纯函数，测试锚点）：平行语料组合
func ruleQaAnswer(question string, refs []paraHit) string {
	if len(refs) == 0 {
		return "暂无对照语料可回答该问题（规则口径）。可在知识库补充沪-普对照后重试。"
	}
	var sb strings.Builder
	for i, r := range refs {
		if i > 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "沪语参考：%s\n普通话：%s", r.Sh, r.Pu)
		if r.En != "" {
			fmt.Fprintf(&sb, "\nEnglish: %s", r.En)
		}
	}
	return sb.String()
}

// truncateRunText2 错误留痕截断（500 字节）。与 text.Truncate 不等价（前者按字节硬截
// 防 DB 列溢出、无省略号；后者按 rune 截断并加「…」），保留本地实现。
func truncateRunText2(s string) string {
	if len(s) > 500 {
		return s[:500]
	}
	return s
}

// DialectQA 方言问答主流程（前置审查→检索注入→双引擎→日志）
func (s *FanhuaService) DialectQA(question string, modelId uint, elderMode bool) (*model.FanhuaDialectQaLog, error) {
	if strings.TrimSpace(question) == "" {
		return nil, errx.New(ErrCodeQaParam, "提问必填")
	}
	// 前置价值观审查（N10 口径复用；词表读取失败不阻断问答，与原口径一致）
	words, _ := s.repo.EnabledWords(s.dbh())
	health := computeHealth(question, words)
	if health.Blocked {
		log := &model.FanhuaDialectQaLog{Question: question, Engine: "rule", ModelId: modelId, WakeCompat: elderMode, Ok: false, ErrMsg: "提问含拦截词，已按价值观规范拒答"}
		_ = s.repo.CreateDialectQaLog(s.dbh(), log)
		return log, errx.New(ErrCodeQaBlocked, "提问包含拦截词，已按价值观规范拒答")
	}
	// 平行语料注入
	refs, _ := s.SearchParallel(question, 3)
	log := &model.FanhuaDialectQaLog{Question: question, ModelId: modelId, WakeCompat: elderMode, CorpusHits: len(refs)}
	started := time.Now()
	var answer string
	var execErr error
	if modelId > 0 {
		content, err := qaChatFn(context.Background(), modelId, buildQaPrompt(question, refs, elderMode))
		if err != nil {
			execErr = errx.Wrap(ErrCodeQaLLM, "模型调用失败", err)
		} else {
			answer = content
			log.Engine = "llm"
		}
	} else {
		answer = ruleQaAnswer(question, refs)
		log.Engine = "rule"
	}
	log.LatencyMs = time.Since(started).Milliseconds()
	log.OverThreshold = log.LatencyMs > qaLatencyThresholdMs
	if elderMode && execErr == nil {
		answer = "【语音助老】" + answer + "\n（可再次语音提问）"
	}
	log.Answer, log.Ok = answer, execErr == nil
	if execErr != nil {
		log.ErrMsg = truncateRunText2(execErr.Error())
	}
	if err := s.repo.CreateDialectQaLog(s.dbh(), log); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "落问答日志失败", err)
	}
	if execErr != nil {
		return log, execErr
	}
	return log, nil
}

// GetDialectQaLogs 问答日志列表（倒序分页）
func (s *FanhuaService) GetDialectQaLogs(page, pageSize int) ([]model.FanhuaDialectQaLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	list, total, err := s.repo.DialectQaLogList(s.dbh(), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询问答日志失败", err)
	}
	return list, total, nil
}
