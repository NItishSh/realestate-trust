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

	echojwt "github.com/labstack/echo-jwt/v5"
	echo "github.com/labstack/echo/v5"
	middleware "github.com/labstack/echo/v5/middleware"
	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	slog.SetDefault(slog.New(&db.SlogCorrelationHandler{
		Handler: slog.NewJSONHandler(os.Stdout, nil),
	}))

	var repo db.FinancingRepository
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		slog.Info("Connecting to database...", "url", dbURL)
		dbPool, err := db.Connect()
		if err != nil {
			slog.Error("Database connection failed", "err", err)
			os.Exit(1)
		}
		defer dbPool.Close()
		repo = db.NewSQLFinancingRepository(dbPool.SQL)
	} else {
		slog.Info("DATABASE_URL is empty. Falling back to InMemoryFinancingRepository.")
		repo = db.NewInMemoryFinancingRepository()
	}

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo loans (APP_ENV != production)...")
		db.SeedLoans(repo)
	}

	handler := db.NewFinancingHandler(repo)

	e := echo.New()
	// Security and global middlewares
	e.Use(db.CorrelationIDMiddleware())
	e.Use(db.RequestLoggerMiddleware())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Correlation-ID"},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
	}))
	e.Use(middleware.BodyLimit(1 << 20))

	e.GET("/api/v1/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	// Webhooks can remain public
	e.POST("/api/v1/loans/webhooks/bank", handler.BankWebhook)

	api := e.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.POST("/loans", handler.ApplyLoan)
	api.GET("/loans", handler.GetLoans)
	api.GET("/loans/:id", handler.GetLoan)
	api.POST("/loans/:id/disburse", handler.DisburseLoan)

	srv := &http.Server{
		Addr:         ":8082",
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
