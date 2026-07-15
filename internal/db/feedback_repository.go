package db

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Feedback struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Message   string    `json:"message"`
	Category  string    `json:"category"` // e.g., "bug", "feature_request", "general"
	Rating    int       `json:"rating"`   // 1 to 5
	CreatedAt time.Time `json:"createdAt"`
}

type FeedbackRepository interface {
	CreateFeedback(userID, message, category string, rating int) (*Feedback, error)
	ListFeedback() ([]*Feedback, error)
}

type InMemoryFeedbackRepository struct {
	mu        sync.RWMutex
	feedbacks map[string]*Feedback
}

func NewInMemoryFeedbackRepository() FeedbackRepository {
	return &InMemoryFeedbackRepository{
		feedbacks: make(map[string]*Feedback),
	}
}

func (r *InMemoryFeedbackRepository) CreateFeedback(userID, message, category string, rating int) (*Feedback, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	id := "fb-" + uuid.New().String()[:8]
	feedback := &Feedback{
		ID:        id,
		UserID:    userID,
		Message:   message,
		Category:  category,
		Rating:    rating,
		CreatedAt: time.Now(),
	}

	r.feedbacks[id] = feedback
	return feedback, nil
}

func (r *InMemoryFeedbackRepository) ListFeedback() ([]*Feedback, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Feedback
	for _, f := range r.feedbacks {
		result = append(result, f)
	}
	if result == nil {
		return []*Feedback{}, nil
	}
	return result, nil
}
