package db

import (
	"sync"

	"github.com/realestate-trust/monorepo/internal/core"
)

// TokenizationRepository defines storage operations for property shares.
type TokenizationRepository interface {
	CreatePool(propertyID string, totalTokens int64, price float64) (*core.FractionalPool, error)
	GetPool(id string) (*core.FractionalPool, error)
	BuyTokens(poolID, investorID string, count int64) (*core.FractionalHolding, error)
	ListPools() ([]*core.FractionalPool, error)
}

// InMemoryTokenizationRepository implements TokenizationRepository.
type InMemoryTokenizationRepository struct {
	mu       sync.RWMutex
	pools    map[string]*core.FractionalPool
	holdings map[string]*core.FractionalHolding
}

func NewInMemoryTokenizationRepository() *InMemoryTokenizationRepository {
	return &InMemoryTokenizationRepository{
		pools:    make(map[string]*core.FractionalPool),
		holdings: make(map[string]*core.FractionalHolding),
	}
}

func (r *InMemoryTokenizationRepository) CreatePool(propertyID string, totalTokens int64, price float64) (*core.FractionalPool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "pool-" + propertyID
	pool := &core.FractionalPool{
		ID:          id,
		PropertyID:  propertyID,
		TotalTokens: totalTokens,
		TokensSold:  0,
		TokenPrice:  price,
	}
	r.pools[id] = pool
	return pool, nil
}

func (r *InMemoryTokenizationRepository) GetPool(id string) (*core.FractionalPool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pool, ok := r.pools[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return pool, nil
}

func (r *InMemoryTokenizationRepository) BuyTokens(poolID, investorID string, count int64) (*core.FractionalHolding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pool, ok := r.pools[poolID]
	if !ok {
		return nil, core.ErrNotFound
	}

	if err := pool.ValidateTokenPurchase(count); err != nil {
		return nil, err
	}

	pool.TokensSold += count

	holdingID := "hold-" + poolID + "-" + investorID
	holding, ok := r.holdings[holdingID]
	if !ok {
		holding = &core.FractionalHolding{
			ID:         holdingID,
			PoolID:     poolID,
			InvestorID: investorID,
			TokenCount: 0,
		}
		r.holdings[holdingID] = holding
	}

	holding.TokenCount += count
	return holding, nil
}

func (r *InMemoryTokenizationRepository) ListPools() ([]*core.FractionalPool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*core.FractionalPool
	for _, p := range r.pools {
		result = append(result, p)
	}
	if result == nil {
		return []*core.FractionalPool{}, nil
	}
	return result, nil
}

type TokenizationRepositoryTest interface {
	InMemoryTokenizationRepository
}
