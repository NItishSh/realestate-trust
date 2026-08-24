package db

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Property struct {
	ID                       string    `json:"id"`
	Address                  string    `json:"address"`
	Description              string    `json:"description"`
	Value                    float64   `json:"value"`
	Thumbnail                string    `json:"thumbnail"`
	OwnerID                  string    `json:"ownerId"`
	Documents                []string  `json:"documents"`
	CreatedAt                time.Time `json:"createdAt"`
	TitleInsuranceStatus     string    `json:"titleInsuranceStatus"`
	TitleInsurancePolicy     string    `json:"titleInsurancePolicy"`
	TitleInsuranceCompany    string    `json:"titleInsuranceCompany"`
	TitleInsuranceVerifiedAt string    `json:"titleInsuranceVerifiedAt"`
	SqFt                     int       `json:"sqft"`
	Bedrooms                 int       `json:"bedrooms"`
	Bathrooms                int       `json:"bathrooms"`
	YearBuilt                int       `json:"yearBuilt"`
	PropertyType             string    `json:"propertyType"`
}

type PropertyRepository interface {
	ListProperties() ([]*Property, error)
	GetProperty(id string) (*Property, error)
	CreateProperty(address, description string, value float64, thumbnail, ownerId string) (*Property, error)
	CreatePropertyWithID(id, address, description string, value float64, thumbnail, ownerId string) (*Property, error)
	UpdateTitleInsurance(id, status, company, policy string) (*Property, error)
	UpdatePropertyDetails(id string, sqft, bedrooms, bathrooms, yearBuilt int, propType string) (*Property, error)
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
		ID:                   id,
		Address:              address,
		Description:          description,
		Value:                value,
		Thumbnail:            thumbnail,
		OwnerID:              ownerId,
		Documents:            []string{"Deed of Trust", "Property Inspection Report", "Title Insurance"},
		CreatedAt:            time.Now(),
		TitleInsuranceStatus: "UNINSURED",
		SqFt:                 1800,
		Bedrooms:             3,
		Bathrooms:            2,
		YearBuilt:            2018,
		PropertyType:         "Residential",
	}

	r.properties[p.ID] = p
	return p, nil
}

func (r *InMemoryPropertyRepository) CreatePropertyWithID(id, address, description string, value float64, thumbnail, ownerId string) (*Property, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == "" {
		id = "prop-" + uuid.New().String()[:8]
	}

	p := &Property{
		ID:                   id,
		Address:              address,
		Description:          description,
		Value:                value,
		Thumbnail:            thumbnail,
		OwnerID:              ownerId,
		Documents:            []string{"Deed of Trust", "Property Inspection Report", "Title Insurance"},
		CreatedAt:            time.Now(),
		TitleInsuranceStatus: "UNINSURED",
		SqFt:                 1800,
		Bedrooms:             3,
		Bathrooms:            2,
		YearBuilt:            2018,
		PropertyType:         "Residential",
	}

	r.properties[p.ID] = p
	return p, nil
}

func (r *InMemoryPropertyRepository) UpdateTitleInsurance(id, status, company, policy string) (*Property, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.properties[id]
	if !exists {
		return nil, errors.New("property not found")
	}

	p.TitleInsuranceStatus = status
	p.TitleInsuranceCompany = company
	p.TitleInsurancePolicy = policy
	p.TitleInsuranceVerifiedAt = time.Now().Format(time.RFC3339)

	return p, nil
}

func (r *InMemoryPropertyRepository) UpdatePropertyDetails(id string, sqft, bedrooms, bathrooms, yearBuilt int, propType string) (*Property, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.properties[id]
	if !exists {
		return nil, errors.New("property not found")
	}

	p.SqFt = sqft
	p.Bedrooms = bedrooms
	p.Bathrooms = bathrooms
	p.YearBuilt = yearBuilt
	p.PropertyType = propType

	return p, nil
}
