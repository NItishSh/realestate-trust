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

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	slog.Info("Starting Immutable Audit Ledger API on :8084...")

	repo := db.NewInMemoryLedgerRepository()

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo ledger (APP_ENV != production)...")
		db.SeedLedger(repo)
	}

	handler := db.NewLedgerHandler(repo)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Security and global middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
	}))
	e.Use(middleware.BodyLimit("1M"))

	e.GET("/api/v1/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	api := e.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.POST("/logs", handler.WriteLog)
	api.GET("/logs", handler.GetLogs)
	api.GET("/logs/:index", handler.GetLog)

	srv := &http.Server{
		Addr:         ":8084",
		Handler:      e,
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
