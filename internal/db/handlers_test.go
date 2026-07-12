package db

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/realestate-trust/monorepo/internal/core"
)

func TestUserHandlersIntegration(t *testing.T) {
	repo := NewInMemoryUserRepository()
	handler := NewUserHandler(repo)

	// Test 1: User Registration
	reqBody, _ := json.Marshal(core.RegisterUserRequest{
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     core.Buyer,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	handler.RegisterUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}

	var user User
	json.NewDecoder(w.Body).Decode(&user)
	if user.Email != "test@example.com" {
		t.Errorf("expected user email test@example.com; got %s", user.Email)
	}

	// Test 2: KYC Submission
	kycBody, _ := json.Marshal(core.KYCSubmissionRequest{
		DocumentType:      "PASSPORT",
		DocumentReference: "P987654321",
	})
	
	// We mimic path parameter mapping since standard ServeMux context values are populated by router
	reqKYC := httptest.NewRequest(http.MethodPost, "/api/v1/users/usr-test@example.com/kyc", bytes.NewBuffer(kycBody))
	reqKYC.SetPathValue("id", "usr-test@example.com")
	wKYC := httptest.NewRecorder()
	handler.SubmitKYC(wKYC, reqKYC)

	if wKYC.Code != http.StatusAccepted {
		t.Errorf("expected status 202 Accepted; got %d", wKYC.Code)
	}
}

func TestTransactionHandlersIntegration(t *testing.T) {
	repo := NewInMemoryTransactionRepository()
	handler := NewTransactionHandler(repo)

	// Test 1: Create Transaction
	txBody, _ := json.Marshal(CreateTxRequest{
		PropertyID:  "prop-123",
		BuyerID:     "usr-1",
		SellerID:    "usr-2",
		TotalAmount: 1000000.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(txBody))
	w := httptest.NewRecorder()
	handler.CreateTransaction(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}

	var tx Transaction
	json.NewDecoder(w.Body).Decode(&tx)
	if tx.TotalAmount != 1000000.0 || tx.Status != core.Draft {
		t.Errorf("unexpected transaction fields: %+v", tx)
	}
}

func TestTokenizationHandlersIntegration(t *testing.T) {
	repo := NewInMemoryTokenizationRepository()
	handler := NewTokenizationHandler(repo)

	// Create Pool
	poolBody, _ := json.Marshal(CreatePoolRequest{
		PropertyID:  "prop-456",
		TotalTokens: 1000,
		TokenPrice:  500.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pools", bytes.NewBuffer(poolBody))
	w := httptest.NewRecorder()
	handler.CreatePool(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}
}

func TestLedgerHandlersIntegration(t *testing.T) {
	repo := NewInMemoryLedgerRepository()
	handler := NewLedgerHandler(repo)

	// Write Log
	logBody, _ := json.Marshal(WriteLogRequest{
		Payload: "Transaction closed prop-456",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewBuffer(logBody))
	w := httptest.NewRecorder()
	handler.WriteLog(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created; got %d", w.Code)
	}
}

