# 第 3 章　契约优先：HTTP / RPC 与 goctl 生成

[返回目录](../README.md) · [边界与类型映射](03-boundaries-remote.md)

## 3.1 最重要的一句话

**先写 `.api` / `.proto`，再用 `goctl` 生成接线，最后在业务扩展点实现。**
工具准确名称是 `goctl`（Go Control 的工具名），不是 `goctrl`；不要把拼写错误写入自动化。
本机帮助已核对 `api validate / api go / rpc protoc` 及其参数；本次没有执行任何生成。

### 环境和可复现性

- `go.mod`：Go 1.26.8、go-zero v1.7.4；本机 goctl 1.10.2。没有发现项目锁定 goctl 的配置。
- API 使用仓库 [template/api](../../template/api/)；RPC 用 [template/rpc](../../template/rpc/)。不要漏 `--home` 而吃个人目录中的模板。
- RPC 另需 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`。本次只发现 protoc 命令位置，未核实插件版本/兼容性，不提供未经验证的安装命令。
- 记录本次生成器版本、模板 diff、工作目录、命令与生成差异。后续若要固定工具版本，应单独评审；不能从框架版本推导工具版本。

## 3.2 文件归属：不能手写什么，可以改什么

| 文件 | 归属 | 如何修改 |
|---|---|---|
| `.api`、`.proto` | 契约源，手写 | 先改这里，兼容性与权限一起评审 |
| `internal/handler/routes.go`、`internal/types` | API 生成物 | 不手改，改契约/模板再生成 |
| HTTP handler | 接线生成物 | 不手建等价 handler；解析/响应共性改 handler.tpl；接口个性业务校验进 logic |
| `pb/*.pb.go`、RPC server/client | RPC 生成物 | 不手改；proto、模板、生成参数解决 |
| `internal/logic` | 生成骨架 + 手写业务 | 允许改真实业务方法，不能把“生成过”理解为永不允许编辑 |
| `internal/svc`、`internal/config`、服务 main | 生成起点 + 业务装配扩展 | 允许必要依赖装配、配置结构、启动清理；核对生成是否保留扩展，不能盲目覆盖 |
| models、remote、业务辅助 | 手写能力 | 遵守本域归属、契约类型与调用边界 |
| `template/` | 公共生成行为源 | 改前说明影响的服务；先小范围验证，不能顺手全仓覆盖 |

废弃接口先修改契约声明并核对调用方；不得擅自删除已有 handler/logic。清理旧文件须明确范围、核实生成差异并获得对应授权。

存量：WS 的握手需要原始 HTTP 对象，已有专用接线；不能把这类差异当作新接口手写 handler 的普遍许可。
模板自身也并非已完全合规：`handler.tpl` 无响应分支的调用签名与 `response.Response` 当前签名不一致；RPC 模板缺完整头。
模板本身有缺陷或冗余时，直接修改模板并用 `goctl` 重生成全部受影响服务，不保留两套模板兼容；当前未运行生成，不能承诺现有契约无差异再生成。

## 3.3 HTTP 契约：任意业务分类 × 保护面/公开面

**用户最新规范（目标）**：不是每服务只有两个 `.api` 文件，也不是整个服务只能有一个 `xxx`。
`xxx` 是可扩展的业务类别，同一类业务放一起。类别数量不限；分类有不同鉴权面时再分开，不机械生成空文件。

| 文件名（目标规范） | 放什么 | 不能怎么用 |
|---|---|---|
| `<业务>.api` | 该类别需要登录鉴权的接口、S2S 调用接口；按本项目受保护链路声明中间件 | 不因增加分类而把保护接口复制到公开面 |
| `<业务>_publish.api` | 该类别无需登录/统一身份闸门的公开接口 | 后缀精确为 **`_publish`**，不是我替用户改成 `_public`；公开不是可泄露数据或绕资源/签名校验 |

目录为**分类扩展示意，不声称这些新文件已经创建**：

```text
account_api/
  account_api.api          同一服务统一 import 入口
  api/
    auth.api               身份业务：登录/S2S 保护面
    auth_publish.api       身份业务：公开面（登录/注册/换票等依其业务语义）
    user.api               用户业务：保护面
    user_publish.api       用户业务：仅确实需要公开能力时增加
    device.api             设备业务：保护面，可继续扩展其它类别
```

**当前事实与目标不要混写**：[account_api.api](../../app/account/account_api/account_api.api) 现在实际 import `auth_public.api`、`auth.api`、`user.api`；
本次窄查主仓发现现有公开分类使用 `_public.api`，没有 `_publish.api`。用户这次明确要求的目标后缀是 `_publish`，文档据此更新，**源码本次未改名**。
后续契约任务直接重命名文件、更新 import 并生成，清掉旧路径，不保留两套后缀做兼容。
文件后缀是源码维护约定，不由 go-zero 自动决定鉴权；**必须核对实际 `@server` 中间件声明**。
本仓典型保护链路为网关校验用户身份并注入机器票，下游 JwtS2s 验机器 Bearer；不要误认为 JwtS2s 直接校验用户登录 JWT。
S2S（服务间调用）属于调用/认证方式，不是“对外公开”的同义词；换票不走已有身份闸门，但仍校验自身凭据。

### 分类、合并、生成的统一规则

1. 主入口只 import（可含 syntax/说明），不塞业务 type/handler；多个业务类别并列 import 到**同一个服务入口**，不是每个类别另造一个服务 main。
2. 同业务同鉴权的接口放一起；不同业务即使鉴权相同也可拆文件。每个业务子文件一个 `@server` 块，保护与公开分开。
   分类是业务 × 鉴权面，不存在“只有鉴权不同才允许拆文件”的限制。
3. 同一服务的不同子契约可沿用同名 `service account`，goctl 汇总路由；`@server group` 决定生成目录，既不自动加 URL 前缀，也不决定新进程/新服务。
   group 应按业务清楚分区；生成 handler/logic 包名遵循仓库模板，工具 common 子包保留自身包名。
4. 每个 `@handler` 一组独立 `type (...)`，含 `XxxReq`、`XxxRes`、接口专属子结构；同文件多接口就有多组，不共用一个 CommonReq/any。
   Res 可空但须声明，不放 code/msg；统一响应层加 `{code,msg,result}`。Go 字段 `UserID`、本仓自有 JSON `userId`，只用 `//` 注释。
5. 拆分/合并文件时 handler 与其 Req/Res/专属子结构整体移动，更新主 import，删除旧声明；不能复制一份留下重复路由/重名 type。
6. 从服务 cwd 对**主入口**运行第 3.5 节的 validate 和 goctl api go，并带本仓模板；不要逐个分类文件运行生成，避免各自覆盖 routes/types 或生成多个 main。
   `template/api/routes.tpl` 的 RegisterHandlers 汇总入口正是此组织方式。分类新增后仍只有一条主服务生成命令，不直接改 handler/routes/types。
7. 公开接口不走登录鉴权不代表不校验：按业务验证输入、验证码/签名/client_secret（适用时）、资源可见性与防滥用；秘密与内部管理字段不得作为公开结果返回。

文件名不是 URL：不能把 `_publish` 自动拼为新公开路径，也不能为目录分类顺手撤掉受保护接口的中间件。
保留必要的当前协议/业务安全支持；不保留只为旧版本存在的兼容路由。分类契约调整要同步消费者并删除废弃路径。

## 3.4 路由与旧材料冲突

新业务沿**已核对的六段**：`/{team}/{surface}/{service}/{group}/v1/{action}`。
team 如 beaver；surface 是 api/admin；service 用合并域 account/relation/message/asset/platform 等；
group 用业务、xxx_public、xxx_s2s 或当前契约已有 machine；action 用 snake_case。
禁止新建 `:id` 路径参数，也不要把 `/v1/qrcode_status` 再拆成 `/v1/qrcode/status`。
旧域 auth/user→account、friend/group→relation、file/emoji→asset、chat→message、org/workbench→platform。

旧 Cursor 入口却说“五段消灭六段”，与旧详细规范和当前路由相反，已记录为被事实纠正的旧句。
旧详细规范要求机器面 `xxx_s2s`，当前部分契约用 `machine`。因本项目无历史包袱、无需兼容旧版线上路径：一旦选定统一命名，直接改契约、重生成并同步所有调用方（本仓/子仓 `remote` 与校验），删除旧路由，**绝不同时保留 `machine` 和 `xxx_s2s` 两套兼容路由**。

## 3.5 新增 HTTP 的七步（主仓 account 示例）

1. 读第 9 章定位，先选业务类别，再明确登录/S2S 保护面还是公开面；不是限制全服务只有一对文件。
2. 同类别保护接口放 `<业务>.api`，公开放 `<业务>_publish.api`；新增分类加入主入口 import，拆合时整体移动 handler/类型组并删除旧声明，此时不手建 Go 路由/类型。
3. 在 **仓根 PowerShell** 进入服务目录并校验（只解析契约）：

```powershell
Push-Location app/account/account_api
goctl api validate --api account_api.api
# 非 0 立即停止，修契约，不带病生成
# 需要用户授权生成，并确认未提交/并发修改后才执行下一条：
goctl api go --api account_api.api --dir . --home ../../../template
Pop-Location
```

4. 查看 git diff（含未跟踪文件）：新增 handler/routes/types 必须可追溯到契约/模板；发现意外额外 main 先确认，再只清理本次误生成者，不能删历史业务文件。
5. 在对应 logic 实现业务，必要时补 svc/config/models/remote；不把 DB/remote 放 handler。
6. 按第 8 章测试矩阵检查鉴权失败、空字段、跨域调用/事务等，核查调用方 DTO。
7. 更新受影响章节/交接，记录生成命令与版本。正式代码任务在独立干净工作区验证再次生成无意外差异；本次文档任务未做此操作。

## 3.6 新增 RPC（account 示例）

核对 [account.proto](../../app/account/account_rpc/account.proto)：`package account_rpc`、`go_package="./pb"`，user/auth 两个 service。
新进程一个 proto 包，按子域多个 service，生成加 `-m`；不能新建一个巨型 service 把不同子域混一起。
字段 snake_case、写明语义；由于无历史二进制兼容性负担，`.proto` 字段与接口可直接做破坏性修改并删去废弃字段，同步重新生成并更新所有消费方，不需要保留旧字段或双字段兼容。

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
message 的真实文件名是 `chat_rpc.proto`，notification/platform 注释中的旧命令有漏 `-m` 等差异，不能盲复制；旧单 service 结构可直接按 `-m` 多 service 拆分重生成并同步调用方，不保留旧单 service 兼容壳。

## 3.7 不要把发布脚本当生成脚本

已查看 build/ 与 scripts/：没有发现统一 API/RPC 生成 wrapper；当前依据是契约注释 + 本章明确的 cwd/命令。
`build/build.sh` 会 Docker 构建并推送，**不是生成器**。不得为了“校验一下”运行。
独立仓 backend/call 的模板相对深度与主仓不同，详见第 11 章。
