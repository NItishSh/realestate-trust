package db

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

type FeedbackHandler struct {
	Repo FeedbackRepository
}

func NewFeedbackHandler(repo FeedbackRepository) *FeedbackHandler {
	return &FeedbackHandler{Repo: repo}
}

type CreateFeedbackRequest struct {
	Message  string `json:"message"`
	Category string `json:"category"`
	Rating   int    `json:"rating"`
}

func (h *FeedbackHandler) CreateFeedback(c *echo.Context) error {
	var req CreateFeedbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message is required"})
	}

	userID := "anonymous"
	// Extract user ID from JWT token if available
	if userVal := c.Get("user"); userVal != nil {
		if userToken, ok := userVal.(*jwt.Token); ok {
			if claims, ok := userToken.Claims.(jwt.MapClaims); ok {
				if sub, ok := claims["sub"].(string); ok {
					userID = sub
				}
			}
		}
	}

	feedback, err := h.Repo.CreateFeedback(userID, req.Message, req.Category, req.Rating)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, feedback)
}

func (h *FeedbackHandler) ListFeedback(c *echo.Context) error {
	feedbacks, err := h.Repo.ListFeedback()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, feedbacks)
}
