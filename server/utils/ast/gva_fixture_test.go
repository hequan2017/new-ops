package ast

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hequan2017/new-ops/server/global"
)

// skipWithoutGvaPlugin 上游 gin-vue-admin 的 AST 集成测试以 gva 示例插件源文件为靶子
// （会真实改写其源码）。本仓库未携带 plugin/gva——目标缺失时跳过，保留上游逻辑以待
// 插件存在时生效；避免全量 go test 因恒定失败红屏（nil file 注入 panic 已另行加守卫）。
func skipWithoutGvaPlugin(t *testing.T) {
	t.Helper()
	dir := filepath.Join(global.GVA_CONFIG.AutoCode.Root, global.GVA_CONFIG.AutoCode.Server, "plugin", "gva")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("本仓库未携带 gva 示例插件，跳过以其源文件为靶子的 AST 集成测试")
	}
}
