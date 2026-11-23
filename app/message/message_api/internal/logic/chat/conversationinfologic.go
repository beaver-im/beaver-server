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

	accttypes "beaver/app/account/account_rpc/pb"
	"beaver/app/message/message_api/internal/svc"
	"beaver/app/message/message_api/internal/types"
	message_models "beaver/app/message/message_models/chat"
	relationtypes "beaver/app/relation/relation_rpc/pb"
	"beaver/common/const/chattypeconst"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
	"beaver/utils/conversation"

	"gorm.io/gorm"
)

type ConversationInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewConversationInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConversationInfoLogic {
	return &ConversationInfoLogic{
		ctx:    ctx,
		logger: beaverlog.New("conversation_info", ctx),
		svcCtx: svcCtx,
	}
}

func (l *ConversationInfoLogic) ConversationInfo(req *types.ConversationInfoReq) (resp *types.ConversationInfoRes, err error) {
	var userConversation message_models.ChatUserConversation
	err = l.svcCtx.DB.Where("conversation_id = ? AND user_id = ?", req.ConversationID, req.UserID).First(&userConversation).Error

	resp = &types.ConversationInfoRes{
		ConversationID: req.ConversationID,
		MsgPreview:     "",
		UpdatedAt:      "",
		IsTop:          false,
	}

	var conversationMeta message_models.ChatConversationMeta
	metaErr := l.svcCtx.DB.Where("conversation_id = ?", req.ConversationID).First(&conversationMeta).Error
	if metaErr != nil {
		l.logger.Error(model.LogMsg{
			Text: "查询会话元数据失败",
			Data: map[string]any{
				"userId":         req.UserID,
				"conversationId": req.ConversationID,
				"err":            metaErr.Error(),
			},
		})
		if errors.Is(metaErr, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在")
		}
		return nil, metaErr
	}

	if err == nil {
		resp.MsgPreview = conversationMeta.LastMessage
		resp.UpdatedAt = userConversation.UpdatedAt.String()
		resp.IsTop = userConversation.IsPinned
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		l.logger.Error(model.LogMsg{
			Text: "查询会话关系失败",
			Data: map[string]any{
				"userId":         req.UserID,
				"conversationId": req.ConversationID,
				"err":            err.Error(),
			},
		})
		return nil, err
	}

	// 会话类型以会话元数据为准，不从会话ID前缀推断
	convType := conversationMeta.Type
	resp.ChatType = convType
	switch {
	case convType == chattypeconst.Circle: // 圈子（名称/头像由客户端补齐）
		resp.NickName = "圈子"
		resp.Avatar = ""
	case chattypeconst.IsP2P(convType): // 两人会话（普通私聊/数字伙伴/机器人）
		ids := conversation.ParseConversationIDs(convType, req.ConversationID)
		opponentID := ""
		if len(ids) == 2 {
			if ids[0] == req.UserID {
				opponentID = ids[1]
			} else {
				opponentID = ids[0]
			}
		} else if len(ids) == 1 {
			opponentID = ids[0]
		}
		userRes, userErr := l.svcCtx.UserRpc.UserInfo(l.ctx, &accttypes.UserInfoReq{
			UserID: opponentID,
		})
		if userErr != nil {
			l.logger.Error(model.LogMsg{
				Text: "获取对方用户信息失败",
				Data: map[string]any{
					"userId":         req.UserID,
					"conversationId": req.ConversationID,
					"opponentId":     opponentID,
					"err":            userErr.Error(),
				},
			})
			return nil, userErr
		}
		if userRes.UserInfo == nil {
			l.logger.Warn(model.LogMsg{
				Text: "用户不存在",
				Data: map[string]any{
					"userId":         req.UserID,
					"conversationId": req.ConversationID,
					"opponentId":     opponentID,
				},
			})
			return nil, errors.New("用户不存在")
		}
		resp.Avatar = userRes.UserInfo.Avatar
		resp.NickName = userRes.UserInfo.NickName
	default: // 群聊
		groupID := conversation.GetTargetIDByConversation(chattypeconst.Group, req.ConversationID, req.UserID)
		groupRes, groupErr := l.svcCtx.GroupRpc.GetGroupsListByIds(l.ctx, &relationtypes.GetGroupsListByIdsReq{
			GroupIDs: []string{groupID},
		})
		if groupErr != nil {
			l.logger.Error(model.LogMsg{
				Text: "获取群组信息失败",
				Data: map[string]any{
					"userId":         req.UserID,
					"conversationId": req.ConversationID,
					"groupId":        groupID,
					"err":            groupErr.Error(),
				},
			})
			return nil, groupErr
		}
		if len(groupRes.Groups) == 0 {
			l.logger.Warn(model.LogMsg{
				Text: "群组不存在",
				Data: map[string]any{
					"userId":         req.UserID,
					"conversationId": req.ConversationID,
					"groupId":        groupID,
				},
			})
			return nil, errors.New("群组不存在")
		}
		group := groupRes.Groups[0]
		resp.Avatar = group.Avatar
		resp.NickName = group.Name
	}

	l.logger.Info(model.LogMsg{
		Text: "查询会话详情成功",
		Data: map[string]any{
			"userId":         req.UserID,
			"conversationId": req.ConversationID,
			"chatType":       resp.ChatType,
		},
	})
	return resp, nil
}
