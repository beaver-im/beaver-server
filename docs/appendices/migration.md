# 附录 A　旧技能逐项迁移与冲突台账

[返回全书目录](../README.md)

## A.1 迁移范围与可审计性

原目录 `skills/beaver-dev` 共 **7 个文件**，SKILL、Cursor 入口与 5 份技术参考均已完整阅读。
原文引用的 `references/global.md` **原目录不存在**，不能假装迁移了它。
本次先在仓库外做逐文件 SHA256 核对的备份，再将原文冻结到本附录，提炼活动规范，检查通过后才移除指定旧目录。
只移除该目录，不移除其它 skills，不处理全局编辑器安装。

| 原文件（相对旧目录） | 冻结原文（非活动规则） | 活动说明 |
|---|---|---|
| SKILL.md | [导航原文](source-material/navigation.md) | AGENTS / 全书目录 / 第 8 章 |
| cursor/beaver-dev.mdc | [Cursor 原文](source-material/cursor-entry.md) | 核心并入薄入口，不再安装第二套内嵌规则 |
| references/go-zero-server.md | [Go 原文](source-material/go-zero-server.md) | 第 2—8 章 |
| references/python-agent.md | [Python 原文](source-material/python-agent.md) | 第 11—12 章 |
| references/electron-desktop.md | [桌面原文](source-material/electron-desktop.md) | 第 12 章，桌面活动规范仍由桌面仓维护 |
| references/flutter-mobile.md | [Flutter 原文](source-material/flutter-mobile.md) | 第 12 章，版本未复核 |
| references/vue-admin.md | [管理端原文](source-material/vue-admin.md) | 第 12 章，版本/脚本未复核 |

原文逐字快照保留版本、示例与所有细节；knowledge-checks.json 记录来源/目标/源 SHA256，检查器提取冻结文本比对哈希。
**冻结文本只供追溯，不重新当指令执行**。下表“采纳”是新代码要求，不表示存量全部达标。

## A.2 Go 参考的主题逐项对照

| 编号 / 原条款主题 | 当前章节（唯一主落点） | 状态与理由 |
|---|---|---|
| G01 三条红线：RPC 仓内、跨仓 S2S、禁止跨库 | [第 4 章](../standards/03-boundaries-remote.md) | 采纳；含仓内互调而非 S2S |
| G02 不擅自 build/启动/打包 | [第 8 章](../workflow.md) | 采纳；测试也需说明编译/缓存成本 |
| G03 不删已有 handler/logic，废弃先停声明 | [第 3 章](../standards/02-contracts-codegen.md) | 采纳；生成误建 main 只能核实后处理本次产物，不自动删历史代码 |
| G04 只改契约/logic 与必要 svc/config/template，生成物不手改 | 第 3 章 | 采纳并澄清；不是禁止编辑所有生成骨架 |
| G05 公共生成行为改模板、完整版权头 | [第 2 章](../standards/01-layout.md)、第 3 章 | 采纳；RPC 模板缺头为存量差异 |
| G06 .api 驼峰/.proto 与表 snake_case，handler 校验/logic 业务 | 第 2、3、5 章 | 采纳；解析校验模板化，个性业务校验在 logic，不为校验手改 handler |
| G07 所有 Go 手写字段缩写全大写，model ID | 第 2 章 | 采纳；pb 生成命名不手改 |
| G08 .api 命名决定生成字段；json lowerCamel，OAuth 固定字段例外 | 第 2、3 章 | 采纳；Go 名与 wire tag 分离 |
| G09 改字段同步跨仓 DTO；编译驱动精确修引用禁全替换 | 第 4、8 章 | 采纳；可直接破坏性改字段并同步全部消费方（build 仍受授权约束） |
| G10 仓内 RPC / 跨仓 remote 调用表 | 第 4 章 | 采纳；本域读自己表不强制自调 RPC |
| G11 remote/peer enter/remote/token/biz + types/remote + commonconfig | 第 4 章 | 采纳；真实目录作证 |
| G12 正常业务或 machine 路径，无 internal 门面，无跨仓 etcd | 第 4 章 | 采纳；具体机器面 group 以现有契约为准 |
| G13 DTO 本仓自持，不 import 对端；每 peer 自带通道 | 第 4 章 | 采纳；不抽统一 RemoteClient |
| G14 Init 只装配，调用不散换票；Bearer 优先/存量静态票 | 第 4 章 | 采纳；不借文档泄露凭据 |
| G15 不重试/聚合/缓存业务，只缓存换票 skew | 第 4 章 | 采纳；业务可靠性在编排层另评审 |
| G16 logic 调 remote，remote.go 不放业务，按能力文件拆 | 第 4 章 | 采纳；静态票对端同样适用 |
| G17 Remote 每接口独立 type 块/Req/Res/Item，token 类型例外 | 第 4 章 | 独立类型组采纳；旧 token 共享豁免被用户最新逐接口 Req/Res 要求覆盖，待按 peer 拆清，稳定领域基础类型可复用 |
| G18 标准信封 code/msg/result，Res 仅 result，ajax 解码 | 第 4、5 章 | 采纳；对端不标准需协同改契约，不私做兼容层 |
| G19 业务 XxxApi 后缀、通道工具例外、URL 写全、存量调用方改名 | 第 4 章 | 采纳；直接改名并同步调用方、删除旧函数名包装，Api 大小写不替换 |
| G20 领域仅 xx_api/xx_models/xx_rpc，可裁剪 | 第 2、11 章 | 采纳为新主仓业务域要求；gateway 与子仓现有结构不伪装已迁移 |
| G21 models 按业务分包、一文件一表、JSON 列辅助只服务本表 | 第 2、6 章 | 采纳；禁止含混 app 杂包和多表文件 |
| G22 API 入口只 import，api 必须有，handler/logic/types 分工 | 第 3 章 | 采纳；附真实 account 样例 |
| G23 子 .api 一个 server，按鉴权分 xxx/xxx_public | 第 3 章 | 用户最新澄清覆盖旧“只有鉴权不同才拆文件”：业务分类任意扩展，目标 `<业务>.api` / `<业务>_publish.api`，不是每服务仅两文件；现有 public 名为未整改事实 |
| G24 C 端与 S2S 同文件，group 分离，不另开 xxx_s2s | 第 3 章 | 采纳文件组织意图；URL 的 xxx_s2s 与 machine 选定统一方案后直接全仓替换、不双轨兼容 |
| G25 换票放 public、靠自身 secret、不另开 token 文件 | 第 3 章 | 目标公开分类名按最新要求 `_publish`；同业务放一起，仍须校验自身凭据，不是免业务安全 |
| G26 合拆 handler 与专属类型整体迁移 | 第 3 章 | 采纳；避免遗漏依赖 |
| G27 handler/logic 包名模板默认、工具 common 保留；重复 service 可合并 | 第 3 章 | 采纳；不无差别改包名 |
| G28 字段改名/tag 同步跨仓 DTO，核对 URL 命中 routes | 第 3、4 章 | 采纳；旧文说的交叉脚本未提供，不伪造现成命令 |
| G29 旧 service→account/relation/message/asset/platform 映射 | 第 3 章 | 采纳；notification/datasync 保持 |
| G30 六段路由、action snake_case、无 :id、不再拆 action | 第 3 章 | 采纳；与 Cursor 五段简写冲突，以详细规范+真实路由为据纠正 |
| G31 网关只转发不鉴权，失败真实 401 | [第 1 章](../architecture.md)、[第 5 章](../standards/04-runtime-security.md) | 部分纠正：当前网关确实鉴权；401 要求保留，不按旧句删校验 |
| G32 每 handler 一组独立 Req/Res/专属子类，不共享，N 对 N | 第 3 章 | 采纳；WS types.api 等未合规处直接拆分整改，不保留共享结构兼容 |
| G33 Res 不含 code/msg、只 //、service 唯一和字段注释 | 第 3、5 章 | 采纳；空 Res 仍要定义 |
| G34 goctl api go + 本仓模板 | 第 3 章 | 采纳并补 validate/cwd/工具版本边界；没有统一生成脚本 |
| G35 RPC 单 proto 包、多 service、-m、go_package ./pb | 第 3 章 | 采纳；存量组合客户端/单 service 路径未全审计 |
| G36 RPC out=. 防 pb/pb，logic 可改，server/client/pb 不手改 | 第 3 章 | 采纳；不复制 help 通用示例中的 ./pb 目标 |
| G37 API types 与 RPC pb 两套，禁止反射/JSON 往返映射 | 第 4 章 | 采纳；旧“common/convert 已移除”与事实冲突，应直接清掉存量使用并移除 convert |
| G38 显式赋值、指针/slice/map/数值转换、自引用递归映射 | 第 4 章 | 采纳；不承诺编译捕获所有语义错误 |
| G39 CRLF/Windows 路径，duplicate type 拆类型，误生成 main | 第 2、3 章 | 采纳实用注意；不强制所有机器使用 F 盘，不自动删除 |
| G40 本仓 beaver import，独立仓各自 module，pb 落点联动 | 第 1、3、11 章 | 采纳；迁生成路径需完整审查，不半改限定符 |

## A.3 导航、Python、客户端逐项对照

| 编号 / 来源主题 | 当前章节 | 状态 / 理由 |
|---|---|---|
| N01 SKILL+Cursor：先列文件再等确认，直接改可跳过 | AGENTS、第 8 章 | 采纳；本次用户已明确授权，不每批再问 |
| N02 不写 md，用户/skill 例外 | 第 8、14 章 | 澄清：本次明确要求手册，允许同任务维护稳定知识；不自动制造报告 |
| N03 头完整、文件小写起始、多词驼峰 | 第 2、12 章 | 采纳通用意图；按 Go/proto/Python/Dart 专门命名澄清例外 |
| N04 五栈导航、global.md 引用 | 全书目录、第 11、12 章 | 迁移；global.md 源缺失，不能声称已有全部通用条款 |
| N05 跨仓红线、团队仓旧名字 doc/ai 等 | 第 4、11 章 | 采纳边界，更正真实仓名 online/ai 与当前目录；不造已不存在 RPC |
| N06 WorkBuddy/Claude/Cursor 安装复制说明 | AGENTS、A.4 | 历史保全，不再维护重复内嵌规则；工具自动加载能力不保证通用 |
| P01 仓根 uv 环境/启动、单 agent_api、无 agent_rpc、paths.boot | 第 11、12 章 | 路径改为真实 beaver-server-ai；单进程由 main/paths 佐证；安装/启动未执行 |
| P02 -db create_all、不手写 ALTER、删库重建 | 第 6、12 章 | 前两项保留意图/需核当前入口；删库降为历史，不当默认，数据销毁需另授权 |
| P03 handler/logic/svc/types/middleware/runtime/models 分层，禁 domain | 第 12 章 | 采纳该域原有规则，不给 Go 新造 domain 层 |
| P04 仓根 core 与 beaver/core 不强并 | 第 12 章 | 保留；具体重构需当前源码再核 |
| P05 Python remote 目录、自持 DTO、peer 通道、Init、Bearer、无重试 | 第 12 章引用第 4 章边界 | 采纳；不把 Go 模板用于 Python |
| P06 Python _api 命名、通道例外、完整 URL、何时建 peer | 第 12 章 | 采纳；原文 Api 字样与 Python 示例差异已按 _api 明确 |
| P07 client_tool/HTTP 回传、不再 host/sandbox | 第 12 章 | 保留约束；事件/具体接口需改动时再核 |
| P08 workspace.path 归属机器、usage ledger、branch、在线归别服 | 第 12 章 | 保留，未重新审计全部业务，不冒充现状验证 |
| P09 本地 middleware、可信头/Bearer、token 豁免、utils/s2s | 第 12 章 | 保留；可信头以认证链为前提，具体机器面实现需读该仓 |
| P10 双电脑工作空间已知问题 | 第 12 章 | 待复核风险，不宣称现在仍有问题，更未授权修桌面 |
| P11 不生本域 pb，遗留 uv grpc_tools/protobuf4/grpcio1.62 | 第 12 章 | 采纳不恢复 RPC；旧工具版本保留历史，使用前核锁文件 |
| E01 Electron 技术栈、4 别名、命名、Less | 第 12 章 + 桌面冻结原文 | 联调参考；当前桌面规范仍在桌面仓，未修改 |
| E02 复用 UI、禁业务重复原生控件、BeaverInput 密码、补 UI 层 | 第 12 章 + 桌面冻结原文 | 完整保留，不移植成服务端规范 |
| E03 beaver-desktop-header-v2 完整头 | 第 12 章 | 保留，不缩减为一个标记 |
| F01 Flutter 版本/依赖与完整 lib 地图 | 第 12 章 + Flutter 冻结原文 | 保留未复核基线，非当前版本断言 |
| F02 Effective Dart/features/bloc/router/get_it/适配 w/h/sp | 第 12 章 | 采纳原约定适用范围，不套 Go 驼峰 |
| V01 Vue 版本/技术栈/目录与 npm 脚本 | 第 12 章 + Vue 冻结原文 | 保留未复核历史；lint --fix 非只读 |
| V02 api 封装、pinia/types、Element Plus、全局命名头 | 第 12 章 | 保留原约定，不改管理端 |

## A.4 最高工程原则覆盖说明（2026-11-22 更新）

用户明确本仓为开源项目且不运营，没有历史包袱，不考虑任何兼容代码，不考虑实现成本，哪里有问题改哪里，不需要的去掉。
据此，原台账中“仅约束新代码、不改旧接口/旧路径/旧命名、保留兼容层或发布兼容窗口”等保守说法均已改写为直接重构并同步全部调用方；同时保留 `.api`/`.proto` → `goctl` 生成铁律，不手写生成物。


## A.5 用户最新澄清：remote 类型与任意业务分类

| 编号 / 新要求 | 当前唯一正文 | 覆盖旧说法的结论 |
|---|---|---|
| U01 remote 调别人接口，业务方法 Api 后缀 | [第 4 章 4.2](../standards/03-boundaries-remote.md) | 精确 XxxApi，通道工具不等于业务端点；不套桌面 I 前缀 |
| U02 每接口一个 type 包裹 XxxReq/XxxRes | [第 4 章 4.3](../standards/03-boundaries-remote.md) | Go type(...) 分组，不是嵌套类/每接口 Go package；含独立空 Res；基础领域值类型不等于可共用业务 Req/Res |
| U03 xxx 可按业务任意扩展，不止两个文件 | [第 3 章 3.3](../standards/02-contracts-codegen.md) | 按业务类别×鉴权面维护，同类放一起；并入单入口生成，不是每分类新进程 |
| U04 xxx.api 保护，xxx_publish.api 公开 | [第 3 章 3.3](../standards/02-contracts-codegen.md) | 后缀 publish 按用户原词，不擅改 public；保护包含登录/S2S；公开不免业务安全检查 |
| U05 服务数据库不共享且物理隔离（如 account/asset 各自独立），禁跨库直连 | [第 6 章 6.1](../standards/05-data-sync.md)、[第 1 章](../architecture.md)、[第 7 章](../standards/06-deployment.md) | 明确按不同服务器物理隔离部署设计：同域 api/rpc 访问本域独占库，跨服务绝不连他库/写跨库 SQL/事务/持他库 DSN；main.go -db 仅本地开发初始化现状 |

当前源码 `_public`、共享 TokenRequest/TokenResponse 和个别 DocId Go 字段是未整改差异，不是最新要求的例外许可。本次只改知识文档，没有改名/生成；以后直接同步整改并删旧，不留兼容双轨。
冻结原文仍逐字保留以维持审计哈希，里面的旧规则不能覆盖本节用户最新澄清；普通 AI 仅按任务读活动主章，无需加载冻结材料。

## A.6 冲突、失效信息与外部引用

已据源码处理：工具 goctl 拼写；六段而非五段；网关当前鉴权；common/convert 仍在；实际子仓名/服务目录；backend/call 注释模板深度不正确。
仍需以后决定或专项核验：goctl 与插件版本锁定；machine 与 xxx_s2s 路由命名统一；WS 共享类型/导入注释与规范差异；根 main 与 database 双迁表入口整合；模板无响应分支/版权头；各子仓旧结构迁移范围。
这些不是本次改业务的授权，本次不通过生成/重构“顺手修完”。

本仓可控的旧路径引用改为手册/冻结证据；WorkBuddy 历史记忆只替换失效路径，不改历史结论。
`.lingma/rules/rules.md` 仍保留其外部 `../../../beaver-skills/beaver-server/rule.md` 引用，只补本仓新入口，不改外部文件。
可能已安装到全局 WorkBuddy/Claude/Cursor 的副本或其它仓引用不在本次范围；未知是否存在，不能宣称全部卸载/同步。
工具是否自动读 AGENTS/各自规则有差异；知识文档只是可导航索引，不能保证自动加载，更不能保证永远无需读代码。
