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

package assertion

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"
)

func testSignerVerifier(t *testing.T) (*Signer, *Verifier, *ecdsa.PrivateKey) {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalECPrivateKey(k)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER}))
	jwks, _ := json.Marshal(jwksDoc{Keys: []jwkKey{{
		Kty: "EC", Crv: "P-256", Alg: "ES256", Use: "sig", Kid: "test-kid",
		X: base64.RawURLEncoding.EncodeToString(k.PublicKey.X.FillBytes(make([]byte, 32))),
		Y: base64.RawURLEncoding.EncodeToString(k.PublicKey.Y.FillBytes(make([]byte, 32))),
	}}})
	signer, err := NewSigner("test-kid", DefaultIssuer, DefaultClientID, time.Minute, privPEM, "")
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewVerifier(DefaultIssuer, string(jwks), "")
	if err != nil {
		t.Fatal(err)
	}
	return signer, verifier, k
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	signer, verifier, _ := testSignerVerifier(t)
	raw, err := signer.Issue("core:api", "u100", "")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := verifier.Verify("core:api", raw)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "u100" || claims.ClientID != DefaultClientID {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyRejectsWrongAudience(t *testing.T) {
	signer, verifier, _ := testSignerVerifier(t)
	raw, err := signer.Issue("core:api", "u100", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.Verify("social:internal", raw); err == nil {
		t.Fatal("expected audience mismatch")
	}
}

func TestVerifyRejectsWrongIssuer(t *testing.T) {
	signer, verifier, _ := testSignerVerifier(t)
	raw, err := signer.Issue("core:api", "u100", "")
	if err != nil {
		t.Fatal(err)
	}
	bad := &Verifier{keys: verifier.keys, issuer: "other-issuer"}
	if _, err := bad.Verify("core:api", raw); err == nil {
		t.Fatal("expected issuer mismatch")
	}
}

func TestVerifyRejectsUnknownKid(t *testing.T) {
	signer, verifier, _ := testSignerVerifier(t)
	raw, err := signer.Issue("core:api", "u100", "")
	if err != nil {
		t.Fatal(err)
	}
	verifier.keys = map[string]*ecdsa.PublicKey{"other": verifier.keys["test-kid"]}
	if _, err := verifier.Verify("core:api", raw); err == nil {
		t.Fatal("expected unknown kid")
	}
}

func TestJWKSPointMustOnCurve(t *testing.T) {
	if _, err := parseJWKS(`{"keys":[{"kty":"EC","crv":"P-256","kid":"x","x":"AA","y":"AA"}]}`); err == nil {
		t.Fatal("expected curve check failure")
	}
}
