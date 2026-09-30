package router

import (
	"github.com/hequan2017/new-ops/server/router/example"
	"github.com/hequan2017/new-ops/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System  system.RouterGroup
	Example example.RouterGroup
}
