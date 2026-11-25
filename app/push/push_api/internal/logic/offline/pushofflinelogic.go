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

	"beaver/app/push/push_api/internal/svc"
	"beaver/app/push/push_api/internal/types"
	"beaver/common/const/mqwsconst"

	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type PushOfflineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewPushOfflineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushOfflineLogic {
	return &PushOfflineLogic{ctx: ctx, svcCtx: svcCtx, logger: beaverlog.New("push_offline", ctx)}
}

func (l *PushOfflineLogic) PushOffline(req *types.PushOfflineReq) (*types.PushOfflineRes, error) {
	if req.UserID == "" || req.MessageID == "" {
		return nil, errors.New("userId 与 messageId 不能为空，离线推送需要幂等键")
	}
	// 直接投递扁平任务体：req 的 json tag 与 Topic 消费契约一致（userId/messageId/isMuted...）。
	if err := l.svcCtx.RocketMQ.SendRawJSON(l.ctx, mqwsconst.TopicOfflinePush, req); err != nil {
		l.logger.Error(model.LogMsg{Text: "离线推送任务投递失败", Data: map[string]any{"messageId": req.MessageID, "err": err.Error()}})
		return nil, err
	}
	return &types.PushOfflineRes{Accepted: true}, nil
}
