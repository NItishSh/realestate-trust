package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Fractional Tokenization Engine API on :8083...")

	repo := db.NewInMemoryTokenizationRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo pools (APP_ENV != production)...")
		db.SeedPools(repo)
	}

	handler := db.NewTokenizationHandler(repo)

	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/pools", db.JWTAuth(http.HandlerFunc(handler.CreatePool)))
	mux.Handle("GET /api/v1/pools", db.JWTAuth(http.HandlerFunc(handler.GetPools)))
	mux.Handle("POST /api/v1/pools/{id}/buy", db.JWTAuth(http.HandlerFunc(handler.BuyShares)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8083",
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
