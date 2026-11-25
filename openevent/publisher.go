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

package openevent

import (
	"context"
	"errors"

	pushcli "beaver/remote/push"
)

// Push 把机器人事件投递到推送中台，由 push_api 转发到 bot 长连接 Topic。
// 业务侧不再持有 MQ 客户端，也不拼实时 Topic payload；出站统一走 remote/push。
func Push(ctx context.Context, digitalID, eventType, conversationID string, body map[string]interface{}) error {
	if digitalID == "" || eventType == "" {
		return errors.New("缺少 digitalId / eventType")
	}
	return pushcli.SendBotApi(ctx, digitalID, eventType, conversationID, body)
}
