package db

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
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
