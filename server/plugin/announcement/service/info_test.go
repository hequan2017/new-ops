package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/announcement/model"
	"github.com/hequan2017/new-ops/server/plugin/announcement/model/request"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/data"
)

func newTestService(t *testing.T) *InfoService {
	t.Helper()
	db, err := data.OpenMemory(&model.Info{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	return NewInfoService(db)
}

func TestInfoCRUD(t *testing.T) {
	s := newTestService(t)
	// 创建
	if err := s.CreateInfo(&model.Info{Title: "标题一", Content: "内容"}); err != nil {
		t.Fatalf("创建: %v", err)
	}
	// 查询
	got, err := s.GetInfo("1")
	if err != nil || got.Title != "标题一" {
		t.Fatalf("查询: err=%v title=%q", err, got.Title)
	}
	// 更新
	got.Title = "标题二"
	if err = s.UpdateInfo(got); err != nil {
		t.Fatalf("更新: %v", err)
	}
	got2, _ := s.GetInfo("1")
	if got2.Title != "标题二" {
		t.Fatalf("更新未生效: %q", got2.Title)
	}
	// 列表（标准分页）
	list, total, err := s.GetInfoInfoList(request.InfoSearch{})
	if err != nil {
		t.Fatalf("列表: %v", err)
	}
	_ = list
	// 分页参数语义：PageSize==0 全量
	if total != 1 {
		t.Fatalf("总数口径不符: %d", total)
	}
	// 批量删除
	if err = s.DeleteInfoByIds([]string{"1"}); err != nil {
		t.Fatalf("批量删除: %v", err)
	}
	if _, err = s.GetInfo("1"); err == nil {
		t.Fatal("删除后应查无此记录")
	}
}

func TestInfoListPaging(t *testing.T) {
	s := newTestService(t)
	for i := 0; i < 5; i++ {
		if err := s.CreateInfo(&model.Info{Title: "t"}); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	// PageSize=2 Page=2 → 第二页 2 条，总数 5
	search := request.InfoSearch{}
	search.Page = 2
	search.PageSize = 2
	list, total, err := s.GetInfoInfoList(search)
	if err != nil || total != 5 || len(list) != 2 {
		t.Fatalf("分页口径不符: err=%v total=%d len=%d", err, total, len(list))
	}
	// PageSize=0 全量（模板既有约定：不设 Limit）
	search2 := request.InfoSearch{}
	list2, total2, err := s.GetInfoInfoList(search2)
	if err != nil || total2 != 5 || len(list2) != 5 {
		t.Fatalf("全量口径不符: err=%v total=%d len=%d", err, total2, len(list2))
	}
}
