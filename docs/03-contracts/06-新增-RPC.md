# 3.6 新增 RPC

[返回本章](README.md) · [全书目录](../README.md)


核对 [account.proto](../../app/account/account_rpc/account.proto)：`package account_rpc`、`go_package="./pb"`，user/auth 两个 service。
新进程一个 proto 包，按子域多个 service，生成加 `-m`；不能新建一个巨型 service 把不同子域混一起。
字段 snake_case、写明语义；`.proto` 字段与接口可直接做破坏性修改并删去废弃字段，同步重新生成并更新所有消费方。

```powershell
Push-Location app/account/account_rpc
# 仅查看帮助，不生成；实际生成前先核实插件与模板版本：
goctl rpc protoc --help
# 授权生成时：
goctl rpc protoc account.proto --go_out=. --go-grpc_out=. -m --zrpc_out=. --home ../../../template
Pop-Location
```

`go_package="./pb"` 时输出根是 `.`，目标为服务下 pb/；不要再写 `--go_out=./pb` 造成 pb/pb。
核对新 client/<子域>、server、logic，业务只补 logic/必要装配。API types 与 RPC pb 不是同一类型，按第 4 章显式转换。
RPC 解析/编译验证在生成链中完成；本仓没有已核实的独立 proto lint 命令，不虚构一条“验证脚本”。
message 的真实文件名是 `chat_rpc.proto`，notification/platform 注释中的旧命令有漏 `-m` 等差异，不能盲复制；旧单 service 结构可按 `-m` 多 service 拆分重生成并同步调用方。
