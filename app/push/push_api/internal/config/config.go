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

package config

import (
	"beaver/commonconfig"
	"beaver/utils/s2s"

	"github.com/zeromicro/go-zero/rest"
)

// Config 是推送网关的配置：鉴权 + 一个 MQ 出口，不持有任何数据表。
// 实时帧与离线任务都按 channel 投递到各自 Topic，消费端在 ws_api / open_api。
type Config struct {
	rest.RestConf
	Etcd     string
	RocketMQ struct {
		Addr string
	}
	// Assertion 北向：JWKS 验网关 ES256 断言（iss + kid）；JwtS2s 强制要求，缺失 fail-closed。
	Assertion commonconfig.AssertionVerify `json:",optional"`
	// S2S：Secret 签发/校验机器票；Clients 是 client_credentials 换票端点的注册表。
	// 换票端点在本服务 POST /beaver/api/push/auth_public/v1/issue；受众 core:api。
	S2S struct {
		Secret  string                     `json:",optional"`
		Clients map[string]s2s.ClientEntry `json:",optional"`
	}
}
