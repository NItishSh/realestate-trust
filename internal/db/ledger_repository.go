package db

import (
	"errors"
	"sync"
	"time"

	"github.com/realestate-trust/monorepo/internal/core"
)

// LedgerRepository defines storage actions for auditing.
type LedgerRepository interface {
	WriteLog(payload string) (*core.AuditEntry, error)
	GetLog(index int64) (*core.AuditEntry, error)
	GetChainLength() int64
	ListLogs() ([]*core.AuditEntry, error)
}

// InMemoryLedgerRepository implements LedgerRepository.
type InMemoryLedgerRepository struct {
	mu    sync.RWMutex
	chain []*core.AuditEntry
}

func NewInMemoryLedgerRepository() *InMemoryLedgerRepository {
	return &InMemoryLedgerRepository{
		chain: make([]*core.AuditEntry, 0),
	}
}

func (r *InMemoryLedgerRepository) WriteLog(payload string) (*core.AuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var prevHash string
	index := int64(len(r.chain))
	if index > 0 {
		prevHash = r.chain[index-1].Hash
	}

	entry := &core.AuditEntry{
		Index:        index,
		Timestamp:    time.Now(),
		Payload:      payload,
		PreviousHash: prevHash,
	}
	entry.Hash = entry.CalculateHash()

	r.chain = append(r.chain, entry)
	return entry, nil
}

func (r *InMemoryLedgerRepository) GetLog(index int64) (*core.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if index < 0 || index >= int64(len(r.chain)) {
		return nil, errors.New("audit log entry not found")
	}
	return r.chain[index], nil
}

func (r *InMemoryLedgerRepository) GetChainLength() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return int64(len(r.chain))
}

func (r *InMemoryLedgerRepository) ListLogs() ([]*core.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid concurrency issues during serialization
	result := make([]*core.AuditEntry, len(r.chain))
	copy(result, r.chain)

	if result == nil {
		return []*core.AuditEntry{}, nil
	}
	return result, nil
}
