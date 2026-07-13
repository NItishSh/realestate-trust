package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Property Registry Service API on :8085...")

	repo := db.NewInMemoryPropertyRepository()

	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo properties (APP_ENV != production)...")
		db.SeedProperties(repo)
	}

	handler := db.NewPropertyHandler(repo)
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/properties", db.JWTAuth(http.HandlerFunc(handler.ListProperties)))
	mux.Handle("GET /api/v1/properties/{id}", db.JWTAuth(http.HandlerFunc(handler.GetProperty)))
	mux.Handle("POST /api/v1/properties/{id}/documents/unlock", db.JWTAuth(http.HandlerFunc(handler.UnlockDocuments)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8085",
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
