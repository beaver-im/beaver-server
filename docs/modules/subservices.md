# 第 11 章　六个独立子服务包

[返回目录](../README.md) · [Python 与客户端](integration.md)

## 11.1 工作区地图（不是 Go workspace，也不是 Git 子模块声明）

以下路径相对本手册文件；六仓有各自清单，不能把主仓 module `beaver` 导入方式复制过去。

| 仓库 / 边界 | 已核实的能力入口 | 适用流程与已知差异 |
|---|---|---|
| [beaver-server-ai](../../../beaver-server-package/beaver-server-ai/)：Agent 配置、知识、运行时 | [pyproject.toml](../../../beaver-server-package/beaver-server-ai/pyproject.toml)、[agent main.py](../../../beaver-server-package/beaver-server-ai/app/ai/agent_api/main.py) | Python≥3.12、uv/FastAPI/LangGraph；单 agent_api；不套 goctl，不新增 agent RPC |
| [beaver-server-backend](../../../beaver-server-package/beaver-server-backend/)：管理面 | [backend_admin.api](../../../beaver-server-package/beaver-server-backend/app/backend_admin/backend_admin.api) | Go；实际为 app/backend_admin + backend_models，不是主仓三层深度 |
| [beaver-server-call](../../../beaver-server-package/beaver-server-call/)：通话 | [call_api.api](../../../beaver-server-package/beaver-server-call/app/call_api/call_api.api) | Go；app/call_api + call_models；call/webhook/types 契约；未发现 RPC 契约 |
| [beaver-server-online](../../../beaver-server-package/beaver-server-online/)：在线文档 | [doc_api.api](../../../beaver-server-package/beaver-server-online/app/doc/doc_api/doc_api.api) | Go；域名是 doc，仓名 online；doc/doc_public 两种鉴权，不虚构 online_rpc |
| [beaver-server-open](../../../beaver-server-package/beaver-server-open/)：开放平台/开发者/机器人 | [open_api.api](../../../beaver-server-package/beaver-server-open/app/open/open_api/open_api.api) | Go；developer/public/machine 等面，以真实契约为准；不可强改成主仓命名 |
| [beaver-server-social](../../../beaver-server-package/beaver-server-social/)：社交圈/动态/帖子 | [social_api.api](../../../beaver-server-package/beaver-server-social/app/social/social_api/social_api.api) | Go；当前 app/social/social_api，circle/moment/post/comment/machine；不是旧的独立 moment/circle RPC 两进程 |

核验范围仅目录、清单、契约入口和 AI 启动/路径代码；没有核对所有业务与配置值。
只单独克隆主仓时这些链接可缺失，文档校验器将它们作为可选同级仓报告，不要求下载。

## 11.2 子仓也先契约、再生成，但 cwd 不能照抄

五个 Go 子仓均有自己的 `.api` 和 template/；本次没有发现它们的 `.proto`。
**不是要求每个服务硬造 RPC**：只有确需仓内 RPC 才采用第 3 章 `.proto`→生成流程。

| 仓根起点进入的服务目录 | 入口 | 指向本仓 template 的参数 |
|---|---|---|
| backend：`app/backend_admin` | backend_admin.api | `--home ../../template` |
| call：`app/call_api` | call_api.api | `--home ../../template` |
| online：`app/doc/doc_api` | doc_api.api | `--home ../../../template` |
| open：`app/open/open_api` | open_api.api | `--home ../../../template` |
| social：`app/social/social_api` | social_api.api | `--home ../../../template` |

先在目标仓根检查状态，进入上述目录：`goctl api validate --api <入口>`；
授权生成后 `goctl api go --api <入口> --dir . --home <表中路径>`，核对真实工具/模板和差异，再写 logic/必要装配。
这些是经目录推算并验证 template 存在的正确相对落点，**本次未实际生成验证兼容性**。
backend/call 原契约注释的 `../../../template` 当前会指向不存在的工作区上层 template，不应盲抄；未越权修改子仓注释/脚本。
不同仓 import 组织也不同（social 根集中 import，部分旧注释声称链式但实际多个 import），不要凭旧注释批量调整。

五个 Go 子仓的后续分类同样遵循[第 3 章](../standards/02-contracts-codegen.md)：业务类别任意扩展，保护 `<业务>.api` / 公开 `<业务>_publish.api`，并入各自服务的单一入口生成；不是每类别新进程。表中 doc_public 等是已核实的现存名字，未在本次改源码，不能当作目标仍叫 `_public`。
出站按[第 4 章](../standards/03-boundaries-remote.md)每接口 XxxApi + 独立 type(...) 的 XxxReq/XxxRes，不复用别仓模型。

## 11.3 跨仓功能的检查顺序

1. 子仓同样遵守数据库独占与物理隔离（[第 6 章](../standards/05-data-sync.md)）；先确认提供者真实 `.api` 路径、鉴权、字段/信封；调用者的 remote 名称不一定与仓名相同（doc/circle/moment/agent/ai 需逐个核对）。
2. 提供者先改契约生成，实现并验证；调用者更新本仓 types/remote 和 XxxApi 方法，不能引对端模型。
3. 列跨仓同步修改清单（提供方、消费方、机器票/权限与错误语义），直接同步更新各端不留兼容层，敏感值不入手册。
4. Python Agent 走下章；跨仓不要 etcd RPC/跨库。没读过的子仓细节标未核实，不套用主仓“已达标”。

本次手册集中在 beaver-server；没有给子仓加规则、改业务或执行生成。
