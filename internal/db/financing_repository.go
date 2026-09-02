package db

import (
	"sync"
	"time"

	"github.com/realestate-trust/monorepo/internal/core"
)

type Loan struct {
	ID              string    `json:"id"`
	TransactionID   string    `json:"transactionId"`
	UserID          string    `json:"userId"`
	RequestedAmount float64   `json:"requestedAmount"`
	ApprovedAmount  *float64  `json:"approvedAmount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Disbursement struct {
	ID                   string    `json:"id"`
	LoanID               string    `json:"loanId"`
	VirtualAccountNumber string    `json:"virtualAccountNumber"`
	DisbursedAmount      float64   `json:"disbursedAmount"`
	TransactionReference string    `json:"transactionReference"`
	Status               string    `json:"status"`
	DisbursedAt          time.Time `json:"disbursedAt"`
}

// FinancingRepository interface defines database actions for loans.
type FinancingRepository interface {
	CreateLoan(txID, userID string, reqAmount float64) (*Loan, error)
	GetLoan(id string) (*Loan, error)
	ApproveLoan(id string, appAmount float64) error
	CreateDisbursement(loanID, vaNumber string, amount float64) (*Disbursement, error)
	ListLoans() ([]*Loan, error)
}

// InMemoryFinancingRepository implements FinancingRepository.
type InMemoryFinancingRepository struct {
	mu            sync.RWMutex
	loans         map[string]*Loan
	disbursements map[string]*Disbursement
}

func NewInMemoryFinancingRepository() *InMemoryFinancingRepository {
	return &InMemoryFinancingRepository{
		loans:         make(map[string]*Loan),
		disbursements: make(map[string]*Disbursement),
	}
}

func (r *InMemoryFinancingRepository) CreateLoan(txID, userID string, reqAmount float64) (*Loan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "loan-" + txID
	loan := &Loan{
		ID:              id,
		TransactionID:   txID,
		UserID:          userID,
		RequestedAmount: reqAmount,
		Status:          "APPLIED",
		CreatedAt:       time.Now(),
	}
	r.loans[id] = loan
	return loan, nil
}

func (r *InMemoryFinancingRepository) GetLoan(id string) (*Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loan, ok := r.loans[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return loan, nil
}

func (r *InMemoryFinancingRepository) ApproveLoan(id string, appAmount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	loan, ok := r.loans[id]
	if !ok {
		return core.ErrNotFound
	}

	loan.ApprovedAmount = &appAmount
	loan.Status = "APPROVED"
	return nil
}

func (r *InMemoryFinancingRepository) CreateDisbursement(loanID, vaNumber string, amount float64) (*Disbursement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := "disb-" + loanID
	disb := &Disbursement{
		ID:                   id,
		LoanID:               loanID,
		VirtualAccountNumber: vaNumber,
		DisbursedAmount:      amount,
		TransactionReference: "REF-" + loanID,
		Status:               "COMPLETED",
		DisbursedAt:          time.Now(),
	}
	r.disbursements[id] = disb

	if loan, ok := r.loans[loanID]; ok {
		loan.Status = "DISBURSED"
	}

	return disb, nil
}

func (r *InMemoryFinancingRepository) ListLoans() ([]*Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Loan
	for _, l := range r.loans {
		result = append(result, l)
	}
	if result == nil {
		return []*Loan{}, nil
	}
	return result, nil
}
