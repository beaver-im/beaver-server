# 10.4 datasync

[返回本章](README.md) · [全书目录](../README.md)


入口：[datasync_api.api](../../app/datasync/datasync_api/datasync_api.api)，子 import 包含 user/friend/group/chat/circle/emoji/notification/agent。
实现：[sync logic](../../app/datasync/datasync_api/internal/logic/sync/)；[ServiceContext](../../app/datasync/datasync_api/internal/svc/servicecontext.go) 装配本域 DB 与各域 RPC / `remote`；关键词 GetSync、DeletedMessages、ReadCursors、Version。
`datasync` 仅拥有本域 `datasync_models`（`beaver_datasync`）独立库；聚合用户/好友/群/消息/表情/通知全走本仓 RPC，外部 circle/agent 走 `remote`（库归属规则见[第 6 章](../06-data/)）。
新增同步项按第 6 章写游标/版本/删除语义，再更新生产者、聚合契约和客户端消费。
