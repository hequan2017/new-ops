package ast

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestMain 在 -short 模式下跳过本包全部测试。
// 原因：utils/ast 的注入/回滚测试会真实修改仓库内源文件（如 initialize/gorm_biz.go），
// 属于代码生成集成测试；CI 与日常快速验证（go test -short ./...）不应触碰工作区源码。
func TestMain(m *testing.M) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.short") || strings.HasPrefix(arg, "--test.short") {
			fmt.Println("SHORT 模式：跳过会修改仓库源文件的 AST 注入/回滚集成测试")
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}
