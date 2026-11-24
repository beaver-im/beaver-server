# 12.1 Python Agent

[返回本章](README.md) · [全书目录](../README.md)


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
- 不跨仓 RPC、不造 internal 门面；通道不重试/聚合/缓存业务，只缓存票；统一走 S2S Bearer，不保留静态令牌兜底。
- 仅本进程读模型/跑图不套 remote；对外部产品才建 peer。对端 URL 和信封仍需对契约，不照抄历史 token 路径。
