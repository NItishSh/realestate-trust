package db

import (
	"net/http"

	echo "github.com/labstack/echo/v5"
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
func (h *FinancingHandler) ApplyLoan(c *echo.Context) error {
	var req CreateLoanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	// Validate underwriting bounds
	if err := core.ValidateRequest(req.RequestedAmount, req.PropertyValue); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	loan, err := h.Repo.CreateLoan(req.TransactionID, req.UserID, req.RequestedAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create loan application"})
	}

	return c.JSON(http.StatusCreated, loan)
}

// GetLoan handles GET /loans/{id}
func (h *FinancingHandler) GetLoan(c *echo.Context) error {
	loanID := c.Param("id")
	if loanID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing loan ID"})
	}

	loan, err := h.Repo.GetLoan(loanID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, loan)
}

// DisburseLoan handles POST /loans/{id}/disburse
func (h *FinancingHandler) DisburseLoan(c *echo.Context) error {
	loanID := c.Param("id")
	if loanID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing loan ID"})
	}

	loan, err := h.Repo.GetLoan(loanID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	if loan.Status != "APPROVED" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Loan must be APPROVED before disbursement"})
	}

	approvedAmount := loan.RequestedAmount
	if loan.ApprovedAmount != nil {
		approvedAmount = *loan.ApprovedAmount
	}

	vaNumber := "VA-YES-" + loan.TransactionID
	disb, err := h.Repo.CreateDisbursement(loanID, vaNumber, approvedAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to disburse loan"})
	}

	return c.JSON(http.StatusOK, disb)
}

// BankWebhook handles POST /loans/webhooks/bank
func (h *FinancingHandler) BankWebhook(c *echo.Context) error {
	var payload struct {
		ApplicationID  string  `json:"applicationId"`
		Status         string  `json:"status"`
		ApprovedAmount float64 `json:"approvedAmount"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid webhook body"})
	}

	if payload.Status == "APPROVED" {
		err := h.Repo.ApproveLoan(payload.ApplicationID, payload.ApprovedAmount)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to approve loan in webhook"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "processed"})
}

// GetLoans handles GET /loans
func (h *FinancingHandler) GetLoans(c *echo.Context) error {
	loans, err := h.Repo.ListLoans()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve loans"})
	}

	return c.JSON(http.StatusOK, loans)
}
