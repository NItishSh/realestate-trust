package db

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/realestate-trust/monorepo/internal/core"
)

type SQLLedgerRepository struct {
	db *sql.DB
	mu sync.Mutex // Mutex to serialize WriteLog operations and prevent race conditions on hash calculations
}

func NewSQLLedgerRepository(db *sql.DB) *SQLLedgerRepository {
	return &SQLLedgerRepository{db: db}
}

func (r *SQLLedgerRepository) WriteLog(eventID, payload string) (*core.AuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Acquire transactional advisory lock to serialize distributed concurrent appends (including empty table bootstrap).
	// This lock automatically releases on tx.Commit() / tx.Rollback() and allows concurrent read-only queries.
	_, err = tx.Exec(`SELECT pg_advisory_xact_lock(73849102)`)
	if err != nil {
		return nil, err
	}

	// Query last entry to find the previous hash and current index
	var lastIndex int64
	var lastHash string

	queryLast := `SELECT log_index, hash FROM ledger_entries ORDER BY log_index DESC LIMIT 1`
	err = tx.QueryRow(queryLast).Scan(&lastIndex, &lastHash)

	var index int64
	var prevHash string

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			index = 0
			prevHash = ""
		} else {
			return nil, err
		}
	} else {
		index = lastIndex + 1
		prevHash = lastHash
	}

	entry := &core.AuditEntry{
		Index:        index,
		Timestamp:    time.Now().UTC(),
		Payload:      payload,
		PreviousHash: prevHash,
		EventID:      eventID,
	}
	entry.Hash = entry.CalculateHash()

	var eventIDVal sql.NullString
	if eventID != "" {
		eventIDVal = sql.NullString{String: eventID, Valid: true}
	}

	insertQuery := `INSERT INTO ledger_entries (log_index, timestamp, payload, previous_hash, hash, event_id)
	                VALUES ($1, $2, $3, $4, $5, $6)
	                RETURNING timestamp`

	err = tx.QueryRow(insertQuery, entry.Index, entry.Timestamp, entry.Payload, entry.PreviousHash, entry.Hash, eventIDVal).Scan(&entry.Timestamp)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entry, nil
}

func (r *SQLLedgerRepository) GetLog(index int64) (*core.AuditEntry, error) {
	query := `SELECT log_index, timestamp, payload, previous_hash, hash, COALESCE(event_id, '') FROM ledger_entries WHERE log_index = $1`
	entry := &core.AuditEntry{}
	err := r.db.QueryRow(query, index).Scan(
		&entry.Index,
		&entry.Timestamp,
		&entry.Payload,
		&entry.PreviousHash,
		&entry.Hash,
		&entry.EventID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("audit log entry not found")
		}
		return nil, err
	}
	return entry, nil
}

func (r *SQLLedgerRepository) GetChainLength() int64 {
	query := `SELECT COUNT(*) FROM ledger_entries`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (r *SQLLedgerRepository) ListLogs() ([]*core.AuditEntry, error) {
	query := `SELECT log_index, timestamp, payload, previous_hash, hash, COALESCE(event_id, '') FROM ledger_entries ORDER BY log_index ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*core.AuditEntry
	for rows.Next() {
		entry := &core.AuditEntry{}
		err := rows.Scan(
			&entry.Index,
			&entry.Timestamp,
			&entry.Payload,
			&entry.PreviousHash,
			&entry.Hash,
			&entry.EventID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*core.AuditEntry{}, nil
	}
	return result, nil
}
