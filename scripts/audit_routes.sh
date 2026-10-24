#!/usr/bin/env bash
# 白泽 · 插件路由与 casbin/SysApi 种子对账
# 原理：从各插件 router/initialize 提取注册路径（枚举 Group 前缀组合），与 initialize 中
# 登记的完整路径（casbin/SysApi 种子）比对；种子中的 :param 动态段规范化为 * 通配。
# 任一组合都无法匹配种子的路由输出 WARN（可能属 public WS/webhook 免 JWT 通道，人工复核）。
# 用法：bash scripts/audit_routes.sh
set -euo pipefail
REPO=$(cd "$(dirname "$0")/.." && pwd)
EXIT=0

for p in asset term job pipeline container k8s gpu dbops monitor workflow agent aiops; do
  dir="$REPO/server/plugin/$p"
  [ -d "$dir" ] || continue
  routes=$(grep -rhoE '\.(GET|POST|PUT|DELETE)\("[^"]*"' \
    "$dir/router" "$dir/initialize/router.go" 2>/dev/null | grep -oE '"[^"]*"' | tr -d '"' | sort -u || true)
  groups=$(grep -rhoE 'Group\("[^"]*"\)' \
    "$dir/router" "$dir/initialize/router.go" 2>/dev/null | grep -oE '"[^"]*"' | tr -d '"' || true)
  # 种子规范化：:id/:token 等动态段 → * 通配
  seeds=$(grep -rhoE "\"/$p/[^\"]*\"" "$dir/initialize" 2>/dev/null | tr -d '"' | sed -E 's/:[^/]+/*/g' | sort -u || true)
  for r in $routes; do
    ok=0
    for g in "" $groups; do
      # 插件间 Group 字面量约定不一（有的自带 "asset/" 前缀，有的只有 "host"）：候选包含两种形式
      if [ -n "$g" ]; then cands=("/$p/$g/$r" "/$g/$r"); else cands=("/$p/$r" "/$r"); fi
      for full in "${cands[@]}"; do
        for pat in $seeds; do
          # shellcheck disable=SC2254  # glob 匹配需要未引用的 pattern
          case "$full" in
            $pat) ok=1; break 3 ;;
          esac
        done
      done
    done
    if [ "$ok" = "0" ]; then
      echo "WARN [$p] $r 无法以任何组前缀匹配 casbin/SysApi 种子"
      EXIT=1
    fi
  done
done
if [ "$EXIT" = "0" ]; then
  echo "✅ 路由与 casbin/SysApi 种子对账无漂移"
else
  echo "（以上 WARN 请人工确认为 public 通道或补种子）"
fi
exit $EXIT
