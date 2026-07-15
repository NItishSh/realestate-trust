package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
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
		"../../cmd/financing-engine/db/migrations/000001_init_financing.up.sql",
		"../../cmd/feedback-service/db/migrations/000001_create_feedback.up.sql",
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
	tables := []string{"users", "kyc_verifications", "transactions", "escrow_accounts", "loans", "disbursements", "feedback"}
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
