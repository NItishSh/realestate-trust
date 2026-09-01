package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	_ "github.com/lib/pq"
)

func TestLiveDatabaseMigrations(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set; skipping live database migration integration tests")
	}

	// Connect to PG
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Verify connection is alive
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// SQL migration files to load sequentially
	// Note: We use relative paths from internal/db context
	migrationPaths := []string{
		"../../cmd/identity-service/db/migrations/000001_init_user.up.sql",
		"../../cmd/transaction-manager/db/migrations/000001_init_escrow.up.sql",
		"../../cmd/transaction-manager/db/migrations/000002_create_outbox.up.sql",
		"../../cmd/financing-engine/db/migrations/000001_init_financing.up.sql",
		"../../cmd/feedback-service/db/migrations/000001_create_feedback.up.sql",
		"../../cmd/ledger-service/db/migrations/000001_init_ledger.up.sql",
		"../../cmd/property-registry-service/db/migrations/000001_init_property.up.sql",
		"../../cmd/tokenization-engine/db/migrations/000001_init_tokenization.up.sql",
	}

	for _, path := range migrationPaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			t.Fatalf("failed to resolve path %s: %v", path, err)
		}

		sqlBytes, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", path, err)
		}

		t.Logf("Running migration schema: %s", filepath.Base(path))
		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("failed to execute migration file %s: %v", path, err)
			}
		}
	}

	// Verify tables were created by running queries
	tables := []string{
		"users", "kyc_verifications", "refresh_token_sessions",
		"transactions", "escrow_accounts", "outbox_events",
		"loans", "disbursements", "feedback",
		"ledger_entries", "properties", "fractional_pools", "fractional_holdings",
	}
	for _, table := range tables {
		query := "SELECT count(*) FROM " + table
		var count int
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			t.Errorf("failed to query table %s: %v", table, err)
		} else {
			t.Logf("Table %s verified. Row count: %d", table, count)
		}
	}
}

func TestSQLLedgerRepository_Idempotency(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set; skipping live integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSQLLedgerRepository(db)

	// Clean up table first to avoid any leftovers from other tests
	_, err = db.Exec("DELETE FROM ledger_entries")
	if err != nil {
		t.Fatalf("failed to clean up table: %v", err)
	}

	// 1. First write should succeed
	entry1, err := repo.WriteLog("event-123", "Payload 1")
	if err != nil {
		t.Fatalf("failed to write log: %v", err)
	}
	if entry1.EventID != "event-123" {
		t.Errorf("expected eventID to be event-123, got %s", entry1.EventID)
	}

	// 2. Second write with the same event_id should fail
	_, err = repo.WriteLog("event-123", "Payload 2")
	if err == nil {
		t.Fatalf("expected unique constraint failure, got nil error")
	}
}

func TestSQLLedgerRepository_ConcurrentDistributedWrites(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set; skipping live integration tests")
	}

	mainDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = mainDB.Close() }()

	// Clean up table first
	_, err = mainDB.Exec("DELETE FROM ledger_entries")
	if err != nil {
		t.Fatalf("failed to clean up table: %v", err)
	}

	numPods := 10
	entriesPerPod := 5
	totalEntries := numPods * entriesPerPod

	var wg sync.WaitGroup
	errChan := make(chan error, totalEntries)

	for p := 0; p < numPods; p++ {
		podID := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each simulated pod has its own distinct DB connection pool / repo instance
			podDB, connErr := sql.Open("postgres", dbURL)
			if connErr != nil {
				errChan <- connErr
				return
			}
			defer func() { _ = podDB.Close() }()

			repo := NewSQLLedgerRepository(podDB)
			for i := 0; i < entriesPerPod; i++ {
				eventID := "event-pod-" + strings.TrimSpace(string(rune('A'+podID))) + "-" + strings.TrimSpace(string(rune('0'+i)))
				payload := "Payload from pod " + string(rune('A'+podID)) + " item " + string(rune('0'+i))
				_, writeErr := repo.WriteLog(eventID, payload)
				if writeErr != nil {
					errChan <- writeErr
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			t.Fatalf("concurrent write failed: %v", err)
		}
	}

	// Validate the entire chain from main repository
	repo := NewSQLLedgerRepository(mainDB)
	logs, err := repo.ListLogs()
	if err != nil {
		t.Fatalf("failed to list logs: %v", err)
	}

	if len(logs) != totalEntries {
		t.Fatalf("expected %d entries, got %d", totalEntries, len(logs))
	}

	// Validate chain continuity and cryptographic integrity
	for i, entry := range logs {
		if entry.Index != int64(i) {
			t.Errorf("entry index mismatch at pos %d: expected index %d, got %d", i, i, entry.Index)
		}
		if i == 0 {
			if entry.PreviousHash != "" {
				t.Errorf("entry 0 must have empty PreviousHash, got: %s", entry.PreviousHash)
			}
		} else {
			prevEntry := logs[i-1]
			if entry.PreviousHash != prevEntry.Hash {
				t.Errorf("hash chain broken between index %d and %d: expected %s, got %s", prevEntry.Index, entry.Index, prevEntry.Hash, entry.PreviousHash)
			}
		}

		// Verify recomputed hash matches stored hash
		recomputed := entry.CalculateHash()
		if recomputed != entry.Hash {
			t.Errorf("hash mismatch at index %d: stored=%s, recomputed=%s", entry.Index, entry.Hash, recomputed)
		}
	}
}

func TestSQLTransactionRepository_OutboxAtomicity(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set; skipping live integration tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSQLTransactionRepository(db)

	// Clean up
	_, _ = db.Exec("DELETE FROM outbox_events")
	_, _ = db.Exec("DELETE FROM escrow_accounts")
	_, _ = db.Exec("DELETE FROM transactions")

	// 1. CreateTransaction should atomically create transaction + escrow + outbox_event
	tx, err := repo.CreateTransaction("prop-outbox-01", "buyer-99", "seller-99", 8500000.0)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if tx.ID == "" {
		t.Errorf("expected non-empty transaction ID")
	}

	var outboxCount int
	var eventType, status, payload string
	err = db.QueryRow("SELECT event_type, status, payload FROM outbox_events WHERE aggregate_id = $1", tx.ID).Scan(&eventType, &status, &payload)
	if err != nil {
		t.Fatalf("expected outbox event to be created atomically in DB, got: %v", err)
	}
	if eventType != "com.realestatetrust.transaction.created.v1" {
		t.Errorf("expected event_type com.realestatetrust.transaction.created.v1, got %s", eventType)
	}
	if status != "PENDING" {
		t.Errorf("expected outbox status PENDING, got %s", status)
	}
	if !strings.Contains(payload, "prop-outbox-01") {
		t.Errorf("expected payload to contain prop-outbox-01, got: %s", payload)
	}

	// 2. FundEscrow should atomically update escrow balance and create an escrow.funded outbox event
	err = repo.FundEscrow(tx.ID, 8500000.0)
	if err != nil {
		t.Fatalf("failed to fund escrow: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1", tx.ID).Scan(&outboxCount)
	if err != nil {
		t.Fatalf("failed to count outbox events: %v", err)
	}
	if outboxCount != 2 {
		t.Errorf("expected 2 outbox events, got %d", outboxCount)
	}
}
