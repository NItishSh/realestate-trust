package db

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
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
func (h *TransactionHandler) CreateTransaction(c *echo.Context) error {
	var req CreateTxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	if req.PropertyID == "" || req.BuyerID == "" || req.SellerID == "" || req.TotalAmount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing or invalid payload inputs"})
	}

	tx, err := h.Repo.CreateTransaction(req.PropertyID, req.BuyerID, req.SellerID, req.TotalAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create transaction"})
	}

	return c.JSON(http.StatusCreated, tx)
}

// GetTransaction handles GET /transactions/{id}
func (h *TransactionHandler) GetTransaction(c *echo.Context) error {
	txID := c.Param("id")
	if txID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing transaction ID"})
	}

	tx, err := h.Repo.GetTransaction(txID)
	if err != nil {
		log.Printf("GetTransaction error: %v\n", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Transaction not found"})
	}

	return c.JSON(http.StatusOK, tx)
}

// UpdateStatus handles PUT /transactions/{id}/status
func (h *TransactionHandler) UpdateStatus(c *echo.Context) error {
	txID := c.Param("id")
	if txID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing transaction ID"})
	}

	var req struct {
		NewState core.TransactionState `json:"newState"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.Repo.UpdateTransactionStatus(txID, req.NewState); err != nil {
		log.Printf("UpdateTransactionStatus error: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to update transaction status"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// FundEscrow handles POST /transactions/{id}/escrow/fund
func (h *TransactionHandler) FundEscrow(c *echo.Context) error {
	txID := c.Param("id")
	if txID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing transaction ID"})
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Amount must be positive"})
	}

	if err := h.Repo.FundEscrow(txID, req.Amount); err != nil {
		log.Printf("FundEscrow error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fund escrow"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "funding_received"})
}

// GetTransactions handles GET /transactions
func (h *TransactionHandler) GetTransactions(c *echo.Context) error {
	txs, err := h.Repo.ListTransactions()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions"})
	}

	return c.JSON(http.StatusOK, txs)
}
