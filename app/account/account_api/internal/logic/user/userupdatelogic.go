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

// UserUpdateLogic 批量更新用户资料（队间；open 同步 bot/digital 工号资料走此口）
type UserUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("user_update", ctx),
	}
}

func (l *UserUpdateLogic) UserUpdate(req *types.UserUpdateReq) (*types.UserUpdateRes, error) {
	rpcReq := &accttypes.UpdateUsersReq{
		UserIds:       req.UserIDs,
		PatchNickName: req.PatchNickName,
		PatchEmail:    req.PatchEmail,
		PatchAvatar:   req.PatchAvatar,
		PatchAbstract: req.PatchAbstract,
	}
	rpcRes, err := l.svcCtx.UserRpc.UpdateUsers(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 有展示字段变更时，按落库后的 version 推 USER_PROFILE
	if rpcRes != nil && rpcRes.AffectedCount > 0 && hasProfilePatch(req) {
		for _, uid := range req.UserIDs {
			if ver := loadUserVersion(l.svcCtx, uid); ver > 0 {
				NotifyUserProfileChange(l.svcCtx, l.logger, uid, ver)
			}
		}
	}

	return &types.UserUpdateRes{AffectedCount: rpcRes.AffectedCount}, nil
}

func hasProfilePatch(req *types.UserUpdateReq) bool {
	return req.PatchNickName != nil || req.PatchAvatar != nil || req.PatchAbstract != nil || req.PatchEmail != nil
}
