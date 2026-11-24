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
	"fmt"
	"path/filepath"
	"time"

	"beaver/app/asset/asset_api/internal/svc"
	"beaver/app/asset/asset_api/internal/types"
	file_models "beaver/app/asset/asset_models/file"
	beaverlog "beaver/utils/beaverlog"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/zeromicro/go-zero/core/logx"

	"gorm.io/gorm"
)

// goctl 会把 logx 写进 imports，占位避免未使用导入。
var _ = logx.WithContext

// 预览业务错误：handler 据此映射 HTTP 状态码，不向客户端暴露内部细节。
var (
	ErrInvalidFileKey    = errors.New("缺少 fileKey 参数")
	ErrFileNotFound      = errors.New("文件不存在")
	ErrUnsupportedSource = errors.New("不支持的文件来源")
)

type PreviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

// 文件预览（公开面）：校验/查库/签名决策在此完成，原始 HTTP 输出交给 handler。
func NewPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewLogic {
	return &PreviewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("Preview", ctx),
	}
}

// Preview 按 fileKey 查文件记录并生成预览决策：
// - qiniu/minio：填 RedirectURL，handler 302 到签名 URL；
// - local：填 LocalPath/ContentType，handler 直接输出文件流。
func (l *PreviewLogic) Preview(req *types.PreviewReq) (*types.PreviewRes, error) {
	if req.FileKey == "" {
		return nil, ErrInvalidFileKey
	}

	var m file_models.FileModel
	if err := l.svcCtx.DB.Take(&m, "file_key = ?", req.FileKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("查询文件记录失败: %w", err)
	}

	res := &types.PreviewRes{
		Source:       string(m.Source),
		OriginalName: m.OriginalName,
	}
	switch m.Source {
	case file_models.QiniuSource:
		mac := qbox.NewMac(l.svcCtx.Config.Qiniu.AK, l.svcCtx.Config.Qiniu.SK)
		expire := l.svcCtx.Config.Qiniu.ExpireTime
		if expire <= 0 {
			expire = 3600
		}
		deadline := time.Now().Add(time.Duration(expire) * time.Second).Unix()
		res.RedirectURL = storage.MakePrivateURL(mac, l.svcCtx.Config.Qiniu.Domain, m.Path, deadline)
	case file_models.MinioSource:
		if l.svcCtx.MinioClient == nil {
			return nil, errors.New("MinIO 未配置")
		}
		u, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, l.svcCtx.Config.Minio.Bucket, m.Path, 3600*time.Second, nil)
		if err != nil {
			return nil, fmt.Errorf("生成 MinIO 预签名 URL 失败: %w", err)
		}
		res.RedirectURL = u.String()
	case file_models.LocalSource:
		res.LocalPath = filepath.Join(l.svcCtx.Config.Local.UploadDir, m.Path)
		res.ContentType = contentTypeFor(m.Type)
	default:
		return nil, ErrUnsupportedSource
	}
	return res, nil
}

// contentTypeFor 按文件类型给出下载/内联展示用的 Content-Type。
func contentTypeFor(fileType string) string {
	switch fileType {
	case "image":
		return "image/*"
	case "video":
		return "video/*"
	case "audio":
		return "audio/*"
	case "document":
		return "application/pdf"
	case "archive":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
