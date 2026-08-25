package db

import (
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v5"
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
func (h *LedgerHandler) WriteLog(c *echo.Context) error {
	var req WriteLogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Payload == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Payload content cannot be empty"})
	}

	entry, err := h.Repo.WriteLog("", req.Payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit log to audit ledger"})
	}

	return c.JSON(http.StatusCreated, entry)
}

// GetLog handles GET /logs/{index}
func (h *LedgerHandler) GetLog(c *echo.Context) error {
	idxStr := c.Param("index")
	if idxStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing index parameter"})
	}

	idx, err := strconv.ParseInt(idxStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid index parameter format"})
	}

	entry, err := h.Repo.GetLog(idx)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, entry)
}

// GetLogs handles GET /logs
func (h *LedgerHandler) GetLogs(c *echo.Context) error {
	logs, err := h.Repo.ListLogs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve logs"})
	}

	return c.JSON(http.StatusOK, logs)
}
