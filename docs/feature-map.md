# 白泽 new-ops · 功能地图与分类总览

> 生成于 2026-11-12 全项目功能审计之后。审计口径：逐插件比对 router 注册、initialize API 记录、Casbin 授权三者一致性；核对错误码分段无冲突。
> 数据速览：24 个插件 · 约 750+ 路由 · 10 段错误码 · 6 大业务域。

---

## 一、功能分类体系（六域）

```
白泽 new-ops
├── A. 平台底座（GVA 上游，不改业务逻辑）
├── B. 基础设施运维域（资产/终端/作业/容器/K8s/GPU/数据库/监控/流水线）
├── C. 大模型平台域（llmops 五件套 + agent/aiops）
├── D. 教学考试域（exam / oplib / announcement / email / auto）
├── E. 千学六插件·业务域（artsci / fanhua / ziqiang / eduadmin / crossauth / policyreg）
└── F. web 前端（按插件分域的页签式工作台）
```

## 二、分域清单

### A. 平台底座
GVA 上游能力：JWT 认证、Casbin v3 鉴权、操作日志中间件、统一响应 `{code,data,msg}`、分页、Swagger、定时任务、多云上传。**约束：插件不修改底座业务逻辑**，扩展仅通过插件 v2 接口自注册。

### B. 基础设施运维域
| 插件 | 职责 | 路由 | 亮点 |
|---|---|---|---|
| asset | 资产/凭据保险库 | 33 | AES-256-GCM 主密钥（环境变量注入），全项目凭据密文落库统一入口 |
| term | WebSSH 终端/SFTP | 11 | WebSocket 终端 + 会话回放 + SFTP 文件管理 |
| job | 分布式定时任务 | 13 | cron 注册/执行记录 |
| pipeline | CI 流水线 | 13 | 构建+Webhook+SSE 日志 |
| container | 容器管理 | 39 | 生命周期/统计/WS 日志与终端 |
| k8s | 集群管理 | 44 | 多集群/工作负载/cordon·drain/事件 |
| gpu | 算力节点 | 11 | GPU 节点/实例+CSV 导出 |
| dbops | 数据库运维 | 13 | 备份/巡检/工单 |
| monitor | 主机监控告警 | 8 | 指标采集/规则/事件导出 |
| workflow | 工作流引擎 | 9 | 状态机编排 |

### C. 大模型平台域
| 插件 | 职责 | 路由 | 亮点 |
|---|---|---|---|
| llmops | 大模型运营五件套 | 143 | 模型/知识库/训练/智能体/网关 chat（全项目 chat 统一出口）|
| agent | 主机代理 | —（通道型） | 任务下发/文件推送/令牌 |
| aiops | 智能运维 | 7 | 告警降噪/根因 |

### D. 教学考试与效率域
| 插件 | 职责 | 路由 | 亮点 |
|---|---|---|---|
| exam | 在线考试 | 40 | 判分/等级/CSV 导出（I07 口径源头）/考试终端 |
| oplib | 运维知识库 | 9 | 脚本库+Python 实执行 |
| announcement | 公告 | 8（官方插件） | 信息发布 |
| email | 邮件 | 2 | SMTP 发送 |
| auto | 代码生成器/AI 工作流/MCP | 56 | 插件脚手架/skills 打包 |

### E. 千学六插件·业务域（本冲刺交付，路由-API-Casbin 100% 一致）
| 插件 | 需求域 | 路由 | 错误码 | 功能分类 |
|---|---|---|---|---|
| artsci | 3.3 学科工具集 | 19 | 2500-2599 | 科学计算（Allen-Cahn/DeepHPMs/DeepONet）·教学工具（批改/组卷/文生图/图像解读）·工具台 |
| fanhua | 3.4 文科大模型 | 51 | 2600-2699 | 语料工程（导入/清洗/分词/CSV 导出）·标注（文本/图片）·三大知识库（海派/价值观/城市）·问答与审查 |
| ziqiang | 3.5 理科大模型 | 37 | 2700-2799 | 鉴权语料（文本/图像+aHash）·法律工具（术语/隐私/分词/抄袭）·图谱·知识库·区块链存证·侵权问答 |
| eduadmin | 3.6 教育管理 | 100 | 2800-2899 | 数据接入（清洗/脱敏分级）·六大知识库·八类智能体工具台（招聘/绩效/预测/预算/报销/预警/资产/采购/能耗/食堂）·驾驶舱·AI-12345 工单全链路（分派/督办/闭环/报表）·跨校协同·同步/身份/政策拉取·审计语料 |
| crossauth | 3.7 认证与安全 | 69 | 2900-2999 | 联邦身份·互认·MFA（TOTP）·SSO·权限（模板/双模式/四级矩阵/生命周期）·数据分级·SM4+字段加密·访问告警·算力防护·五维监控·三级审计·联合运维·应急预案 |
| policyreg | 3.8 制度大模型 | 21 | 3000-3099 | 制度导入（密级脱敏）·清洗·分词属性·实体·标注·角色鉴权·问答 |

### F. web 前端
按插件分域的工作台：`web/src/plugin/<name>/view/*`（动态菜单挂载）+ `api/<name>.js` 封装。本轮六插件工作台：学科工具台 / 繁花工作台 / 自强工作台 / 教育管理工作台+AI-12345 工单台 / 跨校安全工作台 / 制度工作台。

## 三、整合约定（全项目统一口径）

1. **插件六处接线**：plugin.go → initialize.{Gorm, Api, Menu, Casbin, Router} + router/enter.go；路由-API 记录-Casbin 三者必须一致（本轮审计修复 4 个存量缺口，六个新插件 100%）。
2. **错误码分段独占**：每插件一段（见 E 表），通用错误沿用 GVA 7xxx；审计确认无冲突。
3. **凭据安全**：一律 `asset/crypto` AES-GCM 密文落库、接口不回显明文（asset/eduadmin/crossauth 已对齐）。
4. **chat 统一出口**：依赖大模型的功能一律走 llmops 网关 + `chatFn` 变量注入锚点 + 未配置模型规则口径 fallback + 假模型单测（exam J03 → fanhua N13 → policyreg R08 同口径）。
5. **幂等种子**：菜单/API/Casbin/业务种子全部 FirstOrCreate/按名幂等，重启无损。
6. **CSV 导出口径**：strings.Builder + 中文表头 + 纯函数（exam I07 → eduadmin P29 → fanhua corpus 同口径）。
7. **相似度/分词自实现**：bigram（N05/O05/R04）、编辑相似度（N03）、aHash+汉明（O03/O10）、滑窗抄袭（O05）均纯标准库，零第三方依赖。

## 四、本轮整合审计记录（2026-11-12）

| 检查项 | 结果 | 动作 |
|---|---|---|
| 六新插件路由/API/Casbin 一致性 | ✅ 100%（artsci 19、fanhua 51、ziqiang 37、eduadmin 100、crossauth 69、policyreg 21） | — |
| fanhua corpus/export 登记缺失 | ❌ 扩充轮遗漏 | ✅ 补 API+Casbin |
| gpu/instance/export 登记缺失 | ❌ 存量 | ✅ 补 API+Casbin |
| monitor/alert/event/export 登记缺失 | ❌ 存量 | ✅ 补 API+Casbin |
| term sftp ×6 登记缺失 | ❌ 存量 | ✅ 补 API+Casbin |
| 错误码分段冲突 | ✅ 无（仅常量误报） | — |
| WS/Webhook/SSE 特殊路由（agent/container/exam 终端/pipeline/term ws/llmops） | 非 private 组或通道型注册，不属于本审计口径 | 保留现状 |
