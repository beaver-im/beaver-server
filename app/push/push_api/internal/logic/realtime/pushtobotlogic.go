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
	"beaver/common/const/mqwsconst"
	"beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type PushToBotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewPushToBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushToBotLogic {
	return &PushToBotLogic{ctx: ctx, svcCtx: svcCtx, logger: beaverlog.New("push_to_bot", ctx)}
}

// PushToBot 机器人事件固定走 bot Topic；channel 不接受调用方传值。
func (l *PushToBotLogic) PushToBot(req *types.PushToBotReq) (*types.PushToBotRes, error) {
	payload := map[string]interface{}{
		"targetId":       req.DigitalID,
		"command":        "CHAT_MESSAGE",
		"type":           req.EventType,
		"conversationId": req.ConversationID,
		"body":           req.Body,
	}
	if err := l.svcCtx.RocketMQ.SendMessage(l.ctx, mqwsconst.TopicRealtimeBot, payload); err != nil {
		l.logger.Error(model.LogMsg{Text: "机器人推送投递失败", Data: map[string]any{"digitalId": req.DigitalID, "err": err.Error()}})
		return &types.PushToBotRes{FailedTargetIDs: []string{req.DigitalID}}, nil
	}
	return &types.PushToBotRes{}, nil
}
