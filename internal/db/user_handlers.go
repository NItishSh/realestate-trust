package db

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/core"
)

type UserHandler struct {
	Repo UserRepository
}

func NewUserHandler(repo UserRepository) *UserHandler {
	return &UserHandler{Repo: repo}
}

// RegisterUser handles POST /users
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req core.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := core.ValidateRegistration(req); err != nil {
		log.Printf("ValidateRegistration error: %v\n", err)
		http.Error(w, "Invalid registration data", http.StatusBadRequest)
		return
	}

	user, err := h.Repo.CreateUser(req.Email, req.FullName, req.Role)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
	Email string `json:"email"`
}

// Login handles POST /login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	// For demo: Generate a JWT with user ID based on email
	userID := "usr-" + req.Email
	token, err := GenerateJWT(userID)
	if err != nil {
		log.Printf("GenerateJWT error: %v\n", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	user, err := h.Repo.GetUser(userID)
	if err != nil {
		log.Printf("GetUser error: %v\n", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// SubmitKYC handles POST /users/{id}/kyc
func (h *UserHandler) SubmitKYC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	var req core.KYCSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := core.ValidateKYC(req); err != nil {
		log.Printf("ValidateKYC error: %v\n", err)
		http.Error(w, "Invalid KYC data", http.StatusBadRequest)
		return
	}

	kyc, err := h.Repo.SubmitKYC(userID, req.DocumentType, req.DocumentReference)
	if err != nil {
		http.Error(w, "Failed to submit KYC", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(kyc)
}

// GetKYCStatus handles GET /users/{id}/kyc/status
func (h *UserHandler) GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	status, verifiedAt, err := h.Repo.GetKYCStatus(userID)
	if err != nil {
		http.Error(w, "Failed to check KYC status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":     userID,
		"status":     status,
		"verifiedAt": verifiedAt,
	})
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.Repo.ListUsers()
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
