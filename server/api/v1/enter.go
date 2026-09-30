package v1

import (
	"github.com/hequan2017/new-ops/server/api/v1/example"
	"github.com/hequan2017/new-ops/server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	ExampleApiGroup example.ApiGroup
}
