// Package service fanhua：语料库 DB 路径单测（data.OpenMemory 内存库注入，覆盖 repo 会话首参）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// TestCorpusServiceDB 语料 CRUD + 导出（内存库）：保存→列表分页→CSV 导出→删除/不存在
func TestCorpusServiceDB(t *testing.T) {
	db, err := data.OpenMemory(&model.FanhuaCorpus{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	s := NewFanhuaService(db)

	// 保存语料（新建）
	in := &model.FanhuaCorpus{Source: "新闻", Title: "沪语新闻", Lang: "沪语", Content: "今朝天气交关好"}
	out, err := s.SaveCorpus(in)
	if err != nil || out.ID == 0 {
		t.Fatalf("保存语料失败: %v id=%d", err, out.ID)
	}
	// 白名单外来源应拒（2601）
	if _, err = s.SaveCorpus(&model.FanhuaCorpus{Source: "体育", Title: "x", Content: "y"}); errx.CodeOf(err) != ErrCodeCorpusSource {
		t.Fatalf("非法来源应 2601，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	// 再插一条，验证列表过滤+分页
	if _, err = s.SaveCorpus(&model.FanhuaCorpus{Source: "文学", Title: "弄堂闲话", Content: "侬好"}); err != nil {
		t.Fatalf("保存第二条失败: %v", err)
	}
	list, total, err := s.GetCorpora("新闻", "", "", 1, 20)
	if err != nil || total != 1 || len(list) != 1 || list[0].Title != "沪语新闻" {
		t.Fatalf("来源过滤不符: err=%v total=%d list=%+v", err, total, list)
	}
	// CSV 导出：表头 + 全字段清洗口径（csvx）
	csv, err := s.ExportCorpora("", "")
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}
	if !strings.HasPrefix(csv, "来源,标题,语种,内容\n") || !strings.Contains(csv, "沪语新闻") {
		t.Fatalf("CSV 口径不符: %s", csv)
	}
	// 删除后再删 → 2600
	if err := s.DeleteCorpus(out.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if err := s.DeleteCorpus(out.ID); errx.CodeOf(err) != ErrCodeCorpusNotFound {
		t.Fatalf("二次删除应 2600，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
	// 更新路径：ID 存在校验（2600）
	miss := &model.FanhuaCorpus{Source: "新闻", Title: "x", Content: "y"}
	miss.ID = 999
	if _, err = s.SaveCorpus(miss); errx.CodeOf(err) != ErrCodeCorpusNotFound {
		t.Fatalf("更新不存在语料应 2600，实际 code=%d err=%v", errx.CodeOf(err), err)
	}
}
