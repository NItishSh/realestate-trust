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
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	slog.SetDefault(slog.New(&db.SlogCorrelationHandler{
		Handler: slog.NewJSONHandler(os.Stdout, nil),
	}))

	slog.Info("Starting User & Identity Service API on :8081...")

	var repo db.UserRepository
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		slog.Info("Connecting to database...", "url", dbURL)
		dbPool, err := db.Connect()
		if err != nil {
			slog.Error("Database connection failed", "err", err)
			os.Exit(1)
		}
		defer dbPool.Close()
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

	// Health check
	e.GET("/api/v1/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	api := e.Group("/api/v1")

	// Public routes
	api.POST("/users", handler.RegisterUser)
	api.POST("/login", handler.Login)
	api.POST("/refresh", handler.RefreshToken)
	api.POST("/logout", handler.Logout)

	// Protected routes
	protected := api.Group("")
	protected.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	protected.GET("/users", handler.GetUsers)
	protected.GET("/users/:id", handler.GetUser)
	protected.POST("/users/:id/kyc", handler.SubmitKYC)
	protected.GET("/users/:id/kyc/status", handler.GetKYCStatus)

	srv := &http.Server{
		Addr:         ":8081",
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
