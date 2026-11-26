# Beaver 服务端工程手册

给人顺着读、给 AI 按任务查；不是每次全部读入的提示词。

**定位**：Go / go-zero 即时通信后端；`.api` / `.proto` 是接口事实源，goctl 生成接线。

## 文档结构标准（固定）

1. **一个目录 = 一章**，目录名带序号：`01-architecture/`、`02-layout/` …
2. **章入口固定为 `README.md`（章首）**：只写编号、一句话说明、小节目录；**不要把小节正文堆进 README**。
3. **小节 = 章目录下的正文文件**：`01-主题.md` 对应 `1.1`，文件内标题写 `# 1.1 主题`。
4. **再往下拆 = 再开子目录**：对应 `1.1.1`；子目录同样有入口 `README.md`。
5. **编号规则**：章 `1` → 小节 `1.1` → 子节 `1.1.1`。标题与目录名序号一致，便于查找。
6. **不要提前拆**：只有写长了或确有独立主题，才把小节升级成子目录。
7. **全书目录**：本页「目录」必须列出全部章与小节链接；增删小节时同步改本页与章首。

参考页（不算章）：[directory.md](directory.md) 工程目录、[glossary.md](glossary.md) 术语。

变更策略见 [AGENTS.md](../AGENTS.md)。产品需求 / 流水账不进 `docs/`。

## 怎么读

- 找文件 → [工程目录](directory.md)
- 第一次 → [术语](glossary.md) → `1`–`4`、`8` → 负责模块
- 加 HTTP/RPC → `3`；调外部 → `4`；库隔离 → `1`、`6`

## 目录

### 参考
- [工程目录](directory.md)
- [术语](glossary.md)

### 章与小节

1. [架构](01-architecture/)
   - [1.1 项目与工具链](01-architecture/01-项目与工具链.md)
   - [1.2 运行结构](01-architecture/02-运行结构.md)
   - [1.3 进程与基础设施](01-architecture/03-进程与基础设施.md)
   - [1.4 独立仓边界](01-architecture/04-独立仓边界.md)
   - [1.5 本章范围](01-architecture/05-本章范围.md)

2. [目录与命名](02-layout/)
   - [2.1 业务域形状](02-layout/01-业务域形状.md)
   - [2.2 改动落点](02-layout/02-改动落点.md)
   - [2.3 命名与版权](02-layout/03-命名与版权.md)
   - [2.4 正反例](02-layout/04-正反例.md)

3. [契约与生成](03-contracts/)
   - [3.1 原则](03-contracts/01-原则.md)
   - [3.2 文件归属](03-contracts/02-文件归属.md)
   - [3.3 HTTP 契约](03-contracts/03-HTTP-契约.md)
   - [3.4 路由规则](03-contracts/04-路由规则.md)
   - [3.5 新增 HTTP](03-contracts/05-新增-HTTP.md)
   - [3.6 新增 RPC](03-contracts/06-新增-RPC.md)
   - [3.7 生成与发布脚本](03-contracts/07-生成与发布脚本.md)

4. [边界与远程调用](04-boundaries/)
   - [4.1 调用方向](04-boundaries/01-调用方向.md)
   - [4.2 remote](04-boundaries/02-remote.md)
   - [4.3 Req / Res](04-boundaries/03-Req-Res.md)
   - [4.4 API 与 RPC 映射](04-boundaries/04-API-与-RPC-映射.md)

5. [身份与运行时](05-runtime/)
   - [5.1 身份](05-runtime/01-身份.md)
   - [5.2 校验](05-runtime/02-校验.md)
   - [5.3 context 与超时](05-runtime/03-context-与超时.md)
   - [5.4 错误与日志](05-runtime/04-错误与日志.md)
   - [5.5 资源清理](05-runtime/05-资源清理.md)

6. [数据与同步](06-data/)
   - [6.1 库归属](06-data/01-库归属.md)
   - [6.2 事务](06-data/02-事务.md)
   - [6.3 升级](06-data/03-升级.md)
   - [6.4 同步与版本](06-data/04-同步与版本.md)

7. [配置与部署](07-deploy/)
   - [7.1 配置](07-deploy/01-配置.md)
   - [7.2 部署](07-deploy/02-部署.md)
   - [7.3 独立仓发布](07-deploy/03-独立仓发布.md)

8. [实施与验证](08-workflow/)
   - [8.1 任务顺序](08-workflow/01-任务顺序.md)
   - [8.2 验证矩阵](08-workflow/02-验证矩阵.md)
   - [8.3 评审清单](08-workflow/03-评审清单.md)
   - [8.4 文档维护](08-workflow/04-文档维护.md)

9. [业务域](09-domains/)
   - [9.1 域一览](09-domains/01-域一览.md)
   - [9.2 纵切面示例](09-domains/02-纵切面示例.md)
   - [9.3 新增能力](09-domains/03-新增能力.md)

10. [消息与实时](10-realtime/)
   - [10.1 message](10-realtime/01-message.md)
   - [10.2 MQ](10-realtime/02-MQ.md)
   - [10.3 ws](10-realtime/03-ws.md)
   - [10.4 datasync](10-realtime/04-datasync.md)
   - [10.5 验收](10-realtime/05-验收.md)

11. [独立子服务](11-subservices/)
   - [11.1 工作区](11-subservices/01-工作区.md)
   - [11.2 契约与生成](11-subservices/02-契约与生成.md)
   - [11.3 跨仓检查](11-subservices/03-跨仓检查.md)

12. [联调约定](12-integration/)
   - [12.1 Python Agent](12-integration/01-Python-Agent.md)
   - [12.2 Agent 边界](12-integration/02-Agent-边界.md)
   - [12.3 客户端](12-integration/03-客户端.md)

13. [决策记录](13-decisions/)
   - [0001　契约优先的生成边界](13-decisions/0001-contract-first.md)
   - [0002　鉴权平面、公开端点与机器票归属](13-decisions/0002-auth-planes.md)

## 仓库布局

```text
docs/
  README.md                 全书入口：结构标准 + 完整目录（章与小节）
  directory.md / glossary.md  参考（无序号）
  01-architecture/
    README.md               章首
    01-….md                 小节正文（1.1）
  …
  knowledge-checks.json
```

校验：`powershell -NoProfile -File scripts/check-knowledge.ps1`
