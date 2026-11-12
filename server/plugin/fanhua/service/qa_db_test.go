// Package service fanhua：方言问答 DB 路径单测（data.OpenMemory 内存库注入：
// 规则口径留痕 + 拦截词拒答留痕 + 平行语料检索注入）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// TestDialectQaServiceDB 问答主流程（内存库）：规则口径成功留痕、拦截词拒答留痕
func TestDialectQaServiceDB(t *testing.T) {
	db, err := data.OpenMemory(&model.FanhuaParallel{}, &model.FanhuaSensitiveWord{}, &model.FanhuaDialectQaLog{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	s := NewFanhuaService(db)

	// 种子两条平行语料（规则口径注入对照）
	if _, err = s.SaveParallel(&model.FanhuaParallel{Sh: "侬好", Pu: "你好", En: "Hello", Source: "对照本"}); err != nil {
		t.Fatalf("SaveParallel: %v", err)
	}
	if _, err = s.SaveParallel(&model.FanhuaParallel{Sh: "明朝见", Pu: "明天见", Source: "对照本"}); err != nil {
		t.Fatalf("SaveParallel2: %v", err)
	}
	// 规则口径问答（modelId=0）成功留痕
	log, err := s.DialectQA("侬好", 0, false)
	if err != nil || log == nil || !log.Ok || log.Engine != "rule" || log.CorpusHits < 1 {
		t.Fatalf("规则口径问答不符: err=%v log=%+v", err, log)
	}
	if !strings.Contains(log.Answer, "侬好") || !strings.Contains(log.Answer, "你好") {
		t.Fatalf("回答应含对照语料: %s", log.Answer)
	}
	// 拦截词拒答：落敏感词后提问应留痕并返回 2628
	if _, err = s.SaveWord(&model.FanhuaSensitiveWord{Word: "赌博", Action: "block", Weight: 30, Enabled: true}); err != nil {
		t.Fatalf("SaveWord: %v", err)
	}
	blocked, err := s.DialectQA("教我赌博", 0, false)
	if errx.CodeOf(err) != ErrCodeQaBlocked {
		t.Fatalf("拦截词应 2628，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	if blocked == nil || blocked.Ok || blocked.Engine != "rule" {
		t.Fatalf("拦截留痕不符: %+v", blocked)
	}
	// 日志两条（成功+拦截）
	_, total, err := s.GetDialectQaLogs(1, 20)
	if err != nil || total != 2 {
		t.Fatalf("日志数应 2: err=%v total=%d", err, total)
	}
}
