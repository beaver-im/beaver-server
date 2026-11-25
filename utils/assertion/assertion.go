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

// Package assertion 网关北向身份断言（ES256 + JWKS）。
//
// 对标常见边缘身份传播：签发方持私钥并在 JWT 头带 kid，依赖方用 JWKS 按 kid 取公钥验签，
// 并校验 iss / aud / exp。轮换时 JWKS 可同时挂多把公钥，旧 kid 在过期窗口内仍可验。
//
// 与东西向 S2S（HS256 机器票）分离：断言 = 「网关已代表谁」；机器票 = 「哪个服务在调用」。
package assertion

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v4"
)

const (
	DefaultTTL      = 5 * time.Minute
	DefaultIssuer   = "beaver-gateway"
	DefaultClientID = "gateway"
	HeaderClientID  = "X-Client-Id"
	// clockSkew 允许的时钟偏差（exp / nbf）。
	clockSkew = 30 * time.Second
)

// Claims 网关身份断言载荷。
type Claims struct {
	UserID   string `json:"uid,omitempty"`
	AdminID  string `json:"aid,omitempty"`
	ClientID string `json:"cid"`
	jwtlib.RegisteredClaims
}

// Signer 网关侧签发器：一把当前私钥 + kid。
type Signer struct {
	key      *ecdsa.PrivateKey
	keyID    string
	issuer   string
	clientID string
	ttl      time.Duration
}

// Verifier 业务侧验签器：JWKS 公钥集 + 期望 iss。
type Verifier struct {
	keys   map[string]*ecdsa.PublicKey // kid -> key；空 kid 回退到唯一一把
	issuer string
}

// jwksDoc RFC 7517 JWKS。
type jwksDoc struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// NewSigner 构造签发器。PrivateKeyFile 优先；文件不存在且提供了 PEM 时回退 PEM（本地开发）。
func NewSigner(keyID, issuer, clientID string, ttl time.Duration, privateKeyPEM, privateKeyFile string) (*Signer, error) {
	pemBytes := strings.TrimSpace(privateKeyPEM)
	if path := strings.TrimSpace(privateKeyFile); path != "" {
		b, err := os.ReadFile(path)
		if err == nil {
			pemBytes = string(b)
		} else if pemBytes == "" {
			return nil, fmt.Errorf("读取断言私钥文件: %w", err)
		}
	}
	key, err := ParsePrivateKeyPEM(pemBytes)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(keyID) == "" {
		return nil, errors.New("断言 KeyID 不能为空")
	}
	if issuer == "" {
		issuer = DefaultIssuer
	}
	if clientID == "" {
		clientID = DefaultClientID
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Signer{key: key, keyID: keyID, issuer: issuer, clientID: clientID, ttl: ttl}, nil
}

// JWKS 导出当前签发公钥的 RFC 7517 文档（供 /.well-known/jwks.json 或下游静态配置）。
func (s *Signer) JWKS() ([]byte, error) {
	if s == nil || s.key == nil {
		return nil, errors.New("断言签发器未配置")
	}
	doc := jwksDoc{Keys: []jwkKey{{
		Kty: "EC",
		Crv: "P-256",
		Alg: "ES256",
		Use: "sig",
		Kid: s.keyID,
		X:   base64.RawURLEncoding.EncodeToString(s.key.PublicKey.X.FillBytes(make([]byte, 32))),
		Y:   base64.RawURLEncoding.EncodeToString(s.key.PublicKey.Y.FillBytes(make([]byte, 32))),
	}}}
	return json.Marshal(doc)
}

// Issue 签发带 kid 的 ES256 断言。
func (s *Signer) Issue(aud, userID, adminID string) (string, error) {
	if s == nil || s.key == nil {
		return "", errors.New("断言签发器未配置")
	}
	if strings.TrimSpace(aud) == "" {
		return "", errors.New("断言受众不能为空")
	}
	if userID == "" && adminID == "" {
		return "", errors.New("断言必须携带用户或管理员身份")
	}
	now := time.Now()
	sub := userID
	if sub == "" {
		sub = adminID
	}
	jti := make([]byte, 16)
	_, _ = rand.Read(jti)
	claims := Claims{
		UserID:   userID,
		AdminID:  adminID,
		ClientID: s.clientID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   sub,
			Audience:  jwtlib.ClaimStrings{aud},
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.ttl)),
			NotBefore: jwtlib.NewNumericDate(now.Add(-clockSkew)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ID:        base64.RawURLEncoding.EncodeToString(jti),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodES256, claims)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.key)
}

// NewVerifier 从 JWKS JSON 或 JWKS 文件构造验签器。jwksFile 优先。
func NewVerifier(issuer, jwksJSON, jwksFile string) (*Verifier, error) {
	raw := strings.TrimSpace(jwksJSON)
	if path := strings.TrimSpace(jwksFile); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取 JWKS 文件: %w", err)
		}
		raw = string(b)
	}
	keys, err := parseJWKS(raw)
	if err != nil {
		return nil, err
	}
	if issuer == "" {
		issuer = DefaultIssuer
	}
	return &Verifier{keys: keys, issuer: issuer}, nil
}

// Verify 按 kid 选公钥，校验 alg/iss/aud/exp。
func (v *Verifier) Verify(aud, raw string) (*Claims, error) {
	if v == nil || len(v.keys) == 0 {
		return nil, errors.New("断言验签器未配置")
	}
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("缺少断言令牌")
	}
	parsed, err := jwtlib.ParseWithClaims(raw, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("非预期签名算法: %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
		// 无 kid 且 JWKS 仅一把钥匙时回退；多把钥匙必须带 kid，禁止猜。
		if kid == "" && len(v.keys) == 1 {
			for _, key := range v.keys {
				return key, nil
			}
		}
		return nil, fmt.Errorf("未知的断言 kid: %q", kid)
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("无效的身份断言")
	}
	if !c.VerifyIssuer(v.issuer, true) {
		return nil, errors.New("断言签发方不匹配")
	}
	if aud != "" && !c.VerifyAudience(aud, true) {
		return nil, errors.New("断言受众不匹配")
	}
	if c.UserID == "" && c.AdminID == "" {
		return nil, errors.New("断言缺少身份主体")
	}
	return c, nil
}

func parseJWKS(raw string) (map[string]*ecdsa.PublicKey, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("JWKS 为空")
	}
	var doc jwksDoc
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("JWKS JSON 非法: %w", err)
	}
	if len(doc.Keys) == 0 {
		return nil, errors.New("JWKS 不含任何密钥")
	}
	out := make(map[string]*ecdsa.PublicKey, len(doc.Keys))
	for i, k := range doc.Keys {
		if !strings.EqualFold(k.Kty, "EC") || k.Crv != "P-256" {
			return nil, fmt.Errorf("JWKS[%d] 仅支持 EC P-256", i)
		}
		if k.Alg != "" && k.Alg != "ES256" {
			return nil, fmt.Errorf("JWKS[%d] alg 必须是 ES256", i)
		}
		pub, err := ecPublicKeyXY(k.X, k.Y)
		if err != nil {
			return nil, fmt.Errorf("JWKS[%d]: %w", i, err)
		}
		kid := k.Kid
		if kid == "" {
			kid = fmt.Sprintf("key-%d", i)
		}
		out[kid] = pub
	}
	return out, nil
}

func ecPublicKeyXY(xB64, yB64 string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xB64)
	if err != nil {
		return nil, fmt.Errorf("jwk.x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yB64)
	if err != nil {
		return nil, fmt.Errorf("jwk.y: %w", err)
	}
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	if !elliptic.P256().IsOnCurve(x, y) {
		return nil, errors.New("jwk 点不在 P-256 曲线上")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// ParsePrivateKeyPEM 解析 EC PRIVATE KEY / PKCS8 PEM。
func ParsePrivateKeyPEM(pemBytes string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(pemBytes)))
	if block == nil {
		return nil, errors.New("断言私钥 PEM 无法解析")
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("断言私钥非法: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("断言私钥不是 ECDSA")
	}
	return key, nil
}
