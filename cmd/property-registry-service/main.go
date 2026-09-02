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

	cfg, err := core.LoadServiceConfig("property-registry-service", ":8085")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	var repo db.PropertyRepository
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
		repo = db.NewSQLPropertyRepository(dbPool.SQL)
	} else {
		slog.Info("DATABASE_URL is empty. Falling back to InMemoryPropertyRepository.")
		repo = db.NewInMemoryPropertyRepository()
	}

	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo properties (APP_ENV != production)...")
		db.SeedProperties(repo)
	}

	handler := db.NewPropertyHandler(repo)

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
	api.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:    db.GetJWTKeyFunc(),
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.GET("/properties", handler.ListProperties)
	api.POST("/properties", handler.CreateProperty, db.RBACMiddleware(core.Seller, core.Broker, core.Admin))
	api.GET("/properties/:id", handler.GetProperty)
	api.POST("/properties/:id/documents/unlock", handler.UnlockDocuments, db.RBACMiddleware(core.Buyer, core.Broker, core.Admin))
	api.POST("/properties/:id/insurance/verify", handler.VerifyTitleInsurance, db.RBACMiddleware(core.Officer, core.Broker, core.Admin))
	api.PUT("/properties/:id/details", handler.UpdatePropertyDetails, db.RBACMiddleware(core.Seller, core.Broker, core.Admin))

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
