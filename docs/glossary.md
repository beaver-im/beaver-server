# 术语

[返回全书目录](README.md)

本表只收**跨章高频且易混淆**的名词，同一术语只在这里定义一次；各章用链接指回。事实源仍是契约、模型与网关配置，表格与源码冲突时以源码为准并同任务修正（2026-11-23 核对）。

## 小节

- 路由与鉴权
- 身份与实体
- 调用与部署


## 路由与鉴权

| 术语 | 定义 | 易混淆点 |
|---|---|---|
| 六段路由 | `/{team}/{surface}/{service}/{group}/v1/{action}`；网关按第 3 段 service 做服务发现（含 ServiceAlias），第 4 段 group 表业务/鉴权边界 | 旧材料“五段”已作废；`@server group` 名与 URL 段解耦，改路径不必重生成 |
| team / surface | URL 前两段，如 `beaver/api`、`beaver-open/api`、`beaver-admin/api`；决定鉴权首选平面和机器票密钥/受众的查表键 | 不是两个服务，也不代表两个数据库 |
| 外部票 / 机器票 | 外部票分用户/开发者/管理员三类，只在网关验；机器票（S2S）用于服务间调用，下游业务只验机器票 | 机器身份 ≠ 用户身份；网关入站注入与外部出站换票是两条路径 |
| `_public` / `_publish` | `_public` 是 **URL group 段**后缀，命中网关 `PublicPatterns` 才免外部票；`_publish` 是 **`.api` 文件名**后缀，仅表示源码里的公开面 | 文件名不会让网关放行；两者必须同时做对 |
| `xxx_s2s` / `machine` | 当前主仓与子仓并存的机器面 group 命名，均为事实差异；统一方案待裁决 | 不要同时新增两套；统一方案未决前按现有契约，不新增第三套 |

## 身份与实体

| 术语 | 定义 | 易混淆点 |
|---|---|---|
| `conversationType` | 1 私聊 / 2 群聊 / 3 圈子 | “agent 聊天”不是新会话类型，仍是 1/2 + 对方为智能机器人 |
| `UserType` | 1 普通用户；2 通知机器人（单向推送，`bot_public`）；3 智能机器人（双向，`robot` + `event`） | **Agent 不是 UserType**，是产品概念，指智能机器人后面的“大脑” |
| bot / 智能机器人 | 通知机器人只能单向推送，严格说不存在“bot 聊天”；智能机器人才能双向对话 | 群里可放任意个机器人，无需为新机器人改会话代码 |
| `AppID` / `RobotID` | `AppID`=`OpenApp.AppID`，创建应用时生成，用于开放平台鉴权；`RobotID`=`OpenAppRobot.RobotID`=`UserModel.UserID`，IM 内部用（@、会话、`targetId`） | 1:1 但用途不同，两者都不能删 |

## 调用与部署

| 术语 | 定义 | 易混淆点 |
|---|---|---|
| RPC / remote | RPC 是**本仓内**跨域调用（etcd + gRPC）；remote 是**跨独立仓**或第三方出站 HTTP（Bearer/client_credentials），在 `remote/<peer>` + `types/remote/<peer>` | remote 不是本仓 handler，不跨仓 etcd RPC；方法名精确 `XxxApi`（Python `_api`） |
| 域 / 服务 / 仓 | 域（如 account）拥有独占库；域内可有 `_api` 与 `_rpc` 两个进程；独立仓（如 `beaver-server-open`）有自己的契约、库与发布 | 不要从“一个域”推断“一个进程/一个仓”；跨服务绝不连对方库 |
| `datasync` | 按域聚合增量结果的独立服务，实时数据走本仓 RPC / 跨仓 remote，写入本域自有表 | 不是跨库拼表器，不持有他域表模型 |
| 本地开发 `-db` / 生产部署 | 根 `main.go -db` 与 `database/` 是本地建库初始化辅助；生产按 `deploy/<域名>/` 分服务物理隔离部署 | 不能拿本地共机 MySQL 当“生产共库”依据 |

新术语先在对应主章节写清语义，再回填本表；不要在这里复制整段规则。