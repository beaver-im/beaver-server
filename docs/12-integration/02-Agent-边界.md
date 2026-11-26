# 12.2 Agent 边界

[返回本章](README.md) · [全书目录](../README.md)


- 客户端工具链位于 agent_api/beaver/；旧事件 client_tool、HTTP 回传 submit_client_tool_result 的现状需改动时核实，不新造 host/sandbox 通道。
- workspace.path 只描述归属机器上的路径，不据此在服务器或另一台电脑写盘；拥有工作空间的机器才执行本地动作。
- 旧约定用 agent_usage_ledger 记用量，不给消息表补 LlmMeta；消息要 branch_id，话题无分支时建主分支；设备在线状态归其它微服务，不让 Agent 直查设备在线。
- 旧材料记录“双电脑同时执行 client_tool，未校验工作空间归属”的问题：**仅保留待复核风险，不宣称现在仍存在或已修复**，不能凭它直接改桌面。
- 历史 grpc 工具限制是仓内 uv 环境 grpc_tools / protobuf 4.x / grpcio 1.62；当前 pyproject 仍有上界。仅遗留工具需用时复核锁文件，不使用机器全局新 protoc 替代，也不把这条当本域恢复 gRPC 的理由。
- 表结构按 ORM 改到位并删除废弃字段；对本机已有真实数据执行删库/清表须先获授权。
