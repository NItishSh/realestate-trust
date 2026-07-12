package db

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type LedgerHandler struct {
	Repo LedgerRepository
}

func NewLedgerHandler(repo LedgerRepository) *LedgerHandler {
	return &LedgerHandler{Repo: repo}
}

type WriteLogRequest struct {
	Payload string `json:"payload"`
}

// WriteLog handles POST /logs
func (h *LedgerHandler) WriteLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WriteLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Payload == "" {
		http.Error(w, "Payload content cannot be empty", http.StatusBadRequest)
		return
	}

	entry, err := h.Repo.WriteLog(req.Payload)
	if err != nil {
		http.Error(w, "Failed to commit log to audit ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// GetLog handles GET /logs/{index}
func (h *LedgerHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idxStr := r.PathValue("index")
	if idxStr == "" {
		http.Error(w, "Missing index parameter", http.StatusBadRequest)
		return
	}

	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid index parameter format", http.StatusBadRequest)
		return
	}

	entry, err := h.Repo.GetLog(idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// GetLogs handles GET /logs
func (h *LedgerHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := h.Repo.ListLogs()
	if err != nil {
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
