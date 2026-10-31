#!/usr/bin/env bash
# 白泽本地冒烟：server/agent/web 构建+测试汇总。
# 用法：bash scripts/smoke.sh [--skip-web] [--skip-agent]
#   SKIP_WEB=1 / --skip-web    跳过前端构建
#   SKIP_AGENT=1 / --skip-agent 跳过 agent 验证
# 任一步骤失败脚本以非零退出，结尾输出汇总表。
set -u
cd "$(dirname "$0")/.."
ROOT="$(pwd)"

# Go 工具链：go.mod 要求 1.26，PATH 默认可能是过老的 1.18（本机实况），优先取 1.26
if [ -x /usr/local/go1.26/bin/go ]; then
  export PATH="/usr/local/go1.26/bin:$PATH"
fi

for a in "$@"; do
  case "$a" in
    --skip-web) SKIP_WEB=1 ;;
    --skip-agent) SKIP_AGENT=1 ;;
    *) echo "未知参数: $a"; exit 2 ;;
  esac
done

RESULTS=()
step() { # step <名称> <命令...>
  local name="$1"; shift
  printf '==> %s\n' "$name"
  if "$@"; then
    RESULTS+=("PASS  $name")
    echo "    OK"
  else
    RESULTS+=("FAIL  $name")
    echo "    FAILED"
  fi
}

step "server build" bash -c "cd $ROOT/server && go build ./..."
step "server vet"   bash -c "cd $ROOT/server && go vet ./..."
step "server test"  bash -c "cd $ROOT/server && go test ./..."

if [ "${SKIP_AGENT:-0}" != "1" ]; then
  step "agent build" bash -c "cd $ROOT/agent && go build ./..."
  step "agent vet"   bash -c "cd $ROOT/agent && go vet ./..."
  step "agent test"  bash -c "cd $ROOT/agent && go test ./..."
else
  RESULTS+=("SKIP  agent (SKIP_AGENT)")
fi

if [ "${SKIP_WEB:-0}" != "1" ]; then
  if [ -d "$ROOT/web/node_modules" ]; then
    step "web build" bash -c "cd $ROOT/web && npm run build --silent"
  else
    RESULTS+=("SKIP  web build (node_modules 缺失，先 npm install)")
  fi
else
  RESULTS+=("SKIP  web (SKIP_WEB)")
fi

echo
echo "========== 冒烟汇总 =========="
fail=0
for r in "${RESULTS[@]}"; do
  echo "$r"
  case "$r" in FAIL*) fail=$((fail+1)) ;; esac
done
echo "=============================="
if [ "$fail" -gt 0 ]; then
  echo "冒烟未通过：$fail 项失败"
  exit 1
fi
echo "冒烟全部通过"
