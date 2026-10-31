# 白泽 BaiZe · 统一运维开发平台（new-ops）

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**白泽（BaiZe）** 是一个插件化的统一运维开发平台：把 CMDB 资产、WebSSH 终端、批量作业、发布流水线、Docker/K8s 容器管理、监控告警、工单、数据库审核、AI 运维等 20+ 个运维域的能力收敛到**一个底座、一套权限、一份数据**上。

> 命名由来：白泽是中国上古神话中的瑞兽，通晓天下万物之情——愿这套平台也能"通晓"你的全部基础设施。
>
> 基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin) 构建（Apache License 2.0），**业务代码全部以插件形式存在，底座零修改**，可持续跟随上游升级。当前版本 **v0.3.0-rc1**。
>
> 插件开发规范：[docs/plugin-dev-guide.md](docs/plugin-dev-guide.md)

---

## 目录

- [核心能力](#核心能力)
- [功能总览（按插件）](#功能总览按插件)
- [技术栈与架构](#技术栈与架构)
- [权限与数据隔离（三级）](#权限与数据隔离三级)
- [快速开始](#快速开始)
- [安全设计](#安全设计)
- [接口与错误码规范](#接口与错误码规范)
- [仓库结构](#仓库结构)
- [插件开发](#插件开发)
- [路线图](#路线图)
- [License](#license)

---

## 核心能力

| 能力域 | 一句话说明 |
|---|---|
| **资产中心** | 主机/凭据/机房/产品线统一台账，SSH 采集、阿里云同步、网段发现、Excel 导入导出 |
| **终端与文件** | 浏览器直达任意深度内网主机：WebSSH（N 跳级联）、SFTP、远程日志 tail，全程会话审计可回放 |
| **批量作业** | 成百台主机命令/脚本作业，限流防雪崩，SSH/Agent 双执行后端，脚本库 + 变量组参数化 |
| **流水线** | 声明式 Pipeline→Stage→Step，审批 gate、步骤并发、手动/webhook/cron 三通道触发、SSE 实时日志 |
| **容器管理** | Docker 单机/远程全生命周期：容器/镜像/网络/卷/事件/资源统计、Compose 编排、应用模板一键部署 |
| **K8s 管理** | 多集群注册、十一类资源浏览、Pod WebShell、YAML dry-run 下发、Helm 全生命周期、命名空间级授权 |
| **Agent 通道** | 单文件 Go Agent 反向接入：任务执行、文件分发、Docker/K8s 内网代理（免暴露内网 API） |
| **监控告警** | SSH 采集性能指标（CPU/内存/磁盘/负载/网络）、阈值与端口探活告警、钉钉推送 |
| **工单引擎** | 状态机建模任意审批流，release 工单审批通过自动触发流水线构建 |
| **数据库工单** | MySQL 实例纳管 + goInception SQL 审核/执行/备份回滚 |
| **AI 运维** | LLM 网关（OpenAI 兼容）、Pod AI 诊断、MCP Server 接入 AI 助手对话式查平台 |
| **GPU 算力** | 节点/规格/实例台账 + 防超卖引擎（Docker 直通/HAMi 待显卡环境） |

---

## 功能总览（按插件）

### asset 资产中心

全平台的数据心脏：主机、凭据、机房机柜、产品线等基础数据只存一份，term/job/pipeline/container/k8s/monitor 全部按外键引用。

| 功能 | 说明 |
|---|---|
| 主机资产 CRUD | 主机名/IP/系统/硬件/SN/机房/机柜/U 位/产品线多对多/负责人/状态，多条件过滤分页 |
| 凭据保险库 | SSH 密码/私钥、云 AccessKey、Docker TLS、kubeconfig 五类凭据统一纳管；**AES-256-GCM 信封加密**，主密钥仅环境变量注入，接口只回显末 4 位 |
| SSH 现场采集 | 双认证拨号（密码/私钥/keyboard-interactive），采集硬件信息回填并记变更历史 |
| 阿里云 ECS 同步 | OpenAPI RPC V1 签名纯标准库自实现（不引官方 SDK），按 SN upsert |
| Excel 导入导出 | xlsx 导出 / 按 IP upsert 导入 + 逐行校验失败明细 |
| 变更历史 | 增删改事务内记录，时间线抽屉回看 |
| 网段发现 | CIDR 展开 → 并发 SSH banner 探测 → 勾选导入 |

### term 终端与文件

| 功能 | 说明 |
|---|---|
| WebSSH | xterm.js + WebSocket，query token 握手自验，二进制数据帧 + JSON 控制帧，窗口自适应、心跳 |
| ProxyJump 级联 | 跳板机自引用链，≤5 跳、禁环，逐跳 SSH 隧道——免 Agent 打通 N 层跳板 |
| 主机指纹校验 | TOFU 首连自动录入 SHA256 指纹，已录指纹强校验、不匹配拒连 |
| 会话审计 | 会话元数据 / 全量流镜像 / 命令抽取三表落库，回放页 xterm 重演；审计开档失败拒绝开终端 |
| SFTP 文件浏览器 | 列表/上传/下载/删除/重命名/建目录 |
| 远程日志 tail | 复用级联拨号，`tail -n N -F --` 终止选项解析防注入，前端滚动跟随 |

### job 批量作业

| 功能 | 说明 |
|---|---|
| 批量命令执行 | 信号量并发池（限流默认 10 上限 50、单任务超时 5-600s、批次即时取消），数据权限校验，逐主机结果留存 |
| 双执行后端 | SSH 直连 或 **Agent 通道**下发（内网主机免开 SSH 同样可执行） |
| 脚本库 | shell/python/yaml 三类，更新版本自动递增归档 |
| 变量组 | `{{key}}` 占位服务端渲染入批次快照；按主机所在资产组自动注入变量 |
| 文件分发 | 本地文件经 Agent 通道分块下发（SHA256 校验、按声明大小/权限原子落盘） |

### pipeline 流水线

| 功能 | 说明 |
|---|---|
| 三层模型 | Pipeline → Stage → Step；阶段可配审批 gate / 失败继续 / 步骤并发（首败快速中断） |
| 步骤类型 | shell（**经 SSH 或 Agent 在绑定资产上执行**，无本机进程执行面）/ http（4xx/5xx 即失败） |
| 触发快照 | 触发时对定义快照（参数已渲染），构建期间改定义不影响进行中构建 |
| 状态机 | 等待中 → 执行中 →（等待审批 → 放行）→ 成功/失败/已取消；构建序号按流水线递增 |
| 触发器 | 手动 / webhook（令牌即凭据 uuid，错误令牌 404 语义防枚举）/ cron（robfig 5 段 + @every，增删改即时生效） |
| 日志 | 落库分页拉取 + SSE 实时流（lastId 增量、断线续传） |

### container 容器管理

| 功能 | 说明 |
|---|---|
| 接入点纳管 | unix socket 直连 / TCP+TLS / **Agent 代理**（内网 daemon 经 agent WS 通道转发）；30s 合并巡检 |
| 容器生命周期 | 实时列表（不落库）、启停/重启/删除、创建（端口映射/环境变量/挂载/CPU 内存限额/重启策略） |
| 日志流 / exec 终端 | 双 WS 通道：logs -f（stdcopy 去复用）；exec `/bin/sh` TTY 双向桥 |
| 镜像管理 | 列表/异步拉取/删除/打标签/tar 导出导入 |
| 网络与卷 | network 创建（bridge+子网 IPAM、内置保护）/列表/删除；volume 列表/删除；容器端口转发规则管理 |
| 资源统计 | CPU/内存/网络/块 IO one-shot + 5 分钟采样历史（7 天）+ ECharts 趋势图 |
| 事件订阅 | 按水位区间拉取 docker events 落库，离线恢复自动补拉 |
| Compose 编排 | 项目制管理（compose 文件留存）：创建（客户端侧校验）/up/down/restart/更新/删除 |
| 应用模板 | compose 模板库 + `{{参数}}` 渲染一键部署 |

### k8s 集群管理

| 功能 | 说明 |
|---|---|
| 集群注册 | kubeconfig AES-256-GCM 加密落库（不回显）；连接测试、离线状态标记 |
| 集群总览 | 版本/节点统计/命名空间/Pod 数 + metrics-server 资源用量（不可用时降级） |
| Node 管理 | 列表/详情、cordon/uncordon、drain（policy/v1 Eviction，自动跳过 DaemonSet） |
| 资源浏览 | Pods / Deployments / StatefulSets / DaemonSets / Nodes / Services / ConfigMaps / Secrets / PVC / Ingress / Events 十一类只读列表（Secret 值永不回显） |
| Pod 详情/日志/终端 | 详情含容器状态与相关事件、日志 tail、WebShell（client-go SPDY ⇄ WebSocket，query token + casbin 双验） |
| 工作负载写操作 | Deployment 扩缩容（0-500 校验）与滚动重启 |
| YAML 下发 | 三类工作负载 YAML 查看（剔 managedFields）→ 编辑 → **服务端 dry-run diff 红绿预览** → 确认下发（乐观锁 resourceVersion、防跨对象写） |
| Helm 全生命周期 | 列表（全命名空间）/安装升级/卸载/回滚/历史/详情（values+渲染 manifest）；安装支持 tgz 上传或 chart 仓库引用；**values 表单模式**（values.schema.json 驱动动态表单，无 schema 自动回退 YAML 模式） |
| 命名空间授权 | `k8s_ns_grant` 授权表；列表 scoped 包装 + 单对象 GuardNamespace 守卫（越权拒绝） |

### gpu 算力平台

| 功能 | 说明 |
|---|---|
| 节点/规格/实例 | 三表模型：算力节点（资产关联）、产品规格（GPU 数/显存/定价）、实例分配台账 |
| 防超卖引擎 | 事务内 `FOR UPDATE` 行锁 + 运行中占用聚合 + 余量判定（不足精确报错带明细）；销毁释放配额复用 |
| 待环境 | GPU Docker 直通（DeviceRequest）、HAMi 显存切分、实例监控——待真实显卡环境 |

### monitor 监控告警

| 功能 | 说明 |
|---|---|
| 性能采集 | SSH 单命令组合采样：loadavg/CPU delta/meminfo/df/网络双采样速率，@every 5m 定时 |
| 指标留存 | 30 天自动清理；ECharts 折线双 Y 轴（1h/6h/24h/7d 切换） |
| 告警引擎 | metric 阈值（>/< 连续 N 次）+ port TCP 探活；静默窗口去重、恢复自动关闭；钉钉机器人推送 |

### workflow 工单引擎

| 功能 | 说明 |
|---|---|
| 状态机 | Workflow → State → Transition（JSON 灵活建模）；定义校验纯函数 |
| 流转 | 审批节点动作受限、终态自动收档、撤回、迁移表外动作拦截；流转记录时间线 |
| 发版闭环 | release 工单审批完成自动触发流水线构建，结果回写流转行 |
| 工单中心 | 发起/通过/驳回/撤回 + 定义管理；普通用户仅见自己的工单 |

### dbops 数据库工单

| 功能 | 说明 |
|---|---|
| 实例纳管 | MySQL 实例 CRUD，密码 AES-256-GCM 密文（接口永不回显），TCP 探活 |
| SQL 上线工单 | 创建/取消/分页（普通用户仅本人）；有未结束工单的实例拒绝删除 |
| goInception 审核 | 真机闭环：magic 块指定目标库与动作（check/execute），errlevel 三态；执行自动备份生成回滚语句；soar 优化建议与多环境配置待环境 |

### agent Agent 通道

| 功能 | 说明 |
|---|---|
| 接入 | 资产页签发接入令牌（仅存哈希）；单文件二进制出站 WS 反向连接，断线退避重连 |
| 心跳与指标 | 心跳 + CPU/内存/磁盘/负载/网络指标上报，直接入 monitor_metric 复用趋势图 |
| 任务执行 | 任务号幂等、限时执行、输出截断回传、留痕分页 |
| 文件分发 | 分块下发 + SHA256 校验 + 原子落盘（见 job 插件） |
| Docker/K8s 代理 | HTTP-over-WS：内网 docker daemon / 集群内 ServiceAccount 经 agent 通道纳管（一键接入 YAML），无需暴露内网 API |

### aiops AI 运维

| 功能 | 说明 |
|---|---|
| LLM 网关 | 统一 OpenAI 兼容调用；供应商多配置，API Key AES-256-GCM 密文落库；共享连接池 + 请求级超时 |
| Pod AI 诊断 | k8s 采集（详情+事件+日志尾部）→ 固定 prompt 模板（`<snapshot>` 标签防提示注入）→ LLM → 结论留档 |
| MCP Server | 独立进程 `go run ./cmd/mcp`（默认 :8889）：`ops_asset_overview` / `ops_alert_recent` / `ops_ticket_status` 三工具，AI 助手可对话式查平台；身份经 auth_header 透传，不绕过权限体系 |

### 基座能力（gin-vue-admin 自带）

用户/角色/菜单/API 管理（Casbin RBAC + JWT）、代码生成器、表单设计器、定时任务、操作日志、多云对象存储、Swagger、MCP Server 骨架。

---

## 技术栈与架构

| 层 | 技术 |
|---|---|
| 后端 | Go 1.26（agent 兼容 1.24+） · Gin · GORM · Casbin v3 · JWT · Zap · Viper · robfig/cron |
| 前端 | Vue 3.5 · Vite 8 · Element Plus · Pinia · ECharts 5 · xterm.js 6 |
| 实时通道 | WebSocket（WebSSH / SFTP / 容器与 Pod 终端与日志）· SSE（流水线日志） |
| 执行通道 | Go SSH（密码/私钥/keyboard-interactive）· Docker SDK · client-go · Helm SDK v3.22 |
| 数据 | MySQL / SQLite（零依赖部署）· Redis（可选） |
| CI | GitHub Actions（server: build/vet/test；web: build） |

```
┌──────────────── web (Vue3.5 + Vite8 + Element Plus) ────────────────────────┐
│  GVA 基座(superAdmin/systemTools) + 业务插件视图                            │
│  asset 资产/凭据 │ term 终端/SFTP/审计 │ job 批量作业 │ pipeline 流水线       │
│  container 容器 │ k8s 集群/Helm │ monitor 监控告警 │ workflow 工单          │
└──────────▲──────────────────────────────────────────────▲───────────────────┘
           │ REST(JWT+Casbin)                WebSocket / SSE
┌──────────┴──────────────────────────────────────────────┴───────────────────┐
│  server : Gin + GORM + Casbin + Zap + 定时任务 + MCP Server 骨架            │
│  plugin/asset  term  job  pipeline  container  k8s  monitor  workflow      │
│  plugin/dbops  gpu  agent  aiops                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  agent/ : 单文件 Go Agent（反向 WS）——任务/文件分发/Docker-K8s 内网代理     │
│  执行通道：SSH 直连/ProxyJump │ Docker API(unix/TCP+TLS) │ K8s API         │
│  外部依赖：阿里云 OpenAPI │ 钉钉机器人 │ goInception │ LLM(OpenAI 兼容)     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 权限与数据隔离（三级）

| 层级 | 机制 | 生效范围 |
|---|---|---|
| **平台级** | Casbin RBAC：888 管理员全量 / 9528 只读 / 普通用户按授权 | 全部 REST 接口 + WS/SSE 握手 casbin 自验 |
| **资产级** | 资产组（组-主机-用户三表）+ casbin 资源规则 | 资产列表、批量执行、终端选机、监控采集全链路过滤 |
| **命名空间级** | `k8s_ns_grant` 授权表（cluster+namespace+user 唯一） | k8s 资源浏览与 Helm 列表 scoped 包装、Pod/YAML/Helm 单对象守卫（越权拒绝）、总览计数聚合 |

- 888 全量不受限；9528 平台只读；普通用户在 k8s 面**空 ns 聚合授权项、显式 ns 越权返回空、单对象越权直接拒绝**。
- WS/SSE 流端点因 query token 不经中间件，全部在握手时显式 casbin 自验（终端类仅 888，日志流类 888+9528）。

---

## 快速开始

### 方式一：一键部署（推荐，Ubuntu + Docker）

```bash
git clone https://github.com/hequan2017/new-ops.git
cd new-ops
bash scripts/deploy-test.sh                    # 全新安装或增量更新
# 或指定目标机器与端口：
NEW_OPS_TEST_HOST=root@1.2.3.4 NEW_OPS_WEB_PORT=8081 NEW_OPS_API_PORT=8888 bash scripts/deploy-test.sh
```

一条命令完成：交叉编译 → 上传 → systemd + nginx:alpine 容器安装 → SQLite 初始化（无 MySQL/Redis 依赖）→ 冒烟验证。管理员密码随机生成写入远端 `ADMIN_PASSWORD`（chmod 600），不落仓库。

| 默认项 | 值 |
|---|---|
| 前端入口 | `http://<host>:8081`（nginx 容器 `new-ops-web`） |
| 后端 | `:8888`（systemd 服务 `new-ops-server`） |
| 远端目录 | `/opt/new-ops/{bin,web,data,logs,resource,config.yaml}` |
| 数据库 | SQLite `/opt/new-ops/data/new_ops.db` |
| 凭据主密钥 | 环境变量 `NEW_OPS_CREDENTIAL_MASTER_KEY` 注入 unit（更新不丢） |

### 方式二：源码运行（开发）

```bash
# 后端（默认 :8888，SQLite 引导）
cd server && go mod tidy && go run main.go

# 前端（默认 :8080，代理 /api → :8888）
cd web && npm install && npm run serve
```

首次启动走 GVA 数据库引导页；凭据主密钥务必通过环境变量注入，勿写配置文件。

#### 开发环境要求与系统差异

| 项 | 要求 | 说明 |
|---|---|---|
| Go | **≥ 1.21**（建议直接装 1.26+） | `server/go.mod` 声明 `go 1.26.0`：Go 1.21+ 配合默认 `GOTOOLCHAIN=auto` 会自动拉取 1.26 工具链；**1.18 等旧版本连 go.mod 都解析不了**（报 `invalid go version`） |
| Node | ≥ 18 | 前端 Vite 构建 |
| Bash | 任意类 Unix shell | `scripts/smoke.sh`、`deploy-test.sh` 均为 bash 脚本 |

- **macOS**：`brew install go node` 即可；交叉编译目标为 linux/amd64（`make build-server` / `build-agent`），与本地架构无关。
- **Windows**：无 make 环境，直接执行 Makefile 内等价命令（如 `cd server && go build ./...`）；smoke/部署脚本需在 Git Bash / WSL 下运行；前端命令与 macOS 完全一致。
- **一键部署**（方式一）的执行端可以是 macOS/Windows(WSL)，目标机为 Linux（Ubuntu 推荐）。
- **多版本 Go 共存**：机器上装有多个 Go 时注意 PATH 优先级（别名如 `go126` 不会在非交互 shell 生效，CI/脚本用完整路径）。

### 上手路径

1. **凭据保险库**创建 SSH 凭据 → **主机管理**新增主机绑定凭据（或网段发现批量导入）→ 点「采集」回填硬件信息；
2. 「终端」直连或配置**跳板机**体验级联；**批量执行**跑命令，配脚本库/变量组参数化，切换 **Agent 后端**对比；
3. **流水线**编排阶段（shell 选目标主机 / http 步骤），试 webhook/cron 触发与审批 gate；
4. **容器管理**接入 `unix:///var/run/docker.sock` → 容器/镜像/网络/统计/终端；**Compose 模板**一键部署；
5. **K8s 管理**注册集群（k3s 即可）→ 总览看用量 → 资源浏览 → Pod WebShell → YAML dry-run 下发 → **Helm** 装升滚卸 → 888 在「命名空间授权」给普通用户开 ns，换账号验证隔离；
6. **告警规则**配阈值/探活 + 钉钉 webhook，主机页「监控」看趋势；
7. **工单中心**建 release 定义，审批通过看流水线自动构建；
8. **数据库工单**注册 MySQL 实例提 SQL 工单（goInception 审核闭环）；
9. **Agent**：资产页签发令牌 → 目标机跑 `baize-agent -server http://<server>:8888 -token <令牌>`；
10. **AI 运维**：`cd server && go run ./cmd/mcp` 后把 `http://127.0.0.1:8889/mcp` 接入 AI 助手；配置 LLM 供应商后在 Pod 页用 AI 诊断。

#### Agent 接入速查（内网免 SSH 场景）

```bash
# 1) 交叉编译 linux/amd64 单文件二进制
make build-agent          # 产物 agent/baize-agent，scp 到目标机即可（零依赖，无需装运行时）

# 2) 平台侧签发令牌：资产页 → Agent 接入 → 签发（令牌仅此一次显示，请立即保存）

# 3) 目标机运行（令牌无效时进程会明确报错退出而非静默重试）
./baize-agent -server http://<server>:8888 -token <令牌>
# 常用可选参数：
#   -interval 30s     心跳与指标上报周期（默认 30s）
#   -hostname web01   上报主机名（默认取 os.Hostname）
#   -docker-addr unix:///var/run/docker.sock   docker daemon 地址覆盖（默认自动探测 socket）
```

接入成功后：资产页「Agent」标签变绿（在线/离线按心跳时效计算），批量作业可切换 Agent 后端执行，文件分发/Docker/K8s 内网代理经同一 WS 通道。压测验证见 [agent/loadtest/README.md](agent/loadtest/README.md)。

---

## 安全设计

- **凭据一律 AES-256-GCM 加密落库**（信封加密，主密钥仅环境变量注入），任何接口不回显明文（只回末 4 位）；引用方按 ID 取用、内存解密即用即毁
- **SSH 主机公钥指纹校验**：TOFU 首连录入（SHA256），不匹配即拒绝；指纹仅 SSH 链路可写，CRUD 防伪造篡改
- **终端全程审计**：全量流镜像 + 命令抽取落库，回放可重演；审计开档失败拒绝开终端
- **写操作全走操作日志中间件**；危险操作（容器/镜像/卷/Pod 删除、Helm 卸载、工单撤回、批次取消）前端二次确认
- **插件路由显式鉴权**：业务插件路由组显式挂载 JWT + Casbin（不依赖底座隐式继承）
- **WS/SSE 流端点角色收敛**：终端类握手自验仅 888，日志流类 888+9528
- **SSRF 防护**：http 步骤/钉钉/Helm 仓库/LLM 网关 URL 仅允许 http(s)；webhook 令牌错误 404 语义防枚举
- **注入防护**：日志 tail `tail -F --` 终止选项解析 + 绝对路径校验；shell 步骤在远端资产执行，平台进程无本机 exec 面；AI 诊断 prompt 以 `<snapshot>` 隔离不可信数据
- **摘要算法**：全平台统一 SHA-256

---

## 接口与错误码规范

- REST 资源 `/api/<插件>/<资源>`，动作类 POST 子资源；WebSocket `/ws/<插件>/*`（query token 握手自验），SSE `/sse/<插件>/*`
- 响应统一 `{code, data, msg}`，列表分页 `pageInfo`
- **错误码分段**（每插件独占）：asset=1000-1099 · term=1100-1199 · job=1200-1299 · pipeline=1300-1399 · container=1400-1499 · k8s=1500-1599 · gpu=1600-1699 · dbops=1700-1799 · monitor=1800-1899 · workflow=1900-1999 · agent/org/aiops=2000-2199；通用错误沿用 GVA 7xxx 段
- Swagger 注释随代码同步（`/swagger/index.html`）

---

## 仓库结构

```
new-ops/
├── agent/                        # Go Agent 独立二进制（注册/心跳/指标/任务/文件分发/Docker-K8s 代理）
├── server/                       # Go 后端（module github.com/hequan2017/new-ops/server）
│   ├── plugin/
│   │   ├── asset/                # 资产中心 + 凭据保险库 + 网段发现
│   │   ├── term/                 # WebSSH/级联/审计/SFTP/日志 tail
│   │   ├── job/                  # 批量执行（SSH/Agent 双后端）/脚本库/变量组
│   │   ├── pipeline/             # 流水线/执行器/SSE/三通道触发
│   │   ├── container/            # Docker 全生命周期/Compose/应用模板
│   │   ├── k8s/                  # 多集群/资源浏览/WebShell/YAML 下发/Helm/命名空间授权
│   │   ├── monitor/              # 性能采集/告警引擎/钉钉
│   │   ├── workflow/             # 工单引擎 + 发版闭环
│   │   ├── dbops/                # MySQL 纳管/SQL 工单/goInception 审核
│   │   ├── gpu/                  # 算力台账 + 防超卖引擎
│   │   ├── agent/                # Agent 通道服务端网关
│   │   ├── aiops/                # LLM 网关 + Pod AI 诊断；cmd/mcp 独立进程
│   │   └── org/                  # 骨架就绪（钉钉待企业配置）
│   └── ...                       # GVA 底座（零修改）
├── web/src/plugin/               # 各插件前端视图（与 server/plugin 同名对应）
├── docs/plugin-dev-guide.md      # 插件开发规范
├── scripts/deploy-test.sh        # 一键部署/增量更新 + 冒烟
└── .github/workflows/            # CI（server build/vet/test + web build）
```

### 插件开发

单插件内部结构（详见 [plugin-dev-guide](docs/plugin-dev-guide.md)）：

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

## 路线图

| 里程碑 | 内容 | 状态 |
|---|---|---|
| M0-M3 | 底座/资产中心/终端作业/流水线 | ✅ 全部交付 |
| M4 | 容器 C1 全部 + C2 镜像/网络/卷/stats；k8s K1/K2（双 k3s 真集群验证） | ✅ 全部交付 |
| M5 | GPU 防超卖引擎；直通/HAMi/监控 | 🔨 待显卡环境 |
| M6 | monitor 三项 + dbops goInception 审核闭环；soar/多环境 | ✅ 主体交付 |
| M7 | workflow 工单 + aiops LLM 网关/Pod AI 诊断/MCP；org 钉钉待配置 | ✅ 主体交付 |
| M8 | Helm 全生命周期（values 表单）+ 命名空间授权 + Compose/模板化 + AI 诊断 | ✅ 全部交付 |
| M9 | Agent 通道全链（100 并发压测通过、本地 E2E）；集群内 daemon 级 E2E | ✅ 主体交付 |

- 发布：`v0.1.0`（M0-M2）→ `v0.2.0`（+M3/M4/M6 monitor/M7 workflow）→ **`v0.3.0-rc1`**（+M8/M9 全链 + goInception + AI 运维 + 应用模板化）
- 接口仍可能调整；生产部署前请完成安全复核（JWT 密钥、验证码、主密钥轮换）

---

## License

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)（Apache License 2.0）构建，本项目沿用该协议，详见 [LICENSE](LICENSE)。
