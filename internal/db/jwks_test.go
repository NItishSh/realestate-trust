package db

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestJWKSClient_GetKey_StrictKid(t *testing.T) {
	client := NewJWKSClient("http://localhost:9999/certs")

	// Generate a dummy RSA public key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	pubKey := &privateKey.PublicKey

	client.keys["valid-kid-123"] = pubKey
	client.lastFetch = time.Now()

	// 1. Token without kid header should fail
	tokenNoKid := &jwt.Token{
		Method: jwt.SigningMethodRS256,
		Header: map[string]interface{}{
			"alg": "RS256",
		},
	}
	_, err = client.GetKey(tokenNoKid)
	if err == nil {
		t.Fatalf("expected error for token with missing kid header, got nil")
	}

	// 2. Token with valid kid header should succeed
	tokenValidKid := &jwt.Token{
		Method: jwt.SigningMethodRS256,
		Header: map[string]interface{}{
			"alg": "RS256",
			"kid": "valid-kid-123",
		},
	}
	resolvedKey, err := client.GetKey(tokenValidKid)
	if err != nil {
		t.Fatalf("expected key to resolve successfully, got: %v", err)
	}
	if resolvedKey != pubKey {
		t.Fatalf("resolved key does not match expected key")
	}

	// 3. Token with unknown kid should fail (and NOT fall back to arbitrary keys)
	tokenUnknownKid := &jwt.Token{
		Method: jwt.SigningMethodRS256,
		Header: map[string]interface{}{
			"alg": "RS256",
			"kid": "unknown-kid-999",
		},
	}
	_, err = client.GetKey(tokenUnknownKid)
	if err == nil {
		t.Fatalf("expected error for unknown kid without arbitrary fallback, got nil")
	}
}

func TestJWKSClient_GetKey_HMACFallback(t *testing.T) {
	client := NewJWKSClient("http://localhost:9999/certs")

	tokenHMAC := &jwt.Token{
		Method: jwt.SigningMethodHS256,
		Header: map[string]interface{}{
			"alg": "HS256",
		},
	}

	key, err := client.GetKey(tokenHMAC)
	if err != nil {
		t.Fatalf("expected HMAC key to resolve, got: %v", err)
	}
	keyBytes, ok := key.([]byte)
	if !ok || string(keyBytes) != string(JWTSecret) {
		t.Fatalf("expected JWTSecret, got %v", key)
	}
}
