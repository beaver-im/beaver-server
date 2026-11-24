# 3.5 新增 HTTP

[返回本章](README.md) · [全书目录](../README.md)


1. 读第 9 章定位，先选业务类别，再明确登录/S2S 保护面还是公开面；不是限制全服务只有一对文件。
2. 同类别保护接口放 `<业务>.api`，公开放 `<业务>_public.api`（文件名后缀是源码约定；URL group 仍需以 `_public` 结尾才会被网关放行）；新增分类加入主入口 import，拆合时整体移动 handler/类型组并删除旧声明，此时不手建 Go 路由/类型。
3. 在 **仓根 PowerShell** 进入服务目录并校验（只解析契约）：

```powershell
Push-Location app/account/account_api
goctl api validate --api account_api.api
# 非 0 立即停止，修契约，不带病生成
# 需要用户授权生成，并确认未提交/并发修改后才执行下一条：
goctl api go --api account_api.api --dir . --home ../../../template
Pop-Location
```

4. 查看 git diff（含未跟踪文件）：新增 handler/routes/types 必须可追溯到契约/模板；发现意外额外 main 先确认，再只清理本次误生成者，不能删历史业务文件。
5. 在对应 logic 实现业务，必要时补 svc/config/models/remote；不把 DB/remote 放 handler。
6. 按第 8 章测试矩阵检查鉴权失败、空字段、跨域调用/事务等，核查调用方 DTO。
7. 更新受影响章节，记录生成命令与版本。正式代码任务在独立干净工作区验证再次生成无意外差异；本次文档任务未做此操作。
