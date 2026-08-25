package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SQLFinancingRepository struct {
	db *sql.DB
}

func NewSQLFinancingRepository(db *sql.DB) *SQLFinancingRepository {
	return &SQLFinancingRepository{db: db}
}

func (r *SQLFinancingRepository) CreateLoan(txID, userID string, reqAmount float64) (*Loan, error) {
	id := "loan-" + txID
	query := `INSERT INTO loans (id, transaction_id, user_id, requested_amount, status)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, transaction_id, user_id, requested_amount, approved_amount, status, created_at`

	loan := &Loan{}
	err := r.db.QueryRow(query, id, txID, userID, reqAmount, "APPLIED").Scan(
		&loan.ID,
		&loan.TransactionID,
		&loan.UserID,
		&loan.RequestedAmount,
		&loan.ApprovedAmount,
		&loan.Status,
		&loan.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return loan, nil
}

func (r *SQLFinancingRepository) GetLoan(id string) (*Loan, error) {
	query := `SELECT id, transaction_id, user_id, requested_amount, approved_amount, status, created_at FROM loans WHERE id = $1`
	loan := &Loan{}
	err := r.db.QueryRow(query, id).Scan(
		&loan.ID,
		&loan.TransactionID,
		&loan.UserID,
		&loan.RequestedAmount,
		&loan.ApprovedAmount,
		&loan.Status,
		&loan.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("loan not found")
		}
		return nil, err
	}
	return loan, nil
}

func (r *SQLFinancingRepository) ApproveLoan(id string, appAmount float64) error {
	query := `UPDATE loans SET approved_amount = $1, status = $2, updated_at = NOW() WHERE id = $3`
	res, err := r.db.Exec(query, appAmount, "APPROVED", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("loan not found")
	}
	return nil
}

func (r *SQLFinancingRepository) CreateDisbursement(loanID, vaNumber string, amount float64) (*Disbursement, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Update loan status to DISBURSED
	updateQuery := `UPDATE loans SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(updateQuery, "DISBURSED", loanID)
	if err != nil {
		return nil, err
	}

	// Insert Disbursement record
	id := "disb-" + uuid.NewString()[:8]
	txRef := "TXREF-" + uuid.NewString()[:8]

	insertQuery := `INSERT INTO disbursements (id, loan_id, virtual_account_number, disbursed_amount, transaction_reference, status, disbursed_at)
	                VALUES ($1, $2, $3, $4, $5, $6, $7)
	                RETURNING id, loan_id, virtual_account_number, disbursed_amount, transaction_reference, status, disbursed_at`

	disb := &Disbursement{}
	err = tx.QueryRow(insertQuery, id, loanID, vaNumber, amount, txRef, "COMPLETED", time.Now()).Scan(
		&disb.ID,
		&disb.LoanID,
		&disb.VirtualAccountNumber,
		&disb.DisbursedAmount,
		&disb.TransactionReference,
		&disb.Status,
		&disb.DisbursedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return disb, nil
}

func (r *SQLFinancingRepository) ListLoans() ([]*Loan, error) {
	query := `SELECT id, transaction_id, user_id, requested_amount, approved_amount, status, created_at FROM loans`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*Loan
	for rows.Next() {
		loan := &Loan{}
		err := rows.Scan(
			&loan.ID,
			&loan.TransactionID,
			&loan.UserID,
			&loan.RequestedAmount,
			&loan.ApprovedAmount,
			&loan.Status,
			&loan.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, loan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*Loan{}, nil
	}
	return result, nil
}
