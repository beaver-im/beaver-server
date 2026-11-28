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

	"beaver/app/account/account_api/internal/svc"
	"beaver/app/account/account_api/internal/types"
	accttypes "beaver/app/account/account_rpc/pb"
	beaverlog "beaver/utils/beaverlog"
)

// UserUpdateDisplayLogic 更新用户展示信息（队间 / open 工号实体常用）
type UserUpdateDisplayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewUserUpdateDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateDisplayLogic {
	return &UserUpdateDisplayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("user_update_display", ctx),
	}
}

func (l *UserUpdateDisplayLogic) UserUpdateDisplay(req *types.UserUpdateDisplayReq) (*types.UserUpdateDisplayRes, error) {
	rpcReq := &accttypes.UserUpdateDisplayReq{UserId: req.UserID, NickName: req.NickName, Avatar: req.Avatar}
	rpcRes, err := l.svcCtx.UserRpc.UserUpdateDisplay(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	if rpcRes != nil && rpcRes.Version > 0 {
		NotifyUserProfileChange(l.svcCtx, l.logger, req.UserID, rpcRes.Version)
	}
	return &types.UserUpdateDisplayRes{Version: rpcRes.Version}, nil
}
