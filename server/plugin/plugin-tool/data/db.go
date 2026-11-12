package data

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

// WithRowLock 若方言支持（非 sqlite），为查询追加 FOR UPDATE 行锁；sqlite 无此语法，
// 单测内存库为单连接串行、事务天然互斥，直接返回原查询。
// 生产（MySQL 系列）行为与手写 clause.Locking{Strength:"UPDATE"} 完全一致。
func WithRowLock(db *gorm.DB) *gorm.DB {
	if db == nil || db.Dialector == nil || db.Dialector.Name() == "sqlite" {
		return db
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

// OpenMemory 打开内存 sqlite 并 AutoMigrate 给定模型——供单测注入 service/repo 依赖，
// 实现 DB 路径真测。单连接串行（内存库多连接会得到各自独立的空库）。
func OpenMemory(models ...any) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(models...); err != nil {
		return nil, err
	}
	return db, nil
}
