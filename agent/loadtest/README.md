# baize-loadtest · Agent 通道压测工具

M9 验收工具：模拟 N 个虚拟 Agent 并发接入，验证单实例 server 承载能力（历史上 100 并发全链通过）。

## 流程

```
登录(管理员) → 批量创建资产 + 签发接入令牌 → 并发 N 虚拟 agent
（WS 注册 → 周期心跳 + 指标上报）→ 汇总注册/心跳统计与在线数 → 清理（吊销实例 + 删除资产）
```

## 用法

```bash
cd agent
go run ./loadtest -server http://127.0.0.1:8888 -user admin -pass <管理员密码> -n 100 -duration 60s
```

> 建议在测试环境执行：压测会真实创建资产/实例记录（结束时自动清理，`-keep` 除外）。

## 参数

| 参数 | 默认值 | 说明 |
|---|---|---|
| `-server` | `http://127.0.0.1:8888` | server 基地址 |
| `-user` | `admin` | 管理员用户名（需有资产/Agent 管理权限，888） |
| `-pass` | （必填） | 管理员密码 |
| `-n` | `100` | 虚拟 agent 数量 |
| `-duration` | `60s` | 压测持续时长 |
| `-interval` | `3s` | 心跳周期（压测用短周期，生产 agent 默认 30s） |
| `-keep` | `false` | 结束后保留资产与实例记录（调试用，需手工清理） |

## 观察指标

- 脚本输出：注册成功/失败数、心跳成功数、期间在线数
- server 侧：`server/logs/` 慢请求 warn（`slow request`）、agent 实例列表在线状态
- 资源面：server 进程 CPU/内存（`top -p $(pgrep new-ops-server)`）

## 注意

- 心跳/指标 UPDATE 有 15s 节流（见 `server/plugin/agent/service/gateway.go` touchHeartbeat），压测短周期不会线性放大 DB 写压力，属预期。
- 防火墙/反代场景确认 WS 升级路径（`/agent/ws`）可达再压测。
