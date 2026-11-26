# 4.2 remote

[返回本章](README.md) · [全书目录](../README.md)


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
发现旧方法名直接改名并同步调用方，删除旧名包装。

- 业务函数在 `<biz>.go`：拼 `commonconfig.<Peer>Url + 完整路径`，传 ctx/header/Req，调用本 peer 通道，不散落到 logic 拼 URL/取票。
- `remote.go` 仅通道，logic 调业务 XxxApi；handler 不直接调 remote。每 peer 有自己的通道，不先造跨所有线路的大一统 RemoteClient。
- remote 不重试、不聚合、不缓存业务结果，只缓存带过期提前量的换票；重试/幂等决策属于显式业务编排，不能在通道隐藏。
- 主仓与独立仓之间优先 Bearer/client_credentials；不造 `/internal/*` 转发壳，不跨仓 etcd RPC、不导入对端业务包。
  已不需要的静态票兜底直接同步改提供方/消费者并删除，不维持双鉴权。
