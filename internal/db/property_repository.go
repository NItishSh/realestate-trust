package db

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Property struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Thumbnail   string    `json:"thumbnail"`
	OwnerID     string    `json:"ownerId"`
	Documents   []string  `json:"documents"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PropertyRepository interface {
	ListProperties() ([]*Property, error)
	GetProperty(id string) (*Property, error)
	CreateProperty(address, description string, value float64, thumbnail, ownerId string) (*Property, error)
}

type InMemoryPropertyRepository struct {
	properties map[string]*Property
	mu         sync.RWMutex
}

func NewInMemoryPropertyRepository() PropertyRepository {
	return &InMemoryPropertyRepository{
		properties: make(map[string]*Property),
	}
}

func (r *InMemoryPropertyRepository) ListProperties() ([]*Property, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Property, 0, len(r.properties))
	for _, p := range r.properties {
		list = append(list, p)
	}
	return list, nil
}

func (r *InMemoryPropertyRepository) GetProperty(id string) (*Property, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.properties[id]
	if !exists {
		return nil, errors.New("property not found")
	}
	return p, nil
}

func (r *InMemoryPropertyRepository) CreateProperty(address, description string, value float64, thumbnail, ownerId string) (*Property, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "prop-" + uuid.New().String()[:8]

	p := &Property{
		ID:          id,
		Address:     address,
		Description: description,
		Value:       value,
		Thumbnail:   thumbnail,
		OwnerID:     ownerId,
		Documents:   []string{"Deed of Trust", "Property Inspection Report", "Title Insurance"},
		CreatedAt:   time.Now(),
	}

	r.properties[p.ID] = p
	return p, nil
}
