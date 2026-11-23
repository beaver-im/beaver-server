# 10.1 message

[返回本章](README.md) · [全书目录](../README.md)


入口：[message_api.api](../../app/message/message_api/message_api.api)、[chat_rpc.proto](../../app/message/message_rpc/chat_rpc.proto)。
业务：[sendmsglogic.go](../../app/message/message_rpc/internal/logic/message/sendmsglogic.go)，模型在 message_models/chat。
辅助：[core/version](../../core/version/version.go)、[core/corerocketmq](../../core/corerocketmq/rocketmq.go)。
关键词：`SendMsg`、`BuildMsgFromProto`、`notifyMessageUpdateGrouped`、`SendMessage`。

定位流程：API chat logic → 对应 RPC 方法 → message logic → 本域写入/会话更新 → MQ/通知分支。
新增消息类型时核对 `.api/.proto`、`common/models/ctype`、显式映射和客户端载荷；不能只加一个枚举就算完成。
原文反射映射“已移除”的说法与 `common/convert` 实际存在矛盾；全仓禁盲映射，在代码任务中直接改成显式逐字段赋值并移除 `common/convert`。
