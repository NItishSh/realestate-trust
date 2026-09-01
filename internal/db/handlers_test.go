package db

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v5"
	"github.com/realestate-trust/monorepo/internal/core"
)

func TestUserHandlersIntegration(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryUserRepository()
	handler := NewUserHandler(repo)

	// Test 1: User Registration
	reqBody, _ := json.Marshal(core.RegisterUserRequest{
		Email:    "test@example.com",
		FullName: "Test User",
		Password: "secretpassword",
		Role:     core.Buyer,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	err := handler.RegisterUser(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}

	var user User
	if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected user email test@example.com; got %s", user.Email)
	}

	// Test 2: KYC Submission
	kycBody, _ := json.Marshal(core.KYCSubmissionRequest{
		DocumentType:      "PASSPORT",
		DocumentReference: "P987654321",
	})

	reqKYC := httptest.NewRequest(http.MethodPost, "/api/v1/users/usr-test@example.com/kyc", bytes.NewBuffer(kycBody))
	reqKYC.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	wKYC := httptest.NewRecorder()
	cKYC := e.NewContext(reqKYC, wKYC)
	cKYC.SetPathValues(echo.PathValues{echo.PathValue{Name: "id", Value: "usr-test@example.com"}})

	errKYC := handler.SubmitKYC(cKYC)
	if errKYC != nil {
		t.Fatalf("handler returned error: %v", errKYC)
	}

	if wKYC.Code != http.StatusAccepted {
		t.Errorf("expected status 202 Accepted; got %d", wKYC.Code)
	}
}

func TestTransactionHandlersIntegration(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryTransactionRepository()
	handler := NewTransactionHandler(repo, nil)

	// Test 1: Create Transaction
	txBody, _ := json.Marshal(CreateTxRequest{
		PropertyID:  "prop-123",
		BuyerID:     "usr-1",
		SellerID:    "usr-2",
		TotalAmount: 1000000.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(txBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	err := handler.CreateTransaction(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}

	var tx Transaction
	if err := json.NewDecoder(w.Body).Decode(&tx); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if tx.TotalAmount != 1000000.0 || tx.Status != core.Draft {
		t.Errorf("unexpected transaction fields: %+v", tx)
	}
}

func TestTokenizationHandlersIntegration(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryTokenizationRepository()
	handler := NewTokenizationHandler(repo)

	// Create Pool
	poolBody, _ := json.Marshal(CreatePoolRequest{
		PropertyID:  "prop-456",
		TotalTokens: 1000,
		TokenPrice:  500.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pools", bytes.NewBuffer(poolBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	err := handler.CreatePool(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}
}

func TestLedgerHandlersIntegration(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryLedgerRepository()
	handler := NewLedgerHandler(repo)

	// Write Log
	logBody, _ := json.Marshal(WriteLogRequest{
		Payload: "Transaction closed prop-456",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewBuffer(logBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	err := handler.WriteLog(c)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}
}

func TestCorrelationIDMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(CorrelationIDMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "custom-correlation-id")
	w := httptest.NewRecorder()

	e.GET("/test", func(c *echo.Context) error {
		cid := c.Get(string(CorrelationIDContextKey)).(string)
		if cid != "custom-correlation-id" {
			t.Errorf("expected correlation id custom-correlation-id; got %s", cid)
		}
		// check request context
		cidCtx := c.Request().Context().Value(CorrelationIDContextKey).(string)
		if cidCtx != "custom-correlation-id" {
			t.Errorf("expected context correlation id custom-correlation-id; got %s", cidCtx)
		}
		return c.String(http.StatusOK, "OK")
	})

	e.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK; got %d", w.Code)
	}

	cidHeader := w.Header().Get("X-Correlation-ID")
	if cidHeader != "custom-correlation-id" {
		t.Errorf("expected response header custom-correlation-id; got %s", cidHeader)
	}
}

func TestDeleteUser(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryUserRepository()
	handler := NewUserHandler(repo)

	// Seed a user
	_, err := repo.CreateUser("user-to-delete@example.com", "hash", "Delete Me", core.Buyer)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 1. Delete as a different user (Forbidden)
	t.Run("Forbidden Delete", func(t *testing.T) {
		tokenStr, _ := GenerateJWT("usr-other@example.com", core.Buyer)
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/usr-user-to-delete@example.com", nil)
		w := httptest.NewRecorder()
		c := e.NewContext(req, w)
		c.SetPathValues(echo.PathValues{
			echo.PathValue{Name: "id", Value: "usr-user-to-delete@example.com"},
		})
		c.Set("user", token)

		err := handler.DeleteUser(c)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden; got %d", w.Code)
		}
	})

	// 2. Delete own profile (Success)
	t.Run("Delete Own Profile", func(t *testing.T) {
		tokenStr, _ := GenerateJWT("usr-user-to-delete@example.com", core.Buyer)
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/usr-user-to-delete@example.com", nil)
		w := httptest.NewRecorder()
		c := e.NewContext(req, w)
		c.SetPathValues(echo.PathValues{
			echo.PathValue{Name: "id", Value: "usr-user-to-delete@example.com"},
		})
		c.Set("user", token)

		err := handler.DeleteUser(c)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204 No Content; got %d", w.Code)
		}

		// Verify user is gone
		_, err = repo.GetUser("usr-user-to-delete@example.com")
		if err == nil {
			t.Error("expected user to be deleted from repo")
		}
	})
}

func TestKYCEncryption(t *testing.T) {
	original := "my-secret-passport-number-12345"
	encrypted, err := EncryptKYC(original)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if encrypted == original {
		t.Error("encrypted value must not equal plaintext")
	}

	decrypted, err := DecryptKYC(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if decrypted != original {
		t.Errorf("expected decrypted value %s; got %s", original, decrypted)
	}
}

func TestKYCEncryption_FailClosedInProduction(t *testing.T) {
	originalEnv := os.Getenv("APP_ENV")
	defer func() { _ = os.Setenv("APP_ENV", originalEnv) }()

	// 1. Missing Vault in production must fail closed
	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("VAULT_ADDR", "")
	_ = os.Setenv("VAULT_TOKEN", "")

	_, err := EncryptKYC("secret-doc-123")
	if err == nil {
		t.Fatal("expected EncryptKYC to fail closed when Vault is missing in production, got nil error")
	}

	// 2. Unreachable Vault in production must fail closed
	_ = os.Setenv("VAULT_ADDR", "http://127.0.0.1:59999")
	_ = os.Setenv("VAULT_TOKEN", "fake-token")

	_, err = EncryptKYC("secret-doc-123")
	if err == nil {
		t.Fatal("expected EncryptKYC to fail closed when Vault is unreachable in production, got nil error")
	}

	// 3. In non-production, fallback is allowed
	_ = os.Setenv("APP_ENV", "development")
	enc, err := EncryptKYC("secret-doc-123")
	if err != nil {
		t.Fatalf("expected EncryptKYC to fall back to local AES in development, got: %v", err)
	}
	dec, err := DecryptKYC(enc)
	if err != nil {
		t.Fatalf("failed to decrypt local fallback ciphertext: %v", err)
	}
	if dec != "secret-doc-123" {
		t.Errorf("expected secret-doc-123, got: %s", dec)
	}
}
