package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realestate-trust/monorepo/internal/core"
	"github.com/realestate-trust/monorepo/internal/events"
)

type SQLTransactionRepository struct {
	db *sql.DB
}

func NewSQLTransactionRepository(db *sql.DB) *SQLTransactionRepository {
	return &SQLTransactionRepository{db: db}
}

func insertOutboxEvent(tx *sql.Tx, aggregateType, aggregateID, eventType string, payload any) error {
	eventID := uuid.NewString()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status)
	          VALUES ($1, $2, $3, $4, $5, 'PENDING')`
	_, err = tx.Exec(query, eventID, aggregateType, aggregateID, eventType, string(payloadBytes))
	return err
}

func toUUID(s string) string {
	if _, err := uuid.Parse(s); err == nil {
		return s
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s)).String()
}

func (r *SQLTransactionRepository) CreateTransaction(propertyID, buyerID, sellerID string, totalAmount float64) (*Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.NewString()
	propUUID := toUUID(propertyID)
	buyerUUID := toUUID(buyerID)
	sellerUUID := toUUID(sellerID)
	va := "VA-YES-" + id[:8]
	accID := uuid.NewString()

	// Insert Transaction
	txQuery := `INSERT INTO transactions (id, property_id, buyer_id, seller_id, total_amount, status)
	            VALUES ($1, $2, $3, $4, $5, $6)
	            RETURNING id, property_id, buyer_id, seller_id, total_amount, status, created_at`

	transaction := &Transaction{}
	var statusStr string
	err = tx.QueryRow(txQuery, id, propUUID, buyerUUID, sellerUUID, totalAmount, string(core.Draft)).Scan(
		&transaction.ID,
		&transaction.PropertyID,
		&transaction.BuyerID,
		&transaction.SellerID,
		&transaction.TotalAmount,
		&statusStr,
		&transaction.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	transaction.Status = core.TransactionState(statusStr)
	transaction.VirtualAccountNumber = va

	// Insert Escrow Account
	accQuery := `INSERT INTO escrow_accounts (id, transaction_id, virtual_account_number, bank_partner, balance)
	             VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(accQuery, accID, id, va, "YES_BANK", 0.0)
	if err != nil {
		return nil, err
	}

	// Insert CloudEvent in Outbox Table within same ACID transaction
	cloudEvent := events.CloudEvent[events.TransactionCreatedPayload]{
		SpecVersion: "1.0",
		ID:          uuid.NewString(),
		Source:      "realestate-trust/transaction-manager",
		Type:        "com.realestatetrust.transaction.created.v1",
		Time:        time.Now().UTC(),
		Data: events.TransactionCreatedPayload{
			TransactionID: id,
			PropertyID:    propertyID,
			BuyerID:       buyerID,
			SellerID:      sellerID,
			Amount:        totalAmount,
			Currency:      "INR",
		},
	}
	_ = insertOutboxEvent(tx, "transaction", id, "com.realestatetrust.transaction.created.v1", cloudEvent)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (r *SQLTransactionRepository) GetTransaction(id string) (*Transaction, error) {
	query := `SELECT t.id, t.property_id, t.buyer_id, t.seller_id, t.total_amount, t.status, t.created_at, COALESCE(e.virtual_account_number, '')
	          FROM transactions t
	          LEFT JOIN escrow_accounts e ON e.transaction_id = t.id
	          WHERE t.id = $1`

	transaction := &Transaction{}
	var statusStr string
	err := r.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.PropertyID,
		&transaction.BuyerID,
		&transaction.SellerID,
		&transaction.TotalAmount,
		&statusStr,
		&transaction.CreatedAt,
		&transaction.VirtualAccountNumber,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	transaction.Status = core.TransactionState(statusStr)
	return transaction, nil
}

func (r *SQLTransactionRepository) GetEscrow(txID string) (*EscrowAccount, error) {
	query := `SELECT id, transaction_id, virtual_account_number, bank_partner, balance
	          FROM escrow_accounts
	          WHERE transaction_id = $1`

	acc := &EscrowAccount{}
	err := r.db.QueryRow(query, txID).Scan(
		&acc.ID,
		&acc.TransactionID,
		&acc.VirtualAccountNumber,
		&acc.BankPartner,
		&acc.Balance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("escrow account not found")
		}
		return nil, err
	}
	return acc, nil
}

func (r *SQLTransactionRepository) UpdateTransactionStatus(id string, status core.TransactionState) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	res, err := tx.Exec(query, string(status), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("transaction not found")
	}

	cloudEvent := events.CloudEvent[events.TransactionStatusUpdatedPayload]{
		SpecVersion: "1.0",
		ID:          uuid.NewString(),
		Source:      "realestate-trust/transaction-manager",
		Type:        "com.realestatetrust.transaction.updated.v1",
		Time:        time.Now().UTC(),
		Data: events.TransactionStatusUpdatedPayload{
			TransactionID: id,
			NewState:      string(status),
		},
	}
	_ = insertOutboxEvent(tx, "transaction", id, "com.realestatetrust.transaction.updated.v1", cloudEvent)

	return tx.Commit()
}

func (r *SQLTransactionRepository) FundEscrow(id string, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Update Escrow balance
	accQuery := `UPDATE escrow_accounts SET balance = balance + $1, updated_at = NOW() WHERE transaction_id = $2`
	res, err := tx.Exec(accQuery, amount, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("escrow account not found")
	}

	// Update Transaction status to FUNDED
	txQuery := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	res, err = tx.Exec(txQuery, string(core.Funded), id)
	if err != nil {
		return err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("transaction not found")
	}

	cloudEvent := events.CloudEvent[events.EscrowFundedPayload]{
		SpecVersion: "1.0",
		ID:          uuid.NewString(),
		Source:      "realestate-trust/transaction-manager",
		Type:        "com.realestatetrust.escrow.funded.v1",
		Time:        time.Now().UTC(),
		Data: events.EscrowFundedPayload{
			TransactionID: id,
			Amount:        amount,
			Currency:      "INR",
		},
	}
	_ = insertOutboxEvent(tx, "transaction", id, "com.realestatetrust.escrow.funded.v1", cloudEvent)

	return tx.Commit()
}

func (r *SQLTransactionRepository) ListTransactions() ([]*Transaction, error) {
	query := `SELECT t.id, t.property_id, t.buyer_id, t.seller_id, t.total_amount, t.status, t.created_at, COALESCE(e.virtual_account_number, '')
	          FROM transactions t
	          LEFT JOIN escrow_accounts e ON e.transaction_id = t.id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*Transaction
	for rows.Next() {
		transaction := &Transaction{}
		var statusStr string
		err := rows.Scan(
			&transaction.ID,
			&transaction.PropertyID,
			&transaction.BuyerID,
			&transaction.SellerID,
			&transaction.TotalAmount,
			&statusStr,
			&transaction.CreatedAt,
			&transaction.VirtualAccountNumber,
		)
		if err != nil {
			return nil, err
		}
		transaction.Status = core.TransactionState(statusStr)
		result = append(result, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*Transaction{}, nil
	}
	return result, nil
}
