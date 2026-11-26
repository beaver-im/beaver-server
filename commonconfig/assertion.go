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

package commonconfig

// AssertionVerify 业务侧验网关北向断言的配置（JWKS，不是共享 HMAC）。
// Issuer 必须与网关 Assertion.Issuer 一致；JWKS / JWKSFile 二选一（File 优先）。
// 轮换：JWKS 可同时包含新旧 kid，旧票在 TTL 内仍可验。
type AssertionVerify struct {
	Issuer   string `json:",optional"`
	JWKS     string `json:",optional"`
	JWKSFile string `json:",optional"`
}
