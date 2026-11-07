// Package service oplib：脚本算子受限执行（采购件② 自定义算子）
// 口径：解释器白名单（python3/node）；脚本与参数落临时目录执行（args 经 argv[1] 文件路径传入），
// 30s 超时强杀 + 输出 1MB 上限 + temp 目录 defer 清理；进程级隔离（容器级沙箱待运维域接入，注明口径）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
)

// 错误码（oplib 段续用）
const (
	ErrCodeScriptLang = 2311
	ErrCodeScriptExec = 2312
)

// scriptTimeout 脚本执行超时（var 便于测试注入）
var scriptTimeout = 30 * time.Second

const (
	scriptMaxOutput = 1 << 20
	scriptMaxCode   = 64 << 10
)

// ValidScriptLang 解释器白名单
func ValidScriptLang(lang string) bool {
	return lang == "python3" || lang == "node"
}

// RunScriptOperator 脚本算子执行（脚本文件 + 参数文件 → 临时目录 → 执行 → 清理）
func RunScriptOperator(op *model.OplibOperator, argsJSON string) (any, error) {
	if !ValidScriptLang(op.ScriptLang) {
		return nil, fmt.Errorf("[%d]解释器仅支持 python3/node（当前 %s）", ErrCodeScriptLang, op.ScriptLang)
	}
	if len(op.ScriptCode) > scriptMaxCode {
		return nil, fmt.Errorf("[%d]脚本超过 64KB 上限", ErrCodeScriptLang)
	}
	if _, err := exec.LookPath(op.ScriptLang); err != nil {
		return nil, fmt.Errorf("[%d]解释器不可用: %s", ErrCodeScriptExec, op.ScriptLang)
	}
	dir, err := os.MkdirTemp("", "oplib-script-")
	if err != nil {
		return nil, fmt.Errorf("[%d]临时目录创建失败: %w", ErrCodeScriptExec, err)
	}
	defer os.RemoveAll(dir) // 清理护栏
	scriptPath := fmt.Sprintf("%s/main.%s", dir, scriptExt(op.ScriptLang))
	if err := os.WriteFile(scriptPath, []byte(op.ScriptCode), 0o700); err != nil {
		return nil, fmt.Errorf("[%d]脚本写入失败: %w", ErrCodeScriptExec, err)
	}
	argsPath := fmt.Sprintf("%s/args.json", dir)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0o600); err != nil {
		return nil, fmt.Errorf("[%d]参数写入失败: %w", ErrCodeScriptExec, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), scriptTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, op.ScriptLang, scriptPath, argsPath)
	out, err := cmd.CombinedOutput()
	if len(out) > scriptMaxOutput {
		out = out[:scriptMaxOutput]
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("[%d]脚本执行超时(%s)", ErrCodeScriptExec, scriptTimeout)
	}
	if err != nil {
		tail := string(out)
		if len(tail) > 500 {
			tail = tail[len(tail)-500:]
		}
		return nil, fmt.Errorf("[%d]脚本执行失败: %s", ErrCodeScriptExec, truncateSample(tail))
	}
	// 输出契约：脚本须输出 JSON（对象或其它 JSON 值）；非 JSON 原文返回字符串
	var parsed any
	if err := json.Unmarshal(out, &parsed); err != nil {
		return strings.TrimSpace(string(out)), nil
	}
	return parsed, nil
}

func scriptExt(lang string) string {
	if lang == "node" {
		return "js"
	}
	return "py"
}

// saveScriptFields 保存脚本算子的语言/代码校验（SaveOperator 对 script 型调用）
func saveScriptFields(op *model.OplibOperator) error {
	if op.ImplType != "script" {
		return nil
	}
	if !ValidScriptLang(op.ScriptLang) {
		return fmt.Errorf("[%d]脚本算子解释器仅支持 python3/node", ErrCodeScriptLang)
	}
	if strings.TrimSpace(op.ScriptCode) == "" {
		return fmt.Errorf("[%d]脚本算子必须提供代码", ErrCodeScriptLang)
	}
	if len(op.ScriptCode) > scriptMaxCode {
		return fmt.Errorf("[%d]脚本超过 64KB 上限", ErrCodeScriptLang)
	}
	return nil
}
