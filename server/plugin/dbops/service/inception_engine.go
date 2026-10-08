// Package service goInception 引擎客户端沉淀（块提交单一出口；工单载荷即送审对象，非注入面）
package service

import "database/sql"

// queryEngineBlock 向引擎提交一个 magic 块并返回审核结果集
func queryEngineBlock(db *sql.DB, block string) (*sql.Rows, error) {
	return db.Query(block)
}
