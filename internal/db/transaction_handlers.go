package db

import (
	"encoding/json"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/core"
)

type TransactionHandler struct {
	Repo TransactionRepository
}

func NewTransactionHandler(repo TransactionRepository) *TransactionHandler {
	return &TransactionHandler{Repo: repo}
}

type CreateTxRequest struct {
	PropertyID  string  `json:"propertyId"`
	BuyerID     string  `json:"buyerId"`
	SellerID    string  `json:"sellerId"`
	TotalAmount float64 `json:"totalAmount"`
}

// CreateTransaction handles POST /transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.PropertyID == "" || req.BuyerID == "" || req.SellerID == "" || req.TotalAmount <= 0 {
		http.Error(w, "Missing or invalid payload inputs", http.StatusBadRequest)
		return
	}

	tx, err := h.Repo.CreateTransaction(req.PropertyID, req.BuyerID, req.SellerID, req.TotalAmount)
	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

// GetTransaction handles GET /transactions/{id}
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	tx, err := h.Repo.GetTransaction(txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// UpdateStatus handles PUT /transactions/{id}/status
func (h *TransactionHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewState core.TransactionState `json:"newState"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateTransactionStatus(txID, req.NewState); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

// FundEscrow handles POST /transactions/{id}/escrow/fund
func (h *TransactionHandler) FundEscrow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txID := r.PathValue("id")
	if txID == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if err := h.Repo.FundEscrow(txID, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"funding_received"}`))
}

// GetTransactions handles GET /transactions
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txs, err := h.Repo.ListTransactions()
	if err != nil {
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}
