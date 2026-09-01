package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

// DB represents the database connection pool wrapper.
type DB struct {
	SQL *sql.DB
}

// Connect establishes connection pool with PostgreSQL.
func Connect() (*DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Return empty wrapper to support In-Memory Mock mode if no DB string is provided
		return &DB{SQL: nil}, nil
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Production-grade connection pool bounds
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{SQL: db}, nil
}

// Close terminates connection pool.
func (d *DB) Close() error {
	if d.SQL != nil {
		return d.SQL.Close()
	}
	return nil
}
