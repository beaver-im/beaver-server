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
	"fmt"

	"beaver/app/push/push_api/internal/svc"
	"beaver/app/push/push_api/internal/types"
	"beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type PushToUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewPushToUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushToUsersLogic {
	return &PushToUsersLogic{ctx: ctx, svcCtx: svcCtx, logger: beaverlog.New("push_to_users", ctx)}
}

func (l *PushToUsersLogic) PushToUsers(req *types.PushToUsersReq) (*types.PushToUsersRes, error) {
	if len(req.Envelopes) > maxBatchSize {
		return nil, fmt.Errorf("单次最多 %d 个 envelope", maxBatchSize)
	}
	failed := make([]string, 0)
	for _, ref := range req.Envelopes {
		topic, err := topicForChannel(ref.Channel)
		if err != nil {
			failed = append(failed, ref.TargetID)
			continue
		}
		payload := map[string]interface{}{
			"targetId":       ref.TargetID,
			"command":        ref.Command,
			"type":           ref.Type,
			"conversationId": ref.ConversationID,
			"senderId":       ref.SenderID,
			"body":           ref.Body,
		}
		if err := l.svcCtx.RocketMQ.SendMessage(l.ctx, topic, payload); err != nil {
			failed = append(failed, ref.TargetID)
			l.logger.Error(model.LogMsg{Text: "批量实时推送单项失败", Data: map[string]any{"targetId": ref.TargetID, "err": err.Error()}})
		}
	}
	return &types.PushToUsersRes{FailedTargetIDs: failed}, nil
}
