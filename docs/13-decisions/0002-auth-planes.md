# 0002　鉴权平面、公开端点与机器票归属

- 状态：已采纳（2026-11-23，Asia/Shanghai；本次按现有契约与代码核对后补记，不冒充原始作者的设计动机）
- 依据：用户确认的终态；事实源为 [gateway.yaml](../../app/gateway/gateway_api/gateway.yaml)、[core/auth.go](../../app/gateway/gateway_api/core/auth.go)、[types/config.go](../../app/gateway/gateway_api/types/config.go) 与开放平台 `oauth.api` 的 4 个端点。链接里的 gateway.yaml 只作路径证据，本 ADR 不收录任何密钥值。
- 实施细则：[第 5 章](../05-runtime/)。

## 上下文

网关是外部票据唯一的验票点；下游业务只挂 `JwtS2s` 验机器票。系统同时存在三类外部调用方（主站 C 端用户、开放平台开发者、后台管理员），以及一类内部调用方（服务间 S2S）。开放平台下还有 4 个实质上给 IM 客户端用的 oauth 端点，路径平面与票面平面不一致。旧材料称“网关不鉴权”、旧配置曾把机器面写进公开路径，均已与现状冲突。

## 决定

1. **三票三平面**：用户票 `beaver:user`、开发者票 `beaver:developer`、管理员票 `beaver:admin`；验签密钥分三段 `Auth.AuthBeaver` / `Auth.AuthOpen` / `Auth.AuthAdmin`，各不相同，可独立轮换。
2. **三条平级 team 前缀**：`/beaver/` → 用户票、`/beaver-open/` → 开发者票、`/beaver-admin/` → 管理员票（保留 `api` 作 surface 段）。`AuthRules` 只给首选验签顺序，票自身的密钥与受众决定真实平面；未命中默认 `beaver`。
3. **单向跨界兼容**：只有开放平台面额外接受主站用户票，覆盖 `/beaver-open/api/open/oauth/v1/` 的 `h5_authcode`、`qrcode_scan`、`qrcode_confirm`、`qrcode_cancel`。反向一律拒绝。兼容写在 `verifyExternalTicket` 代码里，配置层零 action 特例。
4. **公开端点唯一规范**：URL 的 group 段以 `_public` 结尾即网关放行，`PublicPatterns` 只保留一条通用正则；禁止 action 级特例。`.api` 文件名后缀现同样统一为 `_public`，但仍只是源码组织约定，不参与网关判定。
   变更记录：本 ADR 2026-11-23 补记时文件名后缀写的是 `_publish`，2026-11-24 由用户拍板统一为 `_public`（拼写统一、层级职责不变）；此处按新规范更新规范条目，不改写当时的历史决定与否决理由。
5. **机器票归属按 team/surface 两段**：`teamAndSurface` 取路径前两段，`machineSecret` / `machineAudience` 查表顺序为 `team/surface → team → Default`。机器票优先校验，验过即放行且不覆盖原票（`X-Client-Id` 保留）；网关对外票据验完后另签机器票覆盖 `Authorization` 交给下游。出站 S2S 不经网关，各服务直连目标换票。
6. **身份头按平面分、字段名按域分**：admin 写 `Beaver-Admin-Id`，其余写 `Beaver-User-Id`；下游 `.api` 字段仍是 `UserID` / `AdminID`。

## 备选方案与否决理由

- **迁 oauth 四端点到 `/beaver/` 并给机器票加三段键**：会让路径规范出现特例，也让机器票归属从 team/surface 变成按 action 细分；否决。
- **在 `AuthRules` 写 action 级特例**：规范散落到配置细节，其它 open 端点容易被误放行；否决。改为代码层单向兼容，配置保持三条平级。
- **验票失败后无差别回退所有平面**：开发者票可打主站用户面，构成跨平面越权；否决。
- **把 open 面 `Req.UserID` 改名为 `DeveloperID`**：头名按平面分、字段名按域分是两个正交维度，改名只制造混乱；已回退。

## 后果

- 正面：配置可见层只有三条规则、一条公开正则；轮换某一类票不影响其它类；票自描述平面，无需为每个特例加路由判断。
- 负面：新增开放平台端点若实际面向 C 端，必须先判定是否走那一个单向接受用户票的规则，不能默认放开；机器密钥与受众必须按同一套 team/surface 映射同步维护，配错即全链路验不过。
- 不改变：各服务数据库独占与物理隔离、`.api` 先契约后生成、业务在 logic 做资源归属校验。

## 撤销条件（何时重新评估）

- 网关不再统一验外部票，或下游服务改为直接验多类外部票时；
- 开放平台 C 端自用端点迁出 `/beaver-open/`，单向兼容失去存在理由时；
- 机器票改为由独立授权服务统一签发、且网关不再参与注入时。

满足任一条件应写新的 ADR 取代本条，不直接改写已接受的记录。

## 验证计划与实际结果

- 结构核对：`gateway.yaml` 的 Auth / PublicPatterns / AuthRules 三段、`auth.go` 的验票顺序与单向兼容分支、开放平台 4 条 oauth 路径，均已逐项读源码确认（2026-11-23）。
- 未运行：真实票据穿透测试、错误码与 401 的端到端验证、机器密钥轮换演练。本节不宣称部署已安全。
- 后续变更以[第 8 章验证矩阵](../08-workflow/)选择实际检查，不把路径存在当作验证通过。
