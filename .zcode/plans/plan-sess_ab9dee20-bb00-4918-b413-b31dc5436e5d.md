# 白泽平台 25 插件架构级重写计划

## Context（背景与目标）

现状（3 个探索代理已完成盘点）：25 个插件约 8.3 万行 Go、40k 行 service、207 个纯函数测试文件，存在四类架构级问题：

1. **无数据访问层**：225 个文件直接引用 `global.GVA_DB` 共 1294 处，业务逻辑与 GORM 调用融合，无法单测、无法替换存储。
2. **全局耦合**：service 是无状态单例直取 global 三件套，无依赖注入；交叉插件用具体类型硬依赖（term→asset、workflow→pipeline）。
3. **错误契约碎片化**：三种形态并存（typed struct / 错误码拼字符串 / 丢弃错误），且数字错误码从未到达 HTTP 响应，README 的分段错误码形同虚设；另有 4 处插件内错误码冲突（asset 1012/1013、container 1412、llmops 2298/2299 重复定义）。
4. **跨插件重复**：truncate 助手 10+ 份（2 份逐字节相同）、bigram+topK 排序 6 个插件各写一套、CSV 导出 4 份、脱敏 5 份；104 个文件手写 Offset/Limit 分页。

用户决策：**架构级重写、25 个插件全量挨个来、构建+测试全绿即提交**。对外契约（路由 URL、`{code,data,msg}` 响应结构、casbin/菜单/Api 种子）**保持不变**，前端仅同步必要改动——风险最小且与既有 207 个测试兼容。

## 一、共享内核（Phase 0，1 个提交）

扩展 `server/plugin/plugin-tool/`（各插件已依赖其 utils 包，延续既有惯例）新增 4 个子包：

| 包 | 内容 | 消灭的重复 |
|---|---|---|
| `plugin-tool/errx` | `AppError{Code,Msg}` 实现 error；`Respond(c, err, fallback)` 统一映射到响应（插件错误码首次真正到达前端）；`Code(err)` | 三种错误形态 → 一种 |
| `plugin-tool/text` | `Truncate`（rune 安全）、`Bigrams`、`RankTopK[T](items, score, k)` 通用打分排序 | truncate×10、bigram+rank×6 |
| `plugin-tool/csvx` | `Builder{header}` + `Row` + `String`（固化既有口径：半角逗号→全角、换行→空格、≤5000 行） | CSV 导出×4 |
| `plugin-tool/data` | `Page(db, page, pageSize, dest)` 统一分页；`testdb` 内存 sqlite 测试夹具（glebarez/sqlite 已在 go.mod） | 手写分页×104，并使 DB 层可测 |

提交：`refactor(plugin-tool): 共享内核——errx/text/csvx/data 四子包与 sqlite 测试夹具`。

## 二、逐插件重写模板（核心循环，每个插件 1 个提交）

按统一模板改写，**方法签名与路由对外不变**：

1. **依赖注入**：service 结构体增加 `db *gorm.DB` 字段与构造函数（默认 `global.GVA_DB`），包级单例变量保留 → api 层调用点零改动；交叉依赖改为接口（workflow→pipeline、term→asset）。
2. **repo 抽取**：GORM 访问从 service 迁入 `repo/` 目录（按聚合根一文件）；service 保留业务规则（校验/状态机/行锁决策），repo 只做查询。
3. **错误契约**：各插件局部 `xxxError`/字符串嵌码 → `errx.AppError` + 保留分段错误码常量；api 层 `response.FailWithMessage(err.Error(), c)` → `errx.Respond(c, err, "默认消息")`；顺手修复错误码冲突。
4. **去重接入**：本地 truncate/bigram/rank/CSV/脱敏助手替换为内核包（仅语义逐字节等价者；口径不同者如 ziqiang OOV 单字分词保留本地）。
5. **分页统一**：手写 Offset/Limit → `data.Page`（响应结构不变）。
6. **测试**：既有测试全绿；对关键 DB 路径（防超卖、状态机）用 `data.testdb` 补 sqlite 内存库测试。
7. **验证与提交**：`cd server && go build ./... && go vet ./... && go test ./...` 全绿 → `refactor(xxx): 中文说明` 提交（符合 scripts/check-commit.sh 规范）。

## 三、推进顺序（轻 → 重，试点先行定模式）

| 批次 | 插件 | 说明 |
|---|---|---|
| 试点 | gpu（1600 段）、announcement、workflow、email、auto | gpu 为主试点；workflow 提取 pipeline 依赖接口缝；email 从 v1 旧插件机制迁移到 v2（唯一动 `server/initialize/plugin_biz_v1.go` 的提交，减法迁移）；auto 后端近空、仅接入可适用部分 |
| 轻量纯函数 | policyreg、artsci、fanhua、ziqiang、exam | DB 薄、纯函数厚：DI+errx+去重为主，repo 从简 |
| 中量 | eduadmin、crossauth、oplib、aiops、dbops | 完整模板 |
| 重型收尾 | asset、monitor、agent、job、pipeline、container、k8s、term | 并发敏感（行锁/goroutine/WS 流）代码**原样搬移不"顺手优化"**；term 在 asset 之后做（解 term→asset 具体依赖） |
| 特大 | llmops（9.5k svc 行 + 3086 行 api 文件） | 拆 3-4 个域级提交（网关/数据源 → 知识库与检索 → 评估/智能体 → api 层拆分） |
| 收官 | docs | `docs/plugin-dev-guide.md` 增补新架构章节（repo/DI/errx 惯例）、README 错误码规范段落更新说明 |

预计 30-35 个提交。每个提交独立可构建、可回滚。

## 四、关键文件

- 内核：`server/plugin/plugin-tool/{errx,text,csvx,data}/`（新建）
- 模板参考：`server/plugin/gpu/service/gpu.go`（typed error 现状）→ `gpu/repo/`（新）
- 每插件触及：`service/*.go`（拆改）、`repo/*.go`（新）、`api/v1/*.go`（错误响应替换）、`initialize/`（种子不动）
- 不动：`server/initialize/router.go`、各插件 `router/`、casbin/菜单/Api 种子、`web/src/plugin/*/view/**`（web api js 仅在响应可见变化时同步）

## 五、风险与守护

- **契约守护**：每插件完成后跑 `bash scripts/audit_routes.sh`，路由清单输出必须与改前一致。
- **测试守护**：207 个测试文件不许删除/跳过（既有 t.Skip 环境守卫除外）；口径不同的助手只保留不合并。
- **重型插件**：并发/锁/流式代码逐行搬移，新测试只补纯可测路径；不做无测试背书的"改进"。
- **基础设施**：`plugin-tool/utils/check.go` 种子机制不动；表名/模型不动（AutoMigrate 兼容既有库）。

## 六、验证方式

每单元提交前：
1. `cd server && go build ./... && go vet ./... && go test ./...`
2. 触及 web 时 `cd web && npm run build`；触及 agent/ 时 `make test-agent`
3. `bash scripts/audit_routes.sh` 对比路由无漂移
4. 提交后 `bash scripts/check-commit.sh HEAD 1` 校验提交格式

全量收官：`make test` + `make build-web` + `check-commit.sh` 全量回归 + 确认 25 插件逐个有对应 refactor 提交。