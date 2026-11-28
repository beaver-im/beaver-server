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

	"beaver/app/account/account_api/internal/svc"
	user_models "beaver/app/account/account_models/user"
	relationtypes "beaver/app/relation/relation_rpc/pb"
	"beaver/common/wsEnum/wsCommandConst"
	"beaver/common/wsEnum/wsTypeConst"
	pushcli "beaver/remote/push"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

// NotifyUserProfileChange 用户展示资料变更后推 USER_PROFILE：
// 自己 + 好友 + 共同群成员，body 带 users 表 version，客户端按 version 增量拉取。
// C 端自助改资料、队间 UpdateUsers / UserUpdateDisplay（含 open 工号实体）共用此出口。
func NotifyUserProfileChange(svcCtx *svc.ServiceContext, logger *beaverlog.Logger, userID string, version int64) {
	if svcCtx == nil || userID == "" || version <= 0 {
		return
	}
	if logger == nil {
		logger = beaverlog.New("user_profile_notify", context.Background())
	}

	go func() {
		ctx := context.Background()
		recipients, err := relatedUserIDs(ctx, svcCtx, logger, userID)
		if err != nil {
			logger.Error(model.LogMsg{Text: "获取资料变更推送接收方失败", Data: map[string]interface{}{"userId": userID, "err": err.Error()}})
			return
		}
		logger.Info(model.LogMsg{Text: "推送用户资料变更", Data: map[string]interface{}{
			"userId": userID, "version": version, "count": len(recipients),
		}})
		for _, recipientID := range recipients {
			payload := map[string]interface{}{
				"command":  wsCommandConst.USER_PROFILE,
				"type":     wsTypeConst.UserReceive,
				"senderId": userID,
				"targetId": recipientID,
				"body": map[string]interface{}{
					"table":    "users",
					"version":  int32(version),
					"targetId": userID,
				},
				"conversationId": "",
			}
			pushcli.SendRealtimeApi(context.Background(), "sync", payload)
		}
	}()
}

func relatedUserIDs(ctx context.Context, svcCtx *svc.ServiceContext, logger *beaverlog.Logger, userID string) ([]string, error) {
	userMap := map[string]bool{userID: true}

	friendResp, err := svcCtx.FriendRpc.GetFriendIds(ctx, &relationtypes.GetFriendIdsRequest{UserID: userID})
	if err != nil {
		logger.Error(model.LogMsg{Text: "获取好友列表失败", Data: map[string]interface{}{"userId": userID, "err": err.Error()}})
		return nil, err
	}
	for _, uid := range friendResp.FriendIds {
		userMap[uid] = true
	}

	groupResp, err := svcCtx.GroupRpc.GetUserGroupMembers(ctx, &relationtypes.GetUserGroupMembersReq{UserID: userID})
	if err != nil {
		logger.Error(model.LogMsg{Text: "获取群成员列表失败", Data: map[string]interface{}{"userId": userID, "err": err.Error()}})
		return nil, err
	}
	for _, uid := range groupResp.MemberIDs {
		userMap[uid] = true
	}

	result := make([]string, 0, len(userMap))
	for uid := range userMap {
		result = append(result, uid)
	}
	return result, nil
}

func loadUserVersion(svcCtx *svc.ServiceContext, userID string) int64 {
	var u user_models.UserModel
	if err := svcCtx.DB.Select("version").Where("user_id = ?", userID).Take(&u).Error; err != nil {
		return 0
	}
	return u.Version
}
