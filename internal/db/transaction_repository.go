package db

import (
	"sync"
	"time"

	"github.com/realestate-trust/monorepo/internal/core"
)

type Transaction struct {
	ID                   string                `json:"id"`
	PropertyID           string                `json:"propertyId"`
	BuyerID              string                `json:"buyerId"`
	SellerID             string                `json:"sellerId"`
	TotalAmount          float64               `json:"totalAmount"`
	Status               core.TransactionState `json:"status"`
	VirtualAccountNumber string                `json:"virtualAccountNumber"`
	CreatedAt            time.Time             `json:"createdAt"`
}

type EscrowAccount struct {
	ID                   string    `json:"id"`
	TransactionID        string    `json:"transactionId"`
	VirtualAccountNumber string    `json:"virtualAccountNumber"`
	BankPartner          string    `json:"bankPartner"`
	Balance              float64   `json:"balance"`
	CreatedAt            time.Time `json:"createdAt"`
}

// TransactionRepository interface defines actions for transactions and escrow accounts.
type TransactionRepository interface {
	CreateTransaction(propertyID, buyerID, sellerID string, totalAmount float64) (*Transaction, error)
	GetTransaction(id string) (*Transaction, error)
	GetEscrow(txID string) (*EscrowAccount, error)
	UpdateTransactionStatus(id string, status core.TransactionState) error
	FundEscrow(id string, amount float64) error
	ListTransactions() ([]*Transaction, error)
}

// InMemoryTransactionRepository implements TransactionRepository.
type InMemoryTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	accounts     map[string]*EscrowAccount
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]*Transaction),
		accounts:     make(map[string]*EscrowAccount),
	}
}

func (r *InMemoryTransactionRepository) CreateTransaction(propertyID, buyerID, sellerID string, totalAmount float64) (*Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "tx-" + propertyID
	va := "VA-YES-" + propertyID
	tx := &Transaction{
		ID:                   id,
		PropertyID:           propertyID,
		BuyerID:              buyerID,
		SellerID:             sellerID,
		TotalAmount:          totalAmount,
		Status:               core.Draft,
		VirtualAccountNumber: va,
		CreatedAt:            time.Now(),
	}

	acc := &EscrowAccount{
		ID:                   "acc-" + propertyID,
		TransactionID:        id,
		VirtualAccountNumber: va,
		BankPartner:          "YES_BANK",
		Balance:              0.0,
		CreatedAt:            time.Now(),
	}

	r.transactions[id] = tx
	r.accounts[id] = acc
	return tx, nil
}

func (r *InMemoryTransactionRepository) GetTransaction(id string) (*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.transactions[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return tx, nil
}

func (r *InMemoryTransactionRepository) GetEscrow(txID string) (*EscrowAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	acc, ok := r.accounts[txID]
	if !ok {
		return nil, core.ErrNotFound
	}
	return acc, nil
}

func (r *InMemoryTransactionRepository) UpdateTransactionStatus(id string, status core.TransactionState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, ok := r.transactions[id]
	if !ok {
		return core.ErrNotFound
	}

	next, err := core.Transition(tx.Status, status)
	if err != nil {
		return err
	}

	tx.Status = next
	return nil
}

func (r *InMemoryTransactionRepository) FundEscrow(id string, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, ok := r.transactions[id]
	if !ok {
		return core.ErrNotFound
	}

	acc, ok := r.accounts[id]
	if !ok {
		return core.ErrNotFound
	}

	acc.Balance += amount

	// Automatically advance status to FUNDED if balance matches or exceeds total transaction amount
	if acc.Balance >= tx.TotalAmount && tx.Status == core.Escrow {
		tx.Status = core.Funded
	}

	return nil
}

func (r *InMemoryTransactionRepository) ListTransactions() ([]*Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Transaction
	for _, t := range r.transactions {
		result = append(result, t)
	}
	if result == nil {
		return []*Transaction{}, nil
	}
	return result, nil
}
