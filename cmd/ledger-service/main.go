package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	echojwt "github.com/labstack/echo-jwt/v5"
	echo "github.com/labstack/echo/v5"
	middleware "github.com/labstack/echo/v5/middleware"
	"github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/realestate-trust/monorepo/internal/core"
	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/realestate-trust/monorepo/internal/events"
)

func main() {
	slog.SetDefault(slog.New(&db.SlogCorrelationHandler{
		Handler: slog.NewJSONHandler(os.Stdout, nil),
	}))

	cfg, err := core.LoadServiceConfig("ledger-service", ":8084")
	if err != nil {
		slog.Error("Failed to load configuration", "err", err)
		os.Exit(1)
	}

	var repo db.LedgerRepository
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
		repo = db.NewSQLLedgerRepository(dbPool.SQL)
	} else {
		slog.Info("DATABASE_URL is empty. Falling back to InMemoryLedgerRepository.")
		repo = db.NewInMemoryLedgerRepository()
	}

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo ledger (APP_ENV != production)...")
		db.SeedLedger(repo)
	}

	var rabbitConn *amqp.Connection
	if cfg.RabbitMQURL != "" {
		slog.Info("Connecting to RabbitMQ...", "url", cfg.RabbitMQURL)
		conn, err := events.Connect(cfg.RabbitMQURL)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "err", err)
		} else {
			slog.Info("Connected to RabbitMQ successfully!")
			rabbitConn = conn
			defer func() { _ = rabbitConn.Close() }()

			// Start RabbitMQ background consumer
			err = events.Consume(rabbitConn, "transaction-events", func(ctx context.Context, event events.TransactionEvent) error {
				slog.InfoContext(ctx, "Writing consumed event to immutable ledger", "payload", event.Payload, "id", event.ID)
				_, err := repo.WriteLog(event.ID, event.Payload)
				if err != nil {
					// Handle duplicate event gracefully (database unique constraint or in-memory error)
					var pqErr *pq.Error
					if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
						slog.WarnContext(ctx, "Duplicate event detected (DB constraint), acknowledging to RabbitMQ", "id", event.ID)
						return nil
					}
					if strings.Contains(err.Error(), "duplicate event") {
						slog.WarnContext(ctx, "Duplicate event detected (InMemory check), acknowledging to RabbitMQ", "id", event.ID)
						return nil
					}
					return err
				}
				return nil
			})
			if err != nil {
				slog.Error("Failed to start RabbitMQ consumer", "err", err)
			}
		}
	} else {
		slog.Warn("RABBITMQ_URL is empty. Running without event consumption.")
	}

	handler := db.NewLedgerHandler(repo)

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
	db.RegisterHealthEndpoints(e, sqlDB, rabbitConn)

	api := e.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:    db.GetJWTKeyFunc(),
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.POST("/logs", handler.WriteLog)
	api.GET("/logs", handler.GetLogs)
	api.GET("/logs/:index", handler.GetLog)

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
