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

package agenttopicconst

import "time"

// 会话当前话题的 Redis key 前缀。key = agent:topic:{conversationId}
//
// 一个 IM 会话下可以开多个话题，每个话题是独立的 AI 上下文边界。
// 话题 ID 是服务端编排状态，客户端不持有：
//   - AgentTaskConsumer 有值就续该话题，没值就让 AI 侧新建；
//   - 客户端点「新话题」时由 message_api 删掉这个 key。
const KeyPrefix = "agent:topic:"

// AgentTopicTTL 话题缓存有效期。会话长期不用就让缓存自然过期，下次自动开新话题。
const AgentTopicTTL = 30 * 24 * time.Hour

// Key 返回会话当前话题的缓存 key。
func Key(conversationID string) string {
	return KeyPrefix + conversationID
}
