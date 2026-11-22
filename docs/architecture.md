# 第 1 章　系统、层次与数据流

[返回全书目录](README.md) · 下一章：[目录与归属](standards/01-layout.md)

## 1.1 项目是什么

当前 `go.mod` 的 module 为 `beaver`，Go 指令 `1.26.8`，go-zero 依赖 `v1.7.4`。
本机 `goctl --version` 为 `1.10.2 windows/amd64`：**框架依赖版本不等于生成器版本**。
仓库未发现 goctl 固定版本声明；它是尚待补齐的可复现性边界，不能擅自写成“项目锁定 1.10.2”。
原 README 的 Go 1.24 徽标和旧服务树不能用来判断当前工具链。

## 1.2 实际运行结构

```text
桌面 / 移动端 / 管理端 / 外部机器调用
       ↓ HTTP / WebSocket（具体暴露面按网关配置与契约）
 gateway_api：自定义代理、票据处理、身份头、路由
       ↓ 转发 + 下游身份闸门
 <domain>_api：.api → handler → logic → 本域 models/DB
                                      ├→ 仓内其它域 RPC
                                      ├→ remote/<peer> → 独立仓业务 HTTP
                                      └→ RocketMQ → WS / 异步消费者
 <domain>_rpc：.proto → server → logic → 本域 models/DB
 datasync_api：按领域 RPC / remote 聚合增量结果，不跨域直连表
```

**数据库独占与物理隔离前提**：每个业务服务拥有独立数据库，按不在同一台服务器物理隔离部署设计（例如 `app/account` 拥有独立库 `beaver_account`，`app/asset` 拥有独立库 `beaver_asset`）。同一业务域内的 `_api` 与 `_rpc`（如 `account_api` 与 `account_rpc`）同属该域并访问本域独占库；**不同服务之间禁止连接对方数据库、禁止跨库 SQL/JOIN/事务、禁止持有他服务 DSN**，跨服务只走 RPC 或 `remote`/S2S。不得把开发机偶然共用同一个 MySQL 实例当成数据库共享。

**重要纠偏**：旧材料称“网关只转发、不鉴权”，但当前
[网关 auth.go](../app/gateway/gateway_api/core/auth.go) 有票据验证、身份头剥离/重建、机器票注入；
[JwtS2s](../middleware/jwt/jwtS2s.go) 验证机器 Bearer。不能据旧句删除网关校验，也不能认为请求头天然可信。
参见 [第 5 章](standards/04-runtime-security.md)。

## 1.3 进程入口和基础设施

| 路径 | 职责与边界 |
|---|---|
| `app/*/*_api`、`app/*/*_rpc` | 业务进程，各有自己的 main / internal / 配置结构 |
| [gateway.go](../app/gateway/gateway_api/gateway.go) | 网关入口，现有自定义 core/types，没有同形 `.api` 入口；不是手写业务接口的许可 |
| [account.go](../app/account/account_api/account.go) | 典型 API 入口；其它业务看模块地图 |
| [accountrpc.go](../app/account/account_rpc/accountrpc.go) | 典型 RPC 入口 |
| [main.go](../main.go) | 仅作**本地开发/初始化**的 `-db` 建库迁表工具现状，不是生产共库保证，也不是运行整个服务集群的入口 |
| [database/](../database/) | 本地建库/表集合初始化辅助（与根 main 存在重叠）；生产部署中各服务物理隔离且仅持有自有库权限 |
| [core/](../core/) | GORM、Redis、MQ、搜索、推送、版本等基础设施连接/适配 |
| [common/](../common/) | 常量、身份、响应、拦截器等共享约定，不是跨域业务的收容所 |
| [remote/](../remote/)、[types/remote/](../types/remote/) | 调用别人的 HTTP 接口：每接口 XxxApi 与独立 type(...) 内的 XxxReq/XxxRes；不是本服务生成的 handler/types，见第 4 章 |
| [template/](../template/) | 受版本控制的 goctl 模板；生成行为的修改入口 |

## 1.4 独立仓不是本仓的子目录模块

[六个子服务包索引](modules/subservices.md) 指向同一工作区的独立仓。
不直接 import 对端 Go module 的业务模型，不共享数据库连接，不让 etcd RPC 跨仓。
核心仓、子服务、客户端的契约变更必须列出生产者和消费者；本次只在本仓建文档，没有改那些仓库。

发现层次混杂、失效入口或废弃结构时按第 2—4 章直接重构并删除旧路径，不加兼容包装或临时转发壳。

## 1.5 本书不做的承诺

没有全量复核历史审查报告的“全部归零”，没有运行生成/测试/数据库迁移。
存在命名、共享类型、RPC 单/多 service、鉴权描述等存量差异，详见附录；旧文档中的通过记录不代表这次通过。
