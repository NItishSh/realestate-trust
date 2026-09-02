package main

import (
	"context"
	"database/sql"
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
	"github.com/realestate-trust/monorepo/internal/core"
	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	slog.SetDefault(slog.New(&db.SlogCorrelationHandler{
		Handler: slog.NewJSONHandler(os.Stdout, nil),
	}))

	cfg, err := core.LoadServiceConfig("financing-engine", ":8082")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	var repo db.FinancingRepository
	var dbPool *db.DB
	if cfg.DatabaseURL != "" {
		slog.Info("Connecting to database...", "url", cfg.DatabaseURL)
		pool, err := db.Connect()
		if err != nil {
			slog.Error("Database connection failed", "err", err)
			os.Exit(1)
		}
		dbPool = pool
		defer func() { _ = dbPool.Close() }()
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
		AllowOrigins: cfg.CorsAllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Correlation-ID"},
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(middleware.BodyLimit(1 << 20))

	// Register Kubernetes Liveness, Readiness, and Health endpoints
	var sqlDB = (*sql.DB)(nil)
	if dbPool != nil {
		sqlDB = dbPool.SQL
	}
	db.RegisterHealthEndpoints(e, sqlDB, nil)

	// Webhooks can remain public
	e.POST("/api/v1/loans/webhooks/bank", handler.BankWebhook)

	api := e.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:    db.GetJWTKeyFunc(),
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.POST("/loans", handler.ApplyLoan, db.RBACMiddleware(core.Buyer, core.Admin))
	api.GET("/loans", handler.GetLoans, db.RBACMiddleware(core.Officer, core.Admin))
	api.GET("/loans/:id", handler.GetLoan)
	api.POST("/loans/:id/disburse", handler.DisburseLoan, db.RBACMiddleware(core.Officer, core.Admin))

	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      e,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
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
