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

func TestFeedbackHandler_RatingBounds(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryFeedbackRepository()
	handler := NewFeedbackHandler(repo)

	// 1. Invalid rating < 1
	body, _ := json.Marshal(CreateFeedbackRequest{Message: "Great", Category: "General", Rating: 0})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler.CreateFeedback(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for rating 0, got %d", rec.Code)
	}

	// 2. Invalid rating > 5
	body, _ = json.Marshal(CreateFeedbackRequest{Message: "Great", Category: "General", Rating: 6})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler.CreateFeedback(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for rating 6, got %d", rec.Code)
	}

	// 3. Valid rating 5
	body, _ = json.Marshal(CreateFeedbackRequest{Message: "Great", Category: "General", Rating: 5})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler.CreateFeedback(c)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201 for valid rating 5, got %d", rec.Code)
	}
}

func TestFinancingHandler_BankWebhookAuth(t *testing.T) {
	originalSecret := os.Getenv("BANK_WEBHOOK_SECRET")
	defer func() { _ = os.Setenv("BANK_WEBHOOK_SECRET", originalSecret) }()

	_ = os.Setenv("BANK_WEBHOOK_SECRET", "super-secret-key")

	e := echo.New()
	repo := NewInMemoryFinancingRepository()
	handler := NewFinancingHandler(repo)

	// Create a loan
	loan, _ := repo.CreateLoan("tx-1", "usr-1", 100000.0)

	// 1. Missing secret header (Unauthorized 401)
	webhookPayload := []byte(`{"applicationId":"` + loan.ID + `","status":"APPROVED","approvedAmount":100000.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans/webhooks/bank", bytes.NewBuffer(webhookPayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler.BankWebhook(c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing webhook secret, got %d", rec.Code)
	}

	// 2. Valid secret header (Success 200)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/loans/webhooks/bank", bytes.NewBuffer(webhookPayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Webhook-Secret", "super-secret-key")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	_ = handler.BankWebhook(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid webhook secret, got %d", rec.Code)
	}
}

func TestGetCORSOrigins(t *testing.T) {
	orig := os.Getenv("CORS_ALLOWED_ORIGINS")
	defer func() { _ = os.Setenv("CORS_ALLOWED_ORIGINS", orig) }()

	// 1. Default fallback
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "")
	defaults := GetCORSOrigins()
	if len(defaults) != 2 || defaults[0] != "http://localhost:3000" {
		t.Errorf("expected default localhost:3000, got %v", defaults)
	}

	// 2. Custom comma-separated
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "https://app.realestate-trust.com, https://admin.realestate-trust.com")
	custom := GetCORSOrigins()
	if len(custom) != 2 || custom[0] != "https://app.realestate-trust.com" || custom[1] != "https://admin.realestate-trust.com" {
		t.Errorf("expected parsed custom origins, got %v", custom)
	}
}

func TestAuthRateLimiterMiddleware(t *testing.T) {
	e := echo.New()
	e.POST("/api/v1/auth-test", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}, AuthRateLimiterMiddleware())

	// Send requests within burst limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth-test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for request %d, got %d", i, rec.Code)
		}
	}
}

func TestUserHandlers_ABACEnforcement(t *testing.T) {
	e := echo.New()
	repo := NewInMemoryUserRepository()
	handler := NewUserHandler(repo)

	// Create test users
	userA, _ := repo.CreateUser("user-a@example.com", "User A", "secret", core.Buyer)
	userB, _ := repo.CreateUser("user-b@example.com", "User B", "secret", core.Buyer)

	// 1. User A cannot view User B profile (403 Forbidden)
	tokenStrA, _ := GenerateJWT(userA.ID, core.Buyer)
	tokenA, _ := jwt.Parse(tokenStrA, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userB.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{echo.PathValue{Name: "id", Value: userB.ID}})
	c.Set("user", tokenA)

	_ = handler.GetUser(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User A accessing User B, got %d", rec.Code)
	}

	// 2. Admin can view User B profile (200 OK)
	tokenStrAdmin, _ := GenerateJWT("admin-id", core.Admin)
	tokenAdmin, _ := jwt.Parse(tokenStrAdmin, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userB.ID, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{echo.PathValue{Name: "id", Value: userB.ID}})
	c.Set("user", tokenAdmin)

	_ = handler.GetUser(c)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for Admin accessing User B, got %d", rec.Code)
	}

	// 3. User A cannot submit KYC for User B (403 Forbidden)
	kycBody, _ := json.Marshal(core.KYCSubmissionRequest{
		DocumentType:      "PASSPORT",
		DocumentReference: "PASS12345",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userB.ID+"/kyc", bytes.NewBuffer(kycBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{echo.PathValue{Name: "id", Value: userB.ID}})
	c.Set("user", tokenA)

	_ = handler.SubmitKYC(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for User A submitting KYC for User B, got %d", rec.Code)
	}
}

func TestRBACMiddleware_Enforcement(t *testing.T) {
	e := echo.New()
	e.GET("/api/v1/officer-only", func(c *echo.Context) error {
		return c.String(http.StatusOK, "welcome officer")
	}, RBACMiddleware(core.Officer, core.Admin))

	// 1. Buyer gets 403
	tokenStrBuyer, _ := GenerateJWT("buyer-id", core.Buyer)
	tokenBuyer, _ := jwt.Parse(tokenStrBuyer, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/officer-only", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", tokenBuyer)

	_ = RBACMiddleware(core.Officer, core.Admin)(func(c *echo.Context) error {
		return c.String(http.StatusOK, "welcome officer")
	})(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for Buyer on officer route, got %d", rec.Code)
	}

	// 2. Officer gets 200
	tokenStrOfficer, _ := GenerateJWT("officer-id", core.Officer)
	tokenOfficer, _ := jwt.Parse(tokenStrOfficer, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	req = httptest.NewRequest(http.MethodGet, "/api/v1/officer-only", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user", tokenOfficer)

	_ = RBACMiddleware(core.Officer, core.Admin)(func(c *echo.Context) error {
		return c.String(http.StatusOK, "welcome officer")
	})(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for Officer on officer route, got %d", rec.Code)
	}
}
