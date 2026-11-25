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

// Package s2s 机器票（OAuth2 client_credentials，HS256）的纯逻辑：签发 / 验签 / 调用方凭据校验 / Principal context。
// 不依赖 http.Handler；HTTP 中间件接线在 middleware/jwt/jwtS2s.go。
package s2s

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v4"
)

const (
	// S2STTL 换发的 S2S access token 默认有效期。
	S2STTL = 30 * time.Minute
	// HeaderClientID 验签通过后重建的、不可伪造的调用方标识头（写入下游 context 前剥离外部传入的同名头）。
	HeaderClientID = "X-Client-Id"
	// AudienceCore 本团队（A 主体）全部业务面共用的机器票受众。
	AudienceCore = "core:api"
)

// S2SClaims 是 S2S access token 的载荷。
type S2SClaims struct {
	ClientID string `json:"cid"` // 调用方 client_id
	jwtlib.StandardClaims
}

// Issue 用 secret 给 clientID 签发 audience=aud 的 HS256 JWT access token。
//
// issuer 标识签发方（一般取本团队服务名）。
func Issue(secret, issuer, clientID, aud string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = S2STTL
	}
	now := time.Now()
	claims := S2SClaims{
		ClientID: clientID,
		StandardClaims: jwtlib.StandardClaims{
			Issuer:    issuer,
			Subject:   clientID,
			Audience:  aud,
			ExpiresAt: now.Add(ttl).Unix(),
			IssuedAt:  now.Unix(),
			Id:        randomHex(16),
		},
	}
	signed, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// Verify 验 HS256 JWT：签名 + 受众 + 过期。返回调用方 clientID。
func Verify(secret, aud, raw string) (string, error) {
	parsed, err := jwtlib.ParseWithClaims(raw, &S2SClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期签名算法: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	c, ok := parsed.Claims.(*S2SClaims)
	if !ok || !parsed.Valid {
		return "", errors.New("无效的令牌")
	}
	if !c.VerifyAudience(aud, true) {
		return "", errors.New("令牌受众不匹配")
	}
	return c.ClientID, nil
}

type ctxKey struct{}

// WithPrincipal 把已验签的调用方 clientID 注入 context（不可伪造：外部传入的同名头已被剥离）。
func WithPrincipal(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, clientID)
}

// PrincipalFromContext 取已验签的调用方 clientID（无则空串）。
func PrincipalFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

// ClientEntry 被调方入站注册表里的一个调用方条目。
// 注册表的 key 就是调用方 client_id（= 其 yaml 的 Name），条目不重复记身份；
// 条目即授权：存在即可用，删除条目即下线该调用方，不设开关位。
type ClientEntry struct {
	ClientSecret string
}

// AuthenticateClient 校验 /token 端点的调用方：以 clientID 查注册表、client_secret
// 常量时间比对一致。
//
// fail-closed：空 clientID / 空 clientSecret / 注册表里空口令一律拒绝，避免 ${VAR} 未注入时
// 退化成「两个空串比对相等」而放行。
func AuthenticateClient(clients map[string]ClientEntry, clientID, clientSecret string) bool {
	if clientID == "" || clientSecret == "" {
		return false
	}
	c, ok := clients[clientID]
	if !ok || c.ClientSecret == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(clientSecret), []byte(c.ClientSecret)) == 1
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
