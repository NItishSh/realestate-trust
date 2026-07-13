package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Immutable Audit Ledger API on :8084...")

	repo := db.NewInMemoryLedgerRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo ledger (APP_ENV != production)...")
		db.SeedLedger(repo)
	}

	handler := db.NewLedgerHandler(repo)

	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/logs", db.JWTAuth(http.HandlerFunc(handler.WriteLog)))
	mux.Handle("GET /api/v1/logs", db.JWTAuth(http.HandlerFunc(handler.GetLogs)))
	mux.Handle("GET /api/v1/logs/{index}", db.JWTAuth(http.HandlerFunc(handler.GetLog)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8084",
		Handler:      db.Chain(mux, db.EnableCORS, db.SecurityHeaders, db.MaxBodySize(1<<20)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	fmt.Println("🔒 Security hardening: timeouts, headers, 1MB body limit enabled")
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
