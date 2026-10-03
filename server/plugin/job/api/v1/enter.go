// Package api 白泽批量作业接口层
package api

var Api = new(api)

type api struct {
	BatchExec batchExec
}
