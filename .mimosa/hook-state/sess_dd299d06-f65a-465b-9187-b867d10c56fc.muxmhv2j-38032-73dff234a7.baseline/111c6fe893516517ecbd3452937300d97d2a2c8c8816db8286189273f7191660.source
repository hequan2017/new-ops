package core

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/initialize"
	mcpTool "github.com/hequan2017/new-ops/server/mcp"
	"github.com/hequan2017/new-ops/server/service/system"
	"go.uber.org/zap"
)

func RunServer() {
	if global.GVA_CONFIG.System.UseRedis {
		initialize.Redis()
		if global.GVA_CONFIG.System.UseMultipoint {
			initialize.RedisList()
		}
	}

	if global.GVA_CONFIG.System.UseMongo {
		if err := initialize.Mongo.Initialization(); err != nil {
			zap.L().Error(fmt.Sprintf("%+v", err))
		}
	}

	if global.GVA_DB != nil {
		system.LoadAll()
	}

	Router := initialize.Routers()
	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)
	mcpBaseURL := mcpTool.ResolveMCPServiceURL()

	fmt.Printf(`
	欢迎使用 白泽（BaiZe）· 统一运维开发平台
	当前版本:%s
	项目地址:https://github.com/hequan2017/new-ops
	默认自动化文档地址:http://127.0.0.1%s/swagger/index.html
	MCP 独立服务请手动启动: go run ./cmd/mcp -config ./cmd/mcp/config.yaml
	默认MCP StreamHTTP地址:%s
	默认前端文件运行地址:http://127.0.0.1:8080
`, global.Version, address, mcpBaseURL)

	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}
