# 白泽（BaiZe）冲刺计划 — 2026-10-23 晚 ～ 10-26 09:00

> 本文档是 DEV_PLAN.md 的窗口期冲刺附件；开发会话执行协议、依赖清单（3.5）、错误码规范等一律以 DEV_PLAN 为准。
> 节奏：定时任务每 30 分钟触发一场（2026-10-23 深夜起，至 2026-10-26 09:00 自动停止，约 115 场）。
> 每场 = DEV_PLAN「开发会话执行协议」的一次执行：开场对账 → 小步开发 → 收尾提交推送 + 部署冒烟。
> 两阶段推进：阶段一「开发推进」全部勾选后，自动转入阶段二「优化巡检」（从头逐模块检查是否可以优化）。

## 〇、第一场特殊任务：DEV_PLAN 对账

场 60（10-08）之后已有 4 个功能提交，但 DEV_PLAN 勾选与开发日志未同步，第一场先补记再开工：

| 提交 | 内容 | 待补 |
|---|---|---|
| ba19056 | M9 A1 Agent 通道（agent/ 独立二进制 + 服务端 agent 插件） | M9 清单 A1 勾选 + 开发日志 |
| f755b96 | dbops goInception 审核引擎（审核/执行/备份回滚） | M6 dbops 勾选更新 + 开发日志 |
| 9193340 | M7 aiops LLM 网关 + Pod AI 诊断（M8 K3 项同步交付） | M7 aiops / K3 AI 诊断勾选 + 开发日志 |
| 89514f4 | M8 C3 应用模板化一键部署（compose 模板 + {{参数}} 渲染） | C3 首项勾选 + 开发日志 |

对账口径：以 `git show` 逐提交核实实际交付边界，宁少勾不多勾；场次号从 61 起接续。

## 一、阶段一：开发推进（按序取项，全部完成后进入阶段二）

- [ ] E2E 补验（进行中：本地全链功能 E2E 已过 场67——本机起全新 SQLite server+darwin agent 真进程跑通 7 断言：task/exec 三态、文件分发（嵌套目录+0600+内容一致）、job execBackend=agent 批次、流水线 agent 步骤+工作空间 pwd 实证；剩 deploy-test.sh 部署冒烟与 C3 模板化部署全链，待测试机 192.168.112.138 可达后立即做）。
- [ ] M9 A2 执行代理：job 批量命令/脚本支持 SSH / Agent 双执行后端（可选下发），Agent 消息幂等（任务 ID 去重）与超时重试；文件分发。（进行中：双后端+幂等+超时+文件分发通道及 job 页分发卡片已交付 场62-65，剩真机 E2E）
- [ ] M9 A3 流水线远程执行器：Pipeline step 可指定在资产/Agent 主机执行，工作空间按构建隔离（对齐 new-jenkins README 声明边界的升级）。（已交付 场66，剩真机 E2E，并入全局 E2E 补验项）
- [ ] A4 + C3/K4 联动：Agent 模式纳管——代理所在主机 Docker socket 与集群内 ServiceAccount（一键 Deployment YAML），作为内网 Docker 节点/K8s 集群接入通道；节点池统一设计（Docker 节点为 asset 资产子类型）。（进行中：Docker 侧 场69/70、K8s 集群后端 场71、前端+一键 YAML 场72 已交付；剩 k3s 集群内 daemon 级 E2E 与节点池统一设计）
- [ ] dbops 收尾：soar 优化建议接入审核链（测试机 Docker 可拉镜像时）、SQL 工单多环境配置。
- [x] M9 验收项：单实例 100 Agent 并发心跳/指标上报压测脚本 + 结果记录进开发日志；断线重连幂等回归。（完成 场68——agent/loadtest 本机实测 100/100 注册、45s 300 次心跳、结束 100 在线、零失败，server 日志零 panic；断线退避重连为 A1 已验+任务号幂等单测覆盖；测试机 canonical 部署后可复跑）
- [ ] 收官：若 M9 A2/A3 完成，发布 tag v0.3.0 并更新 README 里程碑口径。

说明：GPU 显卡直通/HAMi、钉钉企业配置为环境阻塞项，本窗口不做；每场只推进一个小单元，宁小勿烂。

## 二、阶段二：优化巡检（阶段一清单全部勾选后触发）

从头开始逐模块检查「是否可以优化」，循环进行直至窗口结束：

1. 顺序：asset → term → job → pipeline → container → k8s → monitor → workflow → dbops → gpu → aiops → agent → web 前端各插件页 → docs/scripts。
2. 每场聚焦一个模块：通读 service/api 关键代码，从六个维度列发现——性能（N+1 查询/索引/并发度）、安全（越权/注入/脱敏）、健壮性（错误处理/超时/资源泄漏）、重复代码、单测缺口、交互体验；挑最有价值的 1-2 项当场实施。
3. 每项优化独立 commit（`refactor(module)` / `perf(module)` / `fix(module)`，中文描述）；行为变化必须真机回归，纯重构不得改行为。
4. 发现与实施记入 DEV_PLAN 开发日志（标注「优化」）；一个模块巡检完成即在本文档对应处记录轮次，进入下一模块；全部模块走完为一轮，可开第二轮。
5. 进度记录：**asset 已巡检（场73）**——GetAssetGroupList 每组 2 次 Pluck 的 N+1 → 两次批量 SELECT + 纯函数内存聚合（单测 2 组），1+2G 次查询 → 3 次；DeleteAssetHostByIds 逐台 4 次查询 → 一次载入 + 批量清关联/软删/记历史（软删与不存在自动排除，语义等价）；凭据加解密/引用计数/数据权限 JOIN 巡检无问题；批量删除 DB 侧行为回归随测试机恢复补验。
6. 进度记录：**term 已巡检（场74）**——修两处 WS 桥接真实缺陷：①写超时 deadline 设在 WriteMessage 之后（本次写入无超时，客户端卡死永久阻塞输出协程）→ 提前到写入前；②输出协程与读循环 pong 应答并发写 WS（gorilla 禁止，容器 exec/PodShell 同类问题已修、term 漏）→ 加 wsWrite 单写者互斥（数据帧/pong/关闭帧统一走锁+超时）；logtail 同步修 deadline 顺序（其本身单写者安全）；-race 下 term 单测全过；待办优化：审计流镜像逐条同步落库在输出热路径（阻塞终端写 DB），后续场次改有界异步批量刷写（行为变化需真机回归）。
7. 进度记录：**job 已巡检（场75）**——CreateBatchExec 主机快照逐台 GetAssetHost（含 Room/Rack/ProductLines 无谓 Preload）→ 一次批量查询+缺失精确报错（语义等价）；自动变量注入逐主机 Pluck 组映射的 N+1 → 批量预载 groupsByHost + 纯函数 renderCommandForHost（单测 1 组：多组首个命中/未命中/无组）；-race 全过。结果逐条 INSERT 保持原样（受主机数上限约束，收益有限）。
8. 进度记录：**pipeline 已巡检（场76）**——修两处日志链路缺陷：①appendLog 字节硬切 16KB 会切出非法 UTF-8 存库（SSH 输出超限时日志乱码）→ truncateLogRunes 字符边界截断（单测：CJK 边界/ASCII/标记）；②runHTTPStep 用单次 Read 读响应体快照（允许短读，日志常为空/不完整）→ readBodySnippet 完整读取上限字节（io.ReadAll+LimitReader）；执行器并发/审批 gate/取消收档逻辑审查无问题（-race 过）。
9. 进度记录：**container 已巡检（场77）**——LogsWS/ExecWS 与 term 同病的两处修复：①日志流输出写入无任何超时（客户端卡死永久阻塞输出协程）→ wsSingleWriter 统一写入前设 10s 超时；②读循环 pong 与输出协程二进制帧**并发写 WS**（gorilla 禁止）→ 单写者互斥（场74 term 同口径）；统计采样/事件水位/生命周期/端口转发审查无问题（agent 模式接入点自动兼容管理面采样）；-race 全过。下一模块：k8s。
10. 进度记录：**k8s 已巡检（场78）**——模块质量整体良好：PodShell 写互斥（k8sWsIO）场55 已做对；GuardNamespace 覆盖核验完整（Pod 日志/YAML/Pod 详情/Helm 历史四端点+agent 集群守卫）；namespaceVisible 出错 fail-closed；scopedList 普通用户多授权 ns 串行扇出为已知低价值点（授权量人工控制，不改）；实施一处加固：k8sWsIO.writeJSON（pong）补写超时（与数据帧同口径，防对端卡死阻塞 stdin 泵）；-race 全过。下一模块：monitor。
11. 进度记录：**monitor 已巡检（场79）**——实施一处优化：EvalAllRules 端口探活规则从串行改有界并发评估（上限 10）——原 N 条不可达端口把采集轮拖长 N×2s；解析纯函数/采集轮并发 5/静默去重/恢复关闭/留存清理/agent 心跳指标复用链路审查无问题；记录功能缺口：钉钉机器人加签方式未支持（仅自定义关键词 webhook），待环境后按需补；-race 全过。下一模块：workflow。下一模块：pipeline。

## 三、每场 DoD（硬性）

1. `go build ./...` + `go vet ./...` 干净；前端改动 `npm run build` 通过；**main 任何时刻可编译**。
2. commit（conventional commits 中文）+ `git push origin main`（失败重试 1 次，仍失败报告原始错误留待人工）。
3. DEV_PLAN 勾选完成项 + 「开发日志」追加一行（场次接续）。
4. `bash scripts/deploy-test.sh` 部署冒烟通过。
5. 开场必做：`git status` / `git log --oneline` / `git pull --rebase`；发现并行会话改动一律以远端最新进度为准，绝不覆盖他人提交；工作区有半成品先接续收尾，不开新头。
6. 凭据/密钥/密码永不入库；验证数据用完即清。

## 四、窗口收尾

- 2026-10-26 09:00 定时任务自动停止（maxRuns 与场数上限对齐）。
- 最后一场除常规收尾外：核对 DEV_PLAN 全部勾选状态与本文档两阶段标记，输出冲刺总结（完成项/遗留项/优化轮次）到开发日志。
