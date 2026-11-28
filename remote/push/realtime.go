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
	"errors"
	"reflect"
	"strings"

	"beaver/commonconfig"
	pushtype "beaver/types/remote/push"
)

// payloadString 取 payload 里的字符串值。调用方常传 wsCommandConst.Command /
// wsTypeConst.Type 这类命名字符串类型，底层虽是 string，但 .(string) 断言会失败，
// 必须用反射按字符串种类取值。
func payloadString(payload map[string]interface{}, key string) string {
	v, ok := payload[key]
	if !ok || v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.String {
		return rv.String()
	}
	return ""
}

// SendRealtimeApi 把实时帧投给 push_api（channel=chat/sync），由后者写 MQ → ws_api。
// payload 约定键：targetId / command / type / senderId / conversationId / body。
func SendRealtimeApi(ctx context.Context, channel string, payload map[string]interface{}) error {
	target := payloadString(payload, "targetId")
	command := payloadString(payload, "command")
	typ := payloadString(payload, "type")
	if target == "" || command == "" || typ == "" {
		return errors.New("推送 payload 缺少 targetId/command/type")
	}
	senderID := payloadString(payload, "senderId")
	conversationID := payloadString(payload, "conversationId")
	req := pushtype.RealtimeReq{
		TargetID: target, Channel: strings.TrimSpace(channel), Command: command, Type: typ,
		ConversationID: conversationID, SenderID: senderID, Body: payload["body"],
	}
	_, err := Post[pushtype.RealtimeRes](ctx,
		commonconfig.PushUrl+"/beaver/api/push/realtime/v1/push_to_user", nil, req)
	return err
}

// SendBotApi 把机器人事件投给 push_api，由后者写 bot Topic → open_api 长连接。
func SendBotApi(ctx context.Context, digitalID, eventType, conversationID string, body interface{}) error {
	if digitalID == "" || eventType == "" {
		return errors.New("缺少 digitalId / eventType")
	}
	req := pushtype.BotReq{
		DigitalID: digitalID, EventType: eventType, ConversationID: conversationID, Body: body,
	}
	_, err := Post[pushtype.BotRes](ctx,
		commonconfig.PushUrl+"/beaver/api/push/realtime/v1/push_to_bot", nil, req)
	return err
}
