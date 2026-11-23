# 3.3 HTTP 契约

[返回本章](README.md) · [全书目录](../README.md)


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

**当前事实与目标不要混写**：[account_api.api](../../app/account/account_api/account_api.api) 现在实际 import `auth_public.api`、`auth.api`、`user.api`，源码后缀未改名；`_publish`（文件名）与 `_public`（URL group）的分工、整改方向见本章 3.4 与[第 5 章](../05-runtime/)。
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
分类契约调整要同步消费者并删除废弃路径；不为已废弃版本单独留路由。
