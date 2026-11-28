/*
 * Copyright (c) 2024-2026 Beaver IM Team
 * SPDX-License-Identifier: MIT
 * Project: beaver-server
 * https://github.com/wsrh8888/beaver-server
 *
 * 中文：
 * 本文件为海狸 IM（Beaver IM）开源项目源代码。
 * 版权所有 © 2024-2026 Beaver IM Team，基于 MIT 协议授权。
 * 禁止删除、篡改或替换本文件头部版权与许可声明。
 * 使用与商业授权说明：https://wsrh8888.github.io/beaver-docs/community/license.html
 *
 * English:
 * This file is part of the Beaver IM open-source project.
 * Copyright (c) 2024-2026 Beaver IM Team. Licensed under the MIT License.
 * Do not remove, alter, or replace this copyright and license header.
 * Usage & commercial licensing: https://wsrh8888.github.io/beaver-docs/community/license.html
 *
 * beaver-server-header-v1
 */

package mqwsconst

// 一套 RocketMQ 集群，靠 Topic / ConsumerGroup 隔离。
// 实时用户侧三个 Topic 仅由 push 服务生产；离线 Topic 仅由 push 服务生产并消费。
// 编排与日志 Topic 不经 push 服务。

type TopicType string

const (
	TopicRealtimeChat TopicType = "im-realtime-chat"
	TopicRealtimeSync TopicType = "im-realtime-sync"
	TopicRealtimeBot  TopicType = "im-realtime-bot"
	TopicOfflinePush  TopicType = "im-offline-push"
	TopicAgentTask    TopicType = "im-agent-task"
	TopicAgentDone    TopicType = "im-agent-done"
	TopicWorkflowTask TopicType = "im-workflow-task"
)

type GroupType string

// ConsumerGroup 与 Topic 一对一镜像，后缀 -cg（consumer group）。
const (
	GroupRealtimeChat GroupType = "im-realtime-chat-cg"
	GroupRealtimeSync GroupType = "im-realtime-sync-cg"
	GroupRealtimeBot  GroupType = "im-realtime-bot-cg"
	GroupOfflinePush  GroupType = "im-offline-push-cg"
	GroupAgentTask    GroupType = "im-agent-task-cg"
	GroupAgentDone    GroupType = "im-agent-done-cg"
	GroupWorkflowTask GroupType = "im-workflow-task-cg"
)
