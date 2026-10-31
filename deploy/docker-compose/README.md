# docker-compose 本地联调说明

用途：本地一键拉起 web + server + MySQL + Redis（开发联调用；生产部署请走 `scripts/deploy-test.sh` 一键脚本，SQLite 单机更轻）。

```bash
cd deploy/docker-compose
cp .env.example .env   # 按需修改（不创建 .env 则全部使用默认值，行为与历史版本一致）
docker compose up -d
```

## 变量表（.env）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `WEB_PORT` | `8080` | 前端入口端口（nginx 容器映射） |
| `API_PORT` | `8888` | 后端 API 端口 |
| `MYSQL_PORT` | `13306` | MySQL 对宿主机暴露端口（容器内 3306） |
| `REDIS_PORT` | `16379` | Redis 对宿主机暴露端口 |
| `MYSQL_PASSWORD` | `Aa@6447985` | MySQL 应用账号密码（**生产/长期运行务必修改**，healthcheck 同步引用该变量） |
| `MYSQL_DATABASE` | `qmPlus` | 初始化数据库名（server config.yaml 需与之对应） |
| `CREDENTIAL_MASTER_KEY` | （空） | 凭据保险库主密钥，注入 server 容器环境变量 `NEW_OPS_CREDENTIAL_MASTER_KEY`；留空则凭据加密功能不可用。生成：`openssl rand -base64 32` |

## 网络与数据

- 固定子网 `177.7.0.0/16`（web .11 / server .12 / mysql .13 / redis .14），容器名互访。
- 持久化卷：`mysql` / `redis`。

## 注意

- server 容器使用镜像内默认 `config.yaml`（开发默认值：SQLite 引导需自行挂载或构建时替换；MySQL 口径需与上表一致）。
- arm64（Apple Silicon）请把 MySQL 镜像换为 `mysql/mysql-server:8.0.21`（yaml 内有注释提示）。
