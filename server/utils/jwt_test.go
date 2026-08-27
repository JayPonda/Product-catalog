package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateSecureToken(t *testing.T) {
	tok, err := GenerateSecureToken(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tok) != 32 { // 16 bytes -> 32 hex chars
		t.Errorf("expected 32 hex chars, got %d", len(tok))
	}

	tok2, _ := GenerateSecureToken(16)
	if tok == tok2 {
		t.Error("expected two tokens to differ")
	}
}

func TestHashToken(t *testing.T) {
	want := sha256.Sum256([]byte("abc"))
	wantHex := hex.EncodeToString(want[:])
	if got := HashToken("abc"); got != wantHex {
		t.Errorf("HashToken = %q, want %q", got, wantHex)
	}
	h1 := HashToken("abc")
	h2 := HashToken("abc")
	if h1 != h2 {
		t.Error("HashToken should be stable")
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := "test-secret"
	tok, err := GenerateAccessToken("user-123", secret, time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	claims, err := ParseAccessToken(tok, secret)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Before(time.Now()) {
		t.Error("expiry should be in the future")
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	tok, _ := GenerateAccessToken("u", "secret1", time.Minute)
	if _, err := ParseAccessToken(tok, "secret2"); err == nil {
		t.Error("expected error with wrong secret")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	tok, _ := GenerateAccessToken("u", "secret", -time.Minute)
	if _, err := ParseAccessToken(tok, "secret"); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == "" || hash == "" {
		t.Error("expected non-empty tokens")
	}
	if hash != HashToken(raw) {
		t.Error("hash does not match raw token")
	}
}

func TestParseAccessToken_UnexpectedSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, TokenClaims{UserID: "user-123"})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to create unsigned token: %v", err)
	}
	_, err = ParseAccessToken(tokenString, "secret")
	if err == nil {
		t.Error("expected error for unexpected signing method, got nil")
	}
}

func TestParseAccessToken_GarbageString(t *testing.T) {
	_, err := ParseAccessToken("not-a-jwt-token", "secret")
	if err == nil {
		t.Error("expected error for garbage token string")
	}
}

func TestParseAccessToken_EmptyString(t *testing.T) {
	_, err := ParseAccessToken("", "secret")
	if err == nil {
		t.Error("expected error for empty token string")
	}
}

func TestGenerateAccessToken_DifferentSecrets(t *testing.T) {
	tok1, _ := GenerateAccessToken("u1", "secret-a", time.Minute)
	tok2, _ := GenerateAccessToken("u1", "secret-b", time.Minute)
	if tok1 == tok2 {
		t.Error("tokens signed with different secrets should differ")
	}
}

func TestParseAccessToken_ValidTokenWithExtraClaims(t *testing.T) {
	tok, _ := GenerateAccessToken("user-456", "mysecret", time.Hour)
	claims, err := ParseAccessToken(tok, "mysecret")
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-456" {
		t.Errorf("Subject = %q, want user-456", claims.Subject)
	}
}

func TestGenerateRefreshToken_DifferentEachTime(t *testing.T) {
	_, hash1, _ := GenerateRefreshToken()
	_, hash2, _ := GenerateRefreshToken()
	if hash1 == hash2 {
		t.Error("two refresh tokens should produce different hashes")
	}
}
