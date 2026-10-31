# 白泽（BaiZe）插件开发规范

> 适用范围：所有业务插件（asset / term / job / pipeline / container / k8s / gpu / dbops / monitor / workflow / org / aiops 等）。
> 功能总览与里程碑口径见仓库根 [README.md](../README.md)；接口与错误码分段规范见本文第 4 节。

## 1. 总原则

1. **业务代码只存在于插件内**：`server/plugin/<name>/` 与 `web/src/plugin/<name>/`，禁止修改底座（server 根包与 web 基座组件）业务逻辑；确需底座增强时单独评估提交，保持上游可合并。
2. **底座能力直接用**：认证(JWT)、鉴权(Casbin v3)、操作日志中间件、统一响应 `{code,data,msg}`、分页、Swagger 注释、定时任务注册、多云上传——不要重复造。
3. **上游跟随**：底座保持与 flipped-aurora/gin-vue-admin main 可合并（merge --allow-unrelated-histories）。插件目录结构对齐官方插件（`plugin/announcement` 为参考实现）。

## 2. 插件目录结构

```shell
server/plugin/<name>/
├── plugin.go            # 注册入口（v2 Plugin 接口，init() 自注册）
├── initialize/          # 路由、菜单/API/casbin 初始化数据（migration，幂等）
│   ├── router.go
│   ├── menu.go          # 菜单种子（参考 plugin/auto/initialize/menu.go）
│   ├── api.go           # API 记录种子（超级管理员→API 管理可见）
│   └── casbin.go        # 角色-API 授权种子
├── api/v1/              # 接口层：参数绑定与响应，禁止业务逻辑
├── router/              # 路由分组（initRouter 挂到 Private/Public Group）
├── service/             # 业务逻辑（单元测试写这里，mock global.GVA_DB）
├── model/               # model / request / response
└── config/              # 插件配置结构（对接 config.yaml）

web/src/plugin/<name>/
├── view/                # 页面组件（动态菜单 Component 路径即此相对路径）
├── api/                 # 后端接口封装
└── router/              # 静态兜底路由（正常走数据库动态菜单）
```

参考已生成的骨架：`server/plugin/asset/plugin.go`（M0 注册 10 个业务插件骨架，后续新增 agent/aiops 通道类插件共 12 个）。

### 2.1 最小可运行插件骨架

以 monitor 为真实样板（完整实现见 `server/plugin/monitor/`），新插件的 `plugin.go` 最小骨架：

```go
// Package <name> 白泽业务插件：<一句话职责>
// 开发规范见 docs/plugin-dev-guide.md
package <name>

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/<name>/initialize"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin) // init() 自注册，无需改底座任何注册表
}

// Register 插件加载时调用：表迁移 → API/菜单种子 → 授权种子 → 路由
func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	initialize.Gorm(ctx)     // AutoMigrate（幂等）
	initialize.Api(ctx)      // API 记录种子
	initialize.Menu(ctx)     // 菜单种子
	initialize.Casbin(ctx)   // 角色-API 授权种子
	initialize.Router(group) // 业务路由（显式挂 JWT + Casbin）
}
```

`initialize/router.go` 路由注册最小样板（摘自 monitor，**显式**挂 JWT + Casbin，不依赖底座隐式继承）：

```go
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private) // 插件内 router/ 包定义业务分组
}
```

要点：`init()` 自注册 + `Register` 内按序初始化；种子一律幂等（见第 3 节）；删除插件时全仓 grep 结构体名/路由名（见第 3 节教训）。

## 3. 初始化数据（菜单/API/Casbin）规范

- 菜单、API、casbin 策略一律通过 **initialize/ 下的幂等种子**下发：先查后插（`errors.Is(db.Where(...).First(&x).Error, gorm.ErrRecordNotFound)` 判重），保证重复启动/初始化安全。
- 菜单 `Component` 路径 = `plugin/<name>/view/xxx.vue`；ParentId 通过 `menuNameMap` 引用已有菜单，避免硬编码 ID。
- 删除功能时：种子与测试库旧菜单都要清理（SQL 直删 `sys_base_menus` + `sys_authority_menus` + `sys_base_menu_btns` 关联行）。
- 教训（2026-09-30）：GVA 示例模块移除时残留 `initialize/gorm.go`、`ensure_tables.go` 注册行导致编译失败——**删模块必须全仓 grep 结构体名/路由名**。

## 4. 编码红线（会话协议第 5 条展开）

| 项 | 规则 |
|---|---|
| 凭据 | 一律 `asset` 凭据保险库（AES-256-GCM）；表只存密文+末4位；接口不回显明文 |
| 审计 | 所有写操作过操作日志中间件；删除类接口软删除 |
| 依赖 | 新增第三方库需评估必要性并保持同类场景唯一；标准库优先 |
| 错误码 | 按插件独占分段（asset=1000-1099 …，全量分段见 README「接口与错误码规范」），0 成功，通用错误用 GVA 7xxx |
| Swagger | 每个导出接口写 swag 注释，`make doc` 可再生成 |
| 命名 | 表名 `<plugin>_<entity>`；避免 GVA_ 前缀（保留给底座） |

## 5. 测试与验收

- 单元测试：service 层为主，**全量 `go test ./...` 必须绿**（场85-87 已根治 ast 注入对称性与环境依赖，不再依赖 `-short` 隔离；CI 已同步放开全量）。**禁止在测试中修改仓库源文件**。
- 依赖全局初始化（GVA_DB/GVA_CONFIG）的测试加环境守卫跳过（不可达 `t.Skip`，参考 mcp/client 与 auto_code 的守卫模式）。
- 每场完成定义（DoD）：代码+编译+单测+Swagger+初始化脚本+push+**测试环境部署冒烟通过**。

## 6. 部署

```bash
bash scripts/deploy-test.sh     # 一键：构建→上传→重启→初始化(仅首次)→冒烟
make deploy                     # 等价
NEW_OPS_TEST_HOST=root@x.x.x.x make deploy-host   # 其他机器
```

配置约定：`config.yaml` 在远端是有状态文件（部署脚本不覆盖）；本地仓库的 `server/config.yaml` 是开发默认值；生产差异（env: public、captcha 开启、MySQL、JWT 密钥、凭据主密钥）记录在 `server/config.prod.yaml.example`。

## 7. 骨架状态（M0 交付 → 当前实况 2026-10-24）

| 插件 | 状态 | 当前阶段 |
|---|---|---|
| asset / term / job / pipeline | **已交付**（M1-M3，含批量执行双后端/流水线远程执行器） | M9 增强（Agent 通道） |
| container / k8s | **已交付**（M4 全部 + C2/C3 + K3 全部 + AI 诊断） | M9 增强（Agent 模式纳管） |
| gpu | 防超卖引擎已交付（行锁修正）；直通/HAMi/监控待显卡环境 | M5 收尾 |
| dbops / monitor | **已交付**（goInception 真机闭环 / 告警引擎全量） | soar/多环境待环境 |
| workflow / aiops | **已交付**（状态机+发版闭环 / MCP+LLM 网关+Pod AI 诊断） | org 钉钉待企业配置 |
| agent | **M9 全链交付**（任务/文件分发/Docker-K8s 代理/100 压测） | 集群内 E2E 待测试机 |
| announcement / email / auto / plugin-tool | GVA 官方插件，保留 | 底座 |

里程碑与各插件功能总览以仓库根 [README.md](../README.md) 为准。
