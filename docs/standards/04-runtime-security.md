# 第 5 章　身份、校验、错误、日志与资源寿命

[返回目录](../README.md) · [数据库与同步](05-data-sync.md)

## 5.1 身份：知道谁写的头，才知道能不能信

**当前事实**：[网关 auth.go](../../app/gateway/gateway_api/core/auth.go) 区分 user/open/admin 外部票平面，剥离身份头、验票、按平面重建身份，再注入机器票。
[主仓 JwtS2s](../../middleware/jwt/jwtS2s.go) 只接受 Bearer 机器票，验证受众/过期等，重建 X-Client-Id 并放入 context；缺密钥或验签失败真实 401。
这不是旧材料所说的“网关无鉴权”。各服务包装中间件见服务自己的 internal/middleware。

**统一要求**：
- 需要登录的接口必须在 `.api` 声明中间件，不能把“读到 Beaver-User-Id”当成校验通过；header 仅能在受信传输/已认证链路上采信。
- 机器身份不等于用户身份；资源归属、群成员、管理员/租户权限在 logic 显式核验。请求 body/query 中的 userId 不能覆盖认证主体。
- 公开面（目标 `<业务>_publish.api`，现有 `_public.api`）只表示不挂统一身份闸门，短信码、client_secret、签名回调仍需专属校验。不要日志输出这些值。
- 不把认证失败伪装 HTTP 200。如需统一或调整业务错误 HTTP 状态码协议，直接改齐响应层与所有客户端/消费方，不留双协议兼容层（仅本次文档任务未改动源码）。

反例：用 req.UserID 任意读用户隐私，未核实是否为当前主体/是否有机器授权。
正向检查：无票→401；伪造身份头不能提权；有效机器票无用户权限仍拒绝用户私有操作；公开入口的专属凭据也必须失败关闭。
本次没有做穿透测试，不能据这段说明宣称部署安全。

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
