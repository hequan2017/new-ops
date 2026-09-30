# new-ops · 统一运维开发平台 — 开发计划

> 底座：[gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin) `main`（e8d675c, 2026-09-20，v2.9.2-stable 之后版本）
> 目标：把本人历史上散落的运维开发平台/工具的功能，整合到一个**插件化的统一运维开发平台**里，一次开发、长期维护。
> 文档版本：v1.0（2026-09-30）

---

## 一、背景与目标

多年来先后独立开发了 20+ 个运维平台与工具（autoops、chain、cmdb、seal、raptor、go-webssh、tianqi、docker-gpu-manage、new-jenkins 等），存在几个共性问题：

1. **技术栈分裂**：Python/Django、Go/Gin、Go/GoFrame 三套后端；Bootstrap、Vue2、Vue3 三代前端，维护成本高。
2. **功能重复**：CMDB、WebSSH、批量执行在 6+ 个仓库里各写了一遍，且互相不通。
3. **大量项目已停止维护**：autoops(365★)、chain(234★)、cmdb(128★)、seal(122★)、raptor(22★)、coypus 等均已标注停止更新，代码只能参考不能再用。
4. **底座重复造轮子**：go-admin(422★)、coypus、go-web-admin、pandaAdmin 各自实现了一遍 JWT+RBAC+菜单，而 gin-vue-admin 的底座（Casbin v3、代码生成器、插件机制、多云上传、MCP）远比自研的完善。

**目标**：以 gin-vue-admin 为唯一底座，历史平台的功能按域拆分、择优重写为 GVA 插件，形成一个数据互通的统一运维开发平台。

**收益**：

- 底座能力（RBAC/审计/代码生成/表单/插件/定时任务/Swagger/MCP）直接继承 GVA main，可持续跟随上游升级；
- 以 CMDB 为核心的数据模型，资产、凭据、终端、发布、容器、K8s、算力共用一套数据；
- 技术栈收敛为 Go 1.24 + Vue 3.5 + Element Plus，单人可维护。

---

## 二、现有平台功能盘点

### 2.1 仓库清册（运维开发相关）

| 仓库 | ★ | 技术栈 | 状态 | 核心功能 | 功能去向 |
|---|---|---|---|---|---|
| go-admin | 422 | Go/Gin | 停止维护 | JWT+Casbin RBAC+菜单骨架 | 被 GVA 底座取代 |
| autoops | 365 | Django | 停止维护 | CMDB、WebSSH、批量命令、监控图表、Docker/K8s 管理、Inception SQL 审核、代码发布、GPU 套餐、知识库、登录/命令审计 | asset/term/container/k8s/db/pipeline/gpu/kb |
| chain | 234 | Django | 停止维护 | 云主机 CMDB、WebSSH、Ansible 批量命令、脚本库(shell/py/yml)、变量组、日志 tail、定时任务、Docker/K8s 登记 | asset/term/job |
| cmdb | 128 | Django | 停止维护 | 资产+Ansible 采集、性能采集(ECharts)、批量命令/脚本、机柜管理、历史命令 | asset/monitor |
| seal | 122 | Django | 停止维护 | 开发模板、K8s 管理+Pod WebSSH、goInception SQL 审核+soar 优化、Celery/dramatiq 双异步 | k8s/db |
| go-webssh | 78 | Go | 可用 | 企业级 WebSSH：三角色、资产组授权、网段自动发现、ProxyJump 级联、AES-256-GCM 凭据、主机指纹校验、会话审计、SFTP | **term 首要参考** |
| husky | 27 | Django | 教程 | 多云 CMDB、DRF、Workflow/State/Transition 工单模型 | asset/workflow |
| zabbix-models | 28 | Python | 模板库 | Zabbix 模板、Grafana 面板、Ansible API 封装、告警机器人 | monitor 参考 |
| raptor | 22 | Go/GVA | 完结 | 钉钉扫码登录+组织同步+机器人告警、云密钥+ECS 同步、产品线、项目管理 | org/asset |
| new-aigc-comfyui | 4 | Go+Vue3 | 可用 | ComfyUI 多卡调度、GPU 看板、任务队列、素材/成片管理 | gpu(独立子域) |
| docker-gpu-manage | 15 | Go/GVA | 活跃 | GPU 实例调度+HAMi 显存切分+SSH 跳板机、K8s 多集群+AI 诊断、PCDN 节点带宽、戴尔资产台账 | gpu/k8s/pcdn/asset |
| DockerGPU | 3 | Go/GVA | 前代 | GPU 算力租赁：规格定价、资源核算防超卖、Web 终端 | gpu |
| k8s-gpu-admin | 0 | Go/GVA | 迭代 | GPU DeviceRequest 直通、overlay2 存储限额、MCP Server | gpu/ai |
| tianqi | 11 | Go/GVA | 活跃 | GPU 平台+数据集管理+SFT/DPO/CPT 训练(Swift)+vLLM 推理+MCP | gpu/train |
| kapigpu | 0 | Go/GVA | 练手 | Docker 集群 TLS 凭证管理 | gpu |
| new-jenkins | 0 | Go/GVA | 开发中 | 声明式流水线(Pipeline→Stage→Step、审批、参数、SSE 日志、cron/webhook)、运维中心(资产/凭据/SFTP/工单发版/巡检/备份/告警/调度/审计) | **pipeline/ops 首要参考** |
| ai-devops | 0 | Go | 迭代 | 资产台账+IPMI 电源、WebSSH、定时采集+告警规则、批量执行、Docker exec 终端、K8s、AI 诊断 | asset/term/monitor/container/k8s/aiops |
| pcfarm-admin | 0 | Go/GVA | 迭代 | 服务器资产、IP 地址池按 MAC 固定分配、PXE 启动策略、IPMI/Redfish 远控、Live Agent 心跳、装机事件 | pxe |
| model-ops | 0 | Go | 迭代 | 设备验收单、安装人员技能矩阵、并发压测引擎(RPS/P95)、72h 烤机、中英双语 | qa(远期) |
| quan-agent | 0 | Go | 迭代 | Windows 单文件 Agent、DPAPI 加密、7 个只读巡检 Skill、公钥校验 | aiops/agent |
| go-webssh / coypus / go-web-admin / pandaAdmin / panda / seal-vue / seal-d2-admin / go-ppp-console / pcdn-p2p / ai-pcdn / ai-manju | - | 各 | 停止/练手 | 底座练习、前端变体、边缘小工具 | 仅作参考，不迁移 |

### 2.2 功能维度归并

| 功能维度 | 出现过的仓库 | 归入插件 |
|---|---|---|
| 资产管理 CMDB | autoops, chain, cmdb, husky, raptor, ai-devops, pcfarm, docker-gpu-manage | `asset` |
| 凭据/密钥管理 | go-webssh, raptor, ai-devops, new-jenkins, kapigpu | `asset`（凭据保险库） |
| WebSSH/跳板机 | autoops, chain, cmdb, seal, go-webssh, ai-devops, tianqi, new-jenkins | `term` |
| 批量命令/脚本作业 | autoops, chain, cmdb, ai-devops | `job` |
| 流水线/代码发布 | autoops, raptor, new-jenkins | `pipeline` |
| Docker 管理 | autoops, chain, ai-devops, GPU 系列 | `container` |
| K8s 管理 | autoops, chain, seal, ai-devops, docker-gpu-manage | `k8s` |
| GPU 算力/租赁/训练 | autoops, tianqi, docker-gpu-manage, DockerGPU, kapigpu, k8s-gpu-admin, comfyui | `gpu` |
| SQL 审核工单 | autoops(Inception), seal(goInception+soar) | `dbops` |
| 监控采集/告警 | cmdb, autoops, ai-devops, raptor(钉钉), zabbix-models | `monitor` |
| 工单/审批流 | husky, new-jenkins(工单发版), autoops(订单) | `workflow` |
| 组织/IM 集成 | raptor(钉钉) | `org` |
| AI/MCP/巡检 | ai-devops, k8s-gpu-admin, tianqi, quan-agent | `aiops` |
| PXE/IPMI 装机 | pcfarm-admin, ai-devops | `pxe`（远期） |
| 边缘网络 PCDN | pcdn-p2p, go-ppp-console, ai-pcdn, docker-gpu-manage | `pcdn`（远期） |

---

## 三、总体设计

### 3.1 架构原则

1. **一切业务皆 GVA 插件**：`server/plugin/<name>` + `web/src/plugin/<name>`，底座目录尽量零修改，保证后续能合并上游 main。已有插件体系（`plugin/announcement`、`plugin/email`、`plugin/plugin-tool`）即模板。
2. **CMDB 是唯一数据心脏**：主机、凭据、产品线、机房等基础数据只存一份，term/job/pipeline/container/k8s/gpu 全部引用资产表外键。
3. **安全红线**：凭据 AES-256-GCM 加密落库、API 永不回显明文；WebSSH 会话可审计（命令记录/录像，参考 go-webssh）；全部写操作进 GVA 操作日志；跳板/终端链路支持主机公钥指纹校验。
4. **先直连后 Agent**：M1-M5 直接用 SSH / Docker API / K8s API 直连（历史项目验证过的路线）；M6 起引入轻量 Go Agent（参考 quan-agent 的只读 Skill 设计）做采集与执行代理。
5. **实时通道统一**：WebSocket（终端/日志 tail）与 SSE（流水线日志，参考 new-jenkins）统一封装在底座网关层，业务插件只注册 channel。

### 3.2 架构总览

```
┌────────────────────────── web (Vue3.5 + Vite8 + Element Plus) ──────────────────────────┐
│   GVA 基座(superAdmin/system/systemTools)   +   业务插件视图(asset/term/job/pipeline/...)  │
└───────────────▲──────────────────────────────────────────────▲──────────────────────────┘
                │ REST / Swagger                               │ WebSocket(xterm) / SSE(日志流)
┌───────────────┴──────────────────────────────────────────────┴──────────────────────────┐
│  server : Gin 1.10 + GORM + Casbin v3 + JWT + Zap + Viper + 定时任务 + MCP Server        │
│  plugin/asset    plugin/term     plugin/job      plugin/pipeline   plugin/container      │
│  plugin/k8s      plugin/gpu      plugin/dbops    plugin/monitor    plugin/workflow       │
│  plugin/org      plugin/aiops    plugin/pxe(远期) plugin/pcdn(远期) plugin/kb(远期)        │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│  执行通道：SSH/Paramiko⇄Go SSH │ Docker API(TLS) │ K8s API(kubeconfig AES-GCM) │ Agent    │
│  外部依赖：MySQL/PG │ Redis │ goInception │ HAMi-core │ Zabbix/Prometheus(适配) │ 钉钉/IM  │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

### 3.3 插件规划与功能明细

**P0 — asset 资产中心（一切的前提）**
- 主机资产 CRUD：主机名/IP/系统/CPU/内存/磁盘/SN/厂商/机房/机柜/U位/产品线多对多/负责人/状态（合并 chain+cmdb+husky+raptor 字段设计）
- 采集三通道：SSH 采集（参考 cmdb 的 Ansible setup 思路，改为 Go SSH）、云同步（阿里云 ECS，参考 raptor 云密钥）、Agent 上报（远期）
- 凭据保险库：SSH 密码/私钥/云 AccessKey/Docker TLS/kubeconfig，AES-256-GCM 加密、按资产组授权引用（合并 go-webssh + raptor + kapigpu）
- 产品线/业务/机房/机柜字典、Excel 导入导出、资产变更历史
- 数据权限：普通用户只见授权资产，管理员全量（沿袭各项目 authorityId 约定）

**P0 — term 终端与文件**
- WebSSH：xterm.js + WebSocket，密码/私钥/键盘交互、窗口自适应、心跳重连（go-webssh 标准）
- 跳板级联 ProxyJump（最多 5 层）、主机指纹 SHA256 校验、会话命令审计
- SFTP 文件管理器（go-webssh/new-jenkins）
- 远程日志 tail（chain）、网段 CIDR 并发探测自动导入资产（go-webssh）

**P1 — job 批量作业**
- 批量命令执行（并发上限、超时控制，参考 ai-devops）
- 脚本库：shell/python/yaml 三类 + 变量组关联资产（chain）+ 执行历史与结果留存
- 底座定时任务承接周期作业

**P1 — pipeline 流水线与发布**
- 声明式流水线 Pipeline→Stage→Step：HTTP/Shell 步骤、人工审批 gate、并行阶段、失败继续、参数体系与变量替换、触发快照（**以 new-jenkins 现成实现为蓝本迁移**）
- 构建：状态机 pending→running→(approval)→success|failed|canceled、构建序号、取消/重跑、SSE 实时日志
- 触发器：cron / webhook / 工单发版审批通过自动触发
- 发布目标绑定资产与凭据（打通 asset）

**P1 — container 容器管理**
- Docker 接入点纳管（unix/TCP+TLS）、连通巡检（GPU 系列成熟代码）
- 容器列表/启停/删除/日志流/exec 交互终端（ai-devops）、镜像/网络/卷管理

**P1 — k8s 集群管理**
- 多集群接入（kubeconfig 加密存储+连接池）、Namespace/Node/工作负载(Deployment/StatefulSet/DaemonSet 扩缩容与滚动重启)/Pod 列表详情日志
- Pod WebShell（seal 验证过的路线）、集群/节点/Pod 指标、AI 故障诊断（状态+事件+日志喂 LLM，docker-gpu-manage 已验证）
- 命名空间级 RBAC（全局/集群/命名空间三级）

**P2 — gpu 算力平台**
- 算力节点纳管、镜像库（上架/下架/显存切分标记）、产品规格与定价、实例全生命周期
- 智能匹配调度：按规格 GPU/显存/CPU/内存/磁盘扣减已占用资源防超卖（DockerGPU/k8s-gpu-admin 算法）
- HAMi 显存切分（tianqi 成熟实现）、实例资源监控、SSH 跳板机（2026 端口交互选容器）、端口转发管理
- 模型训练（远期）：数据集版本管理、SFT/DPO/CPT、Swift+LoRA、vLLM 推理

**P2 — dbops 数据库工单**
- MySQL 实例与账号管理（autoops）、SQL 上线工单：goInception 审核/执行/备份 + soar 优化建议（seal）、多环境配置、工单审批联动 workflow

**P2 — monitor 监控告警**
- 性能采集（CPU/内存/磁盘/网络，定时任务）、指标留存与图表（ECharts）
- 告警规则引擎（ai-devops）+ 钉钉/IM 机器人推送（raptor）、端口探活
- Zabbix/Prometheus 适配器（zabbix-models 的模板与告警机器人经验）

**P2 — workflow 工单引擎**
- 通用工单：Workflow→State→Transition 模型（husky）、审批节点绑定角色、事件钩子（发版/SQL/资源申请）

**P3 — org 组织与 IM**：钉钉扫码登录、部门/用户定时同步、机器人（raptor 直接移植，GVA 用户表打通）

**P3 — aiops AI 运维**：内置 MCP Server（GVA main 已带 mcp/ 骨架）、AI 诊断、只读巡检 Skill 集（quan-agent）

**远期**：pxe（IP 池/PXE/IPMI/Redfish 装机，pcfarm-admin 移植）、pcdn（边缘节点与带宽统计）、kb（运维知识库，autoops）、qa（设备验收/压测烤机，model-ops）、train（训练平台拆分）

---

## 四、里程碑

| 阶段 | 内容 | 预估 | 交付判据 |
|---|---|---|---|
| **M0 底座就绪** | 品牌化与配置、移除 example 演示、docker-compose/Makefile、CI、插件目录骨架与开发规范文档、DEV_PLAN 落库 | 1 周 | compose 一键起；空插件可被加载 |
| **M1 资产中心** | asset 插件：资产 CRUD/采集三通道/凭据保险库/产品线机房/导入导出/数据权限 | 2-3 周 | SSH 采集回填资产；凭据密文存储且接口不回显 |
| **M2 终端作业** | term + job：WebSSH/级联/审计/SFTP/日志 tail、批量命令脚本变量组 | 2-3 周 | 浏览器经跳板连终端且留审计；百台批量执行不雪崩 |
| **M3 流水线** | pipeline：引擎迁移自 new-jenkins + 工单发版闭环 | 2 周 | webhook 触发→审批→发布→SSE 日志全链路 |
| **M4 容器与 K8s** | container + k8s：Docker 全家桶、多集群/Pod WebShell/AI 诊断/三级 RBAC | 3 周 | 两套真实环境纳管通过 |
| **M5 GPU 算力** | gpu：节点/规格/实例/防超卖/HAMi/跳板机 | 3 周 | 按规格开通虚拟 GPU 实例并可达资源上限 |
| **M6 数据库与监控** | dbops + monitor：goInception 工单、采集告警钉钉 | 2 周 | SQL 工单审批上线留备份；告警触达钉钉 |
| **M7 工单与 AI** | workflow + org + aiops：工单引擎、钉钉登录同步、MCP 巡检 | 2 周 | 钉钉扫码登录可用；MCP 工具可被 AI 助手调用 |

M1-M3 完成即可替代 autoops/chain/go-webssh/new-jenkins 的日常使用；M4-M5 替代 Docker/K8s/GPU 系列；之后旧仓库逐一归档标注"功能已并入 new-ops"。

### M1 任务清单（细化到可开工）

- [ ] 定稿资产表/产品线/机房/凭据表模型与 ER 图
- [ ] asset 插件脚手架（api/router/service/model + 前端视图，走代码生成器初版再改造）
- [ ] Go SSH 采集器（hostname/os/cpu/mem/disk 现场回填）
- [ ] 凭据保险库（AES-256-GCM、密钥来自配置+环境变量、写操作审计）
- [ ] 阿里云 ECS 同步任务（密钥管理+定时+手动触发）
- [ ] Excel 导入导出、变更历史
- [ ] 数据权限：资产组 + Casbin 资源规则
- [ ] 单元测试 + Swagger 注释 + 插件开发规范文档

---

## 五、关键设计决策

1. **凭据保险库**：独立表存密文，主密钥从配置/环境变量注入；接口层强制脱敏（只回显末 4 位）；引用方（term/job/pipeline）凭 ID 取用，服务端内存解密即用即毁。
2. **WebSSH 会话审计**：WebSocket 消息镜像落库（命令级正则抽取 + 全量异步录像文件），审计员角色只读回放——直接对齐 go-webssh 的三角色模型。
3. **流水线引擎**：移植 new-jenkins 的三层数据模型与状态机，触发快照保证定义修改不影响进行中的构建；执行器初期仍在本机（与 new-jenkins 现状一致），M6 后接入 Agent 执行器实现工作空间隔离——这是 new-jenkins README 已声明的边界，沿用其演进路线。
4. **GPU 防超卖**：节点可用量 = 物理规格 - Σ(运行实例规格)，分配时行级锁 + 事务；HAMi 虚拟显存以 `CUDA_DEVICE_MEMORY_LIMIT` 注入，节点侧标注是否支持切分。
5. **实时通道**：底座提供 `/ws/*`（xterm、日志 tail、容器 exec）与 `/sse/*`（流水线日志）统一网关与鉴权中间件，业务插件只注册业务 channel，避免每个插件自拉 websocket。
6. **兼容旧平台**：提供只读的数据导入脚本（旧库→新 CMDB），旧仓库 README 顶部统一加"已并入 new-ops"声明与跳转。

---

## 六、目录结构规划（新增部分）

```
new-ops/
├── server/
│   ├── plugin/                  # GVA 既有插件机制
│   │   ├── announcement/ email/ plugin-tool/ auto/   (上游自带)
│   │   ├── asset/               # P0 资产中心 + 凭据保险库
│   │   ├── term/                # P0 WebSSH/SFTP/审计
│   │   ├── job/  pipeline/  container/  k8s/
│   │   ├── gpu/  dbops/  monitor/  workflow/
│   │   └── org/  aiops/         # P3
│   └── task/                    # 采集/巡检定时任务注册
├── web/src/plugin/              # 各插件前端视图（与 server/plugin 同名对应）
├── docs/
│   ├── DEV_PLAN.md              # 本计划
│   └── plugin-dev-guide.md      # M0 产出：插件开发规范
├── deploy/                      # 上游自带，补充 docker-compose.prod.yaml
└── scripts/                     # 旧平台数据导入脚本
```

---

## 七、风险与对策

| 风险 | 对策 |
|---|---|
| 上游 GVA main 演进导致合并冲突 | 底座零修改原则；业务全在 plugin/ 目录；每季度合并一次上游 tag |
| 单人维护 15+ 插件摊子过大 | 严格按里程碑交付，P3/远期插件不提前开工；每个插件 M 完成即发布 tag |
| WebSSH/终端类功能的安全责任重 | 凭据加密+审计回放+指纹校验三项安全红线不裁剪；README 明示安全边界（沿袭 new-jenkins 做法） |
| SSH/Docker/K8s 直连在大规模下不稳 | M1-M5 直连上限明确写文档；Agent 执行器作为 M6 后的硬性演进项 |
| 旧平台数据结构差异大 | 只迁移核心资产表，历史数据留旧库存档，新平台不背历史包袱 |

---

## 八、决议记录

- 2026-09-30：确定以 gin-vue-admin main（e8d675c）为底座新建仓库 `new-ops`，本计划为第一版；功能盘点基于本人 GitHub 各仓库 README（当日版本）。
