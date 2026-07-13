package db

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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
func (h *UserHandler) RegisterUser(c echo.Context) error {
	var req core.RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if err := core.ValidateRegistration(req); err != nil {
		log.Printf("ValidateRegistration error: %v\n", err)
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
func (h *UserHandler) Login(c echo.Context) error {
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
		log.Printf("GenerateJWT error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	user, err := h.Repo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser error: %v\n", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// SubmitKYC handles POST /users/{id}/kyc
func (h *UserHandler) SubmitKYC(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing user ID"})
	}

	var req core.KYCSubmissionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if err := core.ValidateKYC(req); err != nil {
		log.Printf("ValidateKYC error: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid KYC data"})
	}

	kyc, err := h.Repo.SubmitKYC(userID, req.DocumentType, req.DocumentReference)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit KYC"})
	}

	return c.JSON(http.StatusAccepted, kyc)
}

// GetKYCStatus handles GET /users/{id}/kyc/status
func (h *UserHandler) GetKYCStatus(c echo.Context) error {
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
func (h *UserHandler) GetUsers(c echo.Context) error {
	users, err := h.Repo.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve users"})
	}

	return c.JSON(http.StatusOK, users)
}
