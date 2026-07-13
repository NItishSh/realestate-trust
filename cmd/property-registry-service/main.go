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
	slog.Info("Starting Property Registry Service API on :8085...")

	repo := db.NewInMemoryPropertyRepository()

	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo properties (APP_ENV != production)...")
		db.SeedProperties(repo)
	}

	handler := db.NewPropertyHandler(repo)
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/properties", db.JWTAuth(http.HandlerFunc(handler.ListProperties)))
	mux.Handle("GET /api/v1/properties/{id}", db.JWTAuth(http.HandlerFunc(handler.GetProperty)))
	mux.Handle("POST /api/v1/properties/{id}/documents/unlock", db.JWTAuth(http.HandlerFunc(handler.UnlockDocuments)))

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	srv := &http.Server{
		Addr:         ":8085",
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
