# 4.4 API 与 RPC 映射

[返回本章](README.md) · [全书目录](../README.md)


以下为**概念示例**（不保证对应某个现有生成类型）：

```go
// 正例：API 的 UserID 与 protoc 常见 UserId 明确对齐
out := &pb.LookupReq{UserId: req.UserID}
// 调用传递 l.ctx；返回值检查 err 后再组 API 响应
```

不要 `convert.To[T]`、反射按名映射、JSON 往返、整体强转 slice/map。指针显式取址，集合逐元素组装，int32/int64 显式转换。
消息自引用/回复等嵌套结构单独写有终止条件的递归映射，不用反射自动展开。
这样字段、权限裁剪和类型差异可以审查；但不能声称编译能发现所有语义丢字段。
存量 [common/convert](../../common/convert/) **实际仍存在**（旧材料“已移除”不属实）。全仓禁止任何反射/自动盲映射；发现仍在使用时改为显式逐字段赋值，并删除不再需要的 `common/convert` 包及旧调用。
