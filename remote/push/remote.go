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
	"fmt"
	"net/http"
	"time"

	"beaver/utils/ajax"
)

// 本包唯一的出站通道：换票 + 带票发包，其余不碰（不重试、不聚合、不缓存业务结果）。
// 换票走 token.go 的 GetTokenApi → push_api 的 auth_public/v1/issue；过期时间以对端 expiresIn 为准。
const (
	tokenTTL  = 30 * time.Minute
	tokenSkew = 30 * time.Second
)

// Post 发一次 POST 的 JSON 请求，返回值是信封 result 解成的 Res。
func Post[Res, Req any](ctx context.Context, endpoint string, header http.Header, body Req) (*Res, error) {
	h, err := client.authHeader(ctx, header)
	if err != nil {
		return nil, fmt.Errorf("%s：换取出站令牌失败: %w", endpoint, err)
	}
	return ajax.Post[Res](ctx, endpoint, "", h, body)
}

// authHeader 把调用方给的头复制一份，再补上 Authorization。
func (c *Client) authHeader(ctx context.Context, header http.Header) (http.Header, error) {
	bearer, err := c.bearerToken(ctx)
	if err != nil {
		return nil, err
	}
	h := http.Header{}
	for k, vs := range header {
		for _, v := range vs {
			h.Set(k, v)
		}
	}
	h.Set("Authorization", bearer)
	return h, nil
}

// bearerToken 取本次请求要带的 Authorization 值，必要时先换票。
// 缓存到过期前 30 秒；锁住整个换票过程，让同时到达的一批请求只打一次 /issue。
func (c *Client) bearerToken(ctx context.Context) (string, error) {
	if c == nil || c.clientID == "" || c.clientSecret == "" {
		return "", fmt.Errorf("push 出站身份或凭据未配置（c.Name / RemoteProxy.Push.ClientSecret）")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bearer != "" && time.Now().Before(c.expiresAt.Add(-tokenSkew)) {
		return c.bearer, nil
	}
	tok, exp, err := c.exchange(ctx)
	if err != nil {
		return "", err
	}
	c.bearer, c.expiresAt = tok, exp
	return c.bearer, nil
}

// exchange 换票：凭据交给 token.go 的 GetTokenApi 打一次 /issue，这里只算过期时间。
func (c *Client) exchange(ctx context.Context) (string, time.Time, error) {
	token, expiresIn, err := GetTokenApi(ctx, c.clientID, c.clientSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	ttl := tokenTTL
	if expiresIn > 0 {
		ttl = time.Duration(expiresIn) * time.Second
	}
	return "Bearer " + token, time.Now().Add(ttl), nil
}
