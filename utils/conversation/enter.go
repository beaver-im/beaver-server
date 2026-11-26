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

package conversation

import (
	"errors"
	"sort"
	"strings"

	"beaver/common/const/chattypeconst"
)

// 会话类型（chatType）的唯一真相源是 ChatConversationMeta.Type。
// 本包只负责会话ID的构造与解析，绝不从会话ID前缀推断类型——
// 前缀仅用于定位ID里实体段的位置，类型必须由调用方从数据里读出来传进来。

/**
 * @description: 生成两人会话Id
 */
func GenerateConversation(userIds []string) (string, error) {
	if len(userIds) == 1 {
		return "private_" + userIds[0], nil
	} else if len(userIds) == 2 {
		sort.Strings(userIds)
		return "private_" + strings.Join(userIds, "_"), nil
	} else {
		return "", errors.New("userIds must have a length of 1 or 2")
	}
}

/**
 * @description: 按会话类型解析会话ID中的实体ID（已去掉类型前缀）
 * @param chatType: 会话类型，必须来自 ChatConversationMeta.Type
 * @param conversationID: 会话ID
 * @return: 实体ID列表。两人会话返回两个用户ID，群聊/圈子返回一个业务ID
 */
func ParseConversationIDs(chatType int, conversationID string) []string {
	if conversationID == "" {
		return nil
	}

	body := conversationID
	switch {
	case chattypeconst.IsP2P(chatType):
		body = strings.TrimPrefix(conversationID, "private_")
	case chatType == chattypeconst.Group:
		body = strings.TrimPrefix(conversationID, "group_")
	case chatType == chattypeconst.Circle:
		body = strings.TrimPrefix(conversationID, "circle_")
	}

	if body == "" {
		return nil
	}
	return strings.Split(body, "_")
}

/**
 * @description: 从会话ID中提取对应的目标ID（对方UID或群组/圈子ID）
 * @param chatType: 会话类型，必须来自 ChatConversationMeta.Type
 * @param conversationID: 会话ID
 * @param currentUserID: 当前用户ID（两人会话用于区分对端）
 * @return: targetId (对方UID或群组/圈子UUID)
 */
func GetTargetIDByConversation(chatType int, conversationID string, currentUserID string) string {
	ids := ParseConversationIDs(chatType, conversationID)
	if len(ids) == 0 {
		return ""
	}

	if chattypeconst.IsP2P(chatType) {
		if len(ids) == 2 {
			if ids[0] == currentUserID {
				return ids[1]
			}
			return ids[0]
		}
		return ids[0]
	}

	// 群聊 / 圈子：去掉类型前缀后的第一个段就是业务ID
	return ids[0]
}
