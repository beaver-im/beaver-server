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

package svc

import (
	"beaver/app/push/push_api/internal/config"
	"beaver/app/push/push_api/internal/middleware"
	"beaver/core/corerocketmq"

	"github.com/zeromicro/go-zero/rest"
)

// ServiceContext 只装配鉴权与 MQ 出口，不持有数据表、缓存或领域层对象。
type ServiceContext struct {
	Config   config.Config
	RocketMQ *corerocketmq.Client
	// JwtS2s 由 .api 的 @server middleware 声明接线，goctl 生成 WithMiddlewares 注入。
	JwtS2s rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		RocketMQ: corerocketmq.InitRocketMQ(c.RocketMQ.Addr),
		JwtS2s:   middleware.NewJwtS2sMiddleware(c).Handle,
	}
}
