package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	slog.Info("Starting User & Identity Service API on :8081...")

	// Initialize In-Memory Repository (default for local validation / dev)
	repo := db.NewInMemoryUserRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo users (APP_ENV != production)...")
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
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      db.Chain(mux, db.EnableCORS, db.SecurityHeaders, db.MaxBodySize(1<<20)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	slog.Info("🔒 Security hardening: timeouts, headers, 1MB body limit enabled")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("Server stopped")
}
