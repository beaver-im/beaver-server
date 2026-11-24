# 工程目录

[返回目录](README.md) · [1 架构](01-architecture/) · [2 目录与命名](02-layout/)

**这份文件回答三个问题：东西在哪、这个文件是干什么的、我能不能改它。**

它是**现状地图**，不是规范。为什么必须放在这里、新代码该放哪、怎么命名，一律以[第 2 章](02-layout/)为准；本章只描述当前真实存在的目录与文件用途，出现冲突时以源码为准并回来修正本表。

标注约定（全表通用）：

| 标记 | 含义 | 处理方式 |
|---|---|---|
| `生成` | goctl / protoc 产物（handler、routes、types、pb、client） | **不要手改**，改契约后重新生成 |
| `手写` | 业务扩展点（logic、svc、config、models、core） | 可以改，改完遵守对应章节 |
| `配置` | etc/*.yaml、deploy/*.yaml | 改配置不改代码；不含真实密钥 |
| `遗留` | 存在但不作为模板，或位置不当 | 不模仿、不扩散，见第 9 节 |

核对基线：2026-11-23，逐目录 `ls` 与入口文件实读。目录会变，**新增/删除一级目录时同步改本表**。

## 小节

- 1 工作区
- 2 主仓根目录
- 3 app 业务域
- 4 支撑层
- 5 契约与生成
- 6 数据与运维
- 7 docs 地图
- 8 速查
- 9 遗留项


## 1 工作区

工作区根 `F:\code\mine\github` 不是 git 仓，下面 8 个是独立仓，2 个是聚合目录。

| 目录 | 类型 | 是什么 | 技术栈 |
|---|---|---|---|
| `beaver-server` | 独立仓 | **服务端主仓**，本手册的对象。gateway/account/relation/message/asset/notification/platform/datasync/ws | Go + go-zero |
| `beaver-server-package` | **聚合目录（非仓）** | 把 6 个已拆出的服务端仓放在一起，便于同时打开 | — |
| `beaver-server-social` | 独立仓（在 package 下） | 圈子与动态：moment、circle | Go |
| `beaver-server-call` | 独立仓（在 package 下） | 音视频通话：call（依赖 livekit） | Go |
| `beaver-server-open` | 独立仓（在 package 下） | 开放平台 API：open（机器人长连接落这里） | Go |
| `beaver-server-backend` | 独立仓（在 package 下） | 后台管理服务：backend_admin | Go |
| `beaver-server-online` | 独立仓（在 package 下） | 在线文档：office | Go |
| `beaver-server-ai` | 独立仓（在 package 下） | AI 服务与 Python Agent（`app/ai`、`app/agent`） | Python |
| `beaver-desktop` | 独立仓 | Electron 桌面客户端 | Node/Electron |
| `beaver-desktop-package` | **聚合目录（非仓）** | 桌面插件：`beaver-desktop-ai`、`beaver-desktop-online` | Vite/TS |
| `beaver-flutter` | 独立仓 | 移动端 App | Flutter |
| `beaver-manager` | 独立仓 | 管理后台前端（`dist/` 是产物） | Vite 前端 |
| `beaver-open` | 独立仓 | 开放平台控制台前端 | Vite 前端 |
| `beaver-oauth` | 独立仓 | OAuth 授权页前端 | Vite 前端 |
| `beaver-sdk` | 独立仓 | `packages/beaver-js-sdk`、`packages/beaver-openclaw` | TS/pnpm |
| `beaver-docs` | 独立仓 | 对外文档站（`community/`、`commercial/`、`deployment/`…） | 静态站 |

**容易踩错**：`beaver-server-open` 是**服务端**开放平台，`beaver-open` 是**控制台前端**，两个不同仓，别按名字混。
`beaver-server-package` 与 `beaver-desktop-package` 只是容器目录，没有 `.git`。

各拆仓的内部形状与主仓一致（`app/ common/ core/ deploy/ build/`），差异只有业务域和是否带 `remote/`。跨仓协作规则见[第 11 章](11-subservices/)。

---

## 2 主仓根目录

| 路径 | 类型 | 用途 |
|---|---|---|
| `main.go` | 手写 | **只是建库建表工具**：`go run main.go -db` 时创建 7 个库 + AutoMigrate + 灌种子数据；**不带 `-db` 直接 return，不是服务入口** |
| `Dockerfile` | 配置 | 通用构建镜像，服务由 `build/` 脚本选中 |
| `go.mod` / `go.sum` | 配置 | module 名仍是 `beaver`；依赖 go-zero（见第 3 章版本说明） |
| `VERSION` | 配置 | 发布版本号，`build/all.sh` 读它做镜像 tag |
| `AGENTS.md` | 手写 | **AI/协作者薄入口**，不复制整书，只列硬要求 + 任务路由 |
| `README.md` / `README_EN.md` | 手写 | 对外项目介绍；部分版本示例已滞后，以本手册为准 |
| `CONTRIBUTING.md` | 手写 | 贡献流程 |
| `SECURITY.md` / `LEGAL.md` / `LICENSE` | 手写 | 安全披露、授权与许可 |
| `CHANGELOG.md` | 手写 | 版本变更记录 |
| `app/` | 业务 | 9 个业务域，见第 3 节 |
| `common/` | 手写 | 跨域共用、无业务语义的底座，见第 4 节 |
| `core/` | 手写 | 基础设施客户端封装（MySQL/Redis/MQ/etcd/OSS…），见第 4 节 |
| `commonconfig/` | 手写 | 配置文件结构 `config.go`，各服务 `etc` 的公共部分 |
| `middleware/jwt/` | 手写 | 服务端**只挂 S2S 机器票**中间件（`jwtS2s.go`）；外部票在网关验 |
| `remote/` | 手写 | 本仓调用**别人**接口的出站层，按 peer 分目录，见第 4 节 |
| `types/remote/` | 手写 | remote 的请求/响应 DTO，与 `remote/<peer>` 一一对应 |
| `openevent/` | 手写 | 对外开放事件发布（`publisher.go`）与数字化事件（`digital.go`） |
| `utils/` | 手写 | 无状态小工具，见第 4 节 |
| `template/` | 配置 | **本仓定制的 goctl 模板**，`--home ../../../template` 指这里，见第 5 节 |
| `database/` | 手写 | 建库种子数据（file/org/platform/user），**不是 SQL 迁移目录** |
| `deploy/` | 配置 | 按业务域分的 compose 与配置，见第 6 节 |
| `docker/` | 配置 | livekit、rocketmq 等外部依赖的本地编排 |
| `build/` | 脚本 | 构建/发布/本机启动脚本，见第 6 节 |
| `scripts/` | 脚本 | 文档校验、压测、版权头工具 |
| `static/` | 资源 | 默认头像/表情/安装包等静态资源，`.gitignore` 忽略内容 |
| `docs/` | 文档 | 工程手册（不含产品需求），见第 7 节 |
| `server.md` | `遗留` | 服务端口表（内容仍有参考价值，但位置应当并入 docs） |
| `update.md` | `遗留` | 2024 年的手工更新日志，已被 `CHANGELOG.md` 取代 |
| `go-zero规范符合性审查.md` | `遗留` | 一次性审查报告，不属于活动规范 |
| `*.exe`（8 个） | `遗留` | 本机编译产物，`.gitignore` 已忽略、未入库；不要提交 |
| `skills/`、`_gen/` | 空目录 | 当前工作树为空，见第 9 节 |

端口对照（gateway 20800、account 20100/30100、ws 20400…）见仓根 [`server.md`](../server.md)；它是唯一端口表，本章不复制。

---

## 3 app 业务域

### 3.1 九个域一览

| 域 | 含哪几件 | 负责什么 | HTTP 契约入口 |
|---|---|---|---|
| `account` | api + models + rpc | 用户、认证凭据、设备、设置 | `account_api/api/auth.api`、`auth_public.api`、`user.api` |
| `relation` | api + models + rpc | 好友、群、群成员、邀请、拉黑 | `relation_api/api/friend.api`、`group.api` |
| `message` | api + models + rpc | 会话、消息、已读、转发、媒体关联 | `message_api/api/chat.api` |
| `asset` | api + models + rpc | 文件上传、表情包 | `asset_api/api/file.api`、`file_public.api`、`emoji.api` |
| `notification` | api + models + rpc | 事件、收件箱、已读、推送注册 | `notification_api/api/event.api`、`inbox.api`、`read.api`、`push.api` |
| `platform` | api + models + rpc | 组织/部门/员工、工作台、反馈、举报、埋点、版本更新、管理端 | `platform_api/api/org.api`、`workbench.api`、`admin.api`、`feedback.api`、`*_public.api` |
| `datasync` | api + models（**无 rpc**） | 多端增量同步（好友/群/消息/表情/通知/圈子/机器人） | `datasync_api/api/*.api` |
| `ws` | 只有 api | WebSocket 长连接、心跳、消息上下行 | `ws_api/api/socket.api`、`proxy.api`、`types.api` |
| `gateway` | 只有 api（自定义代理） | 唯一对外入口：按路径第二段做服务发现、验票、注入身份头、转发 | 无 `.api`，`gateway.yaml` + `core/` |

域内组合（哪些域有 rpc、哪些没有）由能力决定，不为了对称硬造空服务——规则见[第 2 章](02-layout/)。

### 3.2 一个 `<domain>_api/` 内部

以 `account_api` 为例，其它域同形状：

```text
app/account/account_api/
  account.go                 手写    进程入口 main
  account_api.api            手写    唯一 import 入口，把 api/*.api 汇总
  api/
    auth.api                 手写    业务分类：登录/S2S 保护面
    auth_public.api          手写    业务分类的公开面（文件名后缀统一为 _public）
    user.api                 手写    业务分类
  etc/account_api.yaml       配置    端口、DSN、Redis、etcd、JWT 受众
  internal/
    config/config.go         手写    yaml → struct
    svc/servicecontext.go    手写    连接池/客户端装配，进程内单例
    handler/<业务>/          生成    只做解析 + 调 logic + 统一响应
    logic/<业务>/            生成    **业务主体，生成后自己写**
    types/types.go           生成    .api 派生的 Req/Res
    middleware/              手写    本服务中间件（多数服务为空）
    handler/routes.go        生成    路由注册，勿手补
  logs/                      运行时  go-zero 日志目录，不入库
```

判断"能不能改"只看一条：**`handler`、`types`、`routes` 由 `.api` 决定，改了下次生成就丢；`logic` 是唯一稳定落点。**

### 3.3 一个 `<domain>_rpc/` 内部

```text
app/account/account_rpc/
  accountrpc.go              手写    进程入口
  account.proto              手写    契约事实源
  pb/                        生成    protoc 产物
  client/<子域>/             生成    -m 窄客户端，供其它域 import
  etc/account_rpc.yaml       配置    RPC 端口、DSN、etcd
  internal/
    config/config.go         手写
    svc/servicecontext.go    手写
    server/                  生成    gRPC 注册
    logic/<子域>/            生成    **业务主体**
```

服务端内部调用走 client，**不要跨库**（第 6 章主规范）。

### 3.4 `<domain>_models/` 内部

```text
app/account/account_models/
  auth/    auth_credential_model.go、auth_device_model.go
  user/    user_module.go、user_change_log_model.go、user_setting_model.go
```

- 一文件一表，`xxx_model.go`，`手写`。
- 表定义还决定`main.go -db` 的 AutoMigrate 集合，**加表必须同步改 `main.go`**，否则本地建库缺表。
- 只放本域表，不引他域 model（第 6 章）。

### 3.5 两个特殊域

**`gateway`**（无 `.api`，不走 goctl）：

| 文件 | 用途 |
|---|---|
| `gateway.go` | 进程入口 |
| `gateway.yaml` | 三条平级前缀规则 + 机器票 team/surface 映射 + 上游服务表 |
| `core/gateway_api.go` | 路由匹配与反向代理主流程 |
| `core/auth.go` | 剥身份头 → 验票 → 按平面重建身份头 → 注入机器票 |
| `core/jwtBeaver.go` / `jwtOpen.go` / `jwtAdmin.go` | 三类外部票的验签实现 |
| `core/proxy.go` | 上游发现（etcd）、转发、响应回写 |
| `types/config.go` | 配置结构 |

鉴权顺序与平面定义只在[第 5 章](05-runtime/)与 [ADR 0002](13-decisions/0002-auth-planes.md) 写一次，这里只标文件职责。

**`ws`**（单 API、额外自造类型）：

| 路径 | 用途 |
|---|---|
| `ws_api/api/socket.api` | 连接、心跳、上下行消息帧 |
| `ws_api/api/proxy.api` | 内部代理转发 |
| `ws_api/api/types.api` | WS 帧类型声明 |
| `ws_api/types/` | 手写 WS 结构：`msg.go`、`stream.go`、`typing.go`、`ws.go`、`proxy.go` |
| `ws_api/response/response.go` | WS 专用响应封装，不复用 HTTP 的 |
| `ws_api/internal/logic/websocket/` | 连接管理、心跳、路由 |
| `ws_api/.rocketmq_client_go/` | **RocketMQ 客户端运行时目录**，不是源码 |

---

## 4 支撑层

### 4.1 `common/`（跨域共用，无业务语义）

| 目录 | 内容 | 用途 |
|---|---|---|
| `const/` | `usertypeconst`、`chattypeconst`、`notifyconst`、`webhookconst`、`agenttopicconst`、`mqwsconst`、`websocket` | 枚举与常量，**唯一取值来源** |
| `convert/` | `convert.go`（含测试） | 类型转换工具 |
| `response/` | `enter.go` | HTTP 统一响应信封 |
| `middleware/` | `middleware.go` + `grpc/`、`http/`、`ua/`、`utils/` | 通用中间件 |
| `models/` | `enter.go`、`ctype/` | 公共模型基类（时间戳、软删） |
| `etcd/` | `etcd.go` | 注册中心客户端 |
| `identity/` | `identity.go` | 从 context 取当前身份 |
| `list_query/` | `enter.go` | 列表查询参数归一 |
| `traceid/` | `traceid.go` | 链路 ID |
| `sse/` | `sse.go` | SSE 流式响应 |
| `wsEnum/` | `wsCommandConst/`、`wsTypeConst/` | WS 指令枚举 |
| `zrpc_interceptor/` | 客户端/服务端拦截器 | RPC 身份与调用方透传 |
| `observability/`、`validator/` | 空目录 | 见第 9 节 |

### 4.2 `core/`（基础设施客户端）

| 目录 | 封装 |
|---|---|
| `coregorm/` | GORM 初始化（MySQL） |
| `coreredis/` | Redis |
| `corerocketmq/` | RocketMQ 生产者/消费者 |
| `coreetcd/` | etcd 服务注册发现 |
| `coreopensearch/` | OpenSearch（`opensearch.go`、`knowledge.go`） |
| `corepush/` | 推送通道 |
| `coreonline/` | 在线状态 |
| `coreupdate/` | 版本发布（`release.go`） |
| `corewebhook/` | Webhook 投递 |
| `version/` | 版本号注入 |

`core/coredb/` 是空目录，见第 9 节。

### 4.3 `remote/` 与 `types/remote/`（出站调用）

```text
remote/<peer>/
  remote.go       手写    Client 构造、换票、Post 封装
  <business>.go   手写    业务方法，名字精确以 Api 结尾
types/remote/<peer>/
  <peer>.go       手写    每个接口独立 type(...) 包 XxxReq / XxxRes
types/remote/token.go     换票请求/响应
```

当前 6 个 peer：`agent`、`ai`、`circle`、`doc`、`moment`、`open`，各自对应一个拆仓。类型组规则只在[第 4 章](04-boundaries/)写一次。

### 4.4 `utils/`（无状态工具）

| 目录 | 用途 | 目录 | 用途 |
|---|---|---|---|
| `ajax/` | HTTP 响应助手 | `md5/` | 摘要 |
| `authlock/` | 登录失败锁定 | `pwd/` | 密码哈希 |
| `beaverlog/` | 日志封装（含 `model/`） | `rand/` | 随机 |
| `conversation/` | 会话 ID 与城市辅助 | `s2s/` | S2S 调用辅助 |
| `device/` | 设备与 UA 解析（`device.go`、`register.go`、`uaprofile.go`） | `secret/` | 加解密（含测试） |
| `email/` | 邮件发送 | `uuid/` | ID 生成 |
| `jwts/` | JWT 签发/解析 | `list/` | 列表处理 |
| `authdevice/`、`id_generator/`、`location/` | 空目录 | | |

`utils/` 只放**没有业务语义**的东西；有业务含义的辅助留在本域 logic 的子包（第 2 章 2.2）。

---

## 5 契约与生成

### 5.1 `.api` 文件分布（当前 39 份，含 9 份汇总入口）

```text
app/account/account_api/api/       auth.api  auth_public.api  user.api
app/relation/relation_api/api/     friend.api  group.api
app/message/message_api/api/       chat.api
app/asset/asset_api/api/           file.api  file_public.api  emoji.api
app/notification/notification_api/api/  event.api  inbox.api  push.api  read.api
app/platform/platform_api/api/     admin.api  feedback.api  org.api  workbench.api
                                   instance_public.api  track_public.api  update_public.api
app/datasync/datasync_api/api/     agent  chat  circle  emoji  friend  group  notification  user
app/ws/ws_api/api/                 proxy.api  socket.api  types.api
```

各域顶层的 `<domain>_api.api` 只做 import 汇总，**不写路由**。文件名后缀与 URL group 后缀现已统一为 `_public`：前者是源码组织约定，后者才是网关放行依据；新增公开接口两者都要做对（文件名本身不参与鉴权）——判定规则见[第 3 章](03-contracts/)。

### 5.2 `.proto` 文件分布（6 份）

`account.proto`、`relation.proto`、`asset.proto`、`message_rpc/chat_rpc.proto`、`notification_rpc/notification_rpc.proto`、`platform_rpc/platform_rpc.proto`。`datasync`、`ws`、`gateway` 没有 RPC。

### 5.3 `template/`（定制 goctl 模板）

| 子目录 | 覆盖什么 |
|---|---|
| `api/` | `handler.tpl`、`logic.tpl`、`context.tpl`、`etc.tpl`、`config.tpl`、`main.tpl` |
| `rpc/` | `logic.tpl`、`logic-func.tpl`、`call.tpl`、`config.tpl`、`etc.tpl`、`main.tpl` |
| `model/` | CRUD 方法与错误模板 |
| `mongo/` | Mongo 模型模板 |
| `newapi/` | 新 API 骨架 |
| `docker/` | 镜像模板 |
| `kube/` | deployment / job |
| `gateway/` | 网关 etc 与 main |

生成命令、cwd 与生成物边界见[第 3 章](03-contracts/)；`template/` 是**生成行为的唯一来源**，改模板优先于批量手修生成物。

---

## 6 数据与运维

### 6.1 `database/`（建库种子，不是 SQL）

| 文件 | 用途 |
|---|---|
| `file/init.go` | 默认文件记录 |
| `org/init.go` | 单租户组织初始化（`platform_api` svc 调用） |
| `platform/init.go` | 平台基础数据 |
| `user/seed.go` | 默认账号与凭据 |

以上由 `main.go -db` 与 `platform_api` 启动时调用。**没有版本化 SQL 迁移**，表结构靠 AutoMigrate + 人工升级——风险与规范见[第 6 章](06-data/)。

### 6.2 `deploy/`

```text
deploy/<域>/                account asset datasync gateway message notification platform relation ws
  *_api.yaml / *_rpc.yaml   生产配置
  docker-compose.yaml       单域编排
deploy/dist/                启动产物目录：start.sh stop.sh update.sh config.yaml + 各域
```

部署目录按**业务域名**命名，镜像/容器/二进制仍是 `*_api`。配置项与密钥边界见[第 7 章](07-deploy/)。

### 6.3 脚本

| 脚本 | 用途 | 危险度 |
|---|---|---|
| `build/all.sh` | 全量构建 + 打 tag + 推送镜像 | **高**（会 push） |
| `build/build.sh` | 单服务镜像构建 + 打 tag + 推送镜像 | **高**（会 push） |
| `build/initMysql.bat` | 调 `go run main.go -db` 建库 | **高**（建表/灌数据） |
| `build/start_api.bat` / `start_rpc.bat` | 本机批量起服务 | 中 |
| `scripts/check-knowledge.ps1` | 文档链接/路径/哈希校验 | 无 |
| `scripts/add-copyright-header.mjs` | 批量补版权头 | 中（改源码） |
| `scripts/stress/` | 压测场景、观测、单机脚本 | 中（有环境成本） |

未经授权不运行构建、建库、发布、压测（AGENTS.md 第 5 条）。

### 6.4 `static/` 与 `docker/`

- `static/`：`audio`、`beaver`、`exe`、`image`、`mobile`、`sponsor`、`video`。内容不入库（`.gitignore`），只保留目录约定。
- `docker/livekit/`：`livekit.yaml`，通话依赖。
- `docker/rocketmq/broker/`：MQ 本地编排。

---

## 7 docs 地图

本目录是**工程手册**，不收产品需求 / 功能规格 / 路线图。

| 路径 | 内容 |
|---|---|
| `README.md` | 全书入口 + 结构标准 |
| `directory.md` | 本文件：工程目录（参考） |
| `glossary.md` | 术语（参考） |
| `01-architecture/` … `13-decisions/` | 一章一目录；入口 `README.md`；编号 `1` / `1.1` / `1.1.1` |
| `knowledge-checks.json` | 校验登记 |

---

## 8 速查

| 我要…… | 先去 | 再看 |
|---|---|---|
| 加一个 HTTP 接口 | 对应域 `*_api/api/<业务>.api` | 第 3 章七步法 |
| 改接口返回字段 | `.api` → 重新生成 → `logic` | 第 3 章 |
| 写业务逻辑 | `*_api/internal/logic/<业务>/` | 第 8 章 |
| 加一张表 | `*_models/<业务>/xxx_model.go` **+ `main.go` 的 AutoMigrate** | 第 6 章 |
| 加服务内部 RPC | `*.proto` → 生成 → `internal/logic` | 第 3 章 |
| 调别的仓的接口 | `remote/<peer>/` + `types/remote/<peer>/` | 第 4 章 |
| 改登录/鉴权 | `gateway/core/auth.go` + `gateway.yaml` | 第 5 章、ADR 0002 |
| 改 WS 帧或心跳 | `ws_api/api/*.api` + `ws_api/types/` | 第 10 章 |
| 加 MQ 消费者 | `message_rpc/internal/logic/consumer` 或 `core/corerocketmq` | 第 10 章 |
| 改增量同步 | `datasync_api/api/*.api` + `internal/logic/sync` | 第 10 章 |
| 改端口/DSN/Redis | `app/<域>/*/etc/*.yaml` | 第 7 章 |
| 改部署编排 | `deploy/<域>/` | 第 7 章 |
| 改生成模板 | `template/<类>/` | 第 3 章 |
| 建库建表跑本地 | `main.go -db` / `build/initMysql.bat` | 第 6 章 |
| 补文档 / 查某规则出处 | `docs/README.md` → 对应章 | — |
| 改其它仓（social/call/open/ai…） | `beaver-server-package` 下对应仓 | 第 11 章 |

**通用检索路径**：契约动作 → `@handler` → `internal/logic` → 必要调用者。不要默认扫全仓（AGENTS.md 末段）。

---

## 9 遗留项

这些是**现状问题**，不构成模板；在对应代码任务中直接清掉，不扩散。

| 项 | 问题 | 建议 |
|---|---|---|
| 根目录 8 个 `.exe` | 本机编译产物（`main.exe`、`beaver.exe`、`ws_api.exe` 等），已 gitignore | 保留在本地即可，**不要入库** |
| `server.md` | 端口表内容有效但放在仓根 | 后续并入 `docs/`，仓根只留入口 |
| `update.md` | 2024 年手写日志，与 `CHANGELOG.md` 重复 | 内容并入 CHANGELOG 后删除 |
| `go-zero规范符合性审查.md` | 一次性审查报告挂在根目录 | 结论吸收进第 3 章后删除 |
| `_gen/`、`skills/` | 空目录 | 确认无用途则删除 |
| `core/coredb/`、`common/observability/`、`common/validator/`、`utils/{authdevice,id_generator,location}` | 空目录（go-zero 骨架残留） | 确认后删除，避免误导"这里有实现" |
| `ws_api/.rocketmq_client_go/` | 客户端运行时目录，混在源码树里 | 加进 `.gitignore`，别当源码读 |
| `*/logs/` | 运行时日志目录 | 已忽略；排查问题时才看 |
| 生成物与手写物同目录 | `handler`、`types`、`routes`、`pb` 与手写文件混放 | 只按本文第 3.2/3.3 的标记判断，**改前先确认是不是生成物** |

判定"这是不是生成物"的最快方法：看文件头是否有 `Code generated by goctl/protoc`，以及是否在 `*_api/api` 或 `*.proto` 里有对应声明。
