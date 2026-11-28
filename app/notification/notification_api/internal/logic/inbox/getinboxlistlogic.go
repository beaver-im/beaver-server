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
	"context"
	"time"

	"beaver/app/notification/notification_api/internal/svc"
	"beaver/app/notification/notification_api/internal/types"
	notification_models "beaver/app/notification/notification_models/notification"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

type GetInboxListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

// 按用户分页拉取最近通知：服务端为权威数据源，一次 join 出收件箱 + 事件明细。
// 客户端点开通知入口时直接调用，不再依赖本地库是否被 WS 增量同步完整填充。
func NewGetInboxListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInboxListLogic {
	return &GetInboxListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		logger: beaverlog.New("get_inbox_list", ctx),
	}
}

// 收件箱与事件 join 后的行。两表都有 event_type/category/status/version/created_at，
// 这里统一取收件箱侧（用户视角的状态），事件侧只取展示需要的 payload 和来源。
type inboxListRow struct {
	EventID   string `gorm:"column:event_id"`
	EventType string `gorm:"column:event_type"`
	Category  string `gorm:"column:category"`
	Version   int64  `gorm:"column:version"`
	IsRead    bool   `gorm:"column:is_read"`
	ReadAt    *time.Time `gorm:"column:read_at"`
	Status    int8   `gorm:"column:status"`
	IsDeleted bool   `gorm:"column:is_deleted"`
	Silent    bool   `gorm:"column:silent"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Payload   string `gorm:"column:payload"`
	FromUserID string `gorm:"column:from_user_id"`
	TargetID   string `gorm:"column:target_id"`
	TargetType string `gorm:"column:target_type"`
}

func (l *GetInboxListLogic) GetInboxList(req *types.GetInboxListReq) (resp *types.GetInboxListRes, err error) {
	resp = &types.GetInboxListRes{Items: []types.NotificationListItem{}}

	if req.UserID == "" {
		return resp, nil
	}

	cats := req.Categories
	if len(cats) == 0 {
		cats = []string{
			notification_models.CategorySocial,
			notification_models.CategoryGroup,
			notification_models.CategoryMoment,
		}
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// 多取一条判断 hasMore，不额外 count
	var rows []inboxListRow
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("notification_inboxes AS i").
		Select(`i.event_id, i.event_type, i.category, i.version, i.is_read, i.read_at,
			i.status, i.is_deleted, i.silent, i.created_at, i.updated_at,
			e.payload, e.from_user_id, e.target_id, e.target_type`).
		Joins("JOIN notification_events AS e ON e.event_id = i.event_id").
		Where("i.user_id = ? AND i.category IN ? AND i.is_deleted = ? AND i.status = ?", req.UserID, cats, false, 1).
		Order("i.created_at DESC").
		Limit(limit + 1).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		l.logger.Error(model.LogMsg{
			Text: "拉取最近通知列表失败",
			Data: map[string]any{"userId": req.UserID, "err": err.Error()},
		})
		return nil, err
	}

	resp.HasMore = len(rows) > limit
	if resp.HasMore {
		rows = rows[:limit]
	}

	for _, row := range rows {
		var readAt int64
		if row.ReadAt != nil {
			readAt = row.ReadAt.UnixMilli()
		}
		resp.Items = append(resp.Items, types.NotificationListItem{
			EventID:    row.EventID,
			EventType:  row.EventType,
			Category:   row.Category,
			Version:    row.Version,
			IsRead:     row.IsRead,
			ReadAt:     readAt,
			Status:     int32(row.Status),
			IsDeleted:  row.IsDeleted,
			Silent:     row.Silent,
			Payload:    row.Payload,
			FromUserID: row.FromUserID,
			TargetID:   row.TargetID,
			TargetType: row.TargetType,
			CreatedAt:  row.CreatedAt.UnixMilli(),
			UpdatedAt:  row.UpdatedAt.UnixMilli(),
		})
	}

	return resp, nil
}
