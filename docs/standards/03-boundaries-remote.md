# 第 4 章　模块边界、跨仓 S2S 与显式类型映射

[返回目录](../README.md) · [运行与安全](04-runtime-security.md)

## 4.1 调用方向，先判断再写 import

| 场景 | 允许 | 禁止 |
|---|---|---|
| 本域 API 访问本域表 | logic → 本域 models，经 svc DB | 为“分层”强制绕一圈调自己 RPC |
| 主仓域 A → 域 B | 已有 etcd/gRPC 客户端 | 直接连 B 库、import B 的模型当 DTO |
| 主仓 ↔ 独立仓 | remote/S2S，即业务 HTTP + Bearer 机器票 | 跨仓 etcd RPC、导入对方业务包、不必要 `/internal/*` 门面 |
| 独立仓内部 | 依本仓真实契约与边界 | 将别仓模型/凭据编入本仓 |

各业务域数据库独占、按不同服务器物理隔离部署；跨库直连与他服务 DSN 的禁令见[第 6 章](05-data-sync.md)，本地开发机共机不构成共享前提。

当前实例：[UserInfoLogic](../../app/account/account_api/internal/logic/user/userinfologic.go) 直接查询本域独占表；
[ServiceContext](../../app/account/account_api/internal/svc/servicecontext.go) 装配跨域 RPC 和 open remote；
[asset_api 配置](../../app/asset/asset_api/internal/config/config.go) 只配置本域 `Mysql.DataSource` 并通过 `UserRpc` 访问账号能力。
不要把“仓内互调用 RPC”错解为本服务查询自己的表也必须 RPC；同域的 `_api` 与 `_rpc` 共享本域独占库，跨服务才走 RPC / `remote`。

## 4.2 remote 就是“本仓调用别人接口”的出站层

remote 是调用方向，不是本服务提供 HTTP 接口的 handler，也不是本仓 RPC。
对 Beaver 其它独立仓调用称 remote/S2S；对第三方外部服务同样是出站接口，但认证、HTTP 状态和响应字段必须服从对端真实协议，不能凭空给第三方套本仓信封。

```text
remote/<peer>/
  enter.go          Init：装配本 peer 客户端，不在此执行业务
  remote.go         Post/Get/Delete、带票、换票缓存等通道工具
  token.go          GetTokenApi：如对端有换票接口
  <biz>.go          按同类业务维护 XxxApi 出站方法
 types/remote/<peer>/
  <biz>.go          每接口独立的 type(...)，各自 XxxReq + XxxRes
 commonconfig/     对端地址与配置注册结构；不收录公开凭据值
```

**方法后缀精确为 `Api`**：`GetDocApi` / `GetDocContentApi`，不是 `GetDoc`、`GetDocAPI`、`GetDocapi`，也不是桌面 I 前缀。
所有对外出站接口方法（含 GetTokenApi）都适用；`Post/Get/Delete/Init/authHeader` 是内部通道/装配工具而非对端业务端点，不机械加后缀。
发现旧方法名直接改名并同步调用方，删除旧名包装，不留兼容双入口。

- 业务函数在 `<biz>.go`：拼 `commonconfig.<Peer>Url + 完整路径`，传 ctx/header/Req，调用本 peer 通道，不散落到 logic 拼 URL/取票。
- `remote.go` 仅通道，logic 调业务 XxxApi；handler 不直接调 remote。每 peer 有自己的通道，不先造跨所有线路的大一统 RemoteClient。
- remote 不重试、不聚合、不缓存业务结果，只缓存带过期提前量的换票；重试/幂等决策属于显式业务编排，不能在通道隐藏。
- 主仓与独立仓之间优先 Bearer/client_credentials；不造 `/internal/*` 转发壳，不跨仓 etcd RPC、不导入对端业务包。
  已不需要的静态票兜底直接同步改提供方/消费者并删除，不能以历史兼容维持双鉴权。

## 4.3 每个出站接口必须有独立 Req / Res 类型组

**一个接口 = 一个 `type (...)`，里面是该接口独立的 `XxxReq` 和 `XxxRes`**；空响应也写 `XxxRes struct{}`。
这是 Go 的**分组类型声明**，不是嵌套类、泛型包装或“每接口一个 Go package”。一个业务 `.go` 文件可放多组 type，前提是每接口一组、各有 Req/Res。
类型位于 `types/remote/<peer>/`，不塞 `remote.go`、本服务生成的 `internal/types` 或其它域 models，不给名称加桌面 `I` 前缀。

- 不同接口即使字段一样，也各有自己的 XxxReq/XxxRes，不拿 CommonReq/CommonRes/map[string]any 代替；接口专属 Item 等子结构跟本接口一组。
- **可共享的领域基础类型**仅限语义稳定、归属清楚的枚举、ID 值类型等；不能把整份业务 Req/Res 或随响应演变的 Item 冒充基础类型复用。
  本仓自持外部 DTO，不 import 对端仓业务模型。不是“任何基础类型绝对不可复用”，也不是“所有接口共用一个 any”。
- Go 缩写用 `ID`，**JSON tag/字段语义必须逐项对齐对端真实契约**。例如实际对端是 `doc_id` 时不能只为本仓 lowerCamel 偏好改成 userId/docId。
  本仓自有契约规范化需改齐提供方与消费者，不在一端私改 JSON 后假装兼容。
- 现有 [token.go](../../types/remote/token.go) 仍共用 TokenRequest/TokenResponse，是**当前未整改差异，不是最新规范豁免**；换票也是接口，目标应按 peer/接口定义 GetTokenReq/GetTokenRes 并同步调用方，清掉不再需要的共享旧包装。

### 真接口的类型与请求实现示例

来源为 [remote/doc/content.go](../../remote/doc/content.go) 与 [types/remote/doc/doc.go](../../types/remote/doc/doc.go) 的真实 `GetDocContentApi`。
下面展示按目标 Go 缩写规范整理后的写法（现存源码为 DocId，文档示例用 DocID），**保留当前源码的 JSON tag 与端点不变**，不是新造接口或声称当前 Go 字段已改：

```go
// types/remote/doc/doc.go：一个接口一组，字段取自现有对端 DTO。
type (
    GetDocContentReq struct {
        DocID string `json:"doc_id,omitempty"`
    }
    GetDocContentRes struct {
        DocID       string `json:"doc_id,omitempty"`
        Title       string `json:"title,omitempty"`
        ContentText string `json:"content_text,omitempty"`
        Version     int32  `json:"version,omitempty"`
        Status      int32  `json:"status,omitempty"`
    }
)
// GetDocApi 另有自己独立的 type ( GetDocReq ...; GetDocRes ... )，不复用本组。
```

```go
// remote/doc/content.go：实际函数签名与通道调用（节选；导入在原文件）。
func GetDocContentApi(ctx context.Context, header http.Header, body *doctype.GetDocContentReq) (*doctype.GetDocContentRes, error) {
    return Post[doctype.GetDocContentRes](ctx, commonconfig.DocUrl+"/beaver-doc/api/doc/machine/v1/get_doc_content", header, body)
}
```

### 调用、超时、鉴权、错误各放哪里

1. logic 调 `GetDocContentApi(l.ctx, header, req)` 并检查 error 后使用 Res；传请求 ctx，不用 Background 丢取消。Req 按上例目标字段构造，不能照抄为现存 DocID 已可编译。
2. peer 的 [doc/remote.go](../../remote/doc/remote.go) 负责带票/换票缓存；业务调用方不拼凭据、不把 secret/token 写日志，也不把用户票与机器票混用。
3. [utils/ajax/ajax.go](../../utils/ajax/ajax.go) 的调用预算当前为 10 秒，底层 NewRequestWithContext 保持取消；必要时从父 ctx 再派生更短 deadline 并 cancel，不另建无限超时 HTTP 客户端。
4. Beaver 各仓按标准 `{code,msg,result}`：XxxRes 只描述 result；ajax.decode 把非 2xx 或 HTTP 200 + code 非 0 转 StatusError。不要每个业务方法重复解信封或吞错回空成功。
   第三方有不同协议时按明确的供应商协议解析，不是为旧版专设的兼容分支，不擅改第三方协议。
5. 新增出站接口的顺序：读对端契约 → 建本仓独立类型组 → 写 XxxApi → logic 调用/处理失败 → 验证 URL/tag/信封/超时取消/鉴权失败 → 同步受影响调用方和手册。
   本仓目前没有已验证的统一跨仓契约对比脚本，不虚构命令；依据第 8 章实际测试条件执行，不把存在路径当测试通过。

## 4.4 API ↔ RPC：显式逐字段，禁止盲映射

以下为**概念示例**（不保证对应某个现有生成类型）：

```go
// 正例：API 的 UserID 与 protoc 常见 UserId 明确对齐
out := &pb.LookupReq{UserId: req.UserID}
// 调用传递 l.ctx；返回值检查 err 后再组 API 响应
```

不要 `convert.To[T]`、反射按名映射、JSON 往返、整体强转 slice/map。指针显式取址，集合逐元素组装，int32/int64 显式转换。
消息自引用/回复等嵌套结构单独写有终止条件的递归映射，不用反射自动展开。
这样字段、权限裁剪和类型差异可以审查；但不能声称编译能发现所有语义丢字段。
存量 [common/convert](../../common/convert/) **实际仍存在**（旧材料“已移除”不属实）。全仓禁止任何反射/自动盲映射；发现仍在使用的旧代码直接替换为显式逐字段赋值并删除不再需要的 `common/convert` 包及旧调用，不留兼容兜底。
