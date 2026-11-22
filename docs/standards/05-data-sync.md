# 第 6 章　数据库、事务、升级与增量同步

[返回目录](../README.md) · [配置部署](06-deployment.md)

## 6.1 数据库独占所有权、物理隔离与表归属

**硬性边界**：每个业务服务拥有完全独立的数据库，且按**不在同一台服务器上的物理隔离部署**设计。
例如：`app/account` 是独立数据库（`beaver_account`），`app/asset` 是独立数据库（`beaver_asset`）；**禁止连接别人数据库直接调用**。

- **域与进程的关系（不误定“每进程一库”）**：按仓内现状，`app/<domain>` 下的 `<domain>_api` 与 `<domain>_rpc`（如 `account_api` + `account_rpc`、`asset_api` + `asset_rpc`）同属同一业务域，各自仅配置本域 `Mysql.DataSource`，共同读写本域独占库；而**不同业务域之间**（`account`、`asset`、`relation`、`message`、`notification`、`platform`、`datasync` 及各独立子服务仓）数据库相互完全隔离、不可直连。
- **四项严禁**：
  1. 禁止持有、读取或连接其他服务的数据库 DSN / 连接池；
  2. 禁止跨库 SQL、跨库 JOIN、跨库本地数据库事务；
  3. 禁止 import 其他服务的 `*_models` 去查表或当通用 DTO；
  4. 禁止把开发机偶然共用同一个 MySQL 实例当成数据库共享前提。
- **跨服务数据获取与展示怎么实现**：
  - 实时读取/操作：走对方接口（主仓同仓走对方 **RPC**，例如 [account_api 配置](../../app/account/account_api/internal/config/config.go) 与 [asset_api 配置](../../app/asset/asset_api/internal/config/config.go) 均只配自有 `Mysql.DataSource` 并经 `UserRpc`/`RelationRpc` 等调用他域；跨独立仓走 `remote`/S2S HTTP）。
  - 需要聚合或异步本地视图：通过接口组合（如 `datasync_api` 经领域 RPC 与 `remote` 聚合）或归属清晰的本地投影/异步事件（如既有 RocketMQ 事件与增量同步接口）写入**本域自有表**；禁止把“私自复制同步对方表”当作绕过服务所有权的手段，也不凭空要求仓里不存在的新基础设施。

**正例**：`app/asset` 需要校验或展示用户信息时，通过 `UserRpc` 调用 `app/account`，两域各连各的独立库。
**反例**：在 `asset_api` 配置里新增一套账号库 `DSN`，或通过 GORM 直连 `beaver_account` 表与 `beaver_asset` 表做 SQL JOIN。

模型在 `app/<domain>/<domain>_models/<business>/xxx_model.go`，一文件一表；表名 snake_case。
Go 模型主键写 ID；发现表结构、列名或 GORM tag 有问题时直接修改模型、迁表定义与所有读写代码，不需要保留旧列做兼容双写。
通用基础字段在 [common/models](../../common/models/enter.go)，不代表所有业务模型都可以塞那里。
例：新增账号设备表落 account_models/auth/，登记账号库迁表集合；不能放 relation_models 或 account_api/internal/models。
JSON 列的 Scan/Value 辅助类型可与本表同文件，错误不能被吞成空成功。

## 6.2 事务不等于跨服务原子性

**物理隔离前提**：由于各服务数据库不共享且不在同一台服务器，**跨服务一致性绝不能依靠同机数据库事务实现**。

- **本域单库内部**：需要一致的多步写入使用显式事务；闭包内始终用传入的 `tx`，不混用 `svcCtx.DB`，否则部分写入逃逸。返回任一步错误触发回滚；事务里禁止等待远程 HTTP/RPC 和用户输入造成长锁。
- **跨服务 / DB 与 MQ 之间**：DB 提交与 RPC/`remote`/MQ 发出不是同一个事务。应基于当前真实链路显式说明失败窗口、幂等键、重复消费处理或必要的显式业务补偿（例如先写状态再异步通知、失败重试/对账），未知细节不编造。
- 本仓没有被本次验证的统一 outbox/分布式事务保障，不能写成“已有且可靠”。

以下为 GORM **规范示例**，不是已有方法：
```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // 本次相关写入全部走 tx；任一步失败 return err
    return nil
})
```
不要先发不可撤回通知再写库，失败后仍对外展示成功。必要异步边界见第 10 章。

## 6.3 升级数据库：先分清建表与迁历史数据

当前 [根 main.go](../../main.go) 是唯一的 `-db` 工具，内含建库/AutoMigrate/种子逻辑。
数据库连接凭据只在 main.go 顶部定义一次（`mysqlUser`/`mysqlAddr`，连接别的实例时自行替换），各库 DSN 由 `dbDSN(dbName)` 拼接（`dbName` 为空即连接实例本身，用于 CREATE DATABASE），不在多处重复，也不在源码里保留远程实例凭据；原先重叠且零调用的 `database/migrations.go` 已删除。
**重要定位**：这里仅是**本地开发机集中建库与初始化辅助现状**，绝不表示生产环境共享同一 MySQL 服务器，也不意味着业务服务在运行期可以拿到其他库的连接权限；实际部署中各服务物理隔离且仅持自有库 DSN。

数据库结构变更步骤：
1. 直接改齐目标表模型、迁表登记、索引和读写逻辑；不需要为兼容旧版数据格式编写历史过渡层，废弃表和字段直接删除。
2. 明确全新建库与本地已有开发库的差异；AutoMigrate 不会自动完成旧库删改字段或语义修正。
3. 不得在业务请求中临时执行 ALTER；在**获得明确授权的隔离库**验证建表与读写正确性，不把数据库密码和真实连接串写入文档。
4. 未经明确授权不得直接清空或删除他人本地已有库；本任务未建立版本化迁移框架，不虚构未存在的迁移命令。

旧 Python 参考里“删库再建”仅作为历史原文保留，不是默认操作步骤或删除授权。
没有上线压力也不表示可以删当前用户数据。

## 6.4 同步契约和版本

[core/version](../../core/version/version.go) 提供版本能力；datasync 是增量聚合入口，
[GetSyncChatMessagesLogic](../../app/datasync/datasync_api/internal/logic/sync/getsyncchatmessageslogic.go) 是可定位样例。
用户、关系、消息、通知、媒体等同步分支见 [第 10 章](../modules/realtime.md)。

修改或新增同步能力时写清：游标/版本是何种比较、作用域（用户/会话/表）、排序、分页边界、删除标记、重复返回如何处理。
生产者写版本、RPC/remote 输出、datasync 类型与客户端消费要一起核对；不能只改 HTTP Res。
不得用客户端时间代替服务端版本，也不能在重放/重试时丢掉删除事件。
验收至少包含同版本重取、分页边界、离线后增量、删除与多端并发；具体既有字段语义需读对应模型/logic，手册不伪造一个全局游标算法。
