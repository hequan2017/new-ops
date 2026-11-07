// Package service llmops：干预机制（要求(8)）——会话人工接管 + 敏感词拦截
// 口径：接管中会话拒绝自动回复（返回接管提示）；敏感词两动作——block 命中即拒、
// replace 命中替换后放行；词匹配纯函数注入词表可测；chat 入口前置校验。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeIntervene = 2298

// takeoverNotice 接管中自动回复文案
const takeoverNotice = "本会话已由人工接管，自动回复暂停。请稍候或联系值守人员。"

// TakeOverSession 人工接管
func (s *LlmopsServiceStruct) TakeOverSession(sessionId uint, operator string) (*model.LlmopsAgentSession, error) {
	var sess model.LlmopsAgentSession
	if err := global.GVA_DB.First(&sess, sessionId).Error; err != nil {
		return nil, fmt.Errorf("[%d]会话不存在", ErrCodeSessionMiss)
	}
	if err := global.GVA_DB.Model(&sess).Updates(map[string]any{
		"taken_over": true, "taken_over_by": strings.TrimSpace(operator),
	}).Error; err != nil {
		return nil, err
	}
	sess.TakenOver, sess.TakenOverBy = true, operator
	return &sess, nil
}

// ReleaseSession 解除接管
func (s *LlmopsServiceStruct) ReleaseSession(sessionId uint) (*model.LlmopsAgentSession, error) {
	var sess model.LlmopsAgentSession
	if err := global.GVA_DB.First(&sess, sessionId).Error; err != nil {
		return nil, fmt.Errorf("[%d]会话不存在", ErrCodeSessionMiss)
	}
	if err := global.GVA_DB.Model(&sess).Updates(map[string]any{
		"taken_over": false, "taken_over_by": "",
	}).Error; err != nil {
		return nil, err
	}
	sess.TakenOver, sess.TakenOverBy = false, ""
	return &sess, nil
}

// wordHit 词命中结果（纯函数载体）
type wordHit struct {
	Word        string
	Action      string
	Replacement string
}

// matchSensitiveWords 词表匹配（纯函数，测试锚点）：返回命中的词与替换后文本
func matchSensitiveWords(text string, words []model.LlmopsSensitiveWord) (hits []wordHit, replaced string) {
	replaced = text
	for _, w := range words {
		if !w.Enabled || w.Word == "" {
			continue
		}
		if strings.Contains(text, w.Word) {
			hits = append(hits, wordHit{Word: w.Word, Action: w.Action, Replacement: w.Replacement})
			if w.Action == "replace" && w.Replacement != "" {
				replaced = strings.ReplaceAll(replaced, w.Word, w.Replacement)
			}
		}
	}
	return hits, replaced
}

// CheckTextInput 文本前置校验（chat 入口用）：block 命中 → 拒；replace → 返回净化文本
func (s *LlmopsServiceStruct) CheckTextInput(text string) (blocked bool, cleaned string, hits []string, err error) {
	var words []model.LlmopsSensitiveWord
	if err := global.GVA_DB.Where("enabled = ?", true).Find(&words).Error; err != nil {
		return false, text, nil, err
	}
	wordHits, replaced := matchSensitiveWords(text, words)
	if len(wordHits) == 0 {
		return false, text, nil, nil
	}
	names := make([]string, 0, len(wordHits))
	for _, h := range wordHits {
		names = append(names, h.Word)
	}
	for _, h := range wordHits {
		if h.Action == "block" {
			return true, text, names, fmt.Errorf("[%d]内容含受限词，已被拦截", ErrCodeIntervene)
		}
	}
	return false, replaced, names, nil
}

// ListSensitiveWords 敏感词列表
func (s *LlmopsServiceStruct) ListSensitiveWords() ([]model.LlmopsSensitiveWord, error) {
	var list []model.LlmopsSensitiveWord
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SaveSensitiveWord 新增/更新敏感词
func (s *LlmopsServiceStruct) SaveSensitiveWord(in *model.LlmopsSensitiveWord) (*model.LlmopsSensitiveWord, error) {
	if in == nil || strings.TrimSpace(in.Word) == "" {
		return nil, fmt.Errorf("[%d]敏感词不能为空", ErrCodeIntervene)
	}
	if in.Action != "block" && in.Action != "replace" {
		in.Action = "block"
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsSensitiveWord{}).Where("word = ?", in.Word)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]敏感词已存在", ErrCodeIntervene)
	}
	if in.ID > 0 {
		var w model.LlmopsSensitiveWord
		if err := global.GVA_DB.First(&w, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]敏感词不存在", ErrCodeIntervene)
		}
		err := global.GVA_DB.Model(&w).Updates(map[string]any{
			"word": in.Word, "action": in.Action, "replacement": in.Replacement, "enabled": in.Enabled,
		}).Error
		if err != nil {
			return nil, err
		}
		return &w, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteSensitiveWord 删除敏感词
func (s *LlmopsServiceStruct) DeleteSensitiveWord(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsSensitiveWord{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]敏感词不存在", ErrCodeIntervene)
	}
	return nil
}

// TakeOverIn 接管请求
type TakeOverIn struct {
	SessionId uint   `json:"sessionId" binding:"required"`
	Operator  string `json:"operator" binding:"required"`
}
