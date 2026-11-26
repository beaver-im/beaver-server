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

package push

import (
	"context"

	"beaver/commonconfig"
	pushtype "beaver/types/remote/push"
)

// SendOfflineApi 把离线厂商推送任务入队。messageID 是幂等键，由调用方传入。
// 返回 nil 只表示 push_api 已接收任务，不等于已送达客户端。
func SendOfflineApi(ctx context.Context, userID, title, body, messageID, conversationID string, isMuted bool, data map[string]string) error {
	req := pushtype.OfflineReq{
		UserID: userID, Title: title, Body: body, Data: data,
		ConversationID: conversationID, IsMuted: isMuted, MessageID: messageID,
	}
	_, err := Post[pushtype.OfflineRes](ctx,
		commonconfig.PushUrl+"/beaver/api/push/offline/v1/push_offline", nil, req)
	return err
}
