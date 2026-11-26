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

	"beaver/app/message/message_api/internal/svc"
	"beaver/app/message/message_api/internal/types"
	message_models "beaver/app/message/message_models/chat"
	relationtypes "beaver/app/relation/relation_rpc/pb"
	"beaver/common/const/chattypeconst"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
	"beaver/utils/conversation"
)

type RecentChatListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewRecentChatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecentChatListLogic {
	return &RecentChatListLogic{
		ctx:    ctx,
		logger: beaverlog.New("recent_chat_list", ctx),
		svcCtx: svcCtx,
	}
}

func (l *RecentChatListLogic) RecentChatList(req *types.RecentChatListReq) (resp *types.RecentChatListRes, err error) {
	var userConversations []message_models.ChatUserConversation
	var conversationMetas []message_models.ChatConversationMeta

	// 获取该用户的所有未隐藏会话
	err = l.svcCtx.DB.Where("user_id = ? AND is_hidden = ?", req.UserID, false).
		Order("is_pinned DESC, updated_at DESC").
		Find(&userConversations).Error
	if err != nil {
		l.logger.Error(model.LogMsg{
			Text: "查询最近会话失败",
			Data: map[string]any{"userId": req.UserID, "err": err.Error()},
		})
		return nil, err
	}

	// 获取对应的会话元数据（包含LastMessage）
	conversationIds := make([]string, len(userConversations))
	for i, conv := range userConversations {
		conversationIds[i] = conv.ConversationID
	}

	if len(conversationIds) > 0 {
		err = l.svcCtx.DB.Where("conversation_id IN (?)", conversationIds).
			Find(&conversationMetas).Error
		if err != nil {
			l.logger.Error(model.LogMsg{
				Text: "查询会话元数据失败",
				Data: map[string]any{"userId": req.UserID, "err": err.Error()},
			})
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	// 会话类型（chatType）的唯一来源：ChatConversationMeta.Type
	conversationMetaMap := make(map[string]message_models.ChatConversationMeta, len(conversationMetas))
	for _, meta := range conversationMetas {
		conversationMetaMap[meta.ConversationID] = meta
	}

	var privateChatIds []string
	var groupChatIds []string

	// 按会话元数据里的 chatType 区分两人会话 / 群聊
	for _, convo := range userConversations {
		meta, exists := conversationMetaMap[convo.ConversationID]
		if !exists {
			l.logger.Error(model.LogMsg{
				Text: "会话元数据缺失，跳过该会话",
				Data: map[string]any{"userId": req.UserID, "conversationId": convo.ConversationID},
			})
			continue
		}
		if chattypeconst.IsP2P(meta.Type) {
			privateChatIds = append(privateChatIds, convo.ConversationID)
		} else if meta.Type == chattypeconst.Group {
			groupChatIds = append(groupChatIds, convo.ConversationID)
		}
	}

	// 从两人会话Id中提取用户Id进行去重
	userIdMap := make(map[string]struct{})
	for _, chatID := range privateChatIds {
		for _, userId := range conversation.ParseConversationIDs(conversationMetaMap[chatID].Type, chatID) {
			if userId != req.UserID {
				userIdMap[userId] = struct{}{}
			}
		}
	}

	// 将map转为slice
	var userIds []string
	for id := range userIdMap {
		userIds = append(userIds, id)
	}

	// 通过FriendRpc获取好友详细信息（包含用户基础信息和备注）
	var friendDetails []*relationtypes.FriendDetailItem
	if len(userIds) > 0 {
		friendDetailRes, err := l.svcCtx.FriendRpc.GetFriendDetail(l.ctx, &relationtypes.GetFriendDetailReq{
			UserId:    req.UserID,
			FriendIds: userIds,
		})
		if err != nil {
			l.logger.Error(model.LogMsg{
				Text: "获取好友详情失败",
				Data: map[string]any{"userId": req.UserID, "err": err.Error()},
			})
			return nil, err
		}
		friendDetails = friendDetailRes.Friends
	}

	// 通过RPC批量获取群组信息
	var groupDetails []*relationtypes.GroupListById
	if len(groupChatIds) > 0 {
		// 从群聊会话ID中提取实际的群组ID（去掉group_前缀）
		var actualGroupIds []string
		for _, groupChatId := range groupChatIds {
			ids := conversation.ParseConversationIDs(chattypeconst.Group, groupChatId)
			if len(ids) > 0 {
				actualGroupIds = append(actualGroupIds, ids[0])
			}
		}

		if len(actualGroupIds) > 0 {
			groupRes, err := l.svcCtx.GroupRpc.GetGroupsListByIds(l.ctx, &relationtypes.GetGroupsListByIdsReq{
				GroupIDs: actualGroupIds,
			})
			if err != nil {
				l.logger.Error(model.LogMsg{
					Text: "获取群组详情失败",
					Data: map[string]any{"userId": req.UserID, "err": err.Error()},
				})
				return nil, err
			}
			groupDetails = groupRes.Groups
		}
	}

	// 构建群组数据映射 (key为完整的会话ID)
	groupMap := make(map[string]*relationtypes.GroupListById)
	for i, groupChatId := range groupChatIds {
		if i < len(groupDetails) {
			groupMap[groupChatId] = groupDetails[i]
		}
	}

	// 构建好友详细信息映射
	friendDetailMap := make(map[string]*relationtypes.FriendDetailItem)
	for _, friendDetail := range friendDetails {
		friendDetailMap[friendDetail.UserId] = friendDetail
	}

	var respList []types.ConversationInfoRes
	for _, convo := range userConversations {
		meta, exists := conversationMetaMap[convo.ConversationID]
		if !exists {
			continue
		}

		var chatInfo types.ConversationInfoRes
		chatInfo.MsgPreview = meta.LastMessage
		chatInfo.IsTop = convo.IsPinned
		chatInfo.ConversationID = convo.ConversationID
		chatInfo.UpdatedAt = convo.UpdatedAt.String()
		// 会话类型直接下发：1普通私聊 2群聊 3圈子 4数字伙伴 5机器人
		chatInfo.ChatType = meta.Type

		if chattypeconst.IsP2P(meta.Type) { // 两人会话（普通私聊/数字伙伴/机器人）
			var opponentID string
			if ids := conversation.ParseConversationIDs(meta.Type, convo.ConversationID); len(ids) == 2 {
				if ids[0] == req.UserID {
					opponentID = ids[1]
				} else {
					opponentID = ids[0]
				}
			} else if len(ids) == 1 {
				opponentID = ids[0]
			}

			// 从好友详细信息中获取用户信息和备注
			if friendDetail, exists := friendDetailMap[opponentID]; exists {
				chatInfo.NickName = friendDetail.NickName
				chatInfo.Avatar = friendDetail.Avatar
				chatInfo.Notice = friendDetail.Notice
			} else {
				chatInfo.NickName = "未知用户"
				chatInfo.Avatar = ""
				chatInfo.Notice = ""
			}
		} else if meta.Type == chattypeconst.Circle { // 圈子（名称/头像由客户端用圈子详情补齐）
			chatInfo.NickName = "圈子"
			chatInfo.Avatar = ""
			chatInfo.Notice = ""
		} else { // 群聊
			group, exists := groupMap[convo.ConversationID]
			if exists {
				chatInfo.NickName = group.Name
				chatInfo.Avatar = group.Avatar
			} else {
				chatInfo.NickName = "未知群聊"
				chatInfo.Avatar = ""
			}
			chatInfo.Notice = "" // 群聊暂时没有备注功能
		}

		respList = append(respList, chatInfo)
	}

	l.logger.Info(model.LogMsg{
		Text: "拉取最近会话成功",
		Data: map[string]any{
			"userId": req.UserID,
			"count":  len(respList),
		},
	})
	return &types.RecentChatListRes{
		List:  respList,
		Count: int64(len(respList)),
	}, nil
}
