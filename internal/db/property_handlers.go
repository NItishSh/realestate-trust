package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type PropertyHandler struct {
	Repo PropertyRepository
}

func NewPropertyHandler(repo PropertyRepository) *PropertyHandler {
	return &PropertyHandler{Repo: repo}
}

func (h *PropertyHandler) ListProperties(w http.ResponseWriter, r *http.Request) {
	properties, err := h.Repo.ListProperties()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(properties); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func (h *PropertyHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	p, err := h.Repo.GetProperty(id)
	if err != nil {
		http.Error(w, "property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(p); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// UnlockDocuments Request Payload
type UnlockRequest struct {
	BuyerID string `json:"buyerId"`
}

func (h *PropertyHandler) UnlockDocuments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing property id", http.StatusBadRequest)
		return
	}

	var req UnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	p, err := h.Repo.GetProperty(id)
	if err != nil {
		http.Error(w, "property not found", http.StatusNotFound)
		return
	}

	// 1. Create a mini-escrow for Earnest Money via Transaction Manager
	// Earnest money is fixed at ₹50,000 for document unlock
	earnestMoney := 50000.0

	txPayload := map[string]interface{}{
		"propertyId":  p.ID,
		"buyerId":     req.BuyerID,
		"sellerId":    p.OwnerID,
		"totalAmount": earnestMoney,
	}
	bodyBytes, _ := json.Marshal(txPayload)

	// Call Transaction Manager internally via Kubernetes DNS
	resp, err := http.Post("http://transaction-manager:8080/api/v1/transactions", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error creating escrow: %v\n", err)
		http.Error(w, "failed to initiate earnest money deposit escrow", http.StatusInternalServerError)
		return
	}

	// 2. Return the Data Room documents and a success message
	response := map[string]interface{}{
		"message":   "Earnest money escrow initiated. Legal documents unlocked.",
		"documents": p.Documents,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}
