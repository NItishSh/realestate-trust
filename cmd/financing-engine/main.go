package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Embedded Financing Engine API on :8082...")

	repo := db.NewInMemoryFinancingRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo loans (APP_ENV != production)...")
		db.SeedLoans(repo)
	}

	handler := db.NewFinancingHandler(repo)

	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/loans", db.JWTAuth(http.HandlerFunc(handler.ApplyLoan)))
	mux.Handle("GET /api/v1/loans", db.JWTAuth(http.HandlerFunc(handler.GetLoans)))
	mux.Handle("GET /api/v1/loans/{id}", db.JWTAuth(http.HandlerFunc(handler.GetLoan)))
	mux.Handle("POST /api/v1/loans/{id}/disburse", db.JWTAuth(http.HandlerFunc(handler.DisburseLoan)))

	// Webhooks can remain public or use a different auth scheme, but for simplicity here we keep it public
	mux.HandleFunc("POST /api/v1/loans/webhooks/bank", handler.BankWebhook)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8082",
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
