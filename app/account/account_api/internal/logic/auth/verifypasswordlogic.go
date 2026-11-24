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
)

// VerifyPasswordLogic 校验账号密码
//
// 两种入参形态，都在本端点一次完成，调用方不需要先查用户再验密码：
//   - UserID：已经知道用户 ID，只验密码（管理端代登录、已有会话续期）；
//   - Account + Type：账号（邮箱 / 手机号 / userId）+ 密码，成功时把 userId 与昵称一并返回
//     （开放平台扫码/账密登录，需要拿 userId 去发自己的授权码）。
//
// 「账号不存在」与「密码错误」对外统一表现为 valid=false，不暴露账号是否存在。
type VerifyPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewVerifyPasswordLogic 构造 logic。
func NewVerifyPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyPasswordLogic {
	return &VerifyPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// VerifyPassword 校验账号密码
func (l *VerifyPasswordLogic) VerifyPassword(req *types.VerifyPasswordReq) (*types.VerifyPasswordRes, error) {
	userID, nickName := req.UserID, ""
	if userID == "" {
		resolvedID, resolvedName, ok := l.resolveAccount(req.Account, req.Type)
		if !ok {
			return &types.VerifyPasswordRes{Valid: false}, nil
		}
		userID, nickName = resolvedID, resolvedName
	}
	// 空密码不交给 rpc：rpc 对空参报错，而这里对调用方只应表现为「凭据无效」。
	if req.Password == "" {
		return &types.VerifyPasswordRes{Valid: false}, nil
	}

	rpcRes, err := l.svcCtx.AuthRpc.VerifyPassword(l.ctx, &accttypes.VerifyPasswordReq{
		UserId:   userID,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	if !rpcRes.Valid {
		return &types.VerifyPasswordRes{Valid: false}, nil
	}

	// 走 UserID 形态时没有解析过账号，昵称按需补一次；查不到不影响校验结果。
	if nickName == "" {
		if infoRes, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &accttypes.UserInfoReq{UserID: userID}); err == nil && infoRes.UserInfo != nil {
			nickName = infoRes.UserInfo.NickName
		}
	}

	return &types.VerifyPasswordRes{Valid: true, UserID: userID, NickName: nickName}, nil
}

// resolveAccount 把账号解析成用户：指定 type 时只按该类型查，未指定时按
// 邮箱 → 手机号 → userId 依次尝试。三种都是等值查询，命中即返回。
func (l *VerifyPasswordLogic) resolveAccount(account, accountType string) (userID, nickName string, ok bool) {
	if account == "" {
		return "", "", false
	}

	candidates := []string{accountType}
	if accountType == "" {
		candidates = []string{"email", "phone", "userId"}
	}

	for _, t := range candidates {
		res, err := l.svcCtx.UserRpc.SearchUser(l.ctx, &accttypes.SearchUserReq{
			Keyword: account,
			Type:    t,
		})
		if err == nil && res.UserInfo != nil && res.UserInfo.UserId != "" {
			return res.UserInfo.UserId, res.UserInfo.NickName, true
		}
	}
	return "", "", false
}
