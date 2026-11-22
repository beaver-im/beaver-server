# 第 9 章　账号、关系、资产、通知与平台

[返回目录](../README.md) · [消息与实时链路](realtime.md)

这是一张定位地图，不逐文件复述。共同约束只维护在第 2—6 章。
每域拥有**物理隔离的独立数据库**（如 `account` 对应独立库 `beaver_account`、`asset` 对应独立库 `beaver_asset`；域内 `_api` 与 `_rpc` 共属本域库，跨服务绝不连他库、不持他库 DSN、不共享别域 models）。
每域先打开入口 import，搜动作的 `@handler`，再读对应 logic/模型及调用者，不默认读完整目录。
分类按业务可继续扩展多份 `.api`，不是下表列出的名字或两种鉴权面就限制总文件数：目标 `<业务>.api` 为登录/S2S 保护面，`<业务>_publish.api` 为公开面，统一入口生成；现存 `_public` 文件名只是现状证据，细则以[第 3 章](../standards/02-contracts-codegen.md)为准。
调用子服务/第三方走 remote；每接口 XxxApi 与 types/remote 独立 XxxReq/XxxRes 的主规范见[第 4 章](../standards/03-boundaries-remote.md)。

| 域 / 职责 | 真实入口与关键路径 | 依赖/关键词 | 变更时验证什么 |
|---|---|---|---|
| account：身份、资料、设备、登录 | [account_api.api](../../app/account/account_api/account_api.api)、[account.proto](../../app/account/account_rpc/account.proto)、[account_models](../../app/account/account_models/) | auth_public/auth/user；UserInfoLogic、JwtS2s；RPC 的 user/auth | 公开与受保护面、可信 UserID、隐私字段、设备/登录失败、远端 DTO |
| relation：好友、屏蔽、群、成员 | [relation_api.api](../../app/relation/relation_api/relation_api.api)、[relation_rpc](../../app/relation/relation_rpc/)、[relation_models](../../app/relation/relation_models/) | friend/group，跨域用户信息走账号 RPC，不读账号库 | 成员/管理权限、重复申请、退群/删除同步、通知影响 |
| asset：文件和表情 | [asset_api.api](../../app/asset/asset_api/asset_api.api)、[asset_rpc](../../app/asset/asset_rpc/)、[asset_models](../../app/asset/asset_models/) | file/emoji；存储提供方结构看 config/svc，不读真实密钥 | 所有者、上传大小/类型/路径、重复收藏、签名地址/对象删除失败 |
| notification：事件、收件箱、已读、推送 | [notification_api.api](../../app/notification/notification_api/notification_api.api)、[notification_rpc.proto](../../app/notification/notification_rpc/notification_rpc.proto)、[notification_models](../../app/notification/notification_models/) | event/inbox/read/push；现有组合客户端不等于所有生成文件同形 | 重复事件、接收人范围、已读游标、多端同步和推送失败 |
| platform：组织、工作台、反馈、日志、更新 | [platform_api.api](../../app/platform/platform_api/platform_api.api)、[platform_rpc.proto](../../app/platform/platform_rpc/platform_rpc.proto)、[platform_models](../../app/platform/platform_models/) | org/workbench/update/track 等以入口 import 为准 | 租户/管理员范围、发布策略、版本选择、日志脱敏 |

## 9.1 看一个真实的纵切面

[UserInfoLogic](../../app/account/account_api/internal/logic/user/userinfologic.go) 从 svc DB 读本域 UserModel，再逐字段生成 UserInfoRes。
它说明“本域查询可以直读本域表、业务在 logic、响应显式组装”，不表示它已满足全部安全/ctx/错误处理新规范。
其 DB 查询当前没有显式 WithContext、错误直返等问题，在对应代码任务中应直接修复到位，不留旧写法兼容（仅本次文档任务未改源码）。
装配看 [account ServiceContext](../../app/account/account_api/internal/svc/servicecontext.go)：DB、Redis、RPC、MQ、remote 在这里可定位。

## 9.2 新增一个业务能力的落地路线

例如给现有账号域增加一项资料查询：
1. 确定 user.api 的身份与资源语义；新增独立类型块/handler 声明，按第 3 章生成。
2. logic 仅读取本域独立数据库表；需要群或资产等他域数据时走对方 RPC（如 `account` ↔ `relation`、`asset` → `UserRpc`），禁止连对方库、写跨库 SQL 或 import 别域 model。
3. 真需要新表才放 account_models 对应业务子包并登记迁表；不因接口新增就自动建一个“service/repository”层。
4. 对照 HTTP Res、RPC pb、remote 消费者，分别显式映射。执行第 8 章中与此改动相符的验证，不机械跑所有服务。

测试覆盖以实际 `_test.go` 为准，本章的“验证什么”是评审要求，不是声称现有自动化已经覆盖。
