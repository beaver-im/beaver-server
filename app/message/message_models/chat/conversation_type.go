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

package chat_models

import (
	"gorm.io/gorm"
)

// LoadConversationType 读取会话类型（chatType）。
//
// ChatConversationMeta.Type 是会话类型的唯一真相源，不通过会话ID前缀推断。
// 会话不存在时返回 gorm.ErrRecordNotFound，由调用方决定如何处理。
func LoadConversationType(db *gorm.DB, conversationID string) (int, error) {
	var meta ChatConversationMeta
	if err := db.Select("type").Where("conversation_id = ?", conversationID).First(&meta).Error; err != nil {
		return 0, err
	}
	return meta.Type, nil
}

// LoadConversationTypeMap 批量读取多个会话的类型，返回 conversationId -> chatType。
// 不存在的会话不会出现在结果里。
func LoadConversationTypeMap(db *gorm.DB, conversationIDs []string) (map[string]int, error) {
	result := make(map[string]int, len(conversationIDs))
	if len(conversationIDs) == 0 {
		return result, nil
	}

	var metas []ChatConversationMeta
	if err := db.Select("conversation_id", "type").
		Where("conversation_id IN (?)", conversationIDs).
		Find(&metas).Error; err != nil {
		return result, err
	}
	for _, meta := range metas {
		result[meta.ConversationID] = meta.Type
	}
	return result, nil
}
