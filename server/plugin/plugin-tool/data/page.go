// Package data 白泽插件共享数据访问助手：统一分页查询、方言行锁与单测内存库夹具。
//
// Page 收敛 104 处手写 Count+Offset/Limit 的分页样板；OpenMemory 使 DB 路径可真测
// （glebarez/sqlite 已在 go.mod，内存库单连接串行）；WithRowLock 屏蔽 sqlite 无 FOR UPDATE
// 语法导致的单测不可运行。
package data

import (
	"gorm.io/gorm"
)

// FullPageCap 全量请求（page<=0）时的行数上限（口径同 gpu 实例台账既有约定）。
const FullPageCap = 500

// Page 统一分页查询：先 Count 后 Offset/Limit，返回总数；响应结构与既有列表接口一致。
// page>0：pageSize 钳位 [1,100]（缺省 20）；page<=0：视为全量请求，Limit(FullPageCap) 兜底。
// db 传入 Model+Where+Order 链；Order 先于 Count 对各主流方言均合法（Count 忽略排序语义）。
func Page[T any](db *gorm.DB, page, pageSize int, dest *[]T) (total int64, err error) {
	if err = db.Count(&total).Error; err != nil {
		return 0, err
	}
	if page > 0 {
		if pageSize <= 0 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}
		db = db.Offset((page - 1) * pageSize).Limit(pageSize)
	} else {
		db = db.Limit(FullPageCap)
	}
	if err = db.Find(dest).Error; err != nil {
		return 0, err
	}
	return total, nil
}
