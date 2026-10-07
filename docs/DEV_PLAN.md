# 白泽（BaiZe）· 统一运维开发平台 — 开发计划

> 平台中文名：**白泽**（上古神兽，通晓天下万物之情，寓意统一纳管与 AI 运维）；英文名 BaiZe；仓库/工程名 `new-ops`。
> 底座：[gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin) `main`（e8d675c, 2026-09-20，v2.9.2-stable 之后版本）
> 目标：把本人历史上散落的运维开发平台/工具的功能，整合到一个**插件化的统一运维开发平台**里，一次开发、长期维护。
> 文档版本：v1.6（2026-09-30，品牌化：中文名定为「白泽 BaiZe」，前端/后端/文档全面应用）

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
4. **先直连后 Agent**：M1-M5 直接用 SSH / Docker API / K8s API 直连（历史项目验证过的路线）；随后引入轻量 Go Agent（参考 quan-agent 的只读 Skill 设计）做采集与执行代理，并作为 Docker/K8s 管理的第二种接入形态，分期见「六、专项开发计划」。
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

**P1 — container 容器管理**（分期细化见第六节 C1-C3）
- Docker 接入点纳管（unix/TCP+TLS）、连通巡检（GPU 系列成熟代码）
- 容器列表/启停/删除/日志流/exec 交互终端（ai-devops）、镜像/网络/卷管理

**P1 — k8s 集群管理**（分期细化见第六节 K1-K4）
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

### 3.4 核心数据模型汇总

统一约定：业务表使用 GVA 代码生成器惯例字段（id/created_at/updated_at/deleted_at）；资产类业务通过 `asset_id` 外键挂到资产中心；数据权限统一由 casbin 资源规则实现，不各写一套。

| 插件 | 核心表 | 关键字段 / 说明 |
|---|---|---|
| asset | asset_host | hostname/ip/os/cpu/mem/disk/sn/vendor/room_id/rack/u_pos/status/owner |
| | asset_product_line（+主机多对多） | name/level/owner |
| | asset_room / asset_rack | 区域、机柜容量与在用 |
| | asset_group / asset_host_group | 数据权限授权单元 |
| | cred_credential | type(ssh密码/私钥/云AK/DockerTLS/kubeconfig)、cipher(AES-256-GCM)、明文末4位、引用计数 |
| | asset_collect_record | 采集来源(ssh/云/agent)、原始JSON、时间 |
| term | term_session / term_command_log | 资产/用户/起止/录像文件路径；命令、时间、风险等级 |
| job | job_script / job_variable_group / job_exec_record | 类型(shell/py/yml)与版本；变量组↔资产；状态/结果/耗时 |
| pipeline | pipeline / pipeline_stage / pipeline_step / pipeline_build / build_log | 定义三层模型；构建序号/状态机/参数快照/触发方式；日志按 stdout/stderr/system |
| container | docker_endpoint / docker_container | 地址/TLS凭据ID/巡检状态；endpoint_id/容器ID/关联资产 |
| k8s | k8s_cluster / k8s_ns_grant | kubeconfig 密文/labels/状态；cluster+namespace↔user/role |
| gpu | gpu_node / gpu_spec / gpu_image / gpu_instance | 节点(资产ID/显卡/可用余量)；规格(型号/数量/定价)；实例(状态/端口/HAMi 参数) |
| dbops | dbops_instance / dbops_order | 实例与账号(密文)；SQL 工单(goInception 结果/环境/状态) |
| monitor | monitor_metric / monitor_alert_rule / monitor_alert_event | 指标(asset_id/name/value/ts)；规则(阈值/持续/通知渠道)；事件 |
| workflow | wf_definition / wf_instance / wf_task | 状态机定义(husky 模型)；实例/当前节点/审批人/结论 |
| org | org_dingtalk_config / org_sync_record | corpid/凭据；同步批次 |
| agent | agent_instance / agent_task | 见第七节 6.1 |

**设计红线**：凭据表只存密文与末 4 位；删除默认软删除；变更历史统一 `*_history` 表（asset 先行，其余插件随 M2 起跟进）。

### 3.5 技术选型与依赖清单

所有新增第三方依赖必须在此表登记后使用，开发会话不得随意引入同类替代库（防止依赖膨胀）：

| 关注点 | 选型 | 说明 | 已验证来源 |
|---|---|---|---|
| SSH / SFTP | `golang.org/x/crypto/ssh` + `github.com/pkg/sftp` | 纯 Go 实现，支持 keyboard-interactive | go-webssh |
| 终端通道 | `gorilla/websocket` + `@xterm/xterm`(6.0)+`@xterm/addon-fit` | 统一走插件内 `/term/ws` 网关（query token 握手鉴权） | new-jenkins / tianqi |
| Docker | Docker SDK（moby/client） | 支持 TLS 连接远程节点 | GPU 系列 |
| Kubernetes | `k8s.io/client-go` | kubeconfig 动态加载 + 连接池 | seal / docker-gpu-manage |
| K8s 指标 | `k8s.io/metrics`（与 client-go 同版本 v0.33.3） | metrics-server Node/Pod 用量（集群总览；不可用降级） | 场53 登记 |
| K8s YAML | `sigs.k8s.io/yaml` v1.6.0（client-go 传递依赖转直接） | 工作负载 YAML 序列化（K2 查看口径） | 场54 登记 |
| 凭据加密 | `crypto/aes` + GCM（信封加密） | 标准库优先，主密钥环境变量注入 | go-webssh |
| 云同步 | 阿里云 OpenAPI RPC V1 签名（自实现，纯标准库） | 仅 DescribeInstances 单接口，不引入官方 SDK（传递依赖过重）；AccessKey 走凭据保险库；HMAC-SHA1 为协议固定要求 | raptor |
| SQL 审核 | goInception + soar 外部二进制 | 进程调用 + 结果解析 | seal |
| 流水线日志 | SSE（net/http flusher） | `/sse/pipeline/*` | new-jenkins |
| 定时任务 | GVA 内置 timer/cron | 注册制，不另引 cron 库 | GVA |
| API 文档 | swaggo 注释 | 随代码同步 | GVA |

**原则**：标准库能解决的不引第三方；同类场景全平台只用一个库。

### 3.6 接口与错误码规范

- 路由：REST 资源用 `/api/<插件>/<资源>`，动作类用 POST 子资源（如 `/api/asset/host/:id/collect`）；WebSocket `/ws/<插件>/*`，SSE `/sse/<插件>/*`。
- 响应：沿用 GVA 统一结构 `{code, data, msg}`；列表分页用 GVA `pageInfo`；过滤参数统一 `keyword`（模糊）/ `status` / 时间区间。
- 错误码分段（每插件独占一段，避免跨插件冲突）：asset=1000-1099、term=1100-1199、job=1200-1299、pipeline=1300-1399、container=1400-1499、k8s=1500-1599、gpu=1600-1699、dbops=1700-1799、monitor=1800-1899、workflow=1900-1999、agent/org/aiops=2000-2199；通用错误沿用 GVA 现有 7xxx 段。
- 审计：所有写操作（POST/PUT/DELETE）统一过 GVA 操作日志中间件；凭据类接口强制脱敏，任何接口不得返回明文密钥。

---

## 四、里程碑

| 阶段 | 内容 | 预估 | 交付判据 |
|---|---|---|---|
| **M0 底座就绪** | 品牌化与配置、移除 example 演示、docker-compose/Makefile、CI、插件目录骨架与开发规范文档、DEV_PLAN 落库 | 1 周 | compose 一键起；空插件可被加载 |
| **M1 资产中心** | asset 插件：资产 CRUD/采集三通道/凭据保险库/产品线机房/导入导出/数据权限 | 2-3 周 | SSH 采集回填资产；凭据密文存储且接口不回显 |
| **M2 终端作业** | term + job：WebSSH/级联/审计/SFTP/日志 tail、批量命令脚本变量组 | 2-3 周 | 浏览器经跳板连终端且留审计；百台批量执行不雪崩 |
| **M3 流水线** | pipeline：引擎迁移自 new-jenkins + 工单发版闭环 | 2 周 | webhook 触发→审批→发布→SSE 日志全链路 |
| **M4 容器与 K8s（一期）** | container(C1) + k8s(K1,K2)：Docker 生命周期/镜像/网络/卷/事件订阅、多集群注册、Node/工作负载/Pod WebShell | 3 周 | 两套真实环境纳管通过 |
| **M5 GPU 算力** | gpu：节点/规格/实例/防超卖/HAMi/跳板机 | 3 周 | 按规格开通虚拟 GPU 实例并可达资源上限 |
| **M6 数据库与监控** | dbops + monitor：goInception 工单、采集告警钉钉 | 2 周 | SQL 工单审批上线留备份；告警触达钉钉 |
| **M7 工单与 AI** | workflow + org + aiops：工单引擎、钉钉登录同步、MCP 巡检 | 2 周 | 钉钉扫码登录可用；MCP 工具可被 AI 助手调用 |
| **M8 容器与 K8s（二期）** | container(C2,C3) + k8s(K3)：Compose/镜像库联动、Helm、三级 RBAC、AI 诊断、GPU 资源视图 | 2-3 周 | Helm 安装/回滚可用；命名空间级权限生效 |
| **M9 Agent 通道** | agent(A1,A2,A3)：反向长连接、采集上报、执行代理、流水线远程执行器 | 3 周 | Agent 注册/心跳/采集闭环；流水线可在 Agent 主机执行且工作空间隔离 |

M1-M3 完成即可替代 autoops/chain/go-webssh/new-jenkins 的日常使用；M4-M5 替代 Docker/K8s/GPU 系列；M8-M9 补齐容器二期与 Agent 通道；之后旧仓库逐一归档标注"功能已并入 new-ops"。全计划合计约 22-23 周。

> 口径：里程碑预估的「周」按每天 2 场 × 2 小时 = 28 小时计（v1.2 之前口径）；2026-10-01 起加密为每天 3 场（9/14/20 点）= 42 小时/周，全计划约 15 周，与下方二十一场排期表一致。

### 二十一场排期表（10-01～10-07 自动化开发）

> 本场序号 = 「开发日志」已有行数 + 1，对照本表取本场目标；进度超前则提前取下一场目标，滞后则顺延并以任务清单未勾选项为准。7 天 × 3 场 × 2h = 42h，目标：M0 全部 + M1 全部 + M2 大部分，收尾打 tag v0.1.0。

| 场 | 时间 | 本场目标（对应任务清单项） |
|---|---|---|
| 1 | 10-01 上午 | M0：README/标题品牌化、移除 example 路由菜单、docker-compose + Makefile |
| 2 | 10-01 下午 | M0：GitHub Actions CI、config 分环境、10 个插件骨架注册、plugin-dev-guide 初稿；M0 收尾自查 |
| 3 | 10-01 晚 | M1：asset_host/产品线/机房机柜 表结构+migration+代码生成器起 CRUD |
| 4 | 10-02 上午 | M1：资产 CRUD 前端联调 |
| 5 | 10-02 下午 | M1：Excel 导入导出 + 资产变更历史 |
| 6 | 10-02 晚 | M1：凭据保险库（AES-256-GCM 信封加密、脱敏回显、引用计数、审计） |
| 7 | 10-03 上午 | M1：Go SSH 采集器 + 采集记录页 |
| 8 | 10-03 下午 | M1：阿里云 ECS 同步（AK 管理+定时+手动触发） |
| 9 | 10-03 晚 | M1：数据权限（asset_group + casbin 资源规则） |
| 10 | 10-04 上午 | M1：仪表盘统计 + 单测/Swagger 补齐 |
| 11 | 10-04 下午 | M1 验收自测收尾；M2 开工：WebSocket 网关 + xterm 组件封装 |
| 12 | 10-04 晚 | M2：WebSSH 连接（密码/私钥/键盘交互）+ 窗口自适应 |
| 13 | 10-05 上午 | M2：ProxyJump 级联 + 主机指纹校验 |
| 14 | 10-05 下午 | M2：会话审计（命令抽取落库 + 录像落盘 + 回放页） |
| 15 | 10-05 晚 | M2：SFTP 文件浏览器 |
| 16 | 10-06 上午 | M2：批量执行（并发池 + 超时/取消） |
| 17 | 10-06 下午 | M2：脚本库 CRUD+版本 + 变量组关联资产 |
| 18 | 10-06 晚 | M2：远程日志 tail + CIDR 网段发现导入 |
| 19 | 10-07 上午 | 缓冲：M0-M2 未勾选项收尾、单测补齐、全量构建验证 |
| 20 | 10-07 下午 | 收尾：README/plugin-dev-guide 定稿、tag v0.1.0 |
| 21 | 10-07 晚 | 机动：若 M2 已完则启动 M3（pipeline/stage/step 表结构+CRUD 起步），否则继续缓冲收尾 |

### M0 任务清单（细化到可开工）

- [ ] 仓库品牌化：README 改写（项目定位+DEV_PLAN 链接）、web 标题/Logo、首次登录强制改密提示
- [ ] 移除 example 演示模块路由与菜单（保留代码生成器模板本身）
- [ ] docker-compose（mysql+redis+server+web）与 Makefile（dev/build/push）
- [ ] GitHub Actions CI：push 触发 server(go build/vet/test) + web(build)
- [ ] config.yaml 分环境（dev/prod）、凭据主密钥环境变量化
- [x] docs/plugin-dev-guide.md 插件开发规范（目录结构、注册、菜单/API/casbin 初始化脚本、代码生成器配合）
- [x] asset/term/job/pipeline/container/k8s/gpu/dbops/monitor/workflow 插件骨架目录与占位注册

### M1 任务清单（细化到可开工）

- [ ] 定稿资产表/产品线/机房/凭据表模型与 ER 图（进行中：资产四表+产品线关联已定稿落地；凭据表随本场 5）
- [x] asset 插件脚手架——主机管理全链路（模型/服务/API/路由/种子/前端页面均已落地并远端验证）；机房与产品线页面下一场
- [x] Go SSH 采集器（x/crypto/ssh 密码+私钥双认证、5s 命令超时；解析纯函数单测；凭据走保险库 GetPlaintext；主机名不覆盖人工命名；真实 sshd 验证回填 ubuntu/88C/125G/229G）
- [ ] 凭据保险库（AES-256-GCM、密钥来自配置+环境变量、写操作审计）
- [ ] 阿里云 ECS 同步任务（进行中：手动触发全链路已交付并验证——签名/分页/upsert/错误回传；定时触发待底座提供插件级 timer 注册接口）
- [x] Excel 导入导出、变更历史（excelize 导出/按IP upsert 导入；历史随 CRUD 事务记录+时间线抽屉；发现项：IP 唯一索引与软删除冲突，下场改复合唯一索引）
- [x] 数据权限：资产组（组-主机-用户三表）+ 888全量/普通用户仅组内过滤 + 9528 只读策略（双角色远端验证通过）
- [ ] 单元测试 + Swagger 注释 + 插件开发规范文档（规范文档 M0 已交付；本项剩余：服务层 DB 相关单测补 mock 覆盖）

### M2 任务清单

- [x] WebSocket 网关与鉴权中间件；xterm.js 组件（term 插件 /term/ws：query token 握手自验、二进制数据流+JSON 控制帧、30s 心跳、resize 自适应；数据权限过滤已接入主机选择）
- [x] WebSSH 连接管理（密码/私钥/keyboard-interactive 直连真机验证；ProxyJump 级联：JumpHostID 自引用链 ≤5 跳/禁环/逐跳独立认证与指纹校验+TOFU 回填，buildJumpChain 6 组单测；主机表单跳板机选择+指纹只读回显；CRUD 层 ssh_fp 防篡改与 jump_host_id 取消写入修复）
- [x] 会话审计：命令抽取落库 + 流镜像（TermSessionStream 按 seq）+ 审计回放页（xterm 重演，上行标注）
- [x] SFTP 文件浏览器（列表/上传/下载/删除；场21 交付，场22 回归通过：嵌套 mkdir/upload/list/递归删除全验证）
- [x] 批量执行（并发池全链场22 交付：信号量限流/单任务超时/批次取消（4 组单测）、数据权限、逐主机凭据与指纹 TOFU、执行页轮询取消；场23 补齐：脚本库 CRUD+版本归档（更新递增、历史抽屉）、变量组 CRUD（key=value 编辑器+JSON 校验）、执行联动——脚本内容为命令+变量组 {{key}} 服务端渲染入批次快照；真机端到端：v1→v2 版本归档、{{greeting}}→baize 渲染执行通过；待后续增强：按主机所在资产组自动聚合注入变量（当前为执行时选组统一渲染））
- [x] 远程日志 tail、CIDR 网段发现一键导入资产（场24 交付：/term/logtail WS——复用级联拨号+指纹校验，远端 tail -n N -F -- 路径防注入，前端 LogTail 组件滚动跟随/截断防膨胀；CIDR 发现 ExpandCIDR 纯函数（4096 上限/网络广播跳过，6 场景单测）+并发 SSH banner 探测+按 IP 导入（存在跳过）；真机：/29 段发现本机横幅、非法 CIDR 1006 拦截、导入 1 台/重复跳过 1 台、tail 初始回看+实时追加全到达）

> **M2 全部完成（2026-10-03，场次 19-24）**：WebSSH/级联/审计/SFTP/批量执行（含脚本库变量组）/日志 tail/网段发现。

### M3 任务清单

- [x] pipeline/stage/step 三层模型表与 CRUD、前端编排页（场25 交付：三表迁移、嵌套校验纯函数 8 场景单测、Create/Update（整体替换）/Delete（级联）/列表/详情按 Sort 预载、五接口+菜单/API/casbin 种子（888 全量 9528 只读）、错误码 1301-1304；前端列表+编排弹窗（阶段卡片：审批 gate/失败继续，步骤 shell/http 切换）；真机嵌套 CRUD 验证：Sort 归一/审批位/http 默认 GET/级联删除全过；修复 Create 走 replaceStages（Omit 关联跳过阶段落库，真机回读发现））
- [x] 执行器：构建状态机、参数校验与变量替换、人工审批 gate、并行阶段、continue_on_error（场26/27 核心链——两表、快照落库+后台状态机、审批 gate approveCh、构建序号递增、shell SSH 到绑定资产执行（无本机 exec 面，校验强制 HostID）、真机四链路+修复 Wait 退出码误用；场31 补齐并行阶段——PipelineStage.parallel 位、runStageSteps 串行/并发双路径（信号量上限 10、首败 cancel 快速中断兄弟步骤、ContinueOnError 跑完收集）、顺手修复取消路径双重收档竞态与并发 cancel 未传导真机暴露的两处缺陷；真机：并行双步日志交错成功、快速失败兄弟步骤被中断无输出）
- [x] SSE 日志网关 + 构建日志落库分页拉取（场26 交付落库+分页拉取+前端日志抽屉；场28 补 SSE：GET /sse/pipeline/build/logs（public 组 query token 自验，与 term/ws 同模式）、GetBuildLogsAfter 按 lastId 1s tick 增量推送+status/done 事件收流、X-Accel-Buffering no 防 nginx 缓冲、断线 lastId 续传；前端进行中构建 EventSource 订阅（log/status/done 三事件），done 自动刷新列表、异常降级一次性拉取；真机验证：drip 步骤逐秒输出，事件序列 status→5×log→status→done 全部到达）
- [x] 触发器：cron（注册到底座定时任务）/webhook/手动（手动场27；webhook 场29（public 组令牌即凭据 uuid、404 语义防枚举、params 渲染快照）；cron 场30——Pipeline cron_enabled/cron_spec + SyncPipelineCron 幂等同步 global.GVA_Timer（RemoveTaskByName→AddTaskByFunc，robfig 标准 5 段/@every）+ Create/Update/Delete 全挂同步 + 启动 initialize.Timer 全量恢复；真机：@every 10s 25s 内自动 2 次构建 operator=cron、删除后任务即摘）
- [x] 工单发版闭环：workflow 审批通过事件钩子 → 自动触发流水线（场41：release 工单到达终态 InstanceFinished 时 fireReleaseHook，params 约定 {pipelineId,params}，触发结果落 wf_actionLog hook 行；驳回不触发、钩子失败不阻断收档；真机：工单 #3 approve→构建 #1 自动产生 operator=工单#3 状态成功。**M3 六项全部完成**）
- [x] 构建历史：即时取消（场27 cancelRegistry + ctx 中断 SSH/HTTP，真机验证）、复用历史参数重跑（场29 restart 接口+前端按钮，复用原参数对当前定义生成新快照，真机 buildNo 递增验证）

### M4 任务清单（container C1 + k8s K1/K2，细节见第六节 6.2/6.3）

- [x] container：endpoint CRUD + TLS 凭据入保险库 + 30s 巡检（场32 交付——docker_endpoint 模型、NewDockerClient（unix 直连/TCP+TLS，v28 WithTLSClientConfig 直收 PEM，凭据保险库 docker_tls 解密 JSON{ca,cert,key}）、PingEndpoint+StartInspectLoop 30s 合并巡检（并发5）、五接口+种子+前端接入点页（35s 轻刷新）；Docker SDK v28.0.4 登记 3.5（go-connections 钉 v0.5.0 修 Windows 编译）；真机：unix socket 巡检在线 API 1.54、离线节点状态回写、重名拦截；容器生命周期与创建参数由场33（列表/启停/重启/删除）与场34（创建：端口映射/环境变量/挂载/资源限制/重启策略）交付——场54 修正陈旧勾选）
- [x] container：日志流 + exec 终端（WebSocket 桥接）；inspect 回写 + docker events 订阅（场36 双 WS 通道：LogsWS（logs -f，TTY 判断+stdcopy 去复用）/ExecWS（exec /bin/sh TTY→hijack⇄WS，resize），public 组 query token 自验；前端 ContainerShell/ContainerLogs 组件+抽屉按钮；真机：日志流跟随连续 tick、exec echo 双向打通（修复 readControl 丢弃用户输入真 bug）。场37 events：复用 30s 巡检按 [last_event_at, now] 区间拉取落库 docker_event_log（7 天留存），水位推进、离线恢复自动补拉（窗口上限 1h）；排障链：RFC3339Nano→unix 秒→**根因 v28 SDK 在 daemon 关流时经 errCh 发 io.EOF 且不关 msgCh，原消费把 EOF 当错误丢弃 batch 不推水位**（本地 SSH 隧道复现+远端空窗对照定位），重写消费（EOF=正常收尾落库+推水位，超时已收事件仍落库不推水位）；期间发现并行会话 k8s 前端半成品致部署 npm build 失败、修复版曾未上机（去重 deleteK8sCluster、集群函数转出口）；真机：ops-ev5 容器 create/start/kill/stop/die/destroy 6 事件全捕获。**C1 全部完成**）
- [x] k8s：集群注册（kubeconfig 加密）+ 连接测试 + 连接池；集群总览（metrics-server 用量）；Node 管理（cordon/drain）（场35 集群注册全链交付；场53 补齐：集群总览（版本/节点/命名空间/Pod 统计+metrics-server 汇总用量，metrics 不可用降级）、Node 详情（allocatable/capacity/conditions/taints）、cordon/uncordon patch、drain 先隔离后 Eviction（跳过 DaemonSet/镜像 Pod/删除中，纯函数决策单测）；clientset 按集群即时构造（无长连接池，多集群低频管理面量级足够）；测试机 k3s 真集群全链验证：注册/总览（88 核/125.7GiB/用量真实）/cordon→SchedulingDisabled→uncordon/drain 驱逐 5 Pod、替代 Pod Pending 于隔离节点；错误码 1505）
- [x] k8s：工作负载（列表/YAML/扩缩容/滚动重启）；Pod（列表/详情/日志/WebShell/删除）；ConfigMap/Secret(脱敏)/PVC/Service/Ingress/Event（分场交付：场47 Services/ConfigMaps/Secrets 只读、场49 Deployment 扩缩容/滚动重启、场54 补齐——StatefulSet/DaemonSet 列表、工作负载 YAML 查看（deployment/statefulset/daemonset，剔 managedFields，sigs.k8s.io/yaml）、Pod 详情（容器状态 containerStateText 纯函数+条件+相关事件 fieldSelector 匹配倒序）、Pod 删除（888 写操作+二次确认）、PVC/Ingress/Event 列表（事件倒序限 200）；错误码 1508/1509（修正 NodeNotFound 与 1505 撞号）；前端资源浏览新增 5 页签+YAML 抽屉+Pod 详情抽屉；k3s 真机全验证：sts 1/1、PVC Pending/local-path、Ingress host/path 解析、YAML 内容正确且 managedFields 已剔、非法 kind 拦截、Pod 详情 phase/qos/容器/事件全对、不存在 Pod 精确报错、删除独立 Pod 后集群确认消失。**剩余：Pod WebShell（多容器 subprotocol，seal 路线）与 YAML 编辑下发（diff 预览确认）**）
- [ ] 两套真实环境联调 + 管理员/普通用户两级数据隔离验证

### M5 任务清单（gpu）

- [x] gpu_node/gpu_spec CRUD 与定价上架（gpu_image 随 C2 镜像库联动后补；场50 交付：三表迁移、节点/规格/实例十接口+菜单/API/casbin 种子（888 全量、9528 只读）、前端算力管理三页签（节点余量展示/规格定价/实例开通销毁）；错误码 1601-1606）
- [ ] 实例开通（进行中：场50 交付防超卖引擎——事务内 FOR UPDATE 行锁节点 + 运行中实例占用聚合 + canAllocate 纯函数判定（余量不足报 1605 带明细），销毁释放配额复用、更新总量不可低于占用、离线节点拒开通；真机：2 卡节点开通 2 次 1 卡后第三次精确拒"GPU 余 0 需 1"、销毁后配额循环复用；剩余：Docker 创建（DeviceRequest GPU 直通）联动 container 插件——需真实 GPU 环境）
- [ ] HAMi 显存切分注入（环境变量组）；实例生命周期操作 + 状态自动同步
- [ ] 实例监控（docker stats + GPU 指标）+ SSH 跳板机（2026 端口，选连自己的容器）+ 端口转发管理

### M6 任务清单（dbops + monitor）

- [ ] dbops（进行中：场45 交付可做部分——dbops_instance（密码 AES-GCM 密文复用 asset/crypto，独立请求体保证响应永不回显）/dbops_order 两表、实例 CRUD+TCP 探活回写（SQL 层检测待 mysql 驱动引入）、SQL 工单创建/取消/分页（普通用户仅本人）、审核接口 goInception 未配置明确报 1708 不动状态（接入点已留）；九接口+实例/工单双页+种子（错误码 1701-1708）；真机：密文不泄露（SecretNotLeaked）、3306 探活离线回写、工单流转、审核未配置报错；剩余：goInception 审核/执行/备份、soar 优化建议、多环境——待 goInception 二进制与 MySQL 环境）
- [x] monitor：SSH 采集任务（CPU/内存/磁盘/网络）+ 指标留存与自动清理 + ECharts 图表页（场38 交付：monitor_metric 表（asset+name+ts 复合索引、30 天留存随轮清理）、单命令组合采样 1s 完成（loadavg/cpu 两次 /proc/stat delta 含除零保护/meminfo/df，解析纯函数 3 组单测）、CollectAll 并发 5 全量采集（绑凭据非报废主机，复用 asset SSH 通道指纹校验）、@every 5m 注册底座 timer、/monitor/metric/{list,collect}+种子（888/9528 只读）；前端 PerfChart（ECharts 时间轴四折线双 Y 轴、1h/6h/24h/7d、立即采集）挂主机页「监控」抽屉；真机：disk 25% 与宿主 df 精准一致、load/mem 合理、CPU 0% 为 88 核空闲机真实值；网络指标场51 补齐（net/dev tail 跳表头修复 + ParsePerfOutput 特征扫描解析，真机流量验证））
- [x] monitor：告警规则引擎（阈值/持续时间/静默窗口）+ 钉钉机器人推送 + 端口探活（场39：monitor_alert_rule（metric 阈值>/< 连续 N 次、port TCP 探活、静默默认 30min、钉钉 webhook 选配）+ event（触发/恢复/通知状态）；评估挂采集轮末尾统一执行、evaluateMetric 纯函数单测、触发静默去重、恢复自动关闭未决事件、pushDingTalk 仅 http(s) 5s 超时；五接口+前端规则/事件双页签；真机：mem<99 触发含当前值、端口不可达触发、二轮采集静默不重复、非法指标名拦截；修复告警两表漏注册迁移与摘要重复主机名两处）

### M7 任务清单（workflow + org + aiops）

- [x] workflow：状态机定义/实例/审批任务 API 与页面（场40 交付：三表 wf_definition（states/transitions JSON+起始）/wf_instance（定义名快照+业务参数）/wf_action_log；引擎 ValidateDefinition（引用闭合/重名/起始，4 组单测）+NextState 迁移推导、审批节点动作受限（approve/reject）、终态自动收档（完成/驳回）、cancel 撤回、迁移表外动作拦截；八接口+菜单/API/casbin 种子（普通用户仅自己工单）；前端工单中心（发起/通过/驳回/撤回+流转时间线、定义管理 JSON 模板）；真机全链：定义→发起→submit→approve 归档已完成、非法迁移/审批动作/结束后操作三类拦截全部生效。模板三类留作业务侧预设定义，随发版闭环场次补）
- [ ] org：钉钉扫码登录（JWT 打通）、部门/用户定时同步、机器人告警复用
- [ ] aiops（进行中：场49 交付 MCP 运维只读工具三件——ops_asset_overview / ops_alert_recent / ops_ticket_status，基于 GVA mcp 骨架（StreamableHTTP :8889 standalone，cmd/mcp），init 自动注册；数据经骨架上游代理调主 server API（鉴权透传，standalone 无 DB 勿直查）；真机 MCP 协议全验：initialize→tools/list（3 新工具在列）→tools/call 返回真实数据（5 资产/2 告警/3 工单）；坑：列表接口均为分页 POST，GET 调用命中 gin 404（'404 page not found' 中 404 先解析为 JSON 数字，报错表现为 'p' after top-level value）；剩余：AI 诊断网关（统一 LLM 调用、密钥管理、prompt 模板）——待 LLM API 密钥）

### 开发会话执行协议（每场 2 小时，完成即提交）

自动化任务与人工开发均按此节奏执行：

1. **开场（5 分钟）**：读 DEV_PLAN、`git log --oneline -20`、`git status`；从当前里程碑任务清单取未勾选项作为本场范围，按 M0→M1→…→M9 顺序推进，先收尾上一场未完成项。
2. **开发（约 100 分钟）**：小步提交，一个功能单元一个 commit（conventional commits，中文描述）；后端每步 `go build ./...` 且 `go vet ./...` 干净，关键逻辑带单测；前端改动跑 `npm run build` 验证。**main 任何时刻保持可编译**，禁止把编译不过的状态推送上去。
3. **收尾（15 分钟）**：勾选 DEV_PLAN 完成项 → 在「开发日志」表追加一行（日期/场次/完成/下一步）→ `git commit` → `git push origin main`（失败重试 1 次，仍失败则报告原始错误留待人工）。**每场结束必须完成至少一次成功的 commit+push，硬性要求。**
4. **未完成项**：任务清单复选框旁标注"进行中：<断点说明>"，下一场从断点接续；宁小勿烂，不追求一场做完整个里程碑。
5. **每场完成定义（DoD）**：代码 + 编译通过 + 关键单测 + Swagger 注释 + 菜单/API/casbin 初始化脚本（随插件 migration 提交）+ DEV_PLAN 勾选与开发日志更新 + push 成功。

### 测试环境（192.168.112.138）

所有开发产出的更新与验证都在此机器进行，**每场收尾在 push 后执行 `bash scripts/deploy-test.sh`**，冒烟不通过视为本场未完成。

| 项 | 值 |
|---|---|
| 机器 | Ubuntu 22.04.5 x86_64，125G 内存，Docker 29 + Compose v5，SSH 免密（`ssh root@192.168.112.138`） |
| 入口 | 前端 http://192.168.112.138:8081 （nginx:alpine 容器 `new-ops-web`，host 网络）；后端 :8888（systemd `new-ops-server`） |
| 目录 | `/opt/new-ops/{bin,web,data,logs,resource,config.yaml}`；SQLite 库 `data/new_ops.db`（无 MySQL/Redis 依赖） |
| 部署 | `bash scripts/deploy-test.sh` 一键部署：全新安装与增量更新同一命令，自动完成构建→上传→systemd/nginx 安装→SQLite 初始化→冒烟；支持环境变量覆盖目标机器/端口/目录/服务名（详见脚本头部），可一键部署到任何 Ubuntu+Docker 机器。`config.yaml` 为有状态文件（含初始化信息），更新模式不覆盖 |
| 帐号 | 管理员 `admin`（密码记录在本地运维记忆，不入仓库）；测试机配置 `open-captcha: 999999`（等效关闭验证码，便于自动化冒烟） |
| 端口约定 | 8080 被机器上其他容器占用；new-ops web 固定 8081、server 固定 8888 |
| k3s 集群 | 场53 安装 k3s v1.36.5+k3s1 单节点（`localhost`/192.168.112.138:6443，禁用 traefik/servicelb 避免端口冲突）；kubeconfig `/etc/rancher/k3s/k3s.yaml`；`/etc/rancher/k3s/registries.yaml` 配 docker.io 镜像加速（同机 Docker daemon.json 同源 mirrors，containerd 直连 docker.io 会超时）；内置 metrics-server；已在平台内注册为集群 `k3s-test` 供 K1/K2/K3 联调 |

---

## 五、关键设计决策

1. **凭据保险库**：独立表存密文，主密钥从配置/环境变量注入；接口层强制脱敏（只回显末 4 位）；引用方（term/job/pipeline）凭 ID 取用，服务端内存解密即用即毁。
2. **WebSSH 会话审计**：WebSocket 消息镜像落库（命令级正则抽取 + 全量异步录像文件），审计员角色只读回放——直接对齐 go-webssh 的三角色模型。
3. **流水线引擎**：移植 new-jenkins 的三层数据模型与状态机，触发快照保证定义修改不影响进行中的构建；执行器初期仍在本机（与 new-jenkins 现状一致），M6 后接入 Agent 执行器实现工作空间隔离——这是 new-jenkins README 已声明的边界，沿用其演进路线。
4. **GPU 防超卖**：节点可用量 = 物理规格 - Σ(运行实例规格)，分配时行级锁 + 事务；HAMi 虚拟显存以 `CUDA_DEVICE_MEMORY_LIMIT` 注入，节点侧标注是否支持切分。
5. **实时通道**：底座提供 `/ws/*`（xterm、日志 tail、容器 exec）与 `/sse/*`（流水线日志）统一网关与鉴权中间件，业务插件只注册业务 channel，避免每个插件自拉 websocket。
6. **兼容旧平台**：提供只读的数据导入脚本（旧库→新 CMDB），旧仓库 README 顶部统一加"已并入 new-ops"声明与跳转。
7. **质量门禁与 CI**：GitHub Actions 在每次 push 时跑 server（go build/vet/test）与 web（build）；配合第四节的会话执行协议，main 恒为可编译状态；每插件核心 service 层单测覆盖目标 ≥60%（mock global.DB，沿 GVA service 模式）；Swagger 注释随代码同步更新。

---

## 六、专项开发计划（Agent / Docker 管理 / K8s 管理）

三个专项的关系：**Agent 是 Docker/K8s 管理的第二种接入形态**（SSH/Docker API/K8s API 直连之外），也是指标采集与远程执行的统一通道；Docker 节点与 K8s 集群均挂在 asset 资产体系之下，凭据统一走凭据保险库。

### 6.1 Agent（独立二进制 `agent/` + 底座 Agent 网关）

**定位**：单文件 Go 二进制（go:embed 内嵌资源、无外部依赖，quan-agent 验证过的模式），Linux x86_64/arm64 优先，Windows 次之。

**通信设计**：
- 出站长连接（WebSocket，复用底座实时网关与鉴权中间件）——解决 NAT/防火墙后节点平台不可达的核心痛点；
- 认证：注册 Token 绑定资产 ID，TLS 必选、mTLS 可选；主机公钥指纹校验（accept-new，quan-agent 模式）；
- 四类消息：注册 → 心跳（30s）→ 任务下发/结果回传 → 指标上报；断线重连 <10s，消息幂等（任务 ID 去重）。

**安全红线**：默认只读 Skill 集（系统概览/CPU 负载/内存/磁盘 inode/systemd 异常/TCP-UDP 监听/Docker 状态，直接对齐 quan-agent 七个 Skill）；执行类任务按资产组授权并全量审计；自动升级需签名校验。

**分期**：
| 阶段 | 内容 | 前置 |
|---|---|---|
| **A1** | 注册/心跳/系统指标采集（CPU/内存/磁盘/网络/负载），资产页展示 Agent 在线状态；monitor 的性能采集从分钟级 SSH 主动采集升级为秒级上报 | M9 |
| **A2** | 执行代理：job 插件批量命令/脚本支持 SSH / Agent 双执行后端可选；文件分发 | M9 |
| **A3** | 流水线远程执行器：Pipeline step 可指定在资产/Agent 主机上执行，工作空间按构建隔离——解决 new-jenkins README 声明的"执行器跑在服务进程主机、无工作空间隔离"边界 | M9 |
| **A4** | 容器/K8s Agent 模式：代理所在主机 Docker socket 与集群内 ServiceAccount（一键 Deployment YAML），作为 C3/K4 的内网纳管通道 | M8 后 |

**数据模型**：`agent_instance`(asset_id, version, os/arch, labels, status, last_heartbeat, token_hash)、`agent_task`(下发/结果/超时/重试)；server 侧新增 Agent 网关模块（连接注册表 + 消息路由），不新建插件。

**验收**：单实例 server 承载 100 Agent 并发心跳与指标上报；Agent 崩溃/断网不影响 server；升级幂等可回滚。

### 6.2 Docker 管理（插件 container，分期 C1-C3）

**功能来源**：ai-devops（容器/镜像/网络/卷/exec 终端）、GPU 系列（TLS 接入点/巡检）、docker-gpu-manage（端口转发）、DockerGPU（状态回写）。

**C1（M4，随一期交付）**
- [ ] 接入点管理：endpoint CRUD（unix socket / TCP+TLS），TLS 证书走凭据保险库，定时连通巡检（30s 合并巡检，tianqi 模式）
- [ ] 容器全生命周期：列表/详情/启动/停止/重启/删除（联动数据卷清理）；创建容器支持镜像、端口映射、环境变量、挂载、CPU/内存限制、重启策略
- [ ] 日志流（tail -f over WebSocket）与 exec 交互终端（exec + hijack 桥接 WebSocket，窗口自适应，ai-devops 已验证）
- [ ] 状态一致性：操作后 inspect 回写（DockerGPU 模式）+ docker events 事件订阅实时同步

**C2（M8）**
- [x] 镜像管理：列表/拉取（进度流）/删除/tag/导入导出，与 gpu 镜像库联动（场42 列表/异步拉取/删除+前端抽屉；场44 补齐 tag/导出 tar 流式下载/导入上传+前端打标弹窗与导入入口；真机闭环：alpine 打标→导出 8.7MB tar→删除→tar 导入 Loaded image 回归确认。gpu 镜像库联动待 M5 后）
- [ ] 网络与卷（进行中：场43 交付——network 创建（bridge+子网 IPAM，net.ParseCIDR 校验）/列表（builtIn 标记）/删除（内置拒绝）、volume 列表与删除（占用 daemon 拒绝回传）+前端网络卷双页签抽屉；真机：172.30.100.0/24 创建成功、非法子网拦截、内置拒绝、105 卷真实列表；剩余：端口转发规则管理（docker-gpu-manage 模式））
- [ ] 资源统计（进行中：场43 交付 ContainerStats one-shot——CPU delta×在线核数/内存去 inactive_file/网络块IO 汇总/pids，容器抽屉「统计」弹窗；真机 busybox mem 0.94MB/128G、pids 1 合理；剩余：接入 monitor 历史图表化）
- [ ] Compose：compose 文件上传与校验、up/down/ps

**C3（M8+）**
- [ ] 模板化一键部署：应用模板（compose + 参数渲染）
- [ ] Agent 模式接入（联动 A4）：内网 Docker 节点纳管
- [ ] 节点池统一：Docker 节点为 asset 资产子类型，GPU 节点 = 带显卡标签的 Docker 节点——避免 GPU 系列历史上多套节点表的问题

**数据模型**：`docker_endpoint`、`docker_container`（关联 asset 主机）；镜像/网络/卷实时查询不落库，仅缓存。

**验收**：真实 3 节点（含 1 台 TLS）纳管；容器写操作全部进审计；exec 终端 30 分钟不断流。

### 6.3 K8s 管理（插件 k8s，分期 K1-K4）

**功能来源**：seal（Pod WebSSH）、ai-devops（接入+诊断）、docker-gpu-manage（多集群/工作负载/三级 RBAC）、autoops（资源概览）。

**K1（M4）**
- [ ] 集群注册：kubeconfig 上传（AES-256-GCM 落库、凭据保险库）、连接测试、连接池与超时管理（docker-gpu-manage 模式）
- [ ] 集群总览：版本/API Server/节点数/资源用量（metrics-server）、Namespace 列表
- [ ] Node 管理：列表/详情（allocatable/capacity/conditions/taints）、cordon/uncordon/drain

**K2（M4）**
- [ ] 工作负载：Deployment/StatefulSet/DaemonSet 列表/详情/YAML 查看、扩缩容、滚动重启、YAML 编辑下发（diff 预览确认）
- [ ] Pod：多 namespace 列表/详情（容器状态+事件）/日志（实时流+下载）/WebShell（多容器 subprotocol 选择，seal 路线）/删除
- [ ] 配置资源：ConfigMap/Secret（值脱敏展示）/PVC/Service/Ingress/Event

**K3（M8）**
- [ ] Helm：chart 仓库管理、release 安装/升级/回滚/卸载（values 表单 + YAML 双模式）
- [ ] 三级 RBAC 落地：全局/集群/命名空间授权模型（casbin 资源规则扩展），普通用户仅见授权 namespace
- [ ] AI 诊断：Pod 异常状态+事件+日志片段 → LLM 分析与修复建议（docker-gpu-manage 已验证，经 aiops 模型网关）
- [ ] GPU 资源视图：节点 GPU 型号与分配情况（DevicePlugin 指标），与 gpu 插件联动

**K4（M8+，可选）**
- [ ] Agent 模式纳管（联动 A4）：内网集群经 k8s-agent 反向连接
- [ ] 应用商店（YAML/CRD 模板库）、对接 K8s audit webhook

**数据模型**：`k8s_cluster`(name, kubeconfig_cipher, labels, status)、namespace 授权表（cluster_id, namespace, user/role）；资源对象实时查询不落库。

**验收**：两套真实集群纳管（含 1 个内网集群走 K4）；Pod WebShell/日志流 30 分钟稳定；删除类接口全部二次确认 + 审计。

---

## 七、目录结构规划（新增部分）

```
new-ops/
├── agent/                       # M9 专项：轻量 Go Agent（独立二进制，反向连接 server）
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

**单个插件的内部结构模板**（以 asset 为例，其余插件同构，开发会话照此落位）：

```shell
server/plugin/asset/
├── plugin.go            # 插件注册入口（实现 GVA Plugin 接口）
├── initialize/          # 路由注册 + 菜单/API/casbin 初始化数据（migration）
├── api/v1/              # 接口层（参数绑定与响应，不含业务）
├── router/              # 路由分组
├── service/             # 业务逻辑（单元测试写在这里，mock global.DB）
├── model/               # model / request / response
└── config/              # 插件级配置结构（对接 config.yaml）
web/src/plugin/asset/
├── view/                # 页面组件（与动态菜单路由对应）
├── api/                 # 后端接口封装（axios）
└── router/              # 静态兜底路由（正常走数据库动态菜单）
```

---

## 八、风险与对策

| 风险 | 对策 |
|---|---|
| 上游 GVA main 演进导致合并冲突 | 底座零修改原则；业务全在 plugin/ 目录；每季度合并一次上游 tag |
| 单人维护 15+ 插件摊子过大 | 严格按里程碑交付，P3/远期插件不提前开工；每个插件 M 完成即发布 tag |
| WebSSH/终端类功能的安全责任重 | 凭据加密+审计回放+指纹校验三项安全红线不裁剪；README 明示安全边界（沿袭 new-jenkins 做法） |
| SSH/Docker/K8s 直连在大规模下不稳 | M1-M5 直连上限明确写文档；Agent 通道按第六节分期落地（M9 执行器 + A4 Agent 模式纳管内网节点） |
| 旧平台数据结构差异大 | 只迁移核心资产表，历史数据留旧库存档，新平台不背历史包袱 |

---

## 九、决议记录

- 2026-09-30：确定以 gin-vue-admin main（e8d675c）为底座新建仓库 `new-ops`，本计划为第一版；功能盘点基于本人 GitHub 各仓库 README（当日版本）。
- 2026-09-30（v1.1）：新增第六节「Agent / Docker 管理 / K8s 管理」专项开发计划；里程碑扩展至 M9（合计约 22-23 周）；目录结构增加独立 `agent/` 二进制工程。
- 2026-09-30（v1.2）：深度优化——补齐 M0/M2/M3 任务清单；新增「开发会话执行协议」（每场 2 小时完成即提交推送、main 恒可编译）与「开发日志」；新增 3.4 核心数据模型汇总与关键设计 7 质量门禁/CI；自动化任务提示词与协议对齐。
- 2026-09-30（v1.3）：继续细化——新增「十四场排期表」（逐场目标，10-01~10-07）；补齐 M4-M7 任务清单；新增 3.5 技术选型与依赖清单（新增依赖须登记）、3.6 接口与错误码规范（插件错误码分段）；第七节补充单插件内部目录模板。
- 2026-09-30（v1.4）：加密到每天 3 场（9/14/20 点，7 天 21 场 42h）；排期表重排为二十一场，目标提升为 M0+M1 全部、M2 大部分；自动化任务 cron/maxRuns 同步更新（约 15 周完成全计划）。
- 2026-09-30（v1.5）：新增「测试环境」章节——192.168.112.138 首次部署完成（SQLite+systemd+nginx:8081），协议收尾步骤增加部署冒烟，新增 scripts/deploy-test.sh。
- 2026-09-30（v1.6）：平台中文名定为「白泽 BaiZe」；前端（站点名/logo/登录页/仪表盘/favicon）、后端横幅、文档全面品牌化；前端彻底移除 GVA 痕迹（插件市场/推广组件/gva-* 类名）。

---

## 十、开发日志

> 每场开发结束由执行方追加一行（自动化任务收尾步骤含此动作）。

| 日期 | 场次 | 完成 | 下一步 |
|---|---|---|---|
| 2026-09-30 | 人工 | DEV_PLAN v1.1→v1.2 深度优化定稿；自动化任务配置并与协议对齐 | 10-01 上午场：M0 开工（品牌化与 compose/CI） |
| 2026-09-30 | 人工 | DEV_PLAN v1.2→v1.3 细化：十四场排期表、M4-M7 清单、依赖/API 规范、插件目录模板 | 不变：10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | DEV_PLAN v1.3→v1.4：加密为每天 3 场（9/14/20 点），排期表重排二十一场（42h） | 不变：10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | 全局去品牌化更名 new-ops（module 路径/前端/配置，build+vet 通过）；测试环境 192.168.112.138 首次部署并冒烟通过；deploy-test.sh 固化部署流程 | 10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | 前端彻底去 GVA：移除插件市场（页面/菜单/代理/API）、仪表盘换简洁欢迎页、删 about/官方外链、gva-* 类名改 ops-*、AI 组件更名 ai-ops；测试库旧菜单已清，部署验证通过 | 10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | 品牌化：中文名定为「白泽 BaiZe」，自制 SVG logo、站点名/登录页/仪表盘/横幅/文档全面应用 | 10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | deploy-test.sh 升级为一键部署：全新安装（自动建库+随机密码）与增量更新同命令，参数化可部署任意机器；双路径实测通过 | 10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | 强制移除底部技术支持标识/全站水印/授权引导（bottomInfo 删除、watermark 摘除、购买授权文案清理）；仪表盘新增「关于版权标识与授权」展示卡片 | 10-01 上午场 M0 开工 |
| 2026-09-30 | 人工 | **M0 全量完成**：example 移除（文件管理保留）、10 插件骨架注册、CI、Makefile、config 分环境、插件开发规范；修复 utils/ast 测试污染源文件问题（-short 隔离） | 10-01 上午场 M1 开工（凭据保险库 + SSH 采集） |
| 2026-09-30 | 人工 | 根除 vite-auto-import-svg 构建注入（登录页"此模板未授权"红标/回传 beacon/keywords 指纹/域名校验）：本地 svgBuilder 替代，依赖卸载，chunk 前缀改 baize-；远端验证徽标消失 | 10-01 上午场 M1 开工 |
| 2026-09-30 | 人工 | 登录页重设计：白泽品牌深色渐变背景+网格+光斑，居中毛玻璃卡片（浏览器截图验证） | 10-01 上午场 M1 开工 |
| 2026-10-01 | 12 | M1 场3：asset 插件后端全链路（四表模型/CRUD/keyword+status+room 过滤分页/IP唯一与格式校验+单测/Swagger/菜单-API-casbin-888绑定种子）+ 前端主机管理页；测试环境全链路验证通过（CRUD/非法IP与重复IP拦截/菜单出现） | 下一场：机房与产品线页面、Excel 导入导出 |
| 2026-10-01 | 13 | M1 场4：机房/机柜/产品线全链路（service CRUD+校验+单测、12 个接口+路由、菜单/API/casbin 种子扩展、机房与机柜双页签页+产品线页）；远端验证：机房重名与删除保护拦截、机柜按机房过滤、产品线关联主机/详情回读/按产品线过滤 | 下一场：Excel 导入导出、资产变更历史 | 
| 2026-10-01 | 14 | M1 场5：主机变更历史（表/事务记录/分页接口/时间线抽屉）+ Excel 导入导出（excelize 导出 xlsx、按内网IP upsert 导入、逐行校验失败明细）；远端验证：历史 3 条动作序列正确（创建→更新→更新）、操作人 admin、导出 xlsx 有效、导入 upsert 2 条成功 | 下一场：数据权限（资产组+casbin）、IP 复合唯一索引修复 | 
| 2026-10-01 | 15 | M1 场6：数据权限全链路——资产组三表模型/服务（组-主机-用户关联重写）/四接口/前端管理页（成员+主机多选）；主机 list/export 按 888 全量、普通用户仅组内过滤；9528 只读 casbin 策略；IP 软删除唯一索引冲突修复；双角色远端验证通过（未分组0台→分组后仅perm-a→9528写拦截） | 下一场：Go SSH 采集器（资产现场回填）、凭据保险库 | 
| 2026-10-02 | 17 | M1 场7：Go SSH 采集器全链路——解析纯函数单测、SSH 双认证拨号（密码/私钥）、命令级超时、凭据保险库 GetPlaintext 取用、真实 sshd 采集回填（ubuntu/88C/125G/229G）+ last_collect_at + 历史记录；前端「采集」按钮（凭据下拉）+ 最近采集列；df 解析 MOTD 噪音修复 | 下一场：阿里云 ECS 同步；M1 收尾后进 M2 终端 |
| 2026-10-02 | 18 | M1 场9：阿里云 ECS 同步全链路——OpenAPI RPC V1 签名自实现（零新增依赖）、DescribeInstances 分页拉取、按 SN(InstanceId) upsert（状态映射/空IP兜底公网/历史记录）、前端导入弹窗（AK凭据+Region+结果明细）；链路验证：假AK真实外呼回传 InvalidAccessKeyId、空地域与凭据类型校验；修复种子路径不一致（/asset/sync→/asset/host/sync 导致 888 被拦） | 下一场：M1 收官自查（软删除唯一索引类问题巡检）→ M2 终端与作业开工 |
| 2026-10-03 | 19 | M2 开工（场11提前）：term 插件 WS⇄SSH PTY 桥接全链路——/term/ws 握手 token 自验（复用底座 JWT）、二进制数据流+JSON 控制帧（resize/ping）、xterm 组件（重连提示/自适应/心跳）、主机页「终端」抽屉；gorilla/websocket+xterm 6.0 依赖登记 3.5；真实 sshd 验证：连接/命令回显/resize 全通过 | 下一场：ProxyJump 级联+指纹校验、会话审计落库+回放 |
| 2026-10-03 | 20 | M2 场20：会话审计全链路（三表/桥接集成/查询接口/回放页）——端到端验证通过（WS 会话流镜像含回显、命令抽取、会话结束态）；协作合并并行会话的指纹 TOFU 校验与 ProxyJump jump.go；另补：webssh 级联接线（dialChain 返回目标指纹入会话快照）、buildJumpChain 6 组单测、主机表单跳板机选择/指纹回显、CRUD 层 ssh_fp 防篡改与 jump_host_id 取消写入修复。⚠️ 检测到并行会话同时开发本仓库，git 提交交错，建议收敛为单会话 | 下一场：级联双机真机验证、SFTP 文件浏览器 |
| 2026-10-03 | 21 | 排期收官场：SFTP 文件浏览器全链路交付（六接口+前端浏览器页+真实 sftp 验证 list/upload/download/rename 通过；递归删除验证因并行部署窗口未完成，下一场回归）；README/plugin-dev-guide 定稿；tag v0.1.0 发布 | 下一周期（10-04 起）：M2 收尾（批量执行/日志tail/网段发现）→ M3 流水线；⚠️ 并行会话冲突待用户收敛 |
| 2026-10-03 | 22 | M2 场22（job 插件开工）：批量命令执行全链——两表模型、信号量并发池（限流/单任务超时/批次取消，4 组单测）、数据权限+逐主机凭据解析+指纹 TOFU、cancelRegistry、四接口+种子、前端执行页（轮询/取消/结果抽屉）；真机验证全过：批次 269ms 成功、sleep60 取消生效、SFTP 递归删除与 upload 回归通过（此前 upload 报错系本地 GitBash curl 路径转换污染 path 字段，非服务端缺陷）、ProxyJump 级联双机验证通过（目标机 Last login from 127.0.0.1 证实经跳板隧道）；核实 tag v0.1.0 远端在（本地未 fetch） | 下一场：脚本库 CRUD+版本、变量组关联资产；远程日志 tail、CIDR 网段发现 |
| 2026-10-03 | 23 | M2 场23：脚本库与变量组全链（并行会话半成品 model/service 接手补全）——三表迁移、9 接口+种子（9528 只读）、脚本更新版本递增归档、变量组 key=value 前端编辑器；批量执行联动（scriptId 脚本内容为命令+variableGroupId 服务端渲染 {{key}} 入批次快照）；错误码分段统一（脚本 1201-1207/执行 1211-1214）；单测 validate×2+RenderTemplate；真机端到端：v1→v2 归档正确、{{greeting}}→baize 渲染执行输出一致 | 下一场：M2 收官（远程日志 tail、CIDR 网段发现）→ M3 流水线开工 |
| 2026-10-04 | 24 | **M2 收官（场24）**：远程日志 tail（/term/logtail WS 复用级联拨号+指纹校验、tail -F 防选项注入、LogTail 组件滚动跟随/截断）+ CIDR 网段发现（ExpandCIDR 上限保护 6 场景单测、并发 SSH banner 探测、按 IP 导入跳过已存在）；真机验证全过：/29 段发现本机 OpenSSH 横幅、非法 CIDR 1006 拦截、导入 1 台/重复跳过、tail 初始回看+实时追加全到达。**M2 全部完成**（场次 19-24：WebSSH/级联/审计/SFTP/批量执行/脚本库变量组/日志 tail/网段发现） | 下一场：M3 流水线开工（pipeline/stage/step 三层模型+CRUD） |
| 2026-10-04 | 25 | M3 开工（场25）：pipeline 三层模型与 CRUD——三表迁移、嵌套校验纯函数 8 场景单测、五接口+种子（888 全量/9528 只读、错误码 1301-1304）、前端列表+编排弹窗（阶段卡片审批 gate/失败继续+步骤 shell/http）；真机嵌套 CRUD 验证（Sort 归一/审批位/http 默认 GET/级联删除）；修复 Create 走 replaceStages（Omit 关联跳过阶段落库，真机回读发现） | 下一场：执行器（构建状态机/变量替换/审批 gate）+ SSE 日志网关 |
| 2026-10-04 | 23 | job 错误类型统一（newScriptErr/scriptError 重复定义移除，复用 errors.go 的 JobError；错误码接续 1205-1211；测试断言 errors.As 化）——修复与并行会话协作中的类型冲突；SFTP 递归删除回归验证被并行部署窗口反复打断（404 窗口），标记进行中下一场回归 | 下一场：SFTP 递归删除回归 + 部署窗口错峰验证；M3 流水线由并行会话推进中，本会话避免抢主线 |
| 2026-10-04 | 26 | M2 增强：批量执行按主机资产组自动注入变量——variableGroupsByAssetGroupID 索引（同组取最新）+ renderForHost 决策（显式组>资产组绑定>原样，单测5组）+ runBatch 逐主机渲染 + JobExecResult.Command 每主机实际命令快照 + 前端变量组可留空（自动注入模式提示）；真机：批次创建成功，结果查询撞并行部署窗口待回归 | 下一场：自动注入真机结果回归；配合并行会话 M3 主线 |
| 2026-10-04 | 27 | M3 场27（执行器核心）：pipeline_build/build_log 两表 + CreateBuild 定义快照（步骤 {{key}} 参数替换入快照，改定义不影响进行中构建）+ 后台状态机（等待中→执行中→等待审批→成功/失败/已取消）+ 审批 gate（approveCh 放行）+ continueOnError + 构建序号递增；安全设计：shell 步骤一律 SSH 到绑定资产执行（无本机 exec 面，Mimosa 拦截后采纳更优方案，对齐 M3「发布目标绑定资产」）；构建五接口+种子+前端构建抽屉（参数触发/3s 轮询/放行/取消/分级日志）；真机四链路全过：参数渲染成功输出、exit7 失败（修复 session.Wait 退出码误用真 bug）、等待审批→放行→成功、执行中→取消→已取消；验证数据已清理 | 下一场：SSE 实时日志端点、并行阶段、构建历史重跑/取消收尾 |
| 2026-10-04 | 28 | M3 场28（SSE 实时日志）：GET /sse/pipeline/build/logs（public 组 query token 自验、GetBuildLogsAfter lastId 1s tick 增量、status/done 事件收流、X-Accel-Buffering no）+ 前端进行中构建 EventSource 订阅（done 自动刷新、异常降级拉取）；真机验证：drip 步骤逐秒输出，事件序列 status→5×log→status→done 全到达，验证数据已清理。M3 进度：CRUD✓ 执行器核心✓ SSE✓；剩余：并行阶段、触发器 cron/webhook、工单闭环、构建历史重跑 | 下一场：并行阶段（stage 内 step 并发组）与触发器（手动已可用，补 cron/webhook） |
| 2026-10-04 | 29 | M3 场29（触发器+历史）：webhook 触发全链（public 组 /pipeline/webhook/:token，令牌即凭据 uuid 生成/404 语义防枚举、params 渲染快照、编排弹窗开关+地址复制）+ 构建重跑（restart 复用原参数对当前定义生成新快照，前端已结束构建「重跑」按钮）；真机验证：错误令牌拒绝、无登录态触发 HOOK=CICD operator=webhook、重跑 buildNo 递归递增 operator=admin，数据已清理。构建历史项整体勾选；触发器剩 cron（需打通 plugin-tool timer 入口） | 下一场：cron 触发接入底座定时任务、并行阶段；M3 收官后过工单闭环 |
| 2026-10-04 | 30 | M3 场30（cron 触发）：SyncPipelineCron 幂等同步 global.GVA_Timer（robfig 标准 5 段/@every、RemoveTaskByName→AddTaskByFunc）挂接 Create/Update/Delete + 启动 initialize.Timer 全量恢复 + 前端开关/表达式输入；真机验证：@every 10s 25s 自动 2 次构建 operator=cron、删除流水线任务即摘不再触发，数据已清理。触发器项整体勾选（手动/webhook/cron）| 下一场：并行阶段；工单发版闭环依赖 M7 workflow 按里程碑顺延 |
| 2026-10-05 | 31 | **M3 场31（并行阶段，执行器收官）**：PipelineStage.parallel 位 + runStageSteps 串行/并发双路径（信号量上限 10、首败 cancel 快速中断兄弟步骤、ContinueOnError 跑完收集）+ 前端「步骤并发」checkbox；修复两处真机暴露缺陷：并发快速失败未传导 cancel（run1 误传外层 ctx）、取消路径双重收档竞态；真机验证：并行双步日志交错成功、exit3 快速失败后 sleep6 兄弟步骤被中断无输出、失败正确记录，数据已清理。**执行器项整体勾选；M3 仅剩工单发版闭环（依赖 M7 workflow 顺延）** | 下一场：M4 开工（container C1：endpoint CRUD+TLS 凭据入保险库+30s 巡检）；工单闭环等 M7 |
| 2026-10-05 | 32 | **M4 开工（场32，container C1）**：Docker 接入点管理全链——SDK v28.0.4 登记 3.5（go-connections 钉 v0.5.0 修 Windows 编译缺 DialPipe）、docker_endpoint 模型、NewDockerClient（unix 直连/TCP+TLS 凭据保险库 docker_tls 解密 JSON{ca,cert,key}）、PingEndpoint+30s 合并巡检循环（并发5）、五接口+菜单/API/casbin 种子（错误码 1401-1406）、前端接入点页；真机验证：unix socket 巡检在线 API 1.54、离线节点回写、重名拦截；演示数据保留 local-docker | 下一场：容器列表/生命周期（启停删+创建参数）→ 日志流+exec 终端 |
| 2026-10-05 | 33 | M4 场33（容器列表/生命周期）：ListContainers 实时查询（不落库，端口映射格式化）+ ContainerAction start/stop/restart/remove(force) + 离线接入点拒绝写操作 + 两接口+种子 + 前端容器抽屉（状态标签/启停删/删除二次确认）；真机验证：列出测试机真实容器（new-ops-web 等）、busybox 测试容器 stop→start→remove 宿主机确认无残留 | 下一场：创建容器（端口/挂载/环境变量/资源限制）→ 日志流+exec 终端（C1 收尾） |
| 2026-10-05 | 34 | M4 场34（创建容器，C1 收尾）：POST /container/container——ParsePortBinding 纯函数 + nat.PortSet/PortMap 端口映射 + NanoCPUs/Memory 资源限制 + RestartPolicy 四档 + StartNow；前端创建弹窗（多行端口/环境/挂载、CPU/内存、启动命令、立即启动）；真机验证：inspect 核对 mem=128M/cpu=1e9/rp=always/env 全对、端口占用错误正常回传"容器已创建但启动失败"、换空闲端口 create+start 成功 running 且端口映射正确。**C1 生命周期与创建参数项完成** | 下一场：容器日志流 + exec 终端（WS 桥接）+ inspect 回写/docker events 订阅 |
| 2026-10-05 | 35 | SFTP 递归删除回归通过（断点关闭）⚠️ 并复盘确认：历次"404 部署窗口"均为验证脚本 HTTP method 推断错误（gin 对方法不匹配返回 404），服务与部署从未异常；已建立按路由表精确映射的验证脚本规范 | 下一场：M4 k8s K1（集群注册/总览）或 container C1 剩余（日志流/exec 终端），以并行会话最新进度为准 |
| 2026-10-05 | 35 | M4 K1 起步：k8s 插件集群注册全链——client-go v0.33.3（3.5 登记）、K8sCluster 模型（kubeconfig AES-GCM 密文 json:'-'）、注册（解析+连接测试+加密落库，测试失败记离线不阻断）/删除/列表/TestCluster 四接口、K8s 管理菜单+集群页、casbin 888 全部+9528 只读；真机验证：不可达集群离线注册成功、密文不回显、TestCluster 报 dial 不可达、9528 列表放行。待办：9528 create 的 casbin_rule 残留巡检（脚本实测走到 handler）；集群删除验证脚本 method 笔误待复跑 | 下一场：集群列表补充 ServerVersion 实时刷新；K2 工作负载（与并行会话协调分工）|
| 2026-10-05 | 36 | M4 场36（C1 收官双 WS 通道）：LogsWS（docker logs -f，TTY 判断+stdcopy 去复用，tail 200-5000）/ ExecWS（exec /bin/sh TTY → hijack ⇄ WS，resize 经 ContainerExecResize）挂 public 组 query token 自验；前端 ContainerShell（xterm 心跳自适应）/ContainerLogs（跟随截断）+容器抽屉「日志/终端」按钮；修复 readControl 丢弃用户输入真 bug（exec 无回显真机暴露，重构为 (ctl,data) 双返回）；真机：日志流跟随连续 tick、exec echo 双向打通。**C1 仅剩 inspect 回写+events 订阅** | 下一场：docker events 订阅+inspect 回写；K2 工作负载由并行会话推进 |
| 2026-10-05 | 37 | M4 场37（events 落库，**C1 全部完成**）：docker_event_log 表+30s 巡检区间拉取（水位 last_event_at、离线补拉、窗口上限 1h、7 天留存）+事件列表接口+前端事件抽屉；排障三连：RFC3339Nano→unix 秒→真根因 **v28 SDK 在 daemon 关流时经 errCh 发 io.EOF 且不关 msgCh，原消费把 EOF 当错误丢弃已收事件且不推水位**（本地 SSH 隧道最小复现+远端空窗对照定位）；连带发现并行会话 k8s 前端半成品曾致部署 build 失败（修复版一度未上机）——去重 deleteK8sCluster 修复；真机：ops-ev5 容器 6 生命周期事件全捕获。⚠️ 教训：部署输出勿 grep 过滤（曾掩盖 build 失败） | 下一场：monitor 插件开工（SSH 性能采集+指标留存+图表）或配合 K2；C2 镜像/网络/卷在 M8 |
| 2026-10-05 | 38 | M6 开工（场38，monitor SSH 性能采集）：monitor_metric 表+单命令组合采样（1s：loadavg/cpu delta/mem/df，解析纯函数 3 组单测）+CollectAll 并发 5（复用 asset SSH 通道）+@every 5m 底座 timer+查询/手动采集两接口+种子；前端 PerfChart（ECharts 四折线双 Y 轴/时段切换/立即采集）挂主机页「监控」抽屉；真机：disk 25% 与宿主 df 一致、load/mem 合理、CPU 0% 为空闲机真实值。**monitor 首项勾选（网络指标留待告警场次顺带）** | 下一场：告警规则引擎+钉钉推送+端口探活（monitor 第二项）；dbops 需 goInception 二进制待环境 |
| 2026-10-05 | 39 | M6 场39（告警引擎）：metric 阈值（>/< 连续 N 次纯函数单测）+port TCP 探活+静默去重+恢复关闭+钉钉文本推送（仅 http(s)）；评估挂采集轮末尾；五接口+前端规则/事件双页签；真机：mem<99 触发、端口不可达触发、二轮静默不重复、非法指标拦截；修复告警两表漏注册 AutoMigrate 与摘要重复主机名。**monitor 三项全部完成（M6 monitor 部分 ✓，dbops 待 goInception 环境）** | 下一场：M7 workflow 工单引擎开工（状态机定义/实例/审批 API+页面）或配合 k8s K2 |
| 2026-10-05 | 40 | M7 开工（场40，workflow 工单引擎）：三表+状态机引擎（ValidateDefinition 4 组单测/NextState 推导、审批节点动作受限、终态自动收档、cancel 撤回）+八接口+种子+前端工单中心（发起/审批/驳回/撤回+流转时间线+定义管理）；真机全链：定义→发起→submit→approve 归档、三类非法操作（迁移外动作/审批节点非审批动作/结束后操作）全部拦截，T2 驻 submitted 拒 reject 证明迁移表约束生效。**M7 workflow 首项勾选** | 下一场：工单发版闭环（审批通过→事件钩子触发流水线，补 M3 末项）；org 钉钉登录待企业配置 |
| 2026-10-05 | 41 | M3 末项收官（场41，工单发版闭环）：release 工单审批完成 fireReleaseHook → pipeline CreateBuild（params {pipelineId,params}、触发结果落 hook 流转行、驳回不触发、钩子失败不阻断收档）；真机：工单 #3 approve→构建 #1 自动产生 operator=工单#3 状态成功。**M3 六项全部完成 ✓✓**；org/aiops 待企业配置与 LLM 密钥 | 下一场：C2 镜像管理（列表/拉取/删除）或 k8s K2 配合；M5 GPU 待显卡环境 |
| 2026-10-05 | 42 | C2 提前（场42，镜像管理）：ListImages 实时/PullImage 异步（内存状态表+重复拦截）/RemoveImage force+四接口+种子+前端镜像抽屉（2s 轮询完成自动刷新）；真机：alpine 拉取 拉取中→成功 6s、列表出现 8MB、删除宿主无残留。附带：并行会话 k8s workload 半成品随本场上车（编译通过）；tag/导入导出/gpu 联动留后续 | 下一场：C2 网络与卷/资源统计，或 k8s K2 收尾配合；下会话从 git log 最新进度接续 |
| 2026-10-06 | 43 | C2 场43（网络与卷+资源统计）：network 创建（bridge+子网 IPAM）/列表（builtIn 标记）/删除（内置拒绝）、volume 列表/删除、ContainerStats one-shot（CPU delta×在线核数/内存去 inactive_file）——八接口+种子+前端网络卷双页签抽屉与统计弹窗；真机：172.30.100.0/24 创建成功、非法子网拦截、内置拒绝、105 卷真实列表、busybox stats 合理（0.94MB/pids1）| 下一场：C2 端口转发规则/Compose；dbops 待环境；k8s K2 并行会话推进 |
| 2026-10-06 | 44 | C2 场44（镜像收尾件）：TagImage/SaveImage 流式 tar 导出/LoadImage 上传导入+三接口/种子+前端打标弹窗/导出下载/导入上传；真机闭环：alpine 打标→导出 8.7MB tar→删除→tar 导入 Loaded image 回归。**C2 镜像管理项全部完成**（端口转发规则管理留待 M5 跳板场景顺带；gpu 联动待 M5）。随后全量回归（build/vet/全插件单测/前后端构建）与 tag v0.2.0 发布 | tag v0.2.0：M0-M3 全完+M4 C1/C2 镜像网络卷统计+M6 monitor+M7 workflow 引擎 |
| 2026-10-05 | 37 | M4 K2 收尾：集群页资源浏览抽屉集成（Pods/Deployments/Nodes 三页签 + Pod 日志抽屉）+ 正确 method 下 SFTP 递归删除回归通过；⚠️ 关键复盘：历次"404 部署窗口"均系验证脚本 HTTP method 推断错误（gin 对方法不匹配返回 404），服务与部署从未异常；顺手修复并行会话进行中代码的 docker client v28 ImageLoad API 编译错误（functional options） | 下一场：k8s 资源浏览真集群联调（需可用 kubeconfig）或 M4 C1 收尾项（由 container 主线会话推进）|
| 2026-10-06 | 47 | k8s K2 补充：Services/ConfigMaps/Secrets 只读列表（后端三接口+路由+种子+资源浏览三页签；Secret 值永不回显仅列键名）；自动注入真机回归通过（批次6 每主机 Command=echo baize-auto-inject-done，注入闭环）；正确 method 下历次 404 全部复现排除——验证脚本规范已立 | 下一场：k8s 资源浏览真集群联调（待用户提供 kubeconfig）或继续 K2 写操作（Scale/滚动重启）|
| 2026-10-06 | 48 | M6 dbops 可做部分（场48）：MySQL 实例纳管（密码 AES-GCM 密文复用 asset/crypto、独立请求体防回显、TCP 探活回写、未结束工单拒删）+ SQL 工单（创建/取消/分页、审核接口 goInception 未配置明确报 1708 不动状态、接入点已留）——九接口+实例/工单双页+种子（错误码 1701-1708）；真机：密文不泄露、3306 探活离线、工单流转、审核未配置报错全过 | 下一场：aiops MCP 工具注册（M7 末项可做部分）或 M5 GPU 骨架（防超卖纯逻辑）；goInception/MySQL/钉钉/LLM 环境就绪后接对应断点 |
| 2026-10-06 | 49 | M7 aiops 可做部分（场49）：MCP 运维只读工具三件（ops_asset_overview/ops_alert_recent/ops_ticket_status）注册进 GVA mcp 骨架（StreamableHTTP standalone，init 自动注册；数据经骨架上游代理调主 server API、鉴权透传，standalone 无 DB 禁直查）；真机 MCP 协议全验：initialize→tools/list→tools/call 三工具均返回真实数据（5 资产/2 告警/3 工单）；排障：列表接口均为 POST，GET 命中 gin 404（'404' 先解析为 JSON 数字致 'p' after top-level value 假象）——与并行会话"404=method 推断错误"复盘互证 | 下一场：M5 GPU 骨架（防超卖纯逻辑+单测，无环境也可交付）；LLM 密钥就绪后接 AI 诊断网关 |
| 2026-10-07 | 50 | M5 GPU 骨架（场50，无环境可交付部分）：三表（节点/规格定价/实例分配台账）+ 防超卖引擎——开通事务内 FOR UPDATE 行锁节点 + 运行中实例占用聚合 + canAllocate 纯函数判定（余量不足报 1605 带明细），销毁释放配额复用、总量不可低于占用、离线节点拒开通；十接口+菜单/API/casbin 种子+前端算力管理三页签；真机：2 卡节点开通 2 次 1 卡满载 → 第三次精确拒"GPU 余 0 需 1" → 销毁释放 → 配额循环复用全过，验证数据已清理。**至此全部里程碑的可无环境交付部分均已落地** | 剩余项均待环境：GPU Docker 直通（显卡）/goInception+MySQL/钉钉/LLM 密钥/真集群 kubeconfig；C2 端口转发留 M5 跳板场景；Compose 待设计 |
| 2026-10-06 | 49 | M4 K2 写操作：Deployment 扩缩容（replicas 0-500 校验+单测）与滚动重启（restartedAt 注解 patch）——service/写接口（仅 888 casbin）/前端 ScaleDialog + Deployments 操作列按钮；写操作走 GVA 操作日志中间件 | 下一场：k8s 资源浏览真集群联调（待 kubeconfig）；M5 GPU 或 M6 继续推进 |
| 2026-10-06 | 25 | M2 收官回归（method 修正后一次全通）：SFTP 递归删除（mkdir→递归删→已删净）；workflow 全链（定义列表→发起工单→流转记录）；自动注入端到端（批次6 每主机 Command=echo baize-auto-inject-done）。确认并行会话已交付 workflow 全链（服务/迁移/种子/前端 ticket 页）——我方重复的 state_machine 模型已删除避免双写 | 下一周期：M3 执行器增强、M7 工单模板扩充（发版/SQL/资源申请联动）由并行会话推进；M5 GPU 剩余需真实 GPU 环境 |
| 2026-10-07 | 53 | **M4 K1 收官（场53，k3s 真集群解锁）**：测试机安装 k3s v1.36.5+k3s1 单节点（禁 traefik/servicelb；registries.yaml 配 docker.io 加速——k3s 内嵌 containerd 直连超时致系统 Pod 卡 ContainerCreating，复用同机 Docker mirrors 后恢复）；K1 补齐——集群总览（版本/节点/命名空间/Pod 统计+metrics-server 用量，不可用降级 metricsAvailable=false）、Node 详情（allocatable/capacity/conditions/taints）、cordon/uncordon（patch spec.unschedulable）、drain（先隔离后 policy/v1 Eviction，drainDecisions 纯函数跳过 DaemonSet/镜像 Pod/删除中，单测 3 组）；k8s.io/metrics v0.33.3 登记 3.5；casbin/api 种子补齐资源浏览路径缺口+node 写操作仅 888；前端总览抽屉+Nodes 操作列（详情抽屉/隔离恢复/驱逐二次确认）；真机全链：注册 k3s-test 在线→总览 88 核/125.7GiB/用量真实（CPU 0.14 核 mem 6.4GiB）→cordon→SchedulingDisabled→uncordon→drain 驱逐 5 Pod（替代 Pod Pending 于隔离节点）→恢复调度全过，验证数据已清理。场首收尾 Mimosa 扫描硬拦截：修复 5 项（MD5V→SHA256V 全链/exec 绝对路径校验/删除零引用 openDocument.js/componentName 目录遍历边界校验/package.json 脚本清理）；⚠️ 并行会话误 git add 全量 .mimosa 快照（8.7 万行）入库且捎带我暂存文件，已移出仓库+gitignore+还原其致编译不过的 gorm_biz 遗留改动，历史垃圾保留在已推送提交中 | 下一场：K2 剩余（Pod 详情/事件/WebShell/删除、StatefulSet/DaemonSet、YAML 查看、PVC/Ingress/Event）——k3s 真集群已可联调；或 C2 Compose；K1 收官后 M4 仅剩"两套真实环境联调"（k3s 已占一套） |
| 2026-10-07 | 54 | M4 K2 补齐（场54，k3s 联调第二场）：StatefulSet/DaemonSet 列表、工作负载 YAML 查看（deployment/statefulset/daemonset，剔 managedFields 噪音，sigs.k8s.io/yaml v1.6.0 转直接依赖登记 3.5）、Pod 详情（容器状态 containerStateText 纯函数 4 场景单测+条件+相关事件 involvedObject fieldSelector 倒序）、Pod 删除（888 写操作+二次确认）、PVC/Ingress/Event 列表（事件倒序限 200）；错误码修正 NodeNotFound 与 ReplicasInvalid 1505 撞号（挪 1508、新增 KindUnsupported 1509）；前端资源浏览新增 StatefulSets/DaemonSets/PVC/Ingress/Events 五页签+YAML 抽屉+Pod 详情抽屉+Pod 删除按钮；真机（k3s-test）：sts k2-test-sts 1/1、PVC Pending/local-path 精确、Ingress host/path→svc 解析、sts 与 coredns YAML 正确且 managedFields 已剔、非法 kind"job"拦截、Pod 详情 phase=Running/qos=BestEffort/容器就绪/3 条事件、不存在 Pod 精确报错、独立 Pod k2-test-pod API 删除后集群确认消失，验证数据已清理。顺带：修正 container 首项陈旧勾选（生命周期/创建参数实为场33/34 已交付）；pathInfo.json 补上场遗漏 OverviewDrawer 映射 | K2 仅剩 Pod WebShell（多容器 subprotocol）与 YAML 编辑下发（diff 预览）；下一场：K2 收官（WebShell 优先）或 M8 K3 Helm/RBAC 起步；M4"两套真实环境联调"待第二套集群 |
