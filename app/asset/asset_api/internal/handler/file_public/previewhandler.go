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

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	logic "beaver/app/asset/asset_api/internal/logic/file_public"
	"beaver/app/asset/asset_api/internal/svc"
	"beaver/app/asset/asset_api/internal/types"
	"beaver/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// PreviewHandler 是原始 HTTP 响应特例（文件流 / 302），不走 {code,msg,result} 信封。
// 校验、查库、签名决策在 logic，本函数只负责把决策转成 HTTP 输出。
func PreviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PreviewReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		l := logic.NewPreviewLogic(r.Context(), svcCtx)
		resp, err := l.Preview(&req)
		if err != nil {
			switch {
			case errors.Is(err, logic.ErrInvalidFileKey), errors.Is(err, logic.ErrUnsupportedSource):
				http.Error(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, logic.ErrFileNotFound):
				http.NotFound(w, r)
			default:
				// 内部错误不暴露 SQL / 签名 URL / 配置细节。
				http.Error(w, "文件预览失败", http.StatusInternalServerError)
			}
			return
		}

		if resp.RedirectURL != "" {
			http.Redirect(w, r, resp.RedirectURL, http.StatusFound)
			return
		}

		if resp.LocalPath == "" {
			http.Error(w, "文件预览失败", http.StatusInternalServerError)
			return
		}
		if _, err := os.Stat(resp.LocalPath); err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", resp.ContentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; FileKey=%s", resp.OriginalName))
		http.ServeFile(w, r, resp.LocalPath)
	}
}
