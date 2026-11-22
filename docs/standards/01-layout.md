# 第 2 章　目录、命名与代码归属

[返回目录](../README.md) · [下一章：契约与生成](02-contracts-codegen.md)

## 2.1 新业务域放在哪里

当前 account/asset/message/notification/platform/relation 是 API + models + RPC；datasync 是 API + models；ws 只有 API；gateway 是自定义代理。
新域沿用以下形状，按能力裁剪，不为凑结构创建空服务：

```text
app/<domain>/
  <domain>_api/
    <domain>_api.api             主 import 入口
    api/<business>.api           任意业务分类：登录/S2S 保护面
    api/<business>_publish.api   该分类按需增加的公开面；不是全服务只允许一对文件
    internal/{handler,logic,svc,config,types,middleware}/
    <entry>.go
  <domain>_models/<business>/    本域表，一文件一表
  <domain>_rpc/
    <contract>.proto
    pb/                         protoc 生成
    client/<subdomain>/         -m 窄客户端
    internal/{logic,server,svc,config}/
```

**要求/禁止/理由**：域下面只放 `<domain>_api/_models/_rpc` 三类业务容器（可以缺省），不要另造 `app/`、`common_models/`、含混 `models/` 来藏表。网关既有 core/types 是基础设施现状，不推广为业务域模板。独立仓或主仓有不符合之处时，在相应代码任务中直接按统一结构拆分重构并删除多余旧壳，不留兼容过渡（仅本次文档任务不改代码）。

## 2.2 按改动性质决定落点

| 要做什么 | 放哪里 | 不要放哪里 / 为什么 |
|---|---|---|
| 新增请求字段/路径 | 子 `.api`，再生成 | 不手补 types/routes；下次生成会丢失 |
| 请求解析、统一响应 | handler 模板生成 | 不在 handler 写 DB/remote/复杂业务 |
| 业务、事务编排 | 对应 `internal/logic` | 不把业务塞 svc 构造函数 |
| DB/RPC/MQ 客户端装配 | `internal/svc`；结构在 config | 不在每请求里新建连接池 |
| 表结构 | 本域 `_models/<业务>/xxx_model.go` | 不放 `_api/internal`，严禁引用其它域 models 或配置他服务 DSN 直连别库（如 `app/account` 与 `app/asset` 各自独立库） |
| 本域可复用业务辅助 | 本服务 logic 的语义子包/辅助文件 | 没有第二个稳定调用方，不先下沉全局 utils |
| 调用别人接口（remote） | remote/<peer> 的 XxxApi；types/remote/<peer> 中每接口独立 type(...) 包裹 XxxReq/XxxRes | 不直引对端源码/模型，不复用一个 CommonReq/any 覆盖所有接口；细则见第 4 章 |

生成文件分类、可编辑例外及校验详见第 3 章，不在本章重复另一套规则。

## 2.3 命名与版权

- Go 手写结构体缩写用 `ID/IDs/URL/HTTP/IP/JSON/OS`，例：`UserID`，不是 `UserId`；生成 pb 保持生成器结果，映射时显式处理。
- `.api` 字段同样写 `UserID`，json tag 用 lowerCamel：`json:"userId"`；外部 OAuth 固定 `client_id/grant_type` 按外部协议，不强改。
- `.api` 文件多词驼峰；`.proto`、表名 snake_case；model 文件保持 `xxx_model.go`。手写文件不以大写开头；生成器按默认 gozero 风格输出，不手改名与它对抗。
- `README.md/AGENTS.md` 等工具约定名、Go 的 `_test.go`、Python snake_case 是命名语义的明确例外。旧全端“多词驼峰”不能覆盖语言与契约专门规则。
- 保留邻近文件完整 `beaver-server-header-v1`，不是只留标记行。生成头优先修模板，不批量手修生成物。RPC 模板中仍有缺头者，见附录，不宣称已经统一。

## 2.4 可以直接用于评审的例子

正例：新增用户资料字段，先改 account 的 user.api → 生成 → user logic 显式组响应。
反例：新增 `app/account/common_models/extra.go`，内含 User/Device 两张表，再让 relation 直接查询。
原因：绕开本域表归属，会让目录、迁表集合、调用权限和变更影响无法追踪。
表 JSON 列的 Scan/Value 辅助类型可以与该表同文件，但不能夹带别张表。
