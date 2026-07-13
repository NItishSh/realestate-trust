package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting User & Identity Service API on :8081...")

	// Initialize In-Memory Repository (default for local validation / dev)
	repo := db.NewInMemoryUserRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo users (APP_ENV != production)...")
		db.SeedUsers(repo)
	}

	handler := db.NewUserHandler(repo)

	mux := http.NewServeMux()

	// Endpoints configured with native Go 1.22+ routing features
	// Public routes
	mux.HandleFunc("POST /api/v1/users", handler.RegisterUser)
	mux.HandleFunc("POST /api/v1/login", handler.Login)

	// Protected routes
	mux.Handle("GET /api/v1/users", db.JWTAuth(http.HandlerFunc(handler.GetUsers)))
	mux.Handle("GET /api/v1/users/{id}", db.JWTAuth(http.HandlerFunc(handler.GetUser)))
	mux.Handle("POST /api/v1/users/{id}/kyc", db.JWTAuth(http.HandlerFunc(handler.SubmitKYC)))
	mux.Handle("GET /api/v1/users/{id}/kyc/status", db.JWTAuth(http.HandlerFunc(handler.GetKYCStatus)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8081",
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
