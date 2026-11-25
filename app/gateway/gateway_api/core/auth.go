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

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"beaver/app/gateway/gateway_api/types"
	"beaver/utils/assertion"
	"beaver/utils/jwts"
	"beaver/utils/s2s"
)

// 全仓统一身份头名。验票前一律剥离客户端自带值，验票后由网关按结论重建。
// user / open 平面写 Beaver-User-Id，admin 平面写 Beaver-Admin-Id——头名按平面分，
// 下游 .api 的字段名按域分（UserID / AdminID），两者互不牵连。
const (
	HeaderUserID      = "Beaver-User-Id"
	HeaderAdminID     = "Beaver-Admin-Id"
	HeaderPartnerID   = "X-Partner-Id"
	HeaderServiceName = "X-Service-Name"
	HeaderLegacyUser  = "X-User-Id"

	bearerPrefix = "Bearer "
)

// 客户端可伪造、网关必须先剥后写的身份头。
var spoofableHeaders = []string{
	HeaderUserID, HeaderAdminID, HeaderLegacyUser, HeaderPartnerID, HeaderServiceName,
}

// ticketKind 外部票平面。三类票密钥各自独立成段（见 jwtBeaver/jwtOpen/jwtAdmin.go），
// 受众也各自独立（beaver:user / beaver:developer / beaver:admin），互不通用。
type ticketKind int

const (
	ticketPublic ticketKind = iota
	ticketBeaver
	ticketOpen
	ticketAdmin
)

// stripIdentityHeaders 剥离所有客户端自带身份头（强制处理顺序第 1 步）。
func stripIdentityHeaders(r *http.Request) {
	for _, h := range spoofableHeaders {
		r.Header.Del(h)
	}
}

// compiledAuthRule 预编译后的单条鉴权规则。
type compiledAuthRule struct {
	re   *regexp.Regexp
	kind ticketKind
}

// authMatcher 预编译的鉴权规则集，随 Proxy 生命周期复用，避免每请求重新编译正则。
type authMatcher struct {
	public []*regexp.Regexp
	groups []compiledAuthRule
}

// compileAuthMatcher 启动时一次性编译正则；任一非法即返回错误，由调用方 fail-fast。
// 规则为空是合法降级：全部请求落到默认 beaver 票（与旧配置未命中行为一致）。
func compileAuthMatcher(cfg types.Config) (authMatcher, error) {
	var m authMatcher
	for _, pattern := range cfg.PublicPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return authMatcher{}, fmt.Errorf("PublicPatterns 正则 %q 非法: %w", pattern, err)
		}
		m.public = append(m.public, re)
	}
	for _, rule := range cfg.AuthRules {
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return authMatcher{}, fmt.Errorf("AuthRules 正则 %q 非法: %w", rule.Pattern, err)
		}
		m.groups = append(m.groups, compiledAuthRule{re: re, kind: parseTicketKind(rule.Ticket)})
	}
	return m, nil
}

// parseTicketKind 把配置里的票面名映射到内部枚举；未知值按最严格的 beaver 处理。
func parseTicketKind(name string) ticketKind {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "open":
		return ticketOpen
	case "admin":
		return ticketAdmin
	case "public":
		return ticketPublic
	default:
		return ticketBeaver
	}
}

// routeAuthKind 按路径正则判定要验的外部票类型。
// 优先级：PublicPatterns 任一命中 → public；否则 AuthRules 按声明顺序首个命中 → 对应票；
// 全未命中 → 默认 beaver（主站用户票）。公开必须优先，避免 team 级规则吞掉公开端点。
func (p Proxy) routeAuthKind(path string) ticketKind {
	for _, re := range p.auth.public {
		if re.MatchString(path) {
			return ticketPublic
		}
	}
	for _, rule := range p.auth.groups {
		if rule.re.MatchString(path) {
			return rule.kind
		}
	}
	return ticketBeaver
}

// authenticate 强制处理顺序第 1~3 步：剥头 → 验外部票 → 写可信身份头。
//
// 东西向服务间调用不经网关（直连内网 + S2S）；网关只服务客户端北向流量。
// admin 平面写 Beaver-Admin-Id，其余平面写 Beaver-User-Id。
// public 路由不验票，但同样先剥头，杜绝客户端自带身份头透传到下游。
func (p Proxy) authenticate(r *http.Request) (ticketKind, error) {
	stripIdentityHeaders(r)
	preferred := p.routeAuthKind(r.URL.Path)
	if preferred == ticketPublic {
		return ticketPublic, nil
	}
	kind, claims, err := p.verifyExternalTicket(r, preferred)
	if err != nil {
		return preferred, err
	}
	r.Header.Set(identityHeader(kind), claims.UserID)
	return kind, nil
}

// verifyExternalTicket 验外部票并返回票的真实平面。
//
// 三把外部票的密钥与受众两两独立（beaver:user / beaver:developer / beaver:admin），
// 一张票有且仅有一个平面能验过。AuthRules 给出的首选平面先验，常规请求一次命中。
//
// 唯一的兼容是单向的：开放平台面额外接受主站用户票——IM 客户端自用的 oauth 面挂在
// /beaver-open/ 下却必须用 C 端用户票（UserID 来自 Beaver-User-Id，空即报未登录）。
func (p Proxy) verifyExternalTicket(r *http.Request, preferred ticketKind) (ticketKind, *jwts.CustomClaims, error) {
	claims, err := p.verifyExternalByKind(r, preferred)
	if err == nil {
		return preferred, claims, nil
	}
	if preferred == ticketOpen {
		if userClaims, userErr := p.verifyExternalByKind(r, ticketBeaver); userErr == nil {
			return ticketBeaver, userClaims, nil
		}
	}
	return preferred, nil, err
}

// verifyExternalByKind 按平面分派到 jwtBeaver / jwtOpen / jwtAdmin 各自的验签实现。
func (p Proxy) verifyExternalByKind(r *http.Request, kind ticketKind) (*jwts.CustomClaims, error) {
	switch kind {
	case ticketOpen:
		return p.verifyOpenTicket(r)
	case ticketAdmin:
		return p.verifyAdminTicket(r)
	default:
		return p.verifyBeaverTicket(r)
	}
}

// verifyTicket 三个平面共用的验签口径：取票 → 验签 → 校验非空与受众。
func verifyTicket(r *http.Request, secret, audience, missingSecretMsg string) (*jwts.CustomClaims, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New(missingSecretMsg)
	}
	token := ticketFromRequest(r)
	if token == "" {
		return nil, errors.New("缺少认证令牌")
	}
	claims, err := jwts.ParseToken(token, secret)
	if err != nil || claims.UserID == "" {
		return nil, errors.New("票据无效或已过期")
	}
	if !claims.VerifyAudience(audience, true) {
		return nil, errors.New("票据与访问平面不匹配")
	}
	return claims, nil
}

// identityHeader 按票面类型选择下游可信身份头名。
func identityHeader(kind ticketKind) string {
	if kind == ticketAdmin {
		return HeaderAdminID
	}
	return HeaderUserID
}

// injectGatewayAssertion 强制处理顺序第 5 步：用网关 ES256 私钥签发带 kid 的身份断言，
// 覆盖 Authorization。下游按 JWKS 选公钥验签，并按断言载荷重建身份头。
// 签发器未装配时返回错误（fail-closed，禁止静默降级）。
func (p Proxy) injectGatewayAssertion(r *http.Request, kind ticketKind) error {
	if kind == ticketPublic {
		return nil
	}
	if p.signer == nil {
		return errors.New("网关断言签发器未配置")
	}
	var userID, adminID string
	if kind == ticketAdmin {
		adminID = r.Header.Get(HeaderAdminID)
	} else {
		userID = r.Header.Get(HeaderUserID)
	}
	token, err := p.signer.Issue(p.assertionAudience(r.URL.Path), userID, adminID)
	if err != nil {
		return err
	}
	clientID := p.Config.Assertion.ClientID
	if clientID == "" {
		clientID = assertion.DefaultClientID
	}
	r.Header.Set("Authorization", bearerPrefix+token)
	r.Header.Del(s2s.HeaderClientID)
	r.Header.Set(s2s.HeaderClientID, clientID)
	return nil
}

// assertionAudience 按 team/surface 选断言受众：
// /beaver-admin/api/... -> admin:internal，/beaver-open/api/... -> open:internal，
// 主站 /beaver/api/... -> core:api。
func (p Proxy) assertionAudience(path string) string {
	a := p.Config.Assertion
	aud := a.DefaultAudience
	if aud == "" {
		aud = s2s.AudienceCore
	}
	team, surface := teamAndSurface(path)
	if configured, ok := a.Audiences[team+"/"+surface]; ok {
		return configured
	}
	if configured, ok := a.Audiences[team]; ok {
		return configured
	}
	return aud
}

// teamAndSurface 取六段路径的前两段（team/surface），用于选择断言受众。
func teamAndSurface(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// ticketFromRequest 三种携带口径：Authorization: Bearer / Token 头 / query token。
func ticketFromRequest(r *http.Request) string {
	if raw := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(raw, bearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(raw, bearerPrefix))
	}
	if t := strings.TrimSpace(r.Header.Get("Token")); t != "" {
		return t
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

// writeAuthError 鉴权失败统一 JSON 体（与全局 BaseResponse 一致）。
func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	body, _ := json.Marshal(BaseResponse{Code: 1, Msg: msg})
	_, _ = w.Write(body)
}
