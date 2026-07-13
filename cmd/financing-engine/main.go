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
	slog.Info("Starting Embedded Financing Engine API on :8082...")

	repo := db.NewInMemoryFinancingRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo loans (APP_ENV != production)...")
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
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8082",
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
