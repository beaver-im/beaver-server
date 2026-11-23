# 第 12 章　Python Agent 与客户端联调约定

[返回目录](../README.md) · [原文迁移台账](../appendices/migration.md)

本章只保留旧全栈技能的联调约定，不接管其它仓主手册；未复核的客户端版本/脚本只在冻结材料中。Go 侧规则以[第 3 章](../standards/02-contracts-codegen.md)、[第 4 章](../standards/03-boundaries-remote.md)为准，不要把 Python `_api` 或旧冻结文档的 `_public` 命名反套到 Go。

## 12.1 Python Agent：不是另一个 Go 服务

实际仓名 **beaver-server-ai**，旧参考里的 beaver-ai 路径已过时。
[common/paths.py](../../../beaver-server-package/beaver-server-ai/common/paths.py) 和 main.py 确认：
`app/ai/agent_api` 同进程包含 HTTP、配置/历史与 LangGraph runtime；`agent_models` 在兄弟目录。
启动实际路径（仅说明，**本次未执行，启动/改库需授权**）：目标仓根 `uv run python app/ai/agent_api/main.py`。
旧 `uv sync` 会安装依赖、`main.py -db` 会改库，不能当默认阅读/校验步骤；其具体当前数据库入口本次未复核。

旧约定迁入以下边界，涉及具体功能时仍需在该仓局部复核：
- HTTP `app/handler → app/logic → agent_models/DB 或 remote`；svc 装配、types 定义入出参；不额外造 app/domain/internal/domain。
- `beaver/` 是运行时/工具内核，不与仓根 core 强并；middleware 在 agent_api/app/middleware，工具校验函数可落 utils/s2s。
- 本域不建 agent_rpc，不为它再生 pb 桩。旧 pyproject 描述仍提“对内 gRPC”，不据旧文案恢复第二进程。
- Python remote 在仓根，按 peer 分包；remote.py 只做 Post/Get/Delete/换票，业务函数 `<biz>.py`；本仓 types/remote 自持类型，不 import 对端业务类型。
- Python 出站采用 snake_case **`_api`** 后缀（get_doc_api/get_token_api），通道工具不加；不用 Go 的字面大小写 Api 硬套 Python。
- 不跨仓 RPC、不造 internal 门面；通道不重试/聚合/缓存业务，只缓存票；统一优先 Bearer，不再需要的静态 X-Internal-Token 对端直接切换并删除旧兜底。
- 仅本进程读模型/跑图不套 remote；对外部产品才建 peer。对端 URL 和信封仍需对契约，不照抄历史 token 路径。

## 12.2 Agent 专属业务警戒线（来自旧材料，未全部重新审计）

- 客户端工具链位于 agent_api/beaver/；旧事件 client_tool、HTTP 回传 submit_client_tool_result 的现状需改动时核实，不新造 host/sandbox 通道。
- workspace.path 只描述归属机器上的路径，不据此在服务器或另一台电脑写盘；拥有工作空间的机器才执行本地动作。
- 旧约定用 agent_usage_ledger 记用量，不给消息表补 LlmMeta；消息要 branch_id，话题无分支时建主分支；设备在线状态归其它微服务，不让 Agent 直查设备在线。
- 旧材料记录“双电脑同时执行 client_tool，未校验工作空间归属”的问题：**仅保留待复核风险，不宣称现在仍存在或已修复**，不能凭它直接改桌面。
- 历史 grpc 工具限制是仓内 uv 环境 grpc_tools / protobuf 4.x / grpcio 1.62；当前 pyproject 仍有上界。仅遗留工具需用时复核锁文件，不使用机器全局新 protoc 替代，也不把这条当本域恢复 gRPC 的理由。
- 不写兼容旧表结构的历史过渡层，结构可直接按 ORM 改到位并删除废弃表字段；但对本机已有真实数据或库执行删库/清表仍须先获得明确授权。

## 12.3 客户端约定的适用范围

| 端 | 迁入的协作约定 | 联调时怎么用 |
|---|---|---|
| Electron / Vue3 桌面 | src/main/common/render/preload 的旧别名、Less 嵌套、文件不大写开头、完整 beaver-desktop-header-v2；优先已有 components/ui，密码用 BeaverInput type=password + show-password，缺通用能力先补 UI 层 | 只作为旧资料保全，当前桌面主手册在其仓内；本次不改桌面、不替它裁决规则冲突 |
| Flutter | features 按业务、bloc 状态、router 统一注册、get_it 注入不在 widget new 服务、flutter_screenutil w/h/sp、Effective Dart | 版本/依赖详见冻结原文，当前 pubspec 本次未核对；不拿 Go 驼峰规范覆盖 Dart 文件命名 |
| Vue 管理端 | api 封装请求，pinia 状态，types 类型，Element Plus 优先，Less，保留头/原命名约定 | 旧 npm lint 带 --fix 会改文件，dev/build 会启动/产出；本次不执行，也不声称当前脚本存在 |

完整别名映射、技术栈版本、目录表、脚本原文全部保留在 [冻结材料索引](../appendices/migration.md)，不是丢弃。
新增跨端字段先对 `.api`/DTO，再对各端实际网络封装与消费处；不要以“统一规范”为名把五种技术栈变成同一目录树。
