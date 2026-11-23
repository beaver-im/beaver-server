# 4.3 Req / Res

[返回本章](README.md) · [全书目录](../README.md)


**一个接口 = 一个 `type (...)`，里面是该接口独立的 `XxxReq` 和 `XxxRes`**；空响应也写 `XxxRes struct{}`。
这是 Go 的**分组类型声明**，不是嵌套类、泛型包装或“每接口一个 Go package”。一个业务 `.go` 文件可放多组 type，前提是每接口一组、各有 Req/Res。
类型位于 `types/remote/<peer>/`，不塞 `remote.go`、本服务生成的 `internal/types` 或其它域 models，不给名称加桌面 `I` 前缀。

- 不同接口即使字段一样，也各有自己的 XxxReq/XxxRes，不拿 CommonReq/CommonRes/map[string]any 代替；接口专属 Item 等子结构跟本接口一组。
- **可共享的领域基础类型**仅限语义稳定、归属清楚的枚举、ID 值类型等；不能把整份业务 Req/Res 或随响应演变的 Item 冒充基础类型复用。
  本仓自持外部 DTO，不 import 对端仓业务模型。不是“任何基础类型绝对不可复用”，也不是“所有接口共用一个 any”。
- Go 缩写用 `ID`，**JSON tag/字段语义必须逐项对齐对端真实契约**。例如实际对端是 `doc_id` 时不能只为本仓 lowerCamel 偏好改成 userId/docId。
  本仓自有契约规范化需改齐提供方与消费者，不在一端私改 JSON。
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
   第三方有不同协议时按明确的供应商协议解析，不擅改第三方协议。
5. 新增出站接口的顺序：读对端契约 → 建本仓独立类型组 → 写 XxxApi → logic 调用/处理失败 → 验证 URL/tag/信封/超时取消/鉴权失败 → 同步受影响调用方和手册。
   本仓目前没有已验证的统一跨仓契约对比脚本，不虚构命令；依据第 8 章实际测试条件执行，不把存在路径当测试通过。
