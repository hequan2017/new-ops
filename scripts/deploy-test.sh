#!/usr/bin/env bash
# 白泽 BaiZe · 统一运维开发平台 —— 一键部署脚本
# 同一条命令同时支持：全新安装（自动初始化数据库）与已有环境增量更新（不动配置与数据）。
#
# 用法：
#   bash scripts/deploy-test.sh                                  # 部署/更新到默认测试机
#   NEW_OPS_TEST_HOST=root@1.2.3.4 bash scripts/deploy-test.sh   # 部署到其他 Ubuntu 机器（需 docker）
#
# 管理员密码（仅全新安装时生效）：通过环境变量 NEW_OPS_ADMIN_PASSWORD 传入；
# 不传则随机生成并写入远端 ADMIN_PASSWORD 文件（chmod 600）。密码不落仓库。
#
# 可覆盖参数（环境变量）：
#   NEW_OPS_TEST_HOST      目标机器        默认 root@192.168.112.138
#   NEW_OPS_REMOTE_DIR     远端目录        默认 /opt/new-ops
#   NEW_OPS_WEB_PORT       前端端口        默认 8081
#   NEW_OPS_API_PORT       后端端口        默认 8888
#   NEW_OPS_UNIT           systemd 服务名  默认 new-ops-server
#   NEW_OPS_CONTAINER      nginx 容器名    默认 new-ops-web
#   NEW_OPS_ADMIN_PASSWORD 管理员密码      默认随机生成（仅全新安装）
set -euo pipefail

HOST=${NEW_OPS_TEST_HOST:-root@192.168.112.138}
DIR=${NEW_OPS_REMOTE_DIR:-/opt/new-ops}
WEB_PORT=${NEW_OPS_WEB_PORT:-8081}
API_PORT=${NEW_OPS_API_PORT:-8888}
UNIT=${NEW_OPS_UNIT:-new-ops-server}
CONTAINER=${NEW_OPS_CONTAINER:-new-ops-web}
REPO=$(cd "$(dirname "$0")/.." && pwd)
STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT

echo "==> 白泽一键部署 → $HOST ($DIR, web:$WEB_PORT api:$API_PORT)"

echo "[1/7] 编译后端 (linux/amd64)"
(cd "$REPO/server" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$STAGE/new-ops-server" .)

echo "[2/7] 构建前端"
(cd "$REPO/web" && npm run build --silent >/dev/null)
tar -czf "$STAGE/dist.tar.gz" -C "$REPO/web" dist

echo "[3/7] 上传"
ssh "$HOST" "mkdir -p '$DIR/bin' '$DIR/data' '$DIR/logs'"
scp -q "$STAGE/new-ops-server" "$HOST:$DIR/bin/new-ops-server.new"
scp -q "$STAGE/dist.tar.gz" "$HOST:$DIR/"
# 配置模板随包上传，仅全新安装时启用
scp -q "$REPO/server/config.yaml" "$HOST:$DIR/config.yaml.new"
# 标记本机此前是否已初始化（决定是否走数据库初始化）
ssh "$HOST" "if [ -f '$DIR/config.yaml' ]; then echo update > '$DIR/.deploy_mode'; else echo fresh > '$DIR/.deploy_mode'; fi"

echo "[4/7] 远端安装（幂等）"
ssh "$HOST" DEPLOY_DIR="$DIR" API_PORT="$API_PORT" WEB_PORT="$WEB_PORT" UNIT="$UNIT" CONTAINER="$CONTAINER" bash -s <<'REMOTE'
set -e
cd "$DEPLOY_DIR"
MODE=$(cat .deploy_mode)

# 前端原子切换
rm -rf web.new && mkdir web.new && tar xzf dist.tar.gz -C web.new --strip-components=1
rm -rf web.old && [ -d web ] && mv web web.old
mv web.new web
rm -f dist.tar.gz

# 后端二进制
mv bin/new-ops-server.new bin/new-ops-server && chmod +x bin/new-ops-server

# 全新安装：启用 sqlite 配置 + 关闭验证码（测试环境便利项）
if [ "$MODE" = "fresh" ]; then
  mv config.yaml.new config.yaml
  sed -i 's/db-type: mysql/db-type: sqlite/' config.yaml
  sed -i "s|path: \"\"|path: $DEPLOY_DIR/data|" config.yaml
  sed -i "s/addr: 8888/addr: $API_PORT/" config.yaml
  sed -i 's/open-captcha: 0/open-captcha: 999999/' config.yaml
else
  rm -f config.yaml.new
fi

# systemd 服务
cat > /etc/systemd/system/$UNIT.service <<EOF
[Unit]
Description=BaiZe new-ops server (gin-vue-admin based)
After=network.target

[Service]
Type=simple
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/bin/new-ops-server
Restart=always
RestartSec=3
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now $UNIT >/dev/null 2>&1 || true
systemctl restart $UNIT

# nginx 静态 + API 反代（host 网络）
cat > $DEPLOY_DIR/nginx-default.conf <<EOF
server {
    listen $WEB_PORT;
    server_name _;
    client_max_body_size 1024m;
    location /api/ {
        proxy_pass http://127.0.0.1:$API_PORT/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
    location / {
        root /usr/share/nginx/html;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
if docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  docker rm -f $CONTAINER >/dev/null
fi
docker run -d --name $CONTAINER --network host --restart unless-stopped \
  -v $DEPLOY_DIR/web:/usr/share/nginx/html:ro \
  -v $DEPLOY_DIR/nginx-default.conf:/etc/nginx/conf.d/default.conf:ro \
  nginx:alpine >/dev/null
echo "$MODE" > .deploy_mode.done
REMOTE

MODE=$(ssh "$HOST" "cat '$DIR/.deploy_mode.done'" 2>/dev/null || echo update)

echo "[5/7] 数据库初始化（仅全新安装）"
PW=""
if [ "$MODE" = "fresh" ]; then
  if [ -n "${NEW_OPS_ADMIN_PASSWORD:-}" ]; then
    PW="$NEW_OPS_ADMIN_PASSWORD"
  else
    PW=$(head -c 16 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 12)
  fi
  sleep 2
  ssh "$HOST" "curl -s -X POST http://127.0.0.1:$API_PORT/init/initdb -H 'Content-Type: application/json' \
    -d '{\"dbType\":\"sqlite\",\"dbName\":\"new_ops\",\"dbPath\":\"$DIR/data\",\"adminPassword\":\"$PW\"}'"
  echo
  ssh "$HOST" "umask 077 && echo '$PW' > '$DIR/ADMIN_PASSWORD'"
  echo "    管理员密码已写入远端 $DIR/ADMIN_PASSWORD（chmod 600）"
else
  [ -n "${NEW_OPS_ADMIN_PASSWORD:-}" ] && PW="$NEW_OPS_ADMIN_PASSWORD" || true
fi

echo "[6/7] 冒烟验证"
sleep 2
WEB_CODE=$(ssh "$HOST" "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:$WEB_PORT/")
API_OK=$(ssh "$HOST" "curl -s -X POST http://127.0.0.1:$API_PORT/init/checkdb" | grep -c '"code":0' || true)
LOGIN_OK="skipped"
if [ -n "$PW" ]; then
  LOGIN_OK=$(ssh "$HOST" "curl -s -X POST http://127.0.0.1:$API_PORT/base/login -H 'Content-Type: application/json' \
    -d '{\"username\":\"admin\",\"password\":\"$PW\",\"captcha\":\"\",\"captchaId\":\"\"}'" | grep -c '"code":0' || true)
fi
echo "    web HTTP: $WEB_CODE | checkdb: $API_OK | login: $LOGIN_OK"
if [ "$WEB_CODE" != "200" ] || [ "$API_OK" != "1" ]; then
  echo "!! 冒烟未通过，请查看: ssh $HOST 'journalctl -u $UNIT -n 50 --no-pager'"
  exit 1
fi

echo "[7/7] 完成 ✅"
echo "    访问入口: http://$(echo "$HOST" | cut -d@ -f2):$WEB_PORT"
echo "    管理员: admin"
[ -n "$PW" ] && echo "    密码: $PW" || echo "    密码: 沿用既有密码（全新安装时由脚本生成并写入远端 ADMIN_PASSWORD）"
