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

// Package jwt 业务 HTTP 入站闸门：北向网关断言（ES256+JWKS）或东西向 S2S 机器票（HS256）。
//
// 外部票只在网关验；网关签带 kid 的断言。业务面：
//   - 先验网关断言（JWKS 按 kid 取公钥，校验 iss/aud/exp）→ 按载荷重建身份头；
//   - 否则验本团队 S2S 机器票 → 只认调用方 client，剥离用户身份头。
package jwt

import (
	"net/http"
	"strings"

	"beaver/commonconfig"
	"beaver/utils/assertion"
	"beaver/utils/s2s"
	"github.com/zeromicro/go-zero/rest"
)

const (
	bearerPrefix  = "Bearer "
	headerUserID  = "Beaver-User-Id"
	headerAdminID = "Beaver-Admin-Id"
	headerLegacy  = "X-User-Id"
)

// NewJwtS2s 入站闸门。verifier 为 nil 时 panic（北向必须配置 JWKS，fail-closed）。
func NewJwtS2s(s2sSecret, audience string, verifier *assertion.Verifier) rest.Middleware {
	if verifier == nil {
		panic("网关断言 Verifier 未配置：业务面必须挂 Assertion JWKS")
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return inboundGate(s2sSecret, audience, verifier)(next)
	}
}

// MustVerifier 从业务配置装配验签器；失败直接 panic，避免带病启动。
func MustVerifier(cfg commonconfig.AssertionVerify) *assertion.Verifier {
	v, err := assertion.NewVerifier(cfg.Issuer, cfg.JWKS, cfg.JWKSFile)
	if err != nil {
		panic("Assertion JWKS 非法: " + err.Error())
	}
	return v
}

func inboundGate(s2sSecret, aud string, verifier *assertion.Verifier) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(raw, bearerPrefix) {
				http.Error(w, "缺少 Bearer 令牌", http.StatusUnauthorized)
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(raw, bearerPrefix))

			// 1) 北向：网关 ES256 断言（JWKS）
			if claims, err := verifier.Verify(aud, token); err == nil {
				stripIdentityHeaders(r)
				if claims.AdminID != "" {
					r.Header.Set(headerAdminID, claims.AdminID)
				}
				if claims.UserID != "" {
					r.Header.Set(headerUserID, claims.UserID)
				}
				cid := claims.ClientID
				if cid == "" {
					cid = assertion.DefaultClientID
				}
				r.Header.Del(s2s.HeaderClientID)
				r.Header.Set(s2s.HeaderClientID, cid)
				next(w, r.WithContext(s2s.WithPrincipal(r.Context(), cid)))
				return
			}

			// 2) 东西向：本团队 S2S 机器票
			if s2sSecret == "" {
				http.Error(w, "S2S 密钥未配置，拒绝访问", http.StatusUnauthorized)
				return
			}
			clientID, err := s2s.Verify(s2sSecret, aud, token)
			if err != nil {
				http.Error(w, "调用方令牌校验失败: "+err.Error(), http.StatusUnauthorized)
				return
			}
			stripIdentityHeaders(r)
			r.Header.Del(s2s.HeaderClientID)
			r.Header.Set(s2s.HeaderClientID, clientID)
			next(w, r.WithContext(s2s.WithPrincipal(r.Context(), clientID)))
		}
	}
}

func stripIdentityHeaders(r *http.Request) {
	r.Header.Del(headerUserID)
	r.Header.Del(headerAdminID)
	r.Header.Del(headerLegacy)
}
