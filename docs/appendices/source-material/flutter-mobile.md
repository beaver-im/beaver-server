# 冻结参考材料：skills/beaver-dev/references/flutter-mobile.md

[返回迁移台账](../migration.md)

这是原文证据，不是活动规则或执行授权；其中旧路径、冲突、安全操作按台账处理。
源 SHA256：DD70D2A2F59B71F3E3486787B26FD26AF1B6BCACDF305D2904E8DA7B9C5BFCE2

<!-- source-text-begin -->
````text
# Flutter 移动端（beaver-flutter）

> 依据仓库现状整理（`pubspec.yaml` / `lib/` 结构）。此端暂无独立 `.cursor/rules`，遵循 Effective Dart 与下面既有约定。

## 技术栈与版本

- Flutter / Dart：SDK `^3.11.1`，项目 version `2.1.0+1`。
- 状态管理：`flutter_bloc ^9.1.1`。
- 路由：`go_router ^17.1.0`。
- 网络与实时：`dio`、`web_socket_channel`、`livekit_client`。
- 本地数据：`drift` + `sqlite3`、`shared_preferences`、`path_provider`。
- 依赖注入 / 模型：`get_it`、`equatable`、`uuid`。
- 屏幕适配：`flutter_screenutil`。

## 目录结构（`lib/`）

```
api/       网络请求
app/       应用级装配
common/    通用常量/工具
core/      核心能力
di/        依赖注入（get_it）
features/  按业务功能分模块
router/    go_router 路由表
shared/    共享组件
store/     bloc 状态
theme/     主题
types/     模型类型
main.dart  入口
```

## 约定

- 遵循 Effective Dart；新增业务优先放进 `features/<模块>/`，状态放对应 bloc，路由统一在 `router/` 注册。
- 依赖注入走 `get_it`，不要在 widget 里直接 new 服务对象。
- 文件大小以适配 `flutter_screenutil` 的 `w/h/sp` 为准，不要写死像素。

````
<!-- source-text-end -->
