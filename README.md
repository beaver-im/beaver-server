# 🦫 Beaver IM - 企业级即时通讯平台

> 开发与 AI 协作请先读 **[中文工程手册](docs/README.md)**：go-zero 先写 `.api` / `.proto`，再用 goctl 生成接线。旧版本与目录示例以当前源码和手册核验说明为准。

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-3.0.2-blue.svg)](VERSION)
[![Go](https://img.shields.io/badge/go-1.26.8-blue.svg)](https://golang.org/)
[![Go-Zero](https://img.shields.io/badge/Go--Zero-v1.7.4-green.svg)](https://github.com/zeromicro/go-zero)
[![QQ群](https://img.shields.io/badge/QQ群-1013328597%2B-blue.svg)](https://qm.qq.com/q/82rbf7QBzO)

> **海狸 IM 后端服务** - 基于 Go-Zero 微服务架构，为 Flutter 移动端、Electron 桌面端与后台管理系统提供 REST / WebSocket / gRPC 能力。

**当前版本：[3.0.2](VERSION)**（以仓库根目录 [`VERSION`](VERSION) 为准；最后核对：2026-11-26）

[English](README_EN.md) | [中文](README.md)

---

## ✨ 核心能力

- 🔐 **认证授权** - 手机号 / 邮箱 / 验证码 / 扫码 / OAuth 登录，JWT 外部票经网关验签并以 ES256 断言向业务服务传递身份，服务侧用 JWKS 公钥验签；S2S 机器票、登录设备管理与踢下线
- 💬 **即时通讯** - 私聊与群聊、会话与消息存储、已读回执、消息转发、WebSocket 实时推送与 RocketMQ 异步链路
- 👥 **社交关系** - 好友申请与资料、群组创建与管理、屏蔽；圈子与动态由独立仓 `beaver-social` 提供
- 📁 **文件与表情** - 文件上传下载、表情包与收藏，对象存储支持 MinIO / 七牛云
- 📞 **音视频通话** - 基于 LiveKit 的通话信令与房间管理（`beaver-call`）
- 🔔 **通知与推送** - 系统通知、互动提醒、收件箱与已读；`push_api` 独立推送网关，实时 / 离线两路按 channel 映射 MQ Topic
- 🔄 **跨端同步** - `datasync` 多端增量同步，覆盖好友 / 群 / 消息 / 表情 / 通知 / 圈子 / 机器人等
- 🏢 **组织与工作台** - 部门 / 员工、工作台、反馈 / 举报、埋点、版本发布等平台能力
- 🤖 **AI 与知识** - 数字伙伴 / 机器人、知识源与技能（`SKILL.md`）、Agent 会话增量同步（`beaver-server-ai` + 开放平台）
- 📄 **在线文档** - HTTP API + Yjs WebSocket 同进程协作（`beaver-server-online`）
- 📊 **平台运维** - 客户端日志经 RocketMQ 写入 OpenSearch，OpenSearch 已开启安全认证并支持 Basic Auth 客户端
- 🛠️ **后台管理** - 用户 / 组织 / 版本等管理端接口（`beaver-server-backend` 的 `backend_admin`）

## 🛠️ 技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| 语言 / 框架 | Go 1.26.8 + go-zero 1.7.4 | goctl 是独立工具，版本以本地实际安装为准，见工程手册 1.1 |
| 通信 | gRPC、HTTP S2S、WebSocket、RocketMQ | 本仓服务间走 RPC，跨仓与推送走 HTTP 机器票 |
| 存储 / 检索 | MySQL 8.0、Redis 6.0、ETCD 3.5、OpenSearch | 多库拆分 + 缓存 + 服务发现 + 日志检索 |
| 部署 | Docker Compose、build 脚本 | 配置在 `deploy/<域>/`，外部依赖在 `docker/` |
| 其他 | JWT（ES256 / JWKS）、LiveKit、MinIO / 七牛云 | 鉴权、通话、对象存储 |

## 🏗️ 架构概览

客户端（Flutter / Electron）统一经 `gateway_api :20800` 访问：网关校验外部票（Beaver / Open / Admin 三类）并签发 ES256 断言，注入可信身份头后按路径转发；业务服务以 JWKS 验签，跨服务 / 跨仓调用统一使用 S2S 机器票与 `machine` / `_public` 视图。WebSocket 负责长连接上下行，RocketMQ 承担异步消息与推送投递。后台管理端经网关访问 `backend_admin`。

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Flutter 移动端 │  │ Electron 桌面端│  │  后台管理 Web │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       └─────────────────┼─────────────────┘
                         │
              ┌──────────▼──────────┐
              │  gateway_api :20800  │
              └──────────▼──────────┘
                         │
   ┌─────────────────────┼─────────────────────┐
   │                     │                     │
account / relation   message / ws          asset / notification
platform / datasync      push :20500        （RPC 仅本仓服务间）
   │                     │                     │
   └─────────────────────┼─────────────────────┘
                         │
        ┌────────────────┴────────────────┐
        │ MySQL · Redis · ETCD · OpenSearch │
        └─────────────────────────────────┘

  跨仓调用 ──► social / call / open / office / agent / backend（HTTP 机器票视图）
```

### 服务端口

以 [`server.md`](server.md) 为唯一端口表，实际监听以各服务 `etc/*.yaml` 为准。主仓 9 个业务域 + 独立推送服务；子仓均为单进程，原先的 moment / circle / call_rpc / open_rpc / doc_rpc 端口已不再使用。

| 服务 | API 端口 | RPC 端口 | 所属仓库 | 说明 |
|------|---------:|---------:|------|------|
| gateway_api | 20800 | - | beaver-server | 客户端 API 网关，验票 / 转发 |
| account | 20100 | 30100 | beaver-server | 用户 + 认证（已合并） |
| relation | 20200 | 30200 | beaver-server | 好友 + 群组（已合并） |
| message | 20300 | 30300 | beaver-server | 聊天（chat 更名） |
| ws | 20400 | - | beaver-server | WebSocket 长连接 |
| push | 20500 | - | beaver-server | 独立推送网关（S2S 薄入口） |
| asset | 20600 | 30600 | beaver-server | 文件 + 表情（已合并） |
| notification | 21000 | 31000 | beaver-server | 通知 |
| platform | 21600 | 31100 | beaver-server | 平台（组织 / 反馈 / 埋点 / 更新等） |
| datasync | 21700 | - | beaver-server | 数据同步 |
| call | 21800 | - | beaver-call | 通话（LiveKit，单进程） |
| open | 21900 | - | beaver-open | 开放平台（API / 控制台 / WS / MQ） |
| social | 22100 | - | beaver-social | 圈子 + 动态（单进程双库） |
| agent | 22300 | - | beaver-server-ai | AI / 数字伙伴（Python，agent_api） |
| office | 22400 | - | beaver-server-online | 在线文档（HTTP + Yjs WS） |
| backend_admin | 40000 | - | beaver-server-backend | 后台管理（自带 JWT 鉴权 / CORS） |

## 📁 项目结构

```
beaver-server/
├── app/                    # 9 个业务域：account relation message asset
│   │                       # notification platform datasync ws gateway（+ push）
│   ├── account/            # account_api + account_rpc + account_models
│   ├── relation/           # relation_api + relation_rpc + relation_models
│   ├── message/            # message_api + message_rpc + message_models
│   ├── asset/              # asset_api + asset_rpc + asset_models
│   ├── notification/       # notification_api + notification_rpc + notification_models
│   ├── platform/           # platform_api + platform_rpc + platform_models
│   ├── datasync/           # datasync_api + datasync_models（无 RPC）
│   ├── ws/                 # ws_api（无 RPC）
│   ├── push/               # push_api（无 RPC、无数据表）
│   └── gateway/            # gateway_api（自定义代理，无 .api）
├── common/                 # 常量、身份、响应、拦截器等跨域底座
├── core/                   # MySQL / Redis / MQ / etcd / OpenSearch / 推送等客户端封装
├── commonconfig/           # 各服务 etc 的公共配置结构
├── middleware/             # JWT 等 HTTP 中间件
├── remote/                 # 调用其他仓的出站层，按 peer 分目录
├── types/                  # remote 的请求 / 响应 DTO
├── openevent/              # 对外开放事件发布
├── utils/                  # 日志、JWT、S2S、密码、邮件等无状态工具
├── template/               # 本仓定制的 goctl 模板
├── database/               # 建库种子数据（不是 SQL 迁移目录）
├── deploy/                 # 按业务域的 Compose 与配置
├── build/                  # 构建 / 发布 / 本机启动脚本
├── docs/                   # 工程手册（13 章 + 目录 / 术语）
├── main.go                 # 仅建库建表工具：go run main.go -db
└── server.md               # 服务端口说明
```

## 🚀 快速开始

详细的环境准备、配置说明与部署步骤见文档站及本地工程手册第 7 章：

- 📖 [后端开发文档](https://wsrh8888.github.io/beaver-docs/backend/)
- 📖 [部署说明](https://wsrh8888.github.io/beaver-docs/backend/deploy/build-scripts.html)
- 📖 [本地工程手册 · 配置与部署](docs/07-deploy/README.md)

本地初始化数据库（需先配置 MySQL 连接；`main.go` **只是建库工具，不是服务入口**）：

```bash
go run main.go -db
```

服务构建与编排见 `build/` 与 `deploy/<域>/`；`VERSION` 用于构建脚本打镜像 tag。

## 📚 文档与资源

- 📖 **工程手册**: [docs/README.md](docs/README.md)（架构 / 契约 / 边界 / 身份 / 数据 / 部署 / 决策记录）
- 📖 **详细文档**: [https://wsrh8888.github.io/beaver-docs/](https://wsrh8888.github.io/beaver-docs/)
- 📝 [变更记录](CHANGELOG.md) · [更新说明](update.md)
- 🎥 **视频教程**: [B站频道](https://space.bilibili.com/269553626/lists)
- 📱 **体验包下载**: [海狸 IM Android 体验包](https://github.com/wsrh8888/beaver-docs/releases/download/lastest/latest.apk)
- 💬 **QQ 群**: [1013328597](https://qm.qq.com/q/82rbf7QBzO)

## 🔗 相关项目

以下仓库现统一位于 GitHub `beaver-im` 组织。`beaver-server-package/` 与 `beaver-desktop-package/` 只是子仓的本地容器目录，不是独立仓。

| 类别 | 项目 | 仓库地址 | 说明 |
|------|------|----------|------|
| **主仓** | beaver-server | [GitHub](https://github.com/beaver-im/beaver-server) | 后端微服务主仓（本仓库） |
| **客户端** | beaver-flutter | [GitHub](https://github.com/beaver-im/beaver-flutter) | 移动端（Flutter，推荐） |
| | beaver-desktop | [GitHub](https://github.com/beaver-im/beaver-desktop) | 桌面端（Electron） |
| | beaver-manager | [GitHub](https://github.com/beaver-im/beaver-manager) | 后台管理系统 |
| | beaver-oauth | [GitHub](https://github.com/beaver-im/beaver-oauth) | OAuth 授权登录 |
| **服务端子包** | beaver-server-ai | [GitHub](https://github.com/beaver-im/beaver-server-ai) | AI Agent（Python） |
| | beaver-server-backend | [GitHub](https://github.com/beaver-im/beaver-server-backend) | 后台管理接口 `backend_admin` |
| | beaver-server-call | [GitHub](https://github.com/beaver-im/beaver-server-call) | 音视频通话（LiveKit） |
| | beaver-server-online | [GitHub](https://github.com/beaver-im/beaver-server-online) | 在线文档（office_api） |
| | beaver-server-open | [GitHub](https://github.com/beaver-im/beaver-server-open) | 开放平台 / 机器人 |
| | beaver-server-social | [GitHub](https://github.com/beaver-im/beaver-server-social) | 圈子与动态（social_api） |
| **客户端子包** | beaver-desktop-ai | [GitHub](https://github.com/beaver-im/beaver-desktop-ai) | 桌面端外置 AI 能力包（Vite / TS） |
| | beaver-desktop-online | [GitHub](https://github.com/beaver-im/beaver-desktop-online) | 桌面端外置在线文档能力包（Vite / TS / Vue） |

## 🤝 贡献指南

欢迎通过 Issue / Pull Request 参与贡献，详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📄 开源协议与免责声明

本项目基于 [MIT](LICENSE) 协议开源，详见 [LICENSE](LICENSE)。

**使用要点（摘要）：**

- 闭源自用商用、二次开源均可免费，但须保留根目录 `LICENSE`，上线前端须有「关于」署名（基于海狸 IM + 仓库地址）
- 闭源交付第三方、去掉署名、对外 SaaS 收费等，请采购商业授权（书面合同）
- 无论是否付费，**不得删除或篡改 `LICENSE`**

完整免责与署名要求：[LEGAL.md](LEGAL.md)  
商业授权产品线与报价：[版权与商业授权](https://wsrh8888.github.io/beaver-docs/community/license.html)  
联系：[751135385@qq.com](mailto:751135385@qq.com)

## ⭐ Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=beaver-im/beaver-server&type=Date)](https://star-history.com/#beaver-im/beaver-server&Date)

## ☕ 请作者喝杯茶

如果这个项目对你有帮助，欢迎请作者喝杯茶 ☕

<div align="center">
  <img src="./static/sponsor/wechat.jpg" width="200" alt="微信赞助码"/>
  <img src="./static/sponsor/zhifubao.jpg" width="200" alt="支付宝赞助码"/>
</div>

---

<div align="center">
  <strong>Made with ❤️ by Beaver IM Team</strong><br>
  <em>企业级即时通讯平台</em>
</div>
