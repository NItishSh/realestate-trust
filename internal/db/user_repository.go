package db

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/realestate-trust/monorepo/internal/core"
)

type User struct {
	ID           string        `json:"id"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-"`
	FullName     string        `json:"fullName"`
	Role         core.UserRole `json:"role"`
	CreatedAt    time.Time     `json:"createdAt"`
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

type RefreshTokenSession struct {
	Token     string    `json:"token"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserRepository interface defines data actions for users.
type UserRepository interface {
	CreateUser(email, passwordHash, name string, role core.UserRole) (*User, error)
	GetUser(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	SubmitKYC(userID, docType, docRef string) (*KYCVerification, error)
	GetKYCStatus(userID string) (string, *time.Time, error)
	ListUsers() ([]*User, error)
	CreateSession(userID string) (string, error)
	ValidateSession(token string) (string, error)
	RevokeSession(token string) error
}

// InMemoryUserRepository implements UserRepository using memory stores.
type InMemoryUserRepository struct {
	mu            sync.RWMutex
	users         map[string]*User
	verifications map[string]*KYCVerification
	sessions      map[string]*RefreshTokenSession
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:         make(map[string]*User),
		verifications: make(map[string]*KYCVerification),
		sessions:      make(map[string]*RefreshTokenSession),
	}
}

func (r *InMemoryUserRepository) CreateUser(email, passwordHash, name string, role core.UserRole) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Minimal mock UUID generation
	id := "usr-" + email
	user := &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     name,
		Role:         role,
		CreatedAt:    time.Now(),
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

func (r *InMemoryUserRepository) GetUserByEmail(email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
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

func (r *InMemoryUserRepository) CreateSession(userID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token := uuid.NewString()
	r.sessions[token] = &RefreshTokenSession{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	return token, nil
}

func (r *InMemoryUserRepository) ValidateSession(token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[token]
	if !ok {
		return "", errors.New("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("session expired")
	}
	return session.UserID, nil
}

func (r *InMemoryUserRepository) RevokeSession(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, token)
	return nil
}
