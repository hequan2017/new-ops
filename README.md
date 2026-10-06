# 白泽 BaiZe · 统一运维开发平台（new-ops）

> **v0.2.0+**（2026-10-06）—— M0-M3 全部完成，M4 容器管理主体（C1 全部 + C2 镜像/网络/卷/资源统计），M6 监控告警三项 + 数据库工单可做部分（实例纳管/SQL 工单），M7 工单引擎 + **MCP 运维工具**（AI 助手可直查资产/告警/工单）；全部功能经真实环境端到端验证。接口仍可能调整，生产部署前请完成安全复核（JWT 密钥、验证码、主密钥轮换）。
>
> **命名由来**：白泽是中国上古神话中的瑞兽，通晓天下万物之情——愿这套平台也能"通晓"你的全部基础设施。仓库/工程名沿用 `new-ops`。
>
> 开发总计划（功能盘点 / 架构设计 / 里程碑任务清单 / 逐场排期 / 开发日志）：[docs/DEV_PLAN.md](docs/DEV_PLAN.md) ｜ 插件开发规范：[docs/plugin-dev-guide.md](docs/plugin-dev-guide.md)

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)（main, e8d675c）构建的**插件化统一运维开发平台**：把作者历史开源的 autoops、chain、cmdb、seal、go-webssh、raptor、new-jenkins、GPU 算力系列等 20+ 个运维项目的功能，收敛到一个底座上持续演进，避免多套技术栈、多套权限体系的重复维护。

**核心设计原则**：业务代码只写在 `server/plugin/<name>` 与 `web/src/plugin/<name>`，底座零修改——可持续跟随上游升级，历史项目功能按域拆分、择优重写为插件。

---

## 目录

- [技术栈与架构](#技术栈与架构)
- [功能总览（按插件）](#功能总览按插件)
- [快速开始](#快速开始)
- [一键部署到任意 Ubuntu 机器](#一键部署到任意-ubuntu-机器)
- [安全设计](#安全设计)
- [接口与错误码规范](#接口与错误码规范)
- [仓库结构](#仓库结构)
- [插件开发](#插件开发)
- [里程碑进度与路线图](#里程碑进度与路线图)
- [License](#license)

---

## 技术栈与架构

| 层 | 技术 |
|---|---|
| 后端 | Go 1.24 · Gin · GORM · Casbin v3 · JWT · Zap · Viper |
| 前端 | Vue 3.5 · Vite 8 · Element Plus · Pinia · ECharts 5 · xterm.js 6 |
| 实时通道 | WebSocket（WebSSH / SFTP / 容器终端与日志）· SSE（流水线日志流） |
| 执行通道 | Go SSH（`golang.org/x/crypto/ssh`，密码/私钥/keyboard-interactive）· Docker SDK v28 · client-go |
| 数据 | MySQL / SQLite（零依赖部署）· Redis（可选） |
| CI | GitHub Actions（server: build/vet/test；web: build） |

```
┌───────────────────────── web (Vue3.5 + Vite8 + Element Plus) ─────────────────────────┐
│  GVA 基座(superAdmin/system/systemTools) + 业务插件视图                                │
│  asset 资产/凭据 │ term 终端/SFTP/审计 │ job 批量作业 │ pipeline 流水线                │
│  container 容器 │ k8s 集群 │ monitor 监控告警 │ workflow 工单                          │
└──────────▲──────────────────────────────────────────────▲───────────────────────────┘
           │ REST(JWT+Casbin)                WebSocket / SSE
┌──────────┴──────────────────────────────────────────────┴───────────────────────────┐
│  server : Gin + GORM + Casbin + Zap + 定时任务(robfig/cron) + MCP Server 骨架         │
│  plugin/asset  plugin/term  plugin/job  plugin/pipeline  plugin/container             │
│  plugin/k8s    plugin/monitor  plugin/workflow  plugin/dbops   （gpu/org 骨架就绪）    │
├──────────────────────────────────────────────────────────────────────────────────────┤
│  执行通道：SSH 直连/ProxyJump 级联 │ Docker API(unix/TCP+TLS) │ K8s API(kubeconfig)   │
│  外部依赖：阿里云 OpenAPI(RPC V1 自实现签名) │ 钉钉机器人 webhook │ goInception(规划)   │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 功能总览（按插件）

### asset 资产中心（M1 ✅ 全部完成）

| 功能 | 说明 |
|---|---|
| 主机资产 CRUD | 主机名/IP/系统/CPU/内存/磁盘/SN/厂商/机房/机柜/U 位/产品线多对多/负责人/状态，关键字+状态+机房+产品线过滤分页 |
| 凭据保险库 | SSH 密码/私钥、云 AccessKey、Docker TLS（JSON{ca,cert,key}）、kubeconfig；**AES-256-GCM 信封加密**，主密钥环境变量注入，接口只回显末 4 位与引用计数 |
| SSH 现场采集 | Go SSH 双认证拨号（密码/私钥/keyboard-interactive），命令级超时，采集 hostname/os/cpu/mem/disk 回填并记变更历史 |
| 阿里云 ECS 同步 | OpenAPI RPC V1 签名**纯标准库自实现**（HMAC-SHA1），DescribeInstances 分页，按 SN upsert，空 IP 兜底公网 |
| Excel 导入导出 | excelize 导出 xlsx / 按 IP upsert 导入 + 逐行校验失败明细 |
| 变更历史 | 创建/更新/删除事务内记录，时间线抽屉回看 |
| 数据权限 | 资产组（组-主机-用户三表）+ casbin：888 全量、普通用户仅组内、9528 只读 |
| 网段发现 | CIDR 展开（4096 上限、跳过网络/广播地址）→ 并发 SSH banner 探测 → 勾选按 IP 导入（存在跳过） |

### term 终端与文件（M2 ✅ 全部完成）

| 功能 | 说明 |
|---|---|
| WebSSH | xterm.js + WebSocket，query token 握手自验，二进制数据帧 + JSON 控制帧（resize/ping/close），窗口自适应、30s 心跳 |
| ProxyJump 级联 | 主机 `JumpHostID` 自引用链，≤5 跳、禁环；最外层 TCP 直连，内层逐跳 SSH 隧道（client.Dial + NewClientConn） |
| 主机指纹校验 | **TOFU**（首连/采集自动录入 SHA256 指纹），已录指纹强校验、不匹配拒绝连接；CRUD 层防请求伪造篡改 |
| 会话审计 | 三表：会话元数据 / 全量流镜像（按 seq）/ 命令抽取（ANSI 剥离+退格修正）；回放页 xterm 重演、上行输入标注；审计开档失败拒绝开终端 |
| SFTP 文件浏览器 | 列表/上传/下载/删除（目录递归）/重命名/建目录 |
| 远程日志 tail | `/term/logtail` WS，复用级联拨号，远端 `tail -n N -F --`（`--` 防选项注入、绝对路径校验），前端滚动跟随/上滚暂停 |

### job 批量作业（M2 ✅ 全部完成）

| 功能 | 说明 |
|---|---|
| 批量命令执行 | 信号量并发池（限流默认 10 上限 50、单任务超时 5-600s、批次即时取消）；数据权限校验（非 888 仅授权组内主机）；逐主机结果留存 |
| 脚本库 | shell/python/yaml 三类，更新版本自动递增归档，历史版本抽屉 |
| 变量组 | key=value 编辑器，`{{key}}` 占位**服务端渲染**入批次快照；支持按主机所在资产组自动注入变量 |
| 执行页 | 主机多选（数据权限过滤）、统一凭据可选、3s 静默轮询、批次取消（二次确认）、结果抽屉 |

### pipeline 流水线（M3 ✅ 全部完成）

| 功能 | 说明 |
|---|---|
| 三层模型 | Pipeline → Stage → Step；阶段可配审批 gate / 失败继续 / **步骤并发**（信号量上限 10、首败快速中断兄弟步骤） |
| 步骤类型 | shell（**经 SSH 在绑定资产上执行**，无本机进程执行面，逐跳指纹校验）/ http（4xx/5xx 即失败） |
| 触发快照 | 触发时对定义做快照（参数 `{{key}}` 已渲染），构建期间改定义不影响进行中构建 |
| 状态机 | 等待中 → 执行中 →（等待审批 → approveCh 放行）→ 成功/失败/已取消；构建序号按流水线递增 |
| 触发器 | 手动 / **webhook**（`/pipeline/webhook/:token`，令牌即凭据 uuid、错误令牌 404 语义防枚举）/ **cron**（接入底座 GVA_Timer，robfig 5 段 + @every，增删改即时生效、启动恢复） |
| 日志 | 落库分页拉取 + **SSE 实时流**（lastId 增量、status/done 事件收流、断线续传、X-Accel-Buffering no） |
| 构建历史 | 即时取消（ctx 中断 SSH/HTTP）、复用历史参数重跑 |
| 编排页 | 阶段卡片配置、审批/并发/失败继续开关、shell/http 切换、参数触发、轮询、放行/取消/重跑 |

### container 容器管理（M4 C1 ✅ + C2 主体 ✅）

| 功能 | 说明 |
|---|---|
| 接入点纳管 | unix socket 直连 / TCP+TLS（证书组合走凭据保险库 `docker_tls`）；30s 合并巡检（Ping+版本回写，并发 5） |
| 容器生命周期 | 实时列表（不落库）、启停/重启/删除（force）、离线接入点拒绝写操作 |
| 创建容器 | 端口映射（`host:ct[/proto]` 纯函数解析）、环境变量、挂载 Binds、CPU 核数/内存限额（NanoCPUs/Memory）、重启策略四档、创建即启 |
| 日志流 / exec 终端 | 双 WS 通道：logs -f（TTY 判断 + stdcopy 去复用）；exec `/bin/sh` TTY → hijack 双向桥、resize |
| 事件订阅 | 复用 30s 巡检按 `[水位, now]` 区间拉取 docker events 落库（离线恢复自动补拉、窗口上限 1h、7 天留存） |
| 镜像管理 | 列表/异步拉取（内存状态表+轮询、重复拦截）/删除 force/**打标签/导出 tar 流式下载/导入上传** |
| 网络与卷 | network 创建（bridge+子网 IPAM、内置保护）/列表/删除；volume 列表/删除（占用由 daemon 拒绝回传） |
| 资源统计 | ContainerStats one-shot：CPU（delta×在线核数）/内存（去 inactive_file）/网络/块 IO/进程数 |

### k8s 集群管理（M4 K1/K2 🚧 进行中）

- ✅ 集群注册：kubeconfig **AES-256-GCM 加密落库**（接口不回显）、连接测试（失败不阻断、状态记离线）
- ✅ 资源浏览：Pods / Deployments / Nodes / Services / ConfigMaps / Secrets 只读列表 + Pod 日志（Secret 值永不回显，仅列键名）
- 🚧 写操作（扩缩容/滚动重启）与真集群联调进行中；其余 K2 能力见开发日志

### monitor 监控告警（M6 ✅ 三项全部）

| 功能 | 说明 |
|---|---|
| 性能采集 | SSH 单命令组合采样（1s：loadavg / cpu 两次 /proc/stat delta / meminfo / df），解析纯函数单测；@every 5m 定时，并发 5 |
| 指标留存 | monitor_metric（asset+name+ts 复合索引），30 天自动清理；ECharts 四折线双 Y 轴（cpu/mem/disk % + load1）、1h/6h/24h/7d 切换、立即采集 |
| 告警引擎 | metric 阈值（>/< 连续 N 次）+ port TCP 探活；静默窗口去重（默认 30min）、恢复自动关闭未决事件；**钉钉机器人文本推送**（仅 http(s)）；规则/事件管理页 |

### workflow 工单引擎（M7 ✅ 核心交付）

| 功能 | 说明 |
|---|---|
| 状态机 | Workflow → State → Transition（JSON 灵活建模）；定义校验纯函数（引用闭合/重名/起始状态） |
| 流转 | 审批节点动作受限（approve/reject）、终态自动收档（完成/驳回）、cancel 撤回、迁移表外动作拦截；流转记录时间线 |
| **发版闭环** | `bizType=release` 工单审批完成自动触发流水线构建（params `{pipelineId, params}`），结果落 hook 流转行；驳回不触发 |
| 工单中心 | 发起/通过/驳回/撤回 + 定义管理（JSON 模板）；普通用户仅见自己发起的工单 |

### dbops 数据库工单（M6 🔨 可做部分已交付）

| 功能 | 说明 |
|---|---|
| 实例纳管 | MySQL 实例 CRUD：主机/端口/账号/**密码 AES-256-GCM 密文**（复用 asset/crypto 信封加密；独立请求体保证响应永不回显）；TCP 探活回写在线状态 |
| SQL 上线工单 | 创建（待审核）/取消/分页（普通用户仅本人）；有未结束工单的实例拒绝删除 |
| 审核引擎预留 | goInception 接入点已留——引擎未配置时审核接口明确报错（错误码 1708）且工单状态不变；审核/执行/备份与 soar 优化建议待 goInception + MySQL 环境 |

### aiops MCP 运维工具（M7 🔨 可做部分已交付）

| 功能 | 说明 |
|---|---|
| MCP Server | 复用 GVA 骨架（mark3labs/mcp-go，StreamableHTTP），独立进程 `go run ./cmd/mcp`（默认 :8889，`/mcp` 端点，`/health` 探活） |
| 运维只读工具 | `ops_asset_overview`（资产概览）/ `ops_alert_recent`（最近告警）/ `ops_ticket_status`（工单状态）——AI 助手经 MCP 协议即可直查平台三域概况 |
| 鉴权与数据 | 工具经骨架上游代理调用主 server API，调用者身份经 auth_header 透传（沿用平台 JWT + Casbin 权限）；standalone 进程无 DB 连接，不绕过权限体系 |
| 启动 | `cd server && go run ./cmd/mcp -config ./cmd/mcp/config.yaml`；上游地址/鉴权头/超时在 mcp 段配置 |
| AI 诊断网关 | 统一 LLM 调用/密钥管理/prompt 模板——待 LLM API 密钥后接入 |

### 基座能力（gin-vue-admin 自带 ✅）

用户/角色/菜单/API 管理（Casbin RBAC + JWT）、代码生成器、表单设计器、定时任务、操作日志、多云对象存储、Swagger、MCP Server 骨架。

---

## 快速开始

### 方式一：一键部署（推荐，Ubuntu + Docker）

```bash
git clone https://github.com/hequan2017/new-ops.git
cd new-ops
bash scripts/deploy-test.sh
```

一条命令同时支持**全新安装**与**增量更新**：自动交叉编译 → 上传 → systemd + nginx:alpine 容器安装 → SQLite 初始化（无 MySQL/Redis 依赖）→ 冒烟验证。全新安装的管理员密码随机生成并写入远端 `ADMIN_PASSWORD`（chmod 600），**不落仓库**。

### 一键部署到任意 Ubuntu 机器

```bash
NEW_OPS_TEST_HOST=root@1.2.3.4 \
NEW_OPS_WEB_PORT=8081 NEW_OPS_API_PORT=8888 \
bash scripts/deploy-test.sh
```

| 默认项 | 值 |
|---|---|
| 前端入口 | `http://<host>:8081`（nginx 容器 `new-ops-web`，host 网络） |
| 后端 | `:8888`（systemd 服务 `new-ops-server`） |
| 远端目录 | `/opt/new-ops/{bin,web,data,logs,resource,config.yaml}` |
| 数据库 | SQLite `/opt/new-ops/data/new_ops.db` |
| 凭据主密钥 | 环境变量 `NEW_OPS_CREDENTIAL_MASTER_KEY` 注入 unit（部署时继承，更新不丢） |

### 方式二：源码运行（开发）

```bash
# 后端（默认 :8888，SQLite 引导）
cd server && go mod tidy && go run main.go

# 前端（默认 :8080，代理 /api → :8888）
cd web && npm install && npm run serve
```

首次启动走 GVA 数据库引导页；凭据主密钥务必通过环境变量注入，勿写配置文件。

### 上手路径

1. **凭据保险库**（资产管理 → 凭据）创建 SSH 密码/私钥凭据；
2. **主机管理**新增主机并绑定凭据（或用「网段发现」批量导入），点「采集」回填硬件信息；
3. 「终端」直连或在编辑里配置**跳板机**体验级联；「日志」做远程 tail；
4. **批量执行**选主机跑命令，配**脚本库 + 变量组**做参数化；
5. **流水线**编排阶段（shell 步骤选目标主机 / http 步骤），试 webhook/cron 触发与审批 gate；
6. **工单中心**建定义发起 release 工单，审批通过看流水线自动构建；
7. **容器管理**接入点（`unix:///var/run/docker.sock`）→ 容器/镜像/网络/卷/统计/事件/终端；
8. **告警规则**配阈值或端口探活 + 钉钉 webhook，主机页「监控」看趋势图；
9. **数据库工单**注册 MySQL 实例（密码加密落库、TCP 探活），提 SQL 工单（审核引擎待 goInception 环境）；
10. **MCP 运维工具**：`cd server && go run ./cmd/mcp` 后，把 `http://127.0.0.1:8889/mcp` 接入你的 AI 助手，即可对话式查询资产概览 / 最近告警 / 工单状态。

---

## 安全设计

- **凭据一律 AES-256-GCM 加密落库**（信封加密，主密钥仅环境变量注入），任何接口不回显明文（只回末 4 位）；引用方按 ID 取用、内存解密即用即毁
- **SSH 主机公钥指纹校验**：TOFU 首连录入（SHA256），已录指纹不匹配即拒绝；CRUD 层 `ssh_fp` 防请求伪造篡改（仅 SSH 链路可写）
- **终端全程审计**：全量流镜像 + 命令抽取落库，回放可重演；审计开档失败拒绝开终端（红线不裁剪）
- **写操作全走操作日志中间件**；危险操作（容器/镜像/卷删除、工单撤回、批次取消）前端二次确认
- **数据权限**：资产组 + casbin 资源规则——888 全量、普通用户仅授权组内（资产列表/批量执行/终端主机选择全链路过滤）、9528 只读
- **SSRF 防护**：http 步骤/钉钉 webhook 仅允许 http(s)；webhook 触发令牌即凭据（uuid、错误令牌 404 语义防枚举）
- **注入防护**：日志 tail `tail -F --` 终止选项解析 + 绝对路径校验；shell 步骤在远端资产执行，平台进程无本机 exec 面

---

## 接口与错误码规范

- REST 资源 `/api/<插件>/<资源>`，动作类 POST 子资源；WebSocket `/ws/<插件>/*` 或 public 组端点（query token 握手自验），SSE `/sse/<插件>/*`
- 响应统一 `{code, data, msg}`，列表分页 `pageInfo`；过滤参数 `keyword`/`status`/时间区间
- **错误码分段**（每插件独占）：asset=1000-1099 · term=1100-1199 · job=1200-1299（脚本 1201-1207 / 执行 1211-1214）· pipeline=1300-1399 · container=1400-1499 · k8s=1500-1599 · gpu=1600-1699 · dbops=1700-1799 · monitor=1800-1899 · workflow=1900-1999 · agent/org/aiops=2000-2199；通用错误沿用 GVA 7xxx 段
- Swagger 注释随代码同步（`/swagger/index.html`）

---

## 仓库结构

```
new-ops/
├── server/                        # Go 后端（module github.com/hequan2017/new-ops/server）
│   ├── plugin/
│   │   ├── asset/                 # 资产中心 + 凭据保险库 + 网段发现（M1 ✅）
│   │   ├── term/                  # WebSSH/级联/审计/SFTP/日志tail（M2 ✅）
│   │   ├── job/                   # 批量执行/脚本库/变量组（M2 ✅）
│   │   ├── pipeline/              # 流水线/执行器/SSE/触发器（M3 ✅）
│   │   ├── container/             # Docker 接入点/容器/镜像/网络卷/事件/stats（M4 ✅）
│   │   ├── k8s/                   # 集群注册 + 资源浏览（M4 🚧）
│   │   ├── monitor/               # 性能采集/告警引擎/钉钉（M6 ✅）
│   │   ├── workflow/              # 工单引擎 + 发版闭环（M7 ✅）
│   │   ├── dbops/                 # MySQL 纳管/SQL 工单（M6 🔨 审核引擎待环境）
│   │   ├── mcp/(骨架) + cmd/mcp   # MCP Server 独立进程 + aiops 运维工具（M7 🔨）
│   │   └── gpu/org/               # 骨架就绪（依赖环境/配置）
│   └── ...                        # GVA 底座（零修改）
├── web/src/plugin/                # 各插件前端视图（与 server/plugin 同名对应）
├── docs/
│   ├── DEV_PLAN.md                # 开发总计划 + 逐场开发日志（进度权威来源）
│   └── plugin-dev-guide.md        # 插件开发规范
├── scripts/deploy-test.sh         # 一键部署/增量更新 + 冒烟
└── .github/workflows/             # CI（server go build/vet/test + web build）
```

### 插件开发

单插件内部结构（新插件照此落位，详见 [plugin-dev-guide](docs/plugin-dev-guide.md)）：

```
server/plugin/<name>/
├── plugin.go        # 注册入口（Gorm→Api→Menu→Casbin→Router→Timer）
├── initialize/      # 路由 + 菜单/API/casbin 种子 + 表迁移 + 定时任务
├── api/v1/          # 接口层（参数绑定与响应，不含业务）
├── router/          # 路由分组（private= casbin；public= query token 自验）
├── service/         # 业务逻辑（纯函数优先，单测写这里）
└── model/           # model / request / response
```

前端对应 `web/src/plugin/<name>/{view,api}`，走数据库动态菜单路由。

---

## 里程碑进度与路线图

| 里程碑 | 内容 | 状态 |
|---|---|---|
| M0 | 底座就绪：品牌化、example 移除、compose/Makefile、CI、插件骨架、开发规范 | ✅ |
| M1 | 资产中心：CRUD/SSH 采集/云同步/凭据保险库/数据权限/导入导出 | ✅ |
| M2 | 终端作业：WebSSH/级联/审计/SFTP/批量执行/脚本库变量组/日志tail/网段发现 | ✅ |
| M3 | 流水线：三层模型/执行器（快照+状态机+审批+并发）/SSE/三通道触发/发版闭环 | ✅ |
| M4 | 容器 C1 全部 + C2 镜像/网络/卷/stats；k8s K1/K2 进行中 | 🔨 |
| M5 | GPU 算力：节点/规格/实例/防超卖/HAMi | ⏳ 待显卡环境 |
| M6 | monitor 三项 ✅；dbops 实例纳管 + SQL 工单骨架 ✅（goInception 审核/执行待环境） | 🔨 |
| M7 | workflow 工单引擎 ✅；aiops MCP 运维工具 ✅（AI 诊断网关待 LLM 密钥）；org 钉钉登录 | 🔨 |
| M8 | Compose/端口转发、Helm、三级 RBAC、AI 诊断 | 📋 |
| M9 | 轻量 Go Agent：反向长连接/采集上报/第二执行通道 | 📋 |

- 发布：`v0.1.0`（M0-M2）、`v0.2.0`（+M3/M4 主体/M6 monitor/M7 workflow）
- 逐场进展与排期见 [DEV_PLAN 开发日志](docs/DEV_PLAN.md)；旧仓库功能并入后将逐一归档标注"功能已并入 new-ops"

---

## License

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)（Apache License 2.0）构建，本项目沿用该协议。
