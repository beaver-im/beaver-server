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
	"fmt"

	"beaver/common/const/mqwsconst"
)

const maxBatchSize = 100

// topicForChannel 把推送通道映射到唯一 Topic。channel 由调用方传入，chat/sync/bot 三选一。
func topicForChannel(channel string) (mqwsconst.TopicType, error) {
	switch channel {
	case "chat":
		return mqwsconst.TopicRealtimeChat, nil
	case "sync":
		return mqwsconst.TopicRealtimeSync, nil
	case "bot":
		return mqwsconst.TopicRealtimeBot, nil
	default:
		return "", fmt.Errorf("未知 channel: %s", channel)
	}
}
