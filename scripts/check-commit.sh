#!/usr/bin/env bash
# 白泽提交信息规范检查（只读，不修改任何提交）。
# 用法：
#   bash scripts/check-commit.sh            # 检查最近 20 条提交
#   bash scripts/check-commit.sh <ref> [n]  # 从 <ref> 起向后检查 n 条（默认 20）
# 规范口径：type(scope): 中文说明（type ∈ feat|fix|docs|chore|test|refactor|style|perf|ci；
# scope 可选；允许冒号后带空格；正文换行不校验）。
set -u
START_REF=${1:-HEAD}
COUNT=${2:-20}

PATTERN='^(feat|fix|docs|chore|test|refactor|style|perf|ci)(\([^)]+\))?: .+'

fail=0
total=0
while IFS= read -r line; do
  [ -z "$line" ] && continue
  hash=${line%% *}
  msg=${line#* }
  total=$((total + 1))
  if echo "$msg" | grep -qE "$PATTERN"; then
    printf 'OK    %s %s\n' "$hash" "$msg"
  else
    fail=$((fail + 1))
    printf 'BAD   %s %s\n' "$hash" "$msg"
  fi
done < <(git log --format='%h %s' -n "$COUNT" "$START_REF")

echo "----------------------------------------"
if [ "$fail" -gt 0 ]; then
  echo "规范检查：$total 条中 $fail 条不符合 type(scope): 描述 口径"
  exit 1
fi
echo "规范检查：$total 条全部符合"
