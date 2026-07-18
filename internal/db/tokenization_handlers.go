package db

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
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
func (h *TokenizationHandler) CreatePool(c *echo.Context) error {
	var req CreatePoolRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	if req.PropertyID == "" || req.TotalTokens <= 0 || req.TokenPrice <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid pool payload parameters"})
	}

	pool, err := h.Repo.CreatePool(req.PropertyID, req.TotalTokens, req.TokenPrice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create fractional pool"})
	}

	return c.JSON(http.StatusCreated, pool)
}

// BuyShares handles POST /pools/{id}/buy
func (h *TokenizationHandler) BuyShares(c *echo.Context) error {
	poolID := c.Param("id")
	if poolID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing pool ID"})
	}

	var req BuySharesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
	}

	holding, err := h.Repo.BuyTokens(poolID, req.InvestorID, req.TokenCount)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "BuyTokens error", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to buy shares"})
	}

	return c.JSON(http.StatusOK, holding)
}

// GetPools handles GET /pools
func (h *TokenizationHandler) GetPools(c *echo.Context) error {
	pools, err := h.Repo.ListPools()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve pools"})
	}

	return c.JSON(http.StatusOK, pools)
}
