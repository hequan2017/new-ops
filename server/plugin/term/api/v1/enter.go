// Package api 白泽终端接口层
package api

import (
	"strconv"
)

var Api = new(api)

type api struct {
	Terminal    terminal
	TermSession termSession
	Sftp        sftpApi
}

// parseUintQuery 安全解析 uint 查询参数
func parseUintQuery(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	return uint(v), err
}
