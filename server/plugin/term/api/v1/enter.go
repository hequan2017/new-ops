// Package api 白泽终端接口层
package api

import "github.com/hequan2017/new-ops/server/plugin/term/service"

var Api = new(api)

type api struct {
	Terminal terminal
}

var termService = service.TermService
