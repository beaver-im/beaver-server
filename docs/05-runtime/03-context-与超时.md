# 5.3 context 与超时

[返回本章](README.md) · [全书目录](../README.md)


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
