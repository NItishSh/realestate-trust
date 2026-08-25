package db

import (
	"log/slog"
	"net/http"

	jwt "github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v5"
	"github.com/realestate-trust/monorepo/internal/core"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Repo UserRepository
}

func NewUserHandler(repo UserRepository) *UserHandler {
	return &UserHandler{Repo: repo}
}

// RegisterUser handles POST /users
func (h *UserHandler) RegisterUser(c *echo.Context) error {
	var req core.RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if err := core.ValidateRegistration(req); err != nil {
		slog.ErrorContext(c.Request().Context(), "ValidateRegistration error", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid registration data: " + err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process password"})
	}

	user, err := h.Repo.CreateUser(req.Email, string(hashedPassword), req.FullName, req.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register user"})
	}

	return c.JSON(http.StatusCreated, user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /login
func (h *UserHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password required"})
	}

	user, err := h.Repo.GetUserByEmail(req.Email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, err := GenerateJWT(user.ID, user.Role)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "GenerateJWT error", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	refreshToken, err := h.Repo.CreateSession(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate session"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token":        token,
		"refreshToken": refreshToken,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshToken handles POST /refresh
func (h *UserHandler) RefreshToken(c *echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
	}

	userID, err := h.Repo.ValidateSession(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired refresh token"})
	}

	user, err := h.Repo.GetUser(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	accessToken, err := GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	// Rotate the refresh token
	_ = h.Repo.RevokeSession(req.RefreshToken)
	newRefreshToken, err := h.Repo.CreateSession(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate new refresh token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token":        accessToken,
		"refreshToken": newRefreshToken,
	})
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Logout handles POST /logout
func (h *UserHandler) Logout(c *echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.RefreshToken != "" {
		_ = h.Repo.RevokeSession(req.RefreshToken)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(c *echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	user, err := h.Repo.GetUser(userID)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "GetUser error", "err", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// SubmitKYC handles POST /users/{id}/kyc
func (h *UserHandler) SubmitKYC(c *echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	var req core.KYCSubmissionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if err := core.ValidateKYC(req); err != nil {
		slog.ErrorContext(c.Request().Context(), "ValidateKYC error", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid KYC data"})
	}

	kyc, err := h.Repo.SubmitKYC(userID, req.DocumentType, req.DocumentReference)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit KYC"})
	}

	return c.JSON(http.StatusAccepted, kyc)
}

// GetKYCStatus handles GET /users/{id}/kyc/status
func (h *UserHandler) GetKYCStatus(c *echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	status, verifiedAt, err := h.Repo.GetKYCStatus(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check KYC status"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"userId":     userID,
		"status":     status,
		"verifiedAt": verifiedAt,
	})
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(c *echo.Context) error {
	users, err := h.Repo.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve users"})
	}

	return c.JSON(http.StatusOK, users)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: missing or invalid token"})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: invalid claims"})
	}
	tokenUserID, ok := claims["sub"].(string)
	tokenRole, _ := claims["role"].(string)

	if tokenUserID != userID && tokenRole != "ADMIN" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: you can only delete your own account"})
	}

	err := h.Repo.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		slog.ErrorContext(c.Request().Context(), "DeleteUser error", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.NoContent(http.StatusNoContent)
}
