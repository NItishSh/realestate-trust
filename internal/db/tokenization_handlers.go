package db

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
)

type TokenizationHandler struct {
	Repo TokenizationRepository
}

func NewTokenizationHandler(repo TokenizationRepository) *TokenizationHandler {
	return &TokenizationHandler{Repo: repo}
}

type CreatePoolRequest struct {
	PropertyID  string  `json:"propertyId"`
	TotalTokens int64   `json:"totalTokens"`
	TokenPrice  float64 `json:"tokenPrice"`
}

type BuySharesRequest struct {
	InvestorID string `json:"investorId"`
	TokenCount int64  `json:"tokenCount"`
}

// CreatePool handles POST /pools
func (h *TokenizationHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.PropertyID == "" || req.TotalTokens <= 0 || req.TokenPrice <= 0 {
		http.Error(w, "Invalid pool payload parameters", http.StatusBadRequest)
		return
	}

	pool, err := h.Repo.CreatePool(req.PropertyID, req.TotalTokens, req.TokenPrice)
	if err != nil {
		http.Error(w, "Failed to create fractional pool", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(pool); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// BuyShares handles POST /pools/{id}/buy
func (h *TokenizationHandler) BuyShares(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	poolID := r.PathValue("id")
	if poolID == "" {
		http.Error(w, "Missing pool ID", http.StatusBadRequest)
		return
	}

	var req BuySharesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	holding, err := h.Repo.BuyTokens(poolID, req.InvestorID, req.TokenCount)
	if err != nil {
		log.Printf("BuyTokens error: %v\n", err)
		http.Error(w, "Failed to buy shares", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(holding); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// GetPools handles GET /pools
func (h *TokenizationHandler) GetPools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pools, err := h.Repo.ListPools()
	if err != nil {
		http.Error(w, "Failed to retrieve pools", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pools); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}
