package db

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// JWK represents a JSON Web Key from Keycloak
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSClient manages fetching and caching RSA public keys from Keycloak's JWKS endpoint
type JWKSClient struct {
	jwksURL    string
	httpClient *http.Client
	keys       map[string]*rsa.PublicKey
	mu         sync.RWMutex
	lastFetch  time.Time
	ttl        time.Duration
}

// NewJWKSClient creates a new JWKS client
func NewJWKSClient(jwksURL string) *JWKSClient {
	return &JWKSClient{
		jwksURL: jwksURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		keys: make(map[string]*rsa.PublicKey),
		ttl:  10 * time.Minute,
	}
}

// GetKey resolves the verification key for the given JWT token
func (c *JWKSClient) GetKey(token *jwt.Token) (interface{}, error) {
	// Fallback to local symmetric secret for HMAC (HS256) in tests / legacy callers
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		return JWTSecret, nil
	}

	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, errors.New("missing kid header in JWT token")
	}

	c.mu.RLock()
	key, exists := c.keys[kid]
	fresh := time.Since(c.lastFetch) < c.ttl
	c.mu.RUnlock()

	if exists && fresh {
		return key, nil
	}

	// Fetch or refresh keys
	if err := c.refresh(); err != nil {
		if exists {
			return key, nil // Fall back to stale cached key if fetch fails
		}
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	key, exists = c.keys[kid]
	if !exists {
		return nil, fmt.Errorf("unable to find public key for kid: %s", kid)
	}

	return key, nil
}

func (c *JWKSClient) refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.jwksURL == "" {
		return errors.New("JWKS URL is not configured")
	}

	resp, err := c.httpClient.Get(c.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", c.jwksURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status code %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS JSON: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}
		pubKey, err := parseRSAPublicKey(jwk.N, jwk.E)
		if err != nil {
			slog.Warn("Failed to parse RSA public key from JWKS", "kid", jwk.Kid, "err", err)
			continue
		}
		newKeys[jwk.Kid] = pubKey
	}

	c.keys = newKeys
	c.lastFetch = time.Now()
	return nil
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64url modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64url exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = (e << 8) | int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

var (
	defaultJWKSClient *JWKSClient
	jwksOnce          sync.Once
)

// GetJWTKeyFunc returns a jwt.Keyfunc that dynamically resolves public keys from JWKS or fallback HMAC
func GetJWTKeyFunc() jwt.Keyfunc {
	jwksOnce.Do(func() {
		jwksURL := os.Getenv("JWKS_URL")
		if jwksURL == "" {
			keycloakURL := os.Getenv("KEYCLOAK_URL")
			if keycloakURL == "" {
				keycloakURL = "http://keycloak.realestate-trust.svc.cluster.local:8080"
			}
			jwksURL = fmt.Sprintf("%s/realms/realestate-trust/protocol/openid-connect/certs", keycloakURL)
		}
		defaultJWKSClient = NewJWKSClient(jwksURL)
	})

	return defaultJWKSClient.GetKey
}
