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

package types

import "github.com/zeromicro/go-zero/core/logx"

// Config 网关负责：验外部票 → 写可信身份头 → 签网关身份断言 → 转发。
// 业务用 JWKS 验网关公钥（北向）或本团队 S2S.Secret（东西向）；网关不持有任何团队签发密钥。
type Config struct {
	Name string
	Addr string `json:",optional"`
	Etcd string
	// ServiceAlias URL 前缀 → 实际服务名（服务合并后按旧前缀路由到合并服务）。
	ServiceAlias map[string]string `json:",optional"`
	Log          logx.LogConf
	Limit        LimitConfig

	Auth           AuthConfigs     `json:",optional"`
	Assertion      AssertionConfig `json:",optional"`
	PublicPatterns []string        `json:",optional"`
	AuthRules      []AuthRule      `json:",optional"`
}

// AuthRule 一条鉴权平面规则：路径正则 → 外部票类型（beaver/open/admin/public）。
type AuthRule struct {
	Pattern string `json:",optional"`
	Ticket  string `json:",optional"`
}

// AuthConfigs 三类外部票的验签密钥集合，按平面各占一段、三把独立。
type AuthConfigs struct {
	AuthBeaver AuthConfig `json:",optional"`
	AuthOpen   AuthConfig `json:",optional"`
	AuthAdmin  AuthConfig `json:",optional"`
}

// AuthConfig 单个平面的外部票验签密钥。
type AuthConfig struct {
	AccessSecret string `json:",optional"`
}

// AssertionConfig 网关北向身份断言（ES256）签发配置。
// 私钥只在网关（PrivateKeyFile 优先，其次 PrivateKeyPEM）；JWT 头带 KeyID，下游按 JWKS 选公钥。
// Audiences 按路径选 aud，与各团队 S2S.Secret 无关。
type AssertionConfig struct {
	KeyID           string            `json:",optional"`
	PrivateKeyFile  string            `json:",optional"`
	PrivateKeyPEM   string            `json:",optional"`
	Issuer          string            `json:",optional"`
	ClientID        string            `json:",optional"`
	TTLSeconds      int               `json:",optional"`
	DefaultAudience string            `json:",optional"`
	Audiences       map[string]string `json:",optional"`
}

type LimitConfig struct {
	Enable bool    `json:",default=false"`
	Rate   float64 `json:",default=100"`
	Burst  int     `json:",default=200"`
}
