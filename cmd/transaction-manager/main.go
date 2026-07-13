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
	slog.Info("Starting Transaction & Escrow Manager API on :8080...")

	repo := db.NewInMemoryTransactionRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo transactions (APP_ENV != production)...")
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
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8080",
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
