package db

import (
	"encoding/json"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/core"
)

type FinancingHandler struct {
	Repo FinancingRepository
}

func NewFinancingHandler(repo FinancingRepository) *FinancingHandler {
	return &FinancingHandler{Repo: repo}
}

type CreateLoanRequest struct {
	TransactionID   string  `json:"transactionId"`
	UserID          string  `json:"userId"`
	RequestedAmount float64 `json:"requestedAmount"`
	PropertyValue   float64 `json:"propertyValue"`
}

// ApplyLoan handles POST /loans
func (h *FinancingHandler) ApplyLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Validate underwriting bounds
	if err := core.ValidateRequest(req.RequestedAmount, req.PropertyValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loan, err := h.Repo.CreateLoan(req.TransactionID, req.UserID, req.RequestedAmount)
	if err != nil {
		http.Error(w, "Failed to create loan application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loan)
}

// GetLoan handles GET /loans/{id}
func (h *FinancingHandler) GetLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	loanID := r.PathValue("id")
	if loanID == "" {
		http.Error(w, "Missing loan ID", http.StatusBadRequest)
		return
	}

	loan, err := h.Repo.GetLoan(loanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loan)
}

// DisburseLoan handles POST /loans/{id}/disburse
func (h *FinancingHandler) DisburseLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	loanID := r.PathValue("id")
	if loanID == "" {
		http.Error(w, "Missing loan ID", http.StatusBadRequest)
		return
	}

	loan, err := h.Repo.GetLoan(loanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if loan.Status != "APPROVED" {
		http.Error(w, "Loan must be APPROVED before disbursement", http.StatusBadRequest)
		return
	}

	approvedAmount := loan.RequestedAmount
	if loan.ApprovedAmount != nil {
		approvedAmount = *loan.ApprovedAmount
	}

	vaNumber := "VA-YES-" + loan.TransactionID
	disb, err := h.Repo.CreateDisbursement(loanID, vaNumber, approvedAmount)
	if err != nil {
		http.Error(w, "Failed to disburse loan", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disb)
}

// BankWebhook handles POST /loans/webhooks/bank
func (h *FinancingHandler) BankWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ApplicationID  string  `json:"applicationId"`
		Status         string  `json:"status"`
		ApprovedAmount float64 `json:"approvedAmount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid webhook body", http.StatusBadRequest)
		return
	}

	if payload.Status == "APPROVED" {
		err := h.Repo.ApproveLoan(payload.ApplicationID, payload.ApprovedAmount)
		if err != nil {
			http.Error(w, "Failed to approve loan in webhook", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"processed"}`))
}

// GetLoans handles GET /loans
func (h *FinancingHandler) GetLoans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	loans, err := h.Repo.ListLoans()
	if err != nil {
		http.Error(w, "Failed to retrieve loans", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}
