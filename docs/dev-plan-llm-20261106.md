# 白泽 new-ops · 大模型应用开发平台冲刺开发计划（2026-11-06 23:30 → 2026-11-09 09:00）

> **采购需求对标**：大模型应用开发平台 1 套 = ①大模型超级开发基础平台（8 大能力域：数据管理与知识治理 / 知识库与检索引擎 / 模型接入与管理 / 训练与调优 / 智能体与插件服务 / 技能生态与开放集成 / Prompt 工程 / 基础管控）+ ②算子库 + ③AI 等级考试分布式系统 + ④AI 能力考试终端系统 + ⑤AI 开发数据集。
> **落地形态**：新增 3 个业务插件——`llmops`（基础平台 8 域 + 数据集）、`oplib`（算子库）、`exam`（考试分布式 + 考试终端接入）；前端 `web/src/plugin/<name>/`。
> **执行方式**：定时任务每 15 分钟一轮，每轮认领**一个**未完成条目，做完即构建/测试、勾选、提交。
> 条目状态标记：`- [ ]` 待做（按文件顺序即优先级）｜`- [x]` 完成｜`[!]` 不可行（后缀写明原因）。
> 每条目动手前**先读相关代码确认现状**：若等价功能已存在，改为做相邻的小增强，并在提交信息中注明"原条目已实现，调整为 xxx"。

---

## 轮转规则（每轮必读）

0. **环境注意**：本机 PATH 默认 `go` 为 1.18（无法解析 go.mod 的 1.26）。一律使用**绝对路径** `/usr/local/go1.26/bin/go` 执行 build/vet/test——仅 `export PATH` 不够：zsh 会 hash 缓存旧 `go` 路径，绝对路径免疫。仓库根无 go.mod，`go test ./...` 必须在 server/（或 agent/）目录内执行。当前时间以 `date` 命令实查为准。
1. 开工先 `git status`：工作区脏则 `git stash push -m "auto-stash <时间戳>"`（**不丢弃**），再继续。
2. 取本文件中**第一个** `- [ ]` 条目（主阶段取完自动进入备用池），只做这一项，范围控制在约 15 分钟工作量；偏大的条目可拆**最小可用增量**先交付可提交的部分，剩余留待后续轮次（勾选时注明已完成部分）。
3. 实现 → 验证：server 改动 `cd server && /usr/local/go1.26/bin/go build ./... && /usr/local/go1.26/bin/go test ./...`（**全量必须绿**，当前 25 包基准）；web 改动 `cd web && npm run build`。**不过不提交**。
4. 完成后勾选 `- [x]`，`git add`（含本计划文件），提交信息沿用仓库惯例：`feat(llmops|oplib|exam|web): 一行中文说明`（测试类用 `test(xxx):`，文档用 `docs:`）。**不 push**。
5. 单条目超过 25 分钟仍卡住：标 `[!]` 写明原因；已有完整可用的部分可提交，否则回滚现场，进入下一条目留给下一轮。

## 禁止事项（红线）

- 不 `push`、不 `force`、不 `rebase`、不改写任何历史提交。
- **不新增 go.mod / package.json 第三方依赖**、不升级现有依赖版本；向量/检索/相似度一律标准库实现（余弦、bigram、编辑距离自写）。
- 不改底座（server 根包与 web 基座）业务逻辑；不改 casbin/JWT/登录/权限框架逻辑（插件内种子按规范加）。
- 不做破坏性数据库迁移；不删除现有功能、接口与测试。
- 每轮只做一个条目、一个 commit；不碰与本条目无关的文件。
- 插件开发六处必改清单（漏一步即 403/不可见）见 docs/plugin-dev-guide.md §3.1；新表命名 `<plugin>_<entity>`；错误码分段：llmops=2200-2299、oplib=2300-2399、exam=2400-2499（A01 时在 README 登记，动手前 grep 核实未占用）。
- 密钥类字段一律复用 `plugin/asset/crypto` AES-GCM 密文落库，接口不回显明文（对齐 aiops 口径）。

## 收尾规则

- 主条目 + 备用池全部完成后：不再新增功能，执行全量回归（`cd server && go build/test`、`cd agent && go build/test`、`cd web && npm run build`），修复发现的小问题，在本文件底部"收官总结"追加统计与需求覆盖对照表，提交 `chore: 大模型平台冲刺收官回归`。
- **2026-11-09 09:00 及之后**启动的轮次：直接调用 CronPauseSelf 结束，不再做任何事。

---

## 阶段 A · llmops 骨架 + 数据管理与知识治理（要求(1)，11-06 深夜 → 11-07 上午）

- [x] A01 | llmops: 插件骨架 M0——`server/plugin/llmops/{plugin.go, initialize/initialize.go(Gorm+Api+Menu+Casbin+Router), router/enter.go, api/v1/ping}`，plugin/register.go 空导入接线；菜单种子含父级"大模型平台"；README 错误码分段行补 llmops=2200-2299/oplib=2300-2399/exam=2400-2499；server build+test 全绿
- [x] A02 | llmops: 数据资产 model `LlmopsDataAsset`（name/dataType[doc,image,audio,video,webpage,qa,structured,semistructured]/source/status/rawText/metaJSON/tags/categoryId/size）AutoMigrate + 列表（分页+类型/关键字过滤）/详情/软删除 API + api/casbin 种子 + service 单测
- [x] A03 | llmops: 文档采集——上传接入（复用底座上传工具，落 rawText：txt/md/csv/json 直接读取；pdf 等二进制仅编目元数据并注明解析口径）+ 状态机 draft→parsed + 单测
- [x] A04 | llmops: 网页采集——URL 拉取（net/http 超时+大小上限）+ HTML 正文抽取（去 script/style/标签，纯标准库）入库 webpage 资产 + httptest 单测
- [x] A05 | llmops: 问答对采集——qa 类型 JSON/CSV 解析为 `LlmopsQaPair`（assetId/question/answer）+ 同题去重 + 单测
- [x] A06 | llmops: 结构化数据源接入——`LlmopsDataSource`（type mysql/postgres/dsn 凭据 AES-GCM 密文）+ 连接测试 + 库表清单/采样预览 API（database/sql，driver 已有）
- [x] A07 | llmops: 半结构化解析——JSON/YAML/XML 顶层字段/类型/样本值抽取生成 metaJSON 编目 + 单测（yaml 用 go.mod 已有解析器，先核实）
- [x] A08 | llmops: 编目治理——分类目录树 `LlmopsCategory` + 资产挂分类/打标签 + 生命周期事件流水 `LlmopsAssetEvent` 自动记录 + 树形 API
- [x] A09 | llmops: 音视频资产——元数据编目（大小/MIME）+ ASR 转写适配器接口（默认 NotConfigured）+ 资产多路检索 API（类型/标签/分类/关键字/时间组合）+ 单测
- [x] A10 | llmops: web 数据管理页——`web/src/plugin/llmops/view/data/index.vue`（列表/类型筛选/上传/URL 采集/详情抽屉）+ api/data.js + 菜单种子；web build 绿

## 阶段 B · 模型接入与管理（要求(3)，前置于检索；11-07）

- [x] B01 | llmops: 模型注册表 `LlmopsModel`（name/provider/baseUrl/apiKeyEnc/capability[chat,embed,rerank]/ctxLen/status/isDefault）+ CRUD（密钥不回显，对齐 aiops 供应商口径）+ 单测
- [x] B02 | llmops: 模型网关调用器——OpenAI 兼容 chat/completions 与 embeddings 调用（net/http，参考 aiops gateway）+ 超时/错误码包装 + httptest 单测
- [x] B03 | llmops: 在线测试台 API——chat/embeddings/rerank 三协议测试（rerank 约定 POST /rerank 返回 scores）+ 调用日志 `LlmopsModelCallLog`（model/latency/tokens/ok）
- [x] B04 | llmops: 模型一键部署——部署任务 `LlmopsModelDeploy`（image/端口/参数 JSON）→ 对接 container 插件最小对接点创建容器 + 健康检查 + 预置模型服务种子（3 个内置定义，未部署态）
- [ ] B05 | llmops: 多模型调度——SelectModel(capability) 按 isDefault+近 N 次时延均值路由 + 单测；web 模型管理页（注册表/测试台/调用日志）

## 阶段 C · 知识库与检索引擎（要求(2)；11-07 下午 → 晚上）

- [x] C01 | llmops: 知识库 model `LlmopsKnowledge`（name/type[vector,fulltext,graph,qa]/desc/status）+ CRUD + 文档入库 API（资产/原始文本→指定库）
- [x] C02 | llmops: 切块服务 Chunker——固定长度+重叠+段落边界感知（纯 Go）+ 表驱动单测
- [x] C03 | llmops: 向量库——chunk 表（kbId/docId/content/vectorJSON/dim）+ 向量化（embed 模型调用；未配置模型时确定性哈希向量 fallback 并标注口径）+ 余弦 top-k 检索 + 单测
- [x] C04 | llmops: 全文库——bigram 切分 + 倒排简化打分（简化 BM25）+ 检索 + 单测
- [x] C05 | llmops: 问答对库——qa 检索（精确匹配→编辑相似度阈值排序）+ 单测
- [x] C06 | llmops: 知识图谱库——节点/边表 + BFS 多跳推理 API（startNode→hop N→路径与终点返回）+ 单测
- [x] C07 | llmops: 摘要检索——文档摘要生成（chat 模型异步任务表）+ 摘要检索 API
- [x] C08 | llmops: 混合检索编排——多路召回（向量/全文/问答/图谱/摘要）并发 WaitGroup → RRF 融合 → Rerank 精排（未配置模型跳过）+ mock 单测
- [x] C09 | llmops: 统一检索入口 API `/llmops/knowledge/search`（kbIds/mode/topK）+ 调用留痕
- [x] C10 | llmops: web 知识库页——库管理/文档导入/检索测试台 + 菜单种子；web build 绿

## 阶段 D · 大模型训练与调优（要求(4)；11-07 深夜 → 11-08 上午）

- [x] D01 | llmops: 训练任务 model `LlmopsTrainJob`（type[sft_full,sft_lora,reward,rl,eval,compress]/datasetId/baseModel/hyperJSON/status 状态机）+ CRUD + 状态流转 API + 单测
- [x] D02 | llmops: 训练脚本渲染器——超参→命令行渲染（全参/LoRA 两套模板，参数白名单防注入）+ dry-run 执行器（本地模式仅渲染落库）+ 表驱动单测
- [x] D03 | llmops: RLHF 两阶段——奖励模型训练与强化学习训练任务编排（复用渲染器，阶段依赖 preJobId 校验）+ 单测
- [x] D04 | llmops: 训练加速——混合精度/梯度累积/多卡参数渲染白名单扩展 + 任务启动/停止/日志查询（dry-run 落日志）
- [x] D05 | llmops: 模型评估——`LlmopsEvalRun`（evalSet/指标定义/样本调用生成+判分解析）+ 结果表 + 假模型单测
- [x] D06 | llmops: 模型压缩——量化/剪枝/蒸馏任务类型接入渲染器（占位执行器）+ 压缩产物登记
- [x] D07 | llmops: 产物→预测服务——`LlmopsArtifact` + 训练产物一键注册为模型服务（对接 B04）+ 单测
- [x] D08 | llmops: web 训练页——任务列表/新建向导/详情日志 + 菜单种子；web build 绿

## 阶段 E · 智能体与插件服务（要求(5)；11-08）

- [x] E01 | llmops: 智能体 model `LlmopsAgent`（name/mode[qa,interactive,workflow,multi,agentic]/configJSON/status 六阶段[草稿,调测,评测,发布,运营,监控]/desc）+ CRUD + 版本快照表 `LlmopsAgentVersion` + 单测
- [x] E02 | llmops: 智能体运行时——意图理解+任务自主拆解（chat 结构化 JSON 输出+校验失败重试）+ 会话/消息表 `LlmopsAgentSession/Message`
- [x] E03 | llmops: 工具调用回路——技能/MCP 工具绑定→选择调用→结果评估与自我反思（轮次上限可配）+ 假工具单测
- [x] E04 | llmops: 记忆——短期（最近 N 轮窗口截断）+ 长期（`LlmopsAgentMemory` 条目+检索注入）+ 单测
- [x] E05 | llmops: 工作流模式——节点 JSON 编排（llm/tool/condition/human）+ 顺序执行引擎 + 单测
- [x] E06 | llmops: 多智能体模式——子智能体引用编排（顺序/汇聚）+ 单测
- [x] E07 | llmops: 调测 API——会话调试（含 SSE 流式可选）+ 调试历史
- [x] E08 | llmops: 智能体评测——评测集运行+打分（规则+LLM 评分）+ 结果表 + 单测
- [x] E09 | llmops: 发布与运营——发布/回滚（版本快照切换）+ 调用统计聚合 API（次数/时延/Token）
- [x] E10 | llmops: 智能体监控——成功率/时延分布聚合 + 异常会话列表
- [x] E11 | llmops: web 智能体页——列表/新建（模式选择）/编排表单/调试会话窗 + 菜单种子；web build 绿

## 阶段 F · 技能生态与开放集成（要求(6)；11-08 下午）

- [x] F01 | llmops: Skill model `LlmopsSkill`（name/type[builtin,extended,custom]/defJSON/status）+ CRUD + 内置技能种子 5 个（文本摘要/翻译/SQL 生成/关键词提取/表格问答）
- [x] F02 | llmops: 零代码创建技能——自然语言描述→chat 生成技能定义 JSON→校验→入库 + 假模型单测
- [x] F03 | llmops: 技能包导入导出——JSON 导出/一键安装（校验+冲突提示）
- [x] F04 | llmops: MCP 工具封装——业务系统 API 注册（baseUrl/method/参数模板/鉴权头密文）→ 工具描述生成 + 调用代理（对齐 server/mcp 现状口径）+ httptest 单测
- [x] F05 | llmops: 连接器——ERP/CRM/OA/SCM 连接器配置（类型模板+字段映射 JSON）+ 连通测试 API
- [x] F06 | llmops: 技能执行沙箱——超时/响应大小上限/URL 白名单/参数脱敏 + 单测
- [x] F07 | llmops: 技能鉴权限流——调用令牌（生成/校验/吊销）+ 令牌桶限流（内存）+ 调用审计日志表 + 单测
- [x] F08 | llmops: web 技能中心页——技能列表/零代码向导/MCP 注册/连接器/调用日志 + 菜单种子；web build 绿

## 阶段 G · Prompt 工程（要求(7)；11-08 傍晚）

- [x] G01 | llmops: Prompt 模板 `LlmopsPromptTemplate`（name/content/变量占位符解析校验/版本号）+ CRUD + `{{var}}` 解析单测
- [x] G02 | llmops: 模板渲染 API——POST /llmops/prompt/render（变量填充+缺失报错）+ 版本快照 + 单测
- [x] G03 | llmops: Prompt 在线优化——chat 模型改写（保留原版对比）+ 批量优化任务（队列+逐条结果对比）
- [ ] G04 | llmops: 业务规则/行业术语库/专家思维链——三表 CRUD + 模板引用关联 + web Prompt 工作台页；web build 绿

## 阶段 H · 算子库（采购件②；11-08 晚）

- [ ] H01 | oplib: 插件骨架 + 算子 model `OplibOperator`（name/category[数据处理,特征,检索,后处理,评估]/inputSchema/outputSchema/implType[builtin,script]/version/desc）+ CRUD + 种子（核实 A01 已登记错误码段）
- [ ] H02 | oplib: 内置算子集 8 个——文本清洗/分句/去重/关键词抽取（bigram 频次）/字数统计/JSON↔CSV 转换/正则提取/大小写归一，纯 Go + 单测
- [ ] H03 | oplib: 算子执行引擎——参数校验（简化 schema：type/required/enum）+ 执行 + 留痕 `OplibRunLog` + 单测
- [ ] H04 | oplib: 算子编排——DAG 管道 `OplibPipeline`（节点+边 JSON，拓扑排序校验环）+ 顺序执行引擎 + 单测
- [ ] H05 | oplib: 脚本算子——受限执行（解释器白名单/超时/临时目录清理）+ 测试运行 API + 单测
- [ ] H06 | oplib: web 算子库页——算子列表/执行测试台/管道编排 + 菜单种子；web build 绿

## 阶段 I · AI 等级考试分布式系统（采购件③；11-08 深夜 → 11-09 凌晨）

- [ ] I01 | exam: 插件骨架 + 题库 model `ExamQuestion`（type[单选,多选,判断,简答,实操]/stem/optionsJSON/answer/score/knowledgePoint/difficulty）+ CRUD + JSONL 批量导入 + 单测
- [ ] I02 | exam: 组卷引擎——`ExamPaper`（抽题规则：知识点×难度×题量 JSON）+ 随机抽题（去重+分值合计校验）+ 单测
- [ ] I03 | exam: 考试 model `ExamExam`（场次/时间窗/试卷/考生范围/状态机[草稿,发布,进行中,已结束]）+ CRUD + 状态流转校验
- [ ] I04 | exam: 考生管理——`ExamCandidate`（名单导入/券码唯一生成）+ DB 行锁发券防重复 + 单测
- [ ] I05 | exam: 答题 API——券码+终端令牌登录/拉题/保存答案（幂等+时间窗校验）/交卷（状态机+防重复提交）+ 单测
- [ ] I06 | exam: 判分引擎——客观题自动判分（多选漏选规则可配）+ 主观/实操题 LLM 判分（评分标准 prompt+分数解析容错）+ 判分任务表 + 假模型单测
- [ ] I07 | exam: 成绩与等级——`ExamScore` + 配置化分数线等级判定 + 成绩单列表/CSV 导出 + 单测
- [ ] I08 | exam: 分布式调度——考试节点 `ExamNode`（终端注册/心跳/负载）+ 领卷分片调度（行锁+幂等）+ 单测
- [ ] I09 | exam: 防作弊——行为日志 `ExamBehavior`（切屏/粘贴上报 API+计数）+ 交卷答案相似度抽查（编辑距离抽样）+ 单测
- [ ] I10 | exam: web 考试管理页——题库/试卷/场次/考生/成绩页签 + 菜单种子；web build 绿

## 阶段 J · AI 能力考试终端系统（采购件④；11-09 凌晨 → 早上）

- [ ] J01 | exam: 终端接入——终端注册（机器码+令牌）/心跳/可考列表 API（考生端独立令牌鉴权，不走管理端 JWT）
- [ ] J02 | exam: 断线续考——答案草稿自动保存/恢复接口（版本号幂等）+ 单测
- [ ] J03 | exam: 实操题执行——终端实操任务→调用 llmops 智能体运行时/算子执行，过程留痕（输入输出/耗时）作为判分依据
- [ ] J04 | exam: web 终端答题页——考试列表/答题界面（倒计时/单题导航）/交卷确认 + build 绿
- [ ] J05 | exam: web 实操交互——对话式作答+结果预览+过程日志展示 + build 绿

## 阶段 K · AI 开发数据集（采购件⑤；11-09 凌晨）

- [ ] K01 | llmops: 数据集 model `LlmopsDataset`（name/purpose[sft,rlhf,eval,exam]/format[jsonl,csv]/样本数/version/license/desc）+ CRUD
- [ ] K02 | llmops: 导入导出——JSONL 逐行校验（purpose 对应 schema：sft=instruction+output 等）+ 错误行报告 + 下载 API + 单测
- [ ] K03 | llmops: 种子数据集——仓库内置 3 份教育垂类示例（sft 对话/评测集/考试题集，各 ≥20 样本）+ 幂等种子导入
- [ ] K04 | llmops: 数据集联动——训练/评估/考试任务引用校验（存在性+purpose 匹配）+ 引用计数 + 单测
- [ ] K05 | llmops: web 数据集页——列表/导入/样本预览/下载 + 菜单种子；web build 绿

## 阶段 L · 基础管控收口（要求(8)）+ 收官

- [ ] L01 | llmops: 数据看板 API——各域统计聚合（资产/知识库/模型/任务/智能体/考试计数与近 7 日趋势）+ 单测
- [ ] L02 | llmops: 资源配额——用户/智能体 Token 与调用配额表+扣减校验 + 单测
- [ ] L03 | llmops: 干预机制——会话人工接管标记 API + 敏感词拦截词库（命中检测）+ 单测
- [ ] L04 | llmops: 操作留痕核查——写接口全过操作日志中间件核对清单，补漏
- [ ] L05 | docs: README 新增「大模型平台/算子库/考试系统/数据集」能力小节 + 功能总览表
- [ ] L06 | server: llmops/oplib/exam swagger 注释全量核对补齐
- [ ] L07 | 全量回归：server/agent 测试 + web build 全绿，修复发现的小问题
- [ ] L08 | 收官：本文件底部写总结（完成/跳过统计、采购需求逐项覆盖对照表、遗留清单）

## 备用增强池（主条目全部完成后依次取用）

- [ ] P01 | llmops: 数据资产回收站（软删除恢复 API）
- [ ] P02 | llmops: 知识库文档重复检测（simhash 简化版）
- [ ] P03 | llmops: 智能体调试窗 Markdown 渲染与消息复制
- [ ] P04 | oplib: 算子执行结果缓存（同参数哈希命中）
- [ ] P05 | exam: 考试场次日历视图（前端）
- [ ] P06 | exam: 成绩分布直方图（前端 echarts）
- [ ] P07 | llmops: 训练任务费用/时长估算展示
- [ ] P08 | llmops: 模型调用日志 CSV 导出
