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

package coreopensearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	beaverlog "beaver/utils/beaverlog"
	"beaver/utils/beaverlog/model"
)

// Client 极简 OpenSearch HTTP 写入（扁平 JSON 文档，无 attributes 包装）。
type Client struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

// New 创建客户端。addr 例：https://127.0.0.1:9200；空则返回 nil。
// username/password 为空则不带 Basic Auth（仅用于无安全插件的环境）。
// https 地址默认跳过证书校验（官方镜像自签证书）。
func New(addr, username, password string) *Client {
	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	if addr == "" {
		return nil
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if strings.HasPrefix(strings.ToLower(addr), "https://") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // 本地/演示自签证书
	}
	beaverlog.New("opensearch").Info(model.LogMsg{
		Text: "OpenSearch客户端就绪",
		Data: map[string]any{"addr": addr, "auth": username != ""},
	})
	return &Client{
		baseURL:  addr,
		username: username,
		password: password,
		http:     &http.Client{Timeout: 5 * time.Second, Transport: transport},
	}
}

// IndexRaw 将扁平 JSON 写入指定索引。
// 若缺少 @timestamp，则用文档内 timestamp（毫秒）或当前时间补上，供 Dashboards Discover 时间轴使用。
func (c *Client) IndexRaw(ctx context.Context, index string, body []byte) error {
	if c == nil || c.http == nil {
		return fmt.Errorf("OpenSearch 未初始化")
	}
	index = strings.TrimSpace(index)
	if index == "" {
		return fmt.Errorf("OpenSearch index 为空")
	}
	if len(body) == 0 {
		return nil
	}

	body = ensureAtTimestamp(body)

	url := c.baseURL + "/" + index + "/_doc"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch index failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}

func ensureAtTimestamp(body []byte) []byte {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil || m == nil {
		return body
	}
	if _, ok := m["@timestamp"]; ok {
		return body
	}
	ts := time.Now().UTC()
	switch v := m["timestamp"].(type) {
	case float64:
		if v > 0 {
			ts = time.UnixMilli(int64(v)).UTC()
		}
	case int64:
		if v > 0 {
			ts = time.UnixMilli(v).UTC()
		}
	case json.Number:
		if n, err := v.Int64(); err == nil && n > 0 {
			ts = time.UnixMilli(n).UTC()
		}
	}
	m["@timestamp"] = ts.Format(time.RFC3339Nano)
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return out
}
