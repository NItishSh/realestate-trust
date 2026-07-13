package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

type PropertyHandler struct {
	Repo PropertyRepository
}

func NewPropertyHandler(repo PropertyRepository) *PropertyHandler {
	return &PropertyHandler{Repo: repo}
}

func (h *PropertyHandler) ListProperties(c *echo.Context) error {
	properties, err := h.Repo.ListProperties()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, properties)
}

func (h *PropertyHandler) GetProperty(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing id"})
	}

	p, err := h.Repo.GetProperty(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
	}

	return c.JSON(http.StatusOK, p)
}

// UnlockDocuments Request Payload
type UnlockRequest struct {
	BuyerID string `json:"buyerId"`
}

func (h *PropertyHandler) UnlockDocuments(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing property id"})
	}

	var req UnlockRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	p, err := h.Repo.GetProperty(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to initiate earnest money deposit escrow"})
	}

	// 2. Return the Data Room documents and a success message
	response := map[string]interface{}{
		"message":   "Earnest money escrow initiated. Legal documents unlocked.",
		"documents": p.Documents,
	}

	return c.JSON(http.StatusOK, response)
}
