package main

import (
	"fmt"
	"net/http"

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

	mux.HandleFunc("POST /api/v1/transactions", handler.CreateTransaction)
	mux.HandleFunc("GET /api/v1/transactions", handler.GetTransactions)
	mux.HandleFunc("GET /api/v1/transactions/{id}", handler.GetTransaction)
	mux.HandleFunc("PUT /api/v1/transactions/{id}/status", handler.UpdateStatus)
	mux.HandleFunc("POST /api/v1/transactions/{id}/escrow/fund", handler.FundEscrow)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8080", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
