# 服务端口表

> 当前基线：**VERSION 3.0.2**，核对日期 **2026-11-26**。本表是端口的唯一汇总入口，实际监听值以各服务 `etc/*.yaml`（子仓同路径）为准；配置模板见 `deploy/config.yaml` 与各域 `deploy/<域>/`。

## 主仓 beaver-server

| 服务 | API 端口 | RPC 端口 | 所属仓库 | 备注 |
|------|---------:|---------:|----------|------|
| gateway_api | 20800 | - | beaver-server | 客户端 API 网关：验票 / 注入身份头 / 转发 |
| account | 20100 | 30100 | beaver-server | 用户 + 认证（已合并） |
| relation | 20200 | 30200 | beaver-server | 好友 + 群组（已合并） |
| message | 20300 | 30300 | beaver-server | 聊天（chat 更名） |
| ws | 20400 | - | beaver-server | WebSocket 长连接 |
| push | 20500 | - | beaver-server | 独立推送网关（S2S 薄入口，无数据表），从 notification_rpc 剥离 |
| asset | 20600 | 30600 | beaver-server | 文件 + 表情（已合并） |
| notification | 21000 | 31000 | beaver-server | 通知 |
| platform | 21600 | 31100 | beaver-server | 平台（组织 / 工作台 / 反馈 / 举报 / 埋点 / 更新等） |
| datasync | 21700 | - | beaver-server | 数据同步 |

## 独立子仓

子服务均为**单进程**，没有独立 RPC 端口；跨仓调用走 HTTP 机器票视图（`machine` / `_public`），不跨仓使用 etcd gRPC。

| 服务 | API 端口 | RPC 端口 | 所属仓库 | 备注 |
|------|---------:|---------:|----------|------|
| call | 21800 | - | beaver-call | 通话（LiveKit，原 call_rpc 已并入 call_api） |
| open | 21900 | - | beaver-open | 开放平台 API / 控制台 / WS / MQ（原 open_rpc 已取消） |
| social | 22100 | - | beaver-social | 圈子 + 动态（单进程双库 beaver_circle / beaver_moment） |
| agent | 22300 | - | beaver-server-ai | AI / 数字伙伴（Python，agent_api） |
| office | 22400 | - | beaver-server-online | 在线文档（HTTP + Yjs WS，原 doc_rpc 已取消） |
| backend_admin | 40000 | - | beaver-server-backend | 后台管理（自带 JWT 鉴权 / CORS） |

## 已废弃端口（不要再引用）

| 旧写法 | 现状 |
|--------|------|
| moment 20900 / 30800、circle 32100 | 已合并为 `social` 22100，单进程双库 |
| call 31800 | `call_rpc` 已并入 `call_api` 21800 |
| open 30900 | `open_rpc` 已取消，统一 `open` 21900 |
| doc 31300 | `doc_rpc` 已取消，在线文档为 `office` 22400 |
| group 20500 | `group` 已并入 `relation`；20500 现由 `push` 使用 |

> 端口变更需同步改 `app/<域>/*/etc/*.yaml`（子仓同理）、`deploy/<域>/` 配置与 `gateway.yaml` 的服务别名；工程约定见 [docs/07-deploy](docs/07-deploy/README.md)。
