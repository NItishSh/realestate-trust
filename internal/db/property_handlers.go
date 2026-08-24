package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
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

type CreatePropertyRequest struct {
	Address     string  `json:"address"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Thumbnail   string  `json:"thumbnail"`
}

func (h *PropertyHandler) CreateProperty(c *echo.Context) error {
	var req CreatePropertyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Address == "" || req.Description == "" || req.Value <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing required fields"})
	}

	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: missing or invalid token"})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: invalid claims"})
	}
	ownerID, ok := claims["sub"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: user ID missing from claims"})
	}

	if req.Thumbnail == "" {
		req.Thumbnail = "https://images.unsplash.com/photo-1518780664697-55e3ad937233?w=800"
	}

	p, err := h.Repo.CreateProperty(req.Address, req.Description, req.Value, req.Thumbnail, ownerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create property"})
	}

	return c.JSON(http.StatusCreated, p)
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

type VerifyInsuranceRequest struct {
	Company string `json:"company"`
	Policy  string `json:"policy"`
}

func (h *PropertyHandler) VerifyTitleInsurance(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing property id"})
	}

	var req VerifyInsuranceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Company == "" || req.Policy == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "company and policy are required"})
	}

	p, err := h.Repo.UpdateTitleInsurance(id, "INSURED", req.Company, req.Policy)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
	}

	// Record verification in the ledger log
	payload := fmt.Sprintf("Title Insurance Verified: Policy #%s issued by %s for property at %s", req.Policy, req.Company, p.Address)
	logRequest := map[string]string{"payload": payload}
	bodyBytes, _ := json.Marshal(logRequest)

	resp, postErr := http.Post("http://ledger-service:8084/api/v1/logs", "application/json", bytes.NewBuffer(bodyBytes))
	if postErr != nil {
		fmt.Printf("Error sending ledger log: %v\n", postErr)
	} else {
		defer resp.Body.Close()
	}

	return c.JSON(http.StatusOK, p)
}

type UpdatePropertyDetailsRequest struct {
	SqFt         int    `json:"sqft"`
	Bedrooms     int    `json:"bedrooms"`
	Bathrooms    int    `json:"bathrooms"`
	YearBuilt    int    `json:"yearBuilt"`
	PropertyType string `json:"propertyType"`
}

func (h *PropertyHandler) UpdatePropertyDetails(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing property id"})
	}

	var req UpdatePropertyDetailsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.SqFt <= 0 || req.Bedrooms < 0 || req.Bathrooms < 0 || req.YearBuilt <= 0 || req.PropertyType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid details values"})
	}

	p, err := h.Repo.UpdatePropertyDetails(id, req.SqFt, req.Bedrooms, req.Bathrooms, req.YearBuilt, req.PropertyType)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "property not found"})
	}

	// Record updating in ledger log
	payload := fmt.Sprintf("Property Details Updated: ID %s, SqFt %d, Bed %d, Bath %d, Year %d, Type %s", p.ID, p.SqFt, p.Bedrooms, p.Bathrooms, p.YearBuilt, p.PropertyType)
	logRequest := map[string]string{"payload": payload}
	bodyBytes, _ := json.Marshal(logRequest)

	resp, postErr := http.Post("http://ledger-service:8084/api/v1/logs", "application/json", bytes.NewBuffer(bodyBytes))
	if postErr != nil {
		fmt.Printf("Error sending ledger log: %v\n", postErr)
	} else {
		defer resp.Body.Close()
	}

	return c.JSON(http.StatusOK, p)
}
