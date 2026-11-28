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

package file_public

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"

	"beaver/app/asset/asset_api/internal/handler/common"
	"beaver/app/asset/asset_api/internal/svc"
	"beaver/app/asset/asset_api/internal/types"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/jwts"
)

type FileUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

// 文件上传（经网关；Bearer 须为 beaver:upload）
func NewFileUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadLogic {
	return &FileUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("file_upload", ctx),
	}
}

func (l *FileUploadLogic) FileUpload(req *types.FileUploadReq, file multipart.File, fileHead *multipart.FileHeader, fileInfoStr string) (*types.FileUploadRes, error) {
	_ = req
	store, source, err := common.DefaultStorage(l.svcCtx)
	if err != nil {
		return nil, err
	}
	return common.UploadFile(l.ctx, l.svcCtx, file, fileHead, fileInfoStr, store, source)
}

// VerifyUploadToken 校验第三平面短时上传票，返回归属 userId。
func VerifyUploadToken(svcCtx *svc.ServiceContext, rawToken string) (string, error) {
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return "", errors.New("缺少上传凭证")
	}
	secret := strings.TrimSpace(svcCtx.Config.Upload.AccessSecret)
	if secret == "" {
		return "", errors.New("上传凭证配置不完整")
	}
	claims, err := jwts.ParseToken(token, secret)
	if err != nil || claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return "", errors.New("上传凭证无效或已过期")
	}
	if !claims.VerifyAudience(jwts.AudienceUpload, true) {
		return "", errors.New("上传凭证与能力平面不匹配")
	}
	return claims.UserID, nil
}
