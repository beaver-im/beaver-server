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

package logic

import (
	"context"
	"errors"

	"beaver/app/message/message_api/internal/svc"
	"beaver/app/message/message_api/internal/types"
	message_models "beaver/app/message/message_models/chat"
	"beaver/common/const/agenttopicconst"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

// NewChatTopicLogic 在当前会话开启新话题。
//
// 话题是 AI 上下文边界，topicId 由 AI 侧持有；这里只做一件事：
// 删掉会话当前的 topicId 缓存。下一条 IM 消息经 AgentTaskConsumer 编排时
// 读到空值，就会让 AI 侧新建话题——旧话题的消息仍在，上下文从零开始。
type NewChatTopicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewNewChatTopicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NewChatTopicLogic {
	return &NewChatTopicLogic{
		ctx:    ctx,
		logger: beaverlog.New("new_chat_topic", ctx),
		svcCtx: svcCtx,
	}
}

func (l *NewChatTopicLogic) NewChatTopic(req *types.NewChatTopicReq) (*types.NewChatTopicRes, error) {
	if req.UserID == "" {
		return nil, errors.New("未登录")
	}
	if req.ConversationID == "" {
		return nil, errors.New("conversationId 不能为空")
	}

	// 必须是对话成员，防止清掉别人的话题缓存。
	var count int64
	if err := l.svcCtx.DB.Model(&message_models.ChatUserConversation{}).
		Where("user_id = ? AND conversation_id = ?", req.UserID, req.ConversationID).
		Count(&count).Error; err != nil {
		l.logger.Error(model.LogMsg{Text: "校验会话成员失败", Data: map[string]any{"userId": req.UserID, "conversationId": req.ConversationID, "err": err.Error()}})
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("不是会话成员")
	}

	if err := l.svcCtx.Redis.Del(agenttopicconst.Key(req.ConversationID)).Err(); err != nil {
		l.logger.Error(model.LogMsg{Text: "清除会话当前话题失败", Data: map[string]any{"conversationId": req.ConversationID, "err": err.Error()}})
		return nil, err
	}

	l.logger.Info(model.LogMsg{Text: "已开启新话题", Data: map[string]any{"userId": req.UserID, "conversationId": req.ConversationID}})
	return &types.NewChatTopicRes{}, nil
}
