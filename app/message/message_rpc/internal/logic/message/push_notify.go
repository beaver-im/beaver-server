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

package messagelogic

import (
	pushcli "beaver/remote/push"
	"strconv"

	message_models "beaver/app/message/message_models/chat"
	"beaver/utils/beaverlog/model"
)

// sendOfflinePushIfNeeded 只负责计算调用方已知的 is_muted 并交给推送中台。
// 不查在线态、不查 token、不直接调厂商；这些都在 push 服务的 OfflinePipeline 里。
func (l *SendMsgLogic) sendOfflinePushIfNeeded(
	conversationID, senderID string,
	chatModel message_models.ChatMessage,
	recipientIDs []string,
) {
	sender, err := l.getSenderInfo(chatModel)
	if err != nil {
		l.logger.Error(model.LogMsg{
			Text: "离线推送获取发送者失败",
			Data: map[string]any{"conversationId": conversationID, "err": err.Error()},
		})
		return
	}

	title := sender.NickName
	if title == "" {
		title = "新消息"
	}
	body := chatModel.MsgPreview
	if body == "" {
		body = "你收到一条新消息"
	}
	data := map[string]string{
		"type":           "chat_message",
		"conversationId": conversationID,
		"messageId":      chatModel.MessageID,
		"seq":            strconv.FormatInt(chatModel.Seq, 10),
		"senderId":       senderID,
	}

	for _, recipientID := range recipientIDs {
		if recipientID == senderID {
			continue
		}
		var userConvo message_models.ChatUserConversation
		isMuted := false
		if err := l.svcCtx.DB.Where("user_id = ? AND conversation_id = ?", recipientID, conversationID).
			First(&userConvo).Error; err == nil {
			isMuted = userConvo.IsMuted
		}
		if err := pushcli.SendOfflineApi(l.ctx, recipientID, title, body, chatModel.MessageID, conversationID, isMuted, data); err != nil {
			l.logger.Error(model.LogMsg{Text: "离线推送投递失败", Data: map[string]any{"recipientId": recipientID, "err": err.Error()}})
		}
	}
}
