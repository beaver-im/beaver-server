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

	"beaver/app/push/push_api/internal/svc"
	"beaver/app/push/push_api/internal/types"
	"beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type PushToUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewPushToUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushToUserLogic {
	return &PushToUserLogic{ctx: ctx, svcCtx: svcCtx, logger: beaverlog.New("push_to_user", ctx)}
}

// PushToUser 按 channel 把实时帧投给对应 Topic；Res 只表示 MQ 已接收，不等于已送达客户端。
func (l *PushToUserLogic) PushToUser(req *types.PushToUserReq) (*types.PushToUserRes, error) {
	topic, err := topicForChannel(req.Channel)
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"targetId":       req.TargetID,
		"command":        req.Command,
		"type":           req.Type,
		"conversationId": req.ConversationID,
		"senderId":       req.SenderID,
		"body":           req.Body,
	}
	if err := l.svcCtx.RocketMQ.SendMessage(l.ctx, topic, payload); err != nil {
		l.logger.Error(model.LogMsg{Text: "实时推送投递失败", Data: map[string]any{"targetId": req.TargetID, "err": err.Error()}})
		return &types.PushToUserRes{FailedTargetIDs: []string{req.TargetID}}, nil
	}
	return &types.PushToUserRes{}, nil
}
