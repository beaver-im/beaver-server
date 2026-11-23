# Beaver 服务端工程手册

这是一本给人顺着读、给 AI 按任务查、给评审逐项用的手册，不是要求每次全部读入的提示词。

**项目定位：Go / go-zero 即时通信后端；`.api` / `.proto` 是接口事实源，goctl 负责生成接线。**

核对基线：2026-11-22（Asia/Shanghai），2026-11-23 增补鉴权平面终态与工程目录大全。核对了主仓契约入口、模板、代表性业务链路、7 份旧技能材料和 6 个子服务目录/入口；没有逐行审计全部业务、安全或测试。
2026-11-23 增补：第 5 章补齐网关验票顺序、三条平级平面、公开端点与机器票归属；对应 [ADR 0002](decisions/0002-auth-planes.md)。
## 最高工程原则：无历史包袱，不留兼容代码

本项目开源且不作商业托管：**不考虑兼容代码、不考虑实现成本、哪里有问题改哪里**（完整表述、授权边界见 [AGENTS.md](../AGENTS.md)）。契约、目录、接口与数据模型可直接破坏性调整，但须同步改齐调用方与生成物；`.api`/`.proto` → `goctl` 生成、错误处理、权限安全与测试照常执行。

文中统一区分：**当前事实**（已核查现状）、**统一工程规范**（全仓唯一目标）、**未整改差异/待核实**（现状尚未修完，不代表允许留兼容）。规则正文只在唯一主章节维护，其它章节只留本地摘要与链接——**不要在多处复制同一条规则**。

## 从哪里开始读

- **不知道东西在哪**：先查[工程目录大全](directory.md)（每个目录/文件的用途、哪些是生成物、改什么去哪）。
- 第一次接触：先看[术语表](glossary.md)，再按第 1 → 2 → 3 → 4 → 8 章读，最后进入负责的模块。
- 新增 HTTP：第 3 章“新增 HTTP 的七步”，支持任意业务类别 `.api` / `_publish.api`，不是每服务仅两个文件。
- 调用外部接口：第 4 章“remote 的类型组与实作样例”，逐接口 `XxxApi` + `XxxReq` / `XxxRes`，再按第 8 章评审。
- 新增 RPC：第 3 章“新增 RPC”，不要照搬 HTTP types。
- 跨服务与数据库隔离：第 1 章与第 6 章 6.1；每个服务数据库独立且按不同服务器物理隔离（如 `app/account` 与 `app/asset` 各自独立库），禁止直连别库，必须走 RPC / remote。
- 改用户/关系/文件/平台：第 9 章定位；消息/WS/同步看第 10 章。
- 联调 ai/backend/call/online/open/social：第 11 章；Python 不套 Go 生成流程。
- 追查旧 skill 的某条规则：附录 A；完整原文在附录 A 链接的冻结材料中。

## 全书目录

### 先读
- [工程目录大全](directory.md)：全仓目录与文件用途的现状地图，含"我要改 X 去哪"速查。
- [术语表](glossary.md)：跨章高频且易混淆的名词，只在这里定义一次。

### 第一篇　先看懂工程
1. [第 1 章　系统、层次与数据流](architecture.md)
2. [第 2 章　目录、命名与代码归属](standards/01-layout.md)

### 第二篇　写代码时照着做
3. [第 3 章　契约优先：HTTP / RPC 与 goctl 生成](standards/02-contracts-codegen.md)
4. [第 4 章　模块边界、跨仓 S2S 与显式类型映射](standards/03-boundaries-remote.md)
5. [第 5 章　身份、校验、错误、日志与资源寿命](standards/04-runtime-security.md)
6. [第 6 章　数据库、事务、升级与增量同步](standards/05-data-sync.md)
7. [第 7 章　配置与部署边界](standards/06-deployment.md)
8. [第 8 章　实施、验证、评审和交接](workflow.md)

### 第三篇　按业务查入口
9. [第 9 章　账号、关系、资产、通知与平台](modules/domains.md)
10. [第 10 章　消息、队列、WebSocket 和同步](modules/realtime.md)
11. [第 11 章　六个独立子服务包](modules/subservices.md)
12. [第 12 章　Python Agent 与客户端联调约定](modules/integration.md)

### 第四篇　维护这本书
13. [第 13 章　决策记录：索引与模板](decisions/README.md)
14. [第 14 章　任务交接：模板与归档](tasks/README.md)
15. [附录 A　旧技能逐项迁移与冲突台账](appendices/migration.md)

## 目录与事实源

```text
AGENTS.md                         薄入口，不复制整书
scripts/check-knowledge.ps1        无需安装依赖的文档检查
 docs/
  README.md                       本目录
  directory.md                    工程目录大全（现状地图，文件用途）
  glossary.md                     术语表（单一定义点）
  architecture.md                 系统图
  standards/                      第 2—7 章，具体约束
  workflow.md                     第 8 章，验收矩阵
  modules/                        第 9—12 章，任务定位
  decisions/                      ADR（架构决策记录），只记录有依据的决定
  tasks/                          进行中/归档的交接约定
  appendices/                     迁移台账、冻结原材料（不是活动规范）
  knowledge-checks.json            路径、命令和迁移校验清单
```

文档校验（仓根 PowerShell）：`powershell -NoProfile -File scripts/check-knowledge.ps1`。
仅验证手册链接、登记路径/命令证据、原文迁移完整性和格式；**不证明业务正确、安全合规或生成幂等**。
改契约/目录/命令/跨仓边界时，同一任务更新对应章节与清单；每次真正动代码仍需核实局部源码。

## 既有专题资料与仓库旧文档说明

仓库根 `README.md` / `CONTRIBUTING.md` 中的部分版本和旧目录示例已滞后，请以当前源码、`go.mod` 及本手册核验说明为准。
此前工作区曾存在的若干草案专题文档（部署归属、字段审查、媒体智能、OpenClaw、Webhook 重构、待办）已从当前工作树移除，本目录不再保留无效链接，相关实现直接对照当前代码与契约。
