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

	cfg, err := core.LoadServiceConfig("identity-service", ":8081")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	var repo db.UserRepository
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
		repo = db.NewSQLUserRepository(dbPool.SQL)
	} else {
		slog.Info("DATABASE_URL is empty. Falling back to InMemoryUserRepository.")
		repo = db.NewInMemoryUserRepository()
	}

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo users (APP_ENV != production)...")
		db.SeedUsers(repo)
	}

	handler := db.NewUserHandler(repo)

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

	api := e.Group("/api/v1")

	// Public routes with rate limiting
	api.POST("/users", handler.RegisterUser, db.AuthRateLimiterMiddleware())
	api.POST("/login", handler.Login, db.AuthRateLimiterMiddleware())
	api.POST("/refresh", handler.RefreshToken)
	api.POST("/logout", handler.Logout)

	// Protected routes
	protected := api.Group("")
	protected.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:    db.GetJWTKeyFunc(),
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	protected.GET("/users", handler.GetUsers, db.RBACMiddleware(core.Officer, core.Admin))
	protected.GET("/users/:id", handler.GetUser)
	protected.POST("/users/:id/kyc", handler.SubmitKYC)
	protected.GET("/users/:id/kyc/status", handler.GetKYCStatus)
	protected.DELETE("/users/:id", handler.DeleteUser)

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
