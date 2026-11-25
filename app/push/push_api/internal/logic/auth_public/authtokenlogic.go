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
	"beaver/commonconfig"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
	"beaver/utils/s2s"
)

// AuthTokenLogic 机器票签发：client_credentials → 受众 core:api 的访问凭据。
//
// push_api 自带换票端点（与 account_api 同构）：出站方拿 RemoteProxy.Push 的
// clientID/clientSecret 换票，不再持有签名密钥本地自签。fail-closed：注册表里
// 没有该 client、口令不一致一律拒绝；口令走常量时间比对。
type AuthTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewAuthTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthTokenLogic {
	return &AuthTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("AuthToken", ctx),
	}
}

func (l *AuthTokenLogic) AuthToken(req *types.AuthTokenReq) (resp *types.AuthTokenRes, err error) {
	if req.GrantType != "" && req.GrantType != commonconfig.AuthClientCredentials {
		return nil, errors.New("不支持的授权方式")
	}

	if !s2s.AuthenticateClient(l.svcCtx.Config.S2S.Clients, req.ClientID, req.ClientSecret) {
		l.logger.Error(model.LogMsg{
			Text: "秘钥签发被拒",
			Data: map[string]any{"clientId": req.ClientID},
		})
		return nil, errors.New("invalid_client")
	}

	signed, err := s2s.Issue(
		l.svcCtx.Config.S2S.Secret,
		l.svcCtx.Config.Name,
		req.ClientID,
		s2s.AudienceCore,
		s2s.S2STTL,
	)
	if err != nil {
		l.logger.Error(model.LogMsg{
			Text: "签发访问秘钥失败",
			Data: map[string]any{"clientId": req.ClientID, "err": err.Error()},
		})
		return nil, err
	}

	return &types.AuthTokenRes{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s2s.S2STTL.Seconds()),
	}, nil
}
