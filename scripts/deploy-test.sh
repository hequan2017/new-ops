#!/usr/bin/env bash
# new-ops 测试环境一键部署：构建 → 上传 → 重启 → 冒烟
# 目标机器与目录约定见 docs/DEV_PLAN.md「测试环境」一节
# 注意：config.yaml 是测试机上的有状态文件（含数据库初始化信息），本脚本不覆盖；
#       仅首次部署时需手工下发一次（见 DEV_PLAN 测试环境）。
set -euo pipefail

HOST=${NEW_OPS_TEST_HOST:-root@192.168.112.138}
REMOTE_DIR=/opt/new-ops
REPO=$(cd "$(dirname "$0")/.." && pwd)
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

echo "[1/5] 编译后端 (linux/amd64)"
(cd "$REPO/server" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$STAGE/new-ops-server" .)

echo "[2/5] 构建前端"
(cd "$REPO/web" && npm run build --silent >/dev/null)
tar -czf "$STAGE/dist.tar.gz" -C "$REPO/web" dist

echo "[3/5] 上传"
ssh "$HOST" "mkdir -p $REMOTE_DIR/bin $REMOTE_DIR/data $REMOTE_DIR/logs"
scp -q "$STAGE/new-ops-server" "$HOST:$REMOTE_DIR/bin/new-ops-server.new"
scp -q "$STAGE/dist.tar.gz" "$HOST:$REMOTE_DIR/dist.tar.gz"

echo "[4/5] 切换前端 + 重启服务"
ssh "$HOST" "cd $REMOTE_DIR \
  && rm -rf web.new && mkdir web.new && tar xzf dist.tar.gz -C web.new --strip-components=1 \
  && rm -rf web.old && mv web web.old && mv web.new web \
  && rm -f dist.tar.gz \
  && mv bin/new-ops-server.new bin/new-ops-server && chmod +x bin/new-ops-server \
  && systemctl restart new-ops-server && sleep 3 && systemctl is-active new-ops-server \
  && docker restart new-ops-web >/dev/null && sleep 1"

echo "[5/5] 冒烟验证"
sleep 2
WEB_CODE=$(ssh "$HOST" "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8081/")
API_OK=$(ssh "$HOST" "curl -s -X POST http://127.0.0.1:8888/init/checkdb" | grep -c '"code":0' || true)
echo "web HTTP: $WEB_CODE | checkdb code:0 -> $API_OK"
if [ "$WEB_CODE" != "200" ] || [ "$API_OK" != "1" ]; then
    echo "!! 冒烟未通过，请查看: ssh $HOST 'journalctl -u new-ops-server -n 50 --no-pager'"
    exit 1
fi
echo "部署完成: http://192.168.112.138:8081"
