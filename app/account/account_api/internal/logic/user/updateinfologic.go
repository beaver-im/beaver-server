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
	"errors"
	"fmt"
	"time"

	"beaver/app/account/account_api/internal/svc"
	"beaver/app/account/account_api/internal/types"
	user_models "beaver/app/account/account_models/user"
	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"

	"gorm.io/gorm"
)

type UpdateInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger *beaverlog.Logger
}

func NewUpdateInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateInfoLogic {
	return &UpdateInfoLogic{
		ctx:    ctx,
		logger: beaverlog.New("update_info", ctx),
		svcCtx: svcCtx,
	}
}

func (l *UpdateInfoLogic) UpdateInfo(req *types.UpdateInfoReq) (resp *types.UpdateInfoRes, err error) {
	// 获取要更新的用户信息
	var user user_models.UserModel
	if err := l.svcCtx.DB.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 准备更新的字段
	updateFields := make(map[string]interface{})
	if req.NickName != nil {
		updateFields["nick_name"] = *req.NickName
	}
	if req.Avatar != nil {
		updateFields["avatar"] = *req.Avatar
	}
	if req.Abstract != nil {
		updateFields["abstract"] = *req.Abstract
	}
	if req.Gender != nil {
		updateFields["gender"] = *req.Gender
	}

	var version int64 // 定义version变量在更外层作用域

	// 执行更新操作
	if len(updateFields) > 0 {
		// 获取新版本号（用户独立递增）
		version = l.svcCtx.VersionGen.GetNextVersion("users", "user_id", req.UserID)
		if version == -1 {
			l.logger.Error(model.LogMsg{Text: "获取版本号失败", Data: map[string]interface{}{"userId": req.UserID}})
			return nil, errors.New("获取版本号失败")
		}

		// 添加版本号到更新字段
		updateFields["version"] = version

		err = l.svcCtx.DB.Model(&user).Updates(updateFields).Error
		if err != nil {
			return nil, err
		}

		l.logger.Info(model.LogMsg{
			Text: "用户信息更新成功",
			Data: map[string]interface{}{
				"userId":  req.UserID,
				"version": version,
			},
		})

		// 记录用户变更日志
		l.recordUserChangeLog(req.UserID, version, updateFields)
	}

	if version > 0 {
		NotifyUserProfileChange(l.svcCtx, l.logger, req.UserID, version)
	}

	return &types.UpdateInfoRes{}, nil
}

/**
 * 记录用户变更日志
 */
func (l *UpdateInfoLogic) recordUserChangeLog(userID string, version int64, updateFields map[string]interface{}) {
	var changeLogs []user_models.UserChangeLogModel

	// 为每个变更的字段创建日志记录
	for field, newValue := range updateFields {
		if field == "version" {
			continue // 跳过版本字段
		}

		var changeType string
		switch field {
		case "nick_name":
			changeType = "nickName"
		case "avatar":
			changeType = "avatar"
		case "abstract":
			changeType = "abstract"
		case "gender":
			changeType = "gender"
		case "status":
			changeType = "status"
		default:
			changeType = field
		}

		changeLog := user_models.UserChangeLogModel{
			UserID:     userID,
			ChangeType: changeType,
			NewValue:   fmt.Sprintf("%v", newValue),
			ChangeTime: time.Now().Unix(),
			Version:    version,
		}

		changeLogs = append(changeLogs, changeLog)
	}

	// 批量插入变更日志
	if len(changeLogs) > 0 {
		if err := l.svcCtx.DB.Create(&changeLogs).Error; err != nil {
			l.logger.Error(model.LogMsg{Text: "记录用户变更日志失败", Data: map[string]interface{}{"userId": userID, "err": err.Error()}})
		} else {
			l.logger.Info(model.LogMsg{Text: "用户变更日志记录成功", Data: map[string]interface{}{"userId": userID, "count": len(changeLogs)}})
		}
	}
}
