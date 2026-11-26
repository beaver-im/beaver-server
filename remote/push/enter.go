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

// Package push 是 A 主体推送网关（beaver-server 内部 push_api，:20500）的 S2S 出站客户端。
//
// 与其它 peer 完全一致：域名由 commonconfig.Load 从 yaml 的 RemoteProxy.Push.BaseURL 灌入
// commonconfig.PushUrl，凭据由 Init 装配；clientID 取本服务名（c.Name），不再单独配。
// 换票走 token.go 的 GetTokenApi 打 push_api 自带的 auth_public/v1/issue（client_credentials），
// 缓存挂在 client 上。
// 业务侧直接调包函数 SendRealtimeApi / SendBotApi / SendOfflineApi，不在 logic 里拼 URL 或取票。
//
// Req/Res 放在 types/remote/push，本包不依赖 push_api 模块。
package push

import (
	"sync"
	"time"
)

// client 本包唯一的出站客户端，Init 时装配。
var client *Client

// Init 装配出站凭据（clientID 取服务名 c.Name，clientSecret 取 RemoteProxy.Push）。只在该服务装配依赖时调用一次。
//
// 空身份或空凭据直接崩掉启动：这类配置错误若留到运行期只会退化成 401，排查成本远高于启动失败。
func Init(clientID, clientSecret string) {
	if clientID == "" {
		panic("push 出站身份为空：c.Name 未配置")
	}
	if clientSecret == "" {
		panic("push 出站凭据为空：RemoteProxy.Push.ClientSecret 未配置")
	}
	client = &Client{clientID: clientID, clientSecret: clientSecret}
}

// Client 持凭据与换来的票，换票缓存挂在它上面。
type Client struct {
	clientID     string
	clientSecret string

	mu        sync.Mutex
	bearer    string
	expiresAt time.Time
}
