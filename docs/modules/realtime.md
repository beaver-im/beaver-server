# 第 10 章　消息、队列、WebSocket 和同步

[返回目录](../README.md) · [独立子服务](subservices.md)

## 10.1 message：持久化与消息编排

入口：[message_api.api](../../app/message/message_api/message_api.api)、[chat_rpc.proto](../../app/message/message_rpc/chat_rpc.proto)。
业务：[sendmsglogic.go](../../app/message/message_rpc/internal/logic/message/sendmsglogic.go)，模型在 message_models/chat。
辅助：[core/version](../../core/version/version.go)、[core/corerocketmq](../../core/corerocketmq/rocketmq.go)。
关键词：`SendMsg`、`BuildMsgFromProto`、`notifyMessageUpdateGrouped`、`SendMessage`。

定位流程：API chat logic → 对应 RPC 方法 → message logic → 本域写入/会话更新 → MQ/通知分支。
新增消息类型时核对 `.api/.proto`、`common/models/ctype`、显式映射和客户端载荷；不能只加一个枚举就算完成。
原文反射映射“已移除”的说法与 `common/convert` 实际存在矛盾；全仓禁盲映射，在代码任务中直接改成显式逐字段赋值并移除 `common/convert`。

## 10.2 RocketMQ 和 Agent 异步分支

消费者在 [message consumer](../../app/message/message_rpc/internal/logic/consumer/)：`StartConsumer`、agenttask/agentdone。
MQ 工具提供 SendMessage、RegisterConsumer/RegisterRawConsumer、Shutdown。
需区分：请求上下文内发送、后台独立消费、DB 提交后发消息，这三者的寿命/失败语义不同。
新消费逻辑必须说明幂等键、重投与部分失败；不能消费异常仍一律标成功，也不能把 remote 通道变成无限重试层。
当前 sendmsg 有 `context.Background()` 的推送支路，这是待审查存量，不是第 5 章新代码丢 ctx 的示范。

## 10.3 ws：握手、连接、分发、清理

入口：[ws_api.api](../../app/ws/ws_api/ws_api.api)、[ws.go](../../app/ws/ws_api/ws.go)。
实际接线：[websocket/enter.go](../../app/ws/ws_api/internal/logic/websocket/enter.go)，同目录 auth/conn/handler/heartbeat 分别帮助定位。
[heartbeat/manager.go](../../app/ws/ws_api/internal/logic/websocket/heartbeat/manager.go) 有 Stop，不能据此推断所有退出路径都正确清理。

事实差异：ws_api.api 注释说“单链 import”，当前实际列了 types/socket/proxy 三个 import；types.api 还保留共享结构与 ConversationId 等旧命名。
本次文档任务未改动源码；在对应代码任务中，`ws_api.api` 的共享类型与不一致导入应直接按契约独立规则整改、重新生成并同步调用方，不留旧类型兼容。
新增 WS 命令要列：命令常量、请求结构、身份/设备检查、消息分发、断线清理、重复回包；不要每个 handler 临时建全局连接表。
网关代理需要 Hijack/Flush 透传的事实见 core/proxy.go，改代理不能只验证普通 HTTP。

## 10.4 datasync：按域聚合而不是跨库拼表

入口：[datasync_api.api](../../app/datasync/datasync_api/datasync_api.api)，子 import 包含 user/friend/group/chat/circle/emoji/notification/agent。
实现：[sync logic](../../app/datasync/datasync_api/internal/logic/sync/)；[ServiceContext](../../app/datasync/datasync_api/internal/svc/servicecontext.go) 装配本域 DB 与各域 RPC / `remote`；关键词 GetSync、DeletedMessages、ReadCursors、Version。
`datasync` 仅拥有本域 `datasync_models`（`beaver_datasync`）独立库；聚合用户/好友/群/消息/表情/通知全走本仓 RPC，外部 circle/agent 走 `remote`，严禁连接其他服务数据库或引进对端表模型做 SQL JOIN。
新增同步项按第 6 章写游标/版本/删除语义，再更新生产者、聚合契约和客户端消费。

## 10.5 这条链路怎么验收

按真实改动选择：重复 MessageID、发送失败不假成功、DB 成功/MQ 失败窗口、消费重投、断线重连、心跳超时、资源关闭、撤回/删除、增量补拉和多端并发。
本次只定位链路和代表性源码，没有真实发消息、连 WS、跑队列或做端到端测试。
