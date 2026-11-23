# 第 5 章　身份、校验、错误、日志与资源寿命

[返回目录](../README.md) · [数据库与同步](05-data-sync.md)

## 5.1 身份：网关验外部票，业务只验机器票

本节是 2026-11-23 核对 [网关配置](../../app/gateway/gateway_api/gateway.yaml)、[core/auth.go](../../app/gateway/gateway_api/core/auth.go) 与 [JwtS2s](../../middleware/jwt/jwtS2s.go) 后的可执行规范；旧材料的“网关只转发不鉴权”已作废。决定依据见 [ADR 0002](../decisions/0002-auth-planes.md)。配置里只有密钥**变量名**，手册不收录密钥值。

### 5.1.1 三类外部票与三条平级路径前缀

| 平面 | 首选路径前缀 | 网关验签段 | 票面受众 | 写入下游身份头 |
|---|---|---|---|---|
| 主站 C 端用户 | `/beaver/` | `Auth.AuthBeaver` | `beaver:user` | `Beaver-User-Id` |
| 开放平台开发者 | `/beaver-open/` | `Auth.AuthOpen` | `beaver:developer` | `Beaver-User-Id` |
| 后台管理员 | `/beaver-admin/` | `Auth.AuthAdmin` | `beaver:admin` | `Beaver-Admin-Id` |

- 三票**密钥各不相同、受众各不相同**，一张票只可能被一个平面验过；这是“票自描述平面”的前提。验签实现分在 `core/jwtBeaver.go`、`jwtOpen.go`、`jwtAdmin.go`，`auth.go` 只留流水线。
- `AuthRules` 就是三条平级前缀规则；未命中路径默认按 `beaver` 处理。**首选平面只决定验签顺序**，`verifyExternalTicket` 以票自身的密钥与受众为准判定真实平面；不得改成“无差别回退所有平面”——那会让开发者票也能打主站用户面。
- 网关写头口径：**头名按平面分，字段名按域分**（下游 `.api` 仍是 `UserID` / `AdminID`），不要为每个团队另造 `DeveloperID` 之类的新字段名。

### 5.1.2 单次请求的强制顺序

`core/auth.go` 的 `authenticate` 顺序即规范，改鉴权先读它：

1. **剥离**客户端自带的 `Beaver-User-Id`、`Beaver-Admin-Id`、`X-User-Id`、`X-Partner-Id`、`X-Service-Name`；公开路由也一样剥，杜绝伪造头透传。
2. 判公开：`PublicPatterns` 命中即不验票、不注入，直接转发（仍需业务自身校验）。
3. **机器票优先**：`Authorization: Bearer` 能用目标团队的机器密钥与受众验过，即视为内部调用放行；此时不覆盖原票、保留 `X-Client-Id`，因此 `machine/v1` 一类内部面无需进 `PublicPatterns`。
4. 验外部票：先验路径首选平面，失败后按 5.1.3 的唯一单向兼容再试；失败返回真实 401，不能伪装 200。
5. 写可信身份头（admin 平面写 `Beaver-Admin-Id`，其余写 `Beaver-User-Id`）。
6. 网关注入一张 S2S 机器票**覆盖** `Authorization`，下游业务只挂 `JwtS2s` 验这张票；业务不得直接采信外部票或客户端头。缺机器密钥是开发态降级，生产必须配置。

### 5.1.3 唯一的跨界兼容：开放平台面接受主站用户票

`/beaver-open/api/open/oauth/v1/` 下 `h5_authcode`、`qrcode_scan`、`qrcode_confirm`、`qrcode_cancel` 这 4 个端点是 IM 客户端自用的 C 端能力，`UserID` 取自网关注入的 `Beaver-User-Id`，空即报未登录。

- 兼容写在 **`verifyExternalTicket` 的代码层**：`preferred == open` 且开发者票验不过时，再试 `beaver` 用户票；**只有这一个方向**。反向（主站用户面收开发者票、管理面收其它票）一律拒绝。
- `AuthRules` 与 `PublicPatterns` **不写任何 action 级特例**。曾否决“迁路径到 `/beaver/` + 三段机器键”和“配置里写特例”两案：规范在配置可见层保持三条平级、零特例，业务硬约束留在代码层。
- 改动此兼容前先确认新端点是否真属 C 端；不能因为有兼容就把它当成通用开放平台规则。

### 5.1.4 公开端点与机器票的归属

- **公开端点唯一规范**：六段路由的 group 段（URL 第 4 段）以 `_public` 结尾，由 `PublicPatterns` 一条通用正则放行；**禁止 action 级特例**。换票端点内部直连、不经网关，无需放行。注意 `_publish` 只是 `.api` **文件名**后缀（见[第 3 章](02-contracts-codegen.md)），文件名不会让网关放行。
- **机器票归属按 team/surface 两段**：`teamAndSurface` 取路径前两段，`machineSecret` / `machineAudience` 的查表顺序都是 `team/surface → team → Default`。机器票只用于网关注入；**出站 S2S 不经过网关**，各服务拿 client_credentials 直连目标换票（见[第 4 章](03-boundaries-remote.md)）。一把机器密钥对应一个目标团队，不可混用。

### 5.1.5 统一要求与验收

- 需要登录的接口必须在 `.api` 声明中间件，不能把“读到身份头”当成校验通过；身份头仅能在已认证链路上采信。
- 机器身份不等于用户身份；资源归属、群成员、管理员/租户权限在 logic 显式核验。请求 body/query 中的 `userId` 不能覆盖认证主体。
- 公开面只表示不挂统一身份闸门；短信码、`client_secret`、签名回调仍需专属校验，秘密不入日志。
- 不把认证失败伪装 HTTP 200。调整对外错误码协议要改齐响应层与全部消费方，不留双协议兼容层。

验收至少覆盖：无票→401；伪造身份头不能提权；开发者票打用户面被拒；用户票打 oauth 面可用、反向不可；有效机器票也不能直接操作用户私有资源；公开入口专属凭据失败关闭；机器票验过时原 `X-Client-Id` 不被吞。本次没有做穿透测试，不能据本节宣称部署已安全。

## 5.2 校验：解析成功不等于业务合法

handler.tpl 生成 httpx.Parse 与统一响应；不要手改 handler 叠业务。
格式/必填可放契约标签；范围、状态机、资源归属、幂等约束在 logic 的副作用之前。
错误要区分空值与“明确设置为 0/空串”（更新接口可选指针字段见 user.api），不要用零值误覆盖存量。
新分页/批量请求明确上限和非法值行为，不能把未经限制的数组/页大小直接送 DB。

## 5.3 context、超时与后台任务

- API 用请求 context，RPC 用入站 ctx；继续传给 RPC/remote、GORM WithContext、HTTP 请求。不能为图省事换成 context.Background 丢取消和 trace。
- 子调用需要更短截止时间时从父 ctx 派生，`defer cancel()`；不要用无限超时去掩盖失败。
- 真正需要脱离 HTTP 生命周期的任务要有进程级受控上下文/队列、独立超时和退出机制；先明确持久性、重试与归属，不裸起 goroutine。
- [ajax.go](../../utils/ajax/ajax.go) 有调用超时和 NewRequestWithContext；这不意味着全部 DB/RPC 也都已正确传 ctx。

示例（仅说明标准库语义，周边 repo 调用按实际签名）：
```go
child, cancel := context.WithTimeout(ctx, budget)
defer cancel()
// 用 child 发起本次子调用；错误向上返回，不能偷偷换 Background 后继续写入
```

## 5.4 错误与日志

[response.Response](../../common/response/enter.go) 当前成功 code=0，普通错误 code=1、msg=err.Error()。
**尚未核实统一细分业务错误码登记表**，不要编造 `ErrXXX` API 或宣称已有完整错误体系。
新错误复用本模块真实错误模式，明确给调用方的可见文案；不要把 SQL、内部 URL、凭据或堆栈直接暴露。
若要引入细分码，应先在公共契约处定义并评审消费者，不让各 logic 私造互相冲突数字。

日志沿 [utils/beaverlog](../../utils/beaverlog/logger.go) 与模板：传 ctx、稳定操作名、必要业务 ID；
不要同时多层重复打印同一错误，也不吞错后返回成功；可以补上下文再返回，记录在哪层处理。
禁止把令牌、请求头、完整聊天内容、完整文件内容或未经脱敏配置直接写日志。发现旧代码存在错误直出或敏感日志泄露时直接修改并删除多余输出，不保留旧日志格式兼容；目前未全量审计，不标成已合规。

## 5.5 资源清理

连接池/消费者在 svc/main 的明确装配点创建；请求中只用已装配实例。
创建资源的一方负责关闭；启动失败清理已创建资源。MQ 有 [Shutdown](../../core/corerocketmq/rocketmq.go)，心跳有 [Stop](../../app/ws/ws_api/internal/logic/websocket/heartbeat/manager.go)。
所有 ticker、socket、goroutine、订阅必须有明确取消/关闭路径并测试重复关闭、客户端断线、进程退出；当前 `sendmsg` 等处如有漏传 `ctx` 情况，在代码任务中直接修复。
不要每个业务请求 Shutdown 共享池；不要假设有 Stop 方法就已经在所有退出路径被调用。
