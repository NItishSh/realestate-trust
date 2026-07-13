package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Transaction & Escrow Manager API on :8080...")

	repo := db.NewInMemoryTransactionRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo transactions (APP_ENV != production)...")
		db.SeedTransactions(repo)
	}

	handler := db.NewTransactionHandler(repo)

	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/transactions", db.JWTAuth(http.HandlerFunc(handler.CreateTransaction)))
	mux.Handle("GET /api/v1/transactions", db.JWTAuth(http.HandlerFunc(handler.GetTransactions)))
	mux.Handle("GET /api/v1/transactions/{id}", db.JWTAuth(http.HandlerFunc(handler.GetTransaction)))
	mux.Handle("PUT /api/v1/transactions/{id}/status", db.JWTAuth(http.HandlerFunc(handler.UpdateStatus)))
	mux.Handle("POST /api/v1/transactions/{id}/escrow/fund", db.JWTAuth(http.HandlerFunc(handler.FundEscrow)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8080",
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
