# new-ops · 统一运维开发平台

> 🚧 **正在开发中** —— 项目处于早期开发阶段，功能尚未交付，接口与目录结构可能频繁调整，暂不可用于生产环境。
>
> 开发总计划（功能盘点 / 架构设计 / 里程碑任务清单 / 逐场排期 / 开发日志）：[docs/DEV_PLAN.md](docs/DEV_PLAN.md)

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)（main, e8d675c）构建的**插件化统一运维开发平台**：把作者历史开源的 autoops、chain、cmdb、seal、go-webssh、raptor、new-jenkins、GPU 算力系列等 20+ 个运维项目的功能，收敛到一个底座上持续演进，避免多套技术栈、多套权限体系的重复维护。

## 技术栈

| 端 | 技术 |
|---|---|
| 后端 | Go 1.24 · Gin · GORM · Casbin v3 · JWT · Zap |
| 前端 | Vue 3.5 · Vite 8 · Element Plus · Pinia |
| 实时通道 | WebSocket（终端/日志）· SSE（流水线日志） |
| 架构原则 | 业务全部以 GVA 插件形式开发（`server/plugin/<name>`），底座零修改，可持续跟随上游升级 |

## 基座能力（gin-vue-admin 自带 ✅）

- ✅ 用户 / 角色 / 菜单 / API 管理（Casbin RBAC + JWT）
- ✅ 代码生成器、表单设计器
- ✅ 插件机制（公告 / 邮件）、定时任务
- ✅ 操作日志、多云对象存储上传、Swagger 文档、MCP Server 骨架

## 功能规划（🚧 均为待开发，按里程碑排序）

### M1 资产中心（插件 `asset`）

- 📋 主机资产管理：机房 / 机柜 / 产品线 / 负责人 / 状态，Excel 导入导出
- 📋 凭据保险库：SSH 密码/私钥、云 AccessKey、Docker TLS、kubeconfig 统一 AES-256-GCM 加密，接口永不回显明文
- 📋 资产采集：Go SSH 采集、阿里云 ECS 定时同步、Agent 上报（远期）
- 📋 数据权限：资产组 + casbin 资源规则，普通用户仅见授权资产
- 📋 资产变更历史、仪表盘统计

### M2 终端与作业（插件 `term` / `job`）

- 📋 WebSSH：浏览器终端（xterm.js + WebSocket）、ProxyJump 级联（≤5 层）、主机公钥指纹校验
- 📋 会话审计：命令记录 + 全量录像 + 审计回放
- 📋 SFTP 文件管理器
- 📋 批量命令 / 脚本执行（shell / python / yml），脚本库版本管理、变量组
- 📋 远程日志 tail、CIDR 网段自动发现并导入资产

### M3 流水线与发布（插件 `pipeline`）

- 📋 声明式流水线 Pipeline → Stage → Step（HTTP / Shell 步骤、并行阶段、失败继续）
- 📋 人工审批 gate、参数体系与变量替换、cron / webhook 触发
- 📋 SSE 实时日志推送、构建取消 / 复用参数重跑
- 📋 工单发版：审批通过自动触发流水线

### M4 容器与 K8s（插件 `container` / `k8s`）

- 📋 Docker：多节点纳管（TCP + TLS）、容器全生命周期、日志流、exec 交互终端、镜像/网络/卷管理、Compose、端口转发
- 📋 K8s：多集群注册（kubeconfig 加密）、Node / 工作负载（扩缩容、滚动重启、YAML 下发）、Pod 日志 / WebShell、Helm 安装升级回滚、全局/集群/命名空间三级 RBAC、AI 故障诊断

### M5 GPU 算力（插件 `gpu`）

- 📋 算力节点纳管、镜像库（上架/下架/显存切分标记）、产品规格与定价
- 📋 GPU 容器实例全生命周期、按规格智能匹配调度（资源核算防超卖）、HAMi 显存切分
- 📋 实例资源监控、SSH 跳板机、端口转发管理

### M6 数据库与监控（插件 `dbops` / `monitor`）

- 📋 MySQL 实例与账号纳管、SQL 上线工单（goInception 审核/执行/备份、soar 优化建议）
- 📋 指标采集与图表、告警规则引擎（阈值/持续时间/静默）、钉钉机器人推送、端口探活

### M7 工单与协同（插件 `workflow` / `org`）

- 📋 通用工单引擎（Workflow → State → Transition 状态机、审批节点、事件钩子）
- 📋 钉钉扫码登录、部门/用户定时同步

### M8-M9 AI 与 Agent（插件 `aiops` + 独立二进制 `agent/`）

- 📋 AI 诊断网关、只读巡检 Skill 集、MCP Server 工具集
- 📋 轻量 Go Agent：反向 WebSocket 长连接、系统指标上报、执行代理（SSH 之外的第二执行通道）、流水线远程执行器（工作空间隔离）
- 📋 Agent 模式纳管：内网 Docker 节点 / 内网 K8s 集群经 Agent 反向接入

### 远期

- 📋 PXE / IPMI / Redfish 装机与 IP 地址池（pcfarm-admin 迁移）
- 📋 PCDN 边缘节点与带宽管理
- 📋 运维知识库、GPU 模型训练平台（SFT/DPO/CPT）、ComfyUI 多卡调度

## 开发节奏

- **2026-10-01 ～ 10-07 密集开发期**：每天 9:00 / 14:00 / 20:00 三场自动化开发（每场 2 小时），完成即提交推送，7 天 21 场共 42 小时。
- 7 天目标：**M0 底座就绪 + M1 资产中心全部 + M2 终端作业大部分**，发布 `v0.1.0`。
- 之后按里程碑 M3→M9 滚动推进，全计划约 15 周；每场进展记录在 [DEV_PLAN 开发日志](docs/DEV_PLAN.md)。

## 快速开始（底座部分，随 M0 完善中）

```bash
# 后端（默认 :8888）
cd server && go mod tidy && go run main.go
# 前端（默认 :8080）
cd web && npm install && npm run serve
```

数据库初始化走 GVA 引导页；完整部署（docker-compose / CI）随 M0 里程碑落地。

## License

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)（Apache License 2.0）构建，本项目沿用该协议。
