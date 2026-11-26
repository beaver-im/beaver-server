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

// push 推送中台 S2S 出入参类型。
// 字段名与 json tag 与 push_api 的 api/*.api 定义逐字段一致（lowerCamel），
// 保证 httpx.Parse 能正确接收、ajax 信封能正确解析 result。
package push

// SendRealtimeApi 实时帧推送：Req + Res。
type (
	RealtimeReq struct {
		TargetID       string      `json:"targetId"`
		Channel        string      `json:"channel"`
		Command        string      `json:"command"`
		Type           string      `json:"type"`
		ConversationID string      `json:"conversationId,omitempty"`
		SenderID       string      `json:"senderId,omitempty"`
		Body           interface{} `json:"body"`
	}
	RealtimeRes struct {
		FailedTargetIDs []string `json:"failedTargetIds,omitempty"`
	}
)

// SendBotApi 机器人长连接事件推送：Req + Res。
type (
	BotReq struct {
		DigitalID      string      `json:"digitalId"`
		EventType      string      `json:"eventType"`
		ConversationID string      `json:"conversationId,omitempty"`
		Body           interface{} `json:"body"`
	}
	BotRes struct {
		FailedTargetIDs []string `json:"failedTargetIds,omitempty"`
	}
)
