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

// Package chattypeconst 定义「会话类型」（chatType）。
//
// chatType 是会话的唯一类型字段，落库在 ChatConversationMeta.Type。
// 数字伙伴 / 机器人都是独立的会话类型，与普通私聊并列，不靠会话ID前缀推断。
//
// 新增类型：在此加常量 + 在 conversation.InitializeConversation 里补建会话时的判定。
package chattypeconst

const (
	Private int = 1 // 普通私聊（好友之间）
	Group   int = 2 // 群聊
	Circle  int = 3 // 圈子
	Digital int = 4 // 数字伙伴（与 Agent 的两人会话）
	Bot     int = 5 // 机器人（与 Bot 的两人会话）
)

// IsP2P 判断是否为两人会话形态（普通私聊 / 数字伙伴 / 机器人）。
// 这三者都是「一个人 + 一个对端」，在成员、已读、typing 等逻辑上一致。
func IsP2P(t int) bool {
	return t == Private || t == Digital || t == Bot
}

// IsValid 判断是否为已定义的会话类型。
func IsValid(t int) bool {
	return t >= Private && t <= Bot
}
