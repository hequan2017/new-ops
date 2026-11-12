package data

import (
	"testing"

	"gorm.io/gorm"
)

// demoRow 分页/内存库共用演示模型（sqlite 兼容的极简列）
type demoRow struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:64"`
	Score int
}

func mustOpen(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := OpenMemory(&demoRow{})
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	return db
}

func TestPage(t *testing.T) {
	db := mustOpen(t)
	for i := 1; i <= 25; i++ {
		if err := db.Create(&demoRow{Name: "n", Score: i}).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// 标准分页（契约：db 须已设 Model，Count 才能定表）
	var page2 []demoRow
	total, err := Page(db.Model(&demoRow{}).Where("name = ?", "n"), 2, 10, &page2)
	if err != nil {
		t.Fatalf("Page: %v", err)
	}
	if total != 25 || len(page2) != 10 || page2[0].Score != 11 || page2[9].Score != 20 {
		t.Fatalf("分页口径不符: total=%d len=%d first=%d", total, len(page2), page2[0].Score)
	}

	// pageSize 钳位：超 100 收 100
	var capped []demoRow
	total, err = Page(db.Model(&demoRow{}), 1, 999, &capped)
	if err != nil || total != 25 || len(capped) != 25 {
		t.Fatalf("钳位口径不符: err=%v total=%d len=%d", err, total, len(capped))
	}

	// pageSize 缺省 20
	var defaultSz []demoRow
	if _, err = Page(db.Model(&demoRow{}), 2, 0, &defaultSz); err != nil || len(defaultSz) != 5 {
		t.Fatalf("缺省 pageSize 口径不符: err=%v len=%d", err, len(defaultSz))
	}

	// page<=0 视为全量，上限 FullPageCap
	var all []demoRow
	total, err = Page(db.Model(&demoRow{}), 0, 10, &all)
	if err != nil || total != 25 || len(all) != 25 {
		t.Fatalf("全量口径不符: err=%v total=%d len=%d", err, total, len(all))
	}

	// Order 先于 Count 合法且不影响计数
	var ordered []demoRow
	total, err = Page(db.Model(&demoRow{}).Order("score DESC"), 1, 3, &ordered)
	if err != nil || total != 25 || ordered[0].Score != 25 || len(ordered) != 3 {
		t.Fatalf("排序+分页口径不符: err=%v total=%d len=%d first=%d", err, total, len(ordered), ordered[0].Score)
	}
}

func TestOpenMemoryIsolation(t *testing.T) {
	// 两次 OpenMemory 互不串库（各自独立内存库）
	dbA, dbB := mustOpen(t), mustOpen(t)
	if err := dbA.Create(&demoRow{Name: "onlyA"}).Error; err != nil {
		t.Fatalf("seed A: %v", err)
	}
	var cntA, cntB int64
	if err := dbA.Model(&demoRow{}).Count(&cntA).Error; err != nil {
		t.Fatalf("count A: %v", err)
	}
	if err := dbB.Model(&demoRow{}).Count(&cntB).Error; err != nil {
		t.Fatalf("count B: %v", err)
	}
	if cntA != 1 || cntB != 0 {
		t.Fatalf("内存库隔离失效: A=%d B=%d", cntA, cntB)
	}
}

func TestWithRowLock(t *testing.T) {
	// sqlite 方言：不追加 FOR UPDATE（语法不支持），查询仍可执行
	db := mustOpen(t)
	if err := db.Create(&demoRow{Name: "onlyA"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	var row demoRow
	if err := WithRowLock(db.Model(&demoRow{}).Where("name = ?", "onlyA")).First(&row).Error; err != nil {
		t.Fatalf("sqlite 行锁豁免查询失败: %v", err)
	}
	if row.Name != "onlyA" {
		t.Fatalf("查询结果不符: %q", row.Name)
	}
	// nil 入参安全
	if WithRowLock(nil) != nil {
		t.Fatal("nil db 应原样返回")
	}
}
