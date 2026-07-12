package db

import (
	"errors"
	"sync"
	"time"

	"github.com/realestate-trust/monorepo/internal/core"
)

type User struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	FullName  string        `json:"fullName"`
	Role      core.UserRole `json:"role"`
	CreatedAt time.Time     `json:"createdAt"`
}

type KYCVerification struct {
	ID                string     `json:"id"`
	UserID            string     `json:"userId"`
	DocumentType      string     `json:"documentType"`
	DocumentReference string     `json:"documentReference"`
	Status            string     `json:"status"`
	VerifiedAt        *time.Time `json:"verifiedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// UserRepository interface defines data actions for users.
type UserRepository interface {
	CreateUser(email, name string, role core.UserRole) (*User, error)
	GetUser(id string) (*User, error)
	SubmitKYC(userID, docType, docRef string) (*KYCVerification, error)
	GetKYCStatus(userID string) (string, *time.Time, error)
	ListUsers() ([]*User, error)
}

// InMemoryUserRepository implements UserRepository using memory stores.
type InMemoryUserRepository struct {
	mu            sync.RWMutex
	users         map[string]*User
	verifications map[string]*KYCVerification
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:         make(map[string]*User),
		verifications: make(map[string]*KYCVerification),
	}
}

func (r *InMemoryUserRepository) CreateUser(email, name string, role core.UserRole) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Minimal mock UUID generation
	id := "usr-" + email
	user := &User{
		ID:        id,
		Email:     email,
		FullName:  name,
		Role:      role,
		CreatedAt: time.Now(),
	}
	r.users[id] = user
	return user, nil
}

func (r *InMemoryUserRepository) GetUser(id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) SubmitKYC(userID, docType, docRef string) (*KYCVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "kyc-" + userID
	kyc := &KYCVerification{
		ID:                id,
		UserID:            userID,
		DocumentType:      docType,
		DocumentReference: docRef,
		Status:            "PENDING",
		CreatedAt:         time.Now(),
	}
	r.verifications[userID] = kyc
	return kyc, nil
}

func (r *InMemoryUserRepository) GetKYCStatus(userID string) (string, *time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	kyc, ok := r.verifications[userID]
	if !ok {
		return "PENDING", nil, nil
	}
	return kyc.Status, kyc.VerifiedAt, nil
}

func (r *InMemoryUserRepository) ListUsers() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*User
	for _, u := range r.users {
		result = append(result, u)
	}
	if result == nil {
		return []*User{}, nil
	}
	return result, nil
}
