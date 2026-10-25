package auto

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/plugin/auto/initialize"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Dictionary(ctx)
	initialize.Gorm(ctx)
	initialize.Router(engine)
}
