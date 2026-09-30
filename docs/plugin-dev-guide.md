# 白泽（BaiZe）插件开发规范

> 适用范围：所有业务插件（asset / term / job / pipeline / container / k8s / gpu / dbops / monitor / workflow / org / aiops 等）。
> 开发计划与里程碑见 [DEV_PLAN.md](DEV_PLAN.md)。本规范与 DEV_PLAN 3.5（依赖清单）、3.6（接口与错误码）配合使用。

## 1. 总原则

1. **业务代码只存在于插件内**：`server/plugin/<name>/` 与 `web/src/plugin/<name>/`，禁止修改底座（server 根包与 web 基座组件）业务逻辑；需要底座增强时在 DEV_PLAN 立项后单独提交。
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

参考已生成的骨架：`server/plugin/asset/plugin.go`（M0 已注册 10 个插件骨架）。

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
| 依赖 | 新增第三方库先登记 DEV_PLAN 3.5；标准库优先；同类场景唯一库 |
| 错误码 | 按 DEV_PLAN 3.6 分段（asset=1000-1099 …），0 成功，通用错误用 GVA 7xxx |
| Swagger | 每个导出接口写 swag 注释，`make doc` 可再生成 |
| 命名 | 表名 `<plugin>_<entity>`；避免 GVA_ 前缀（保留给底座） |

## 5. 测试与验收

- 单元测试：service 层为主，`go test -short ./...` 必须全绿；**禁止在测试中修改仓库源文件**（GVA utils/ast 的注入测试即是反例，已用 TestMain + -short 隔离）。
- 依赖全局初始化（GVA_DB/GVA_CONFIG）的测试加 `if testing.Short() { t.Skip(...) }`。
- 每场完成定义（DoD）：代码+编译+单测+Swagger+初始化脚本+DEV_PLAN 勾选+开发日志+push+**测试环境部署冒烟通过**。

## 6. 部署

```bash
bash scripts/deploy-test.sh     # 一键：构建→上传→重启→初始化(仅首次)→冒烟
make deploy                     # 等价
NEW_OPS_TEST_HOST=root@x.x.x.x make deploy-host   # 其他机器
```

配置约定：`config.yaml` 在远端是有状态文件（部署脚本不覆盖）；本地仓库的 `server/config.yaml` 是开发默认值；生产差异（env: public、captcha 开启、MySQL、JWT 密钥、凭据主密钥）记录在 `server/config.prod.yaml.example`。

## 7. 当前骨架状态（M0 交付）

| 插件 | 状态 | 目标里程碑 |
|---|---|---|
| asset / term / job | 骨架已注册 | M1-M2 |
| pipeline | 骨架已注册 | M3（以 new-jenkins 为蓝本迁移） |
| container / k8s | 骨架已注册 | M4 |
| gpu | 骨架已注册 | M5 |
| dbops / monitor | 骨架已注册 | M6 |
| workflow / org / aiops | 骨架已注册 | M7 |
| announcement / email / auto / plugin-tool | GVA 官方插件，保留 | 底座 |
