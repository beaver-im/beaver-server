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
	"strings"
	"time"

	"beaver/app/asset/asset_api/internal/svc"
	"beaver/app/asset/asset_api/internal/types"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
	"beaver/utils/jwts"
	"beaver/utils/s2s"
)

const uploadTokenClientID = "backend_admin"

type GetUploadTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

// 队间：签发文件上传用户票
func NewGetUploadTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadTokenLogic {
	return &GetUploadTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("GetUploadToken", ctx),
	}
}

func (l *GetUploadTokenLogic) GetUploadToken(req *types.GetUploadTokenReq) (*types.GetUploadTokenRes, error) {
	if s2s.PrincipalFromContext(l.ctx) != uploadTokenClientID {
		return nil, errors.New("无权获取文件上传 token")
	}
	if strings.TrimSpace(req.UserID) == "" {
		return nil, errors.New("缺少上传用户身份")
	}
	secret := strings.TrimSpace(l.svcCtx.Config.Auth.AccessSecret)
	if secret == "" || l.svcCtx.Config.Auth.AccessExpire <= 0 {
		return nil, errors.New("上传凭证配置不完整")
	}

	ttl := time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second
	token, err := jwts.GenTokenTTL(jwts.JwtPayLoad{UserID: req.UserID}, secret, ttl, jwts.AudienceUser)
	if err != nil {
		l.logger.Error(model.LogMsg{Text: "签发文件上传 token 失败", Data: map[string]interface{}{"err": err.Error()}})
		return nil, errors.New("签发文件上传 token 失败")
	}

	l.logger.Info(model.LogMsg{Text: "签发文件上传 token 成功", Data: map[string]interface{}{"expiresIn": int64(ttl.Seconds())}})
	return &types.GetUploadTokenRes{UploadToken: token, ExpiresIn: int64(ttl.Seconds())}, nil
}
