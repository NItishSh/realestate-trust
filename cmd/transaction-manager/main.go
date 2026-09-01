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
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/realestate-trust/monorepo/internal/events"
)

func main() {
	slog.SetDefault(slog.New(&db.SlogCorrelationHandler{
		Handler: slog.NewJSONHandler(os.Stdout, nil),
	}))

	slog.Info("Starting Transaction & Escrow Manager API on :8080...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var rabbitConn *amqp.Connection
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL != "" {
		slog.Info("Connecting to RabbitMQ...", "url", rabbitURL)
		conn, err := events.Connect(rabbitURL)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ, running without event publishing", "err", err)
		} else {
			slog.Info("Connected to RabbitMQ successfully!")
			rabbitConn = conn
			defer func() { _ = rabbitConn.Close() }()
		}
	} else {
		slog.Warn("RABBITMQ_URL is empty. Running without event publishing.")
	}

	var repo db.TransactionRepository
	var dbPool *db.DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		slog.Info("Connecting to database...", "url", dbURL)
		pool, err := db.Connect()
		if err != nil {
			slog.Error("Database connection failed", "err", err)
			os.Exit(1)
		}
		dbPool = pool
		defer func() { _ = dbPool.Close() }()
		repo = db.NewSQLTransactionRepository(dbPool.SQL)
	} else {
		slog.Info("DATABASE_URL is empty. Falling back to InMemoryTransactionRepository.")
		repo = db.NewInMemoryTransactionRepository()
	}

	// Start Transactional Outbox Relay if DB and RabbitMQ are available
	var outboxRelay *events.OutboxRelay
	if dbPool != nil && dbPool.SQL != nil && rabbitConn != nil {
		relay, err := events.NewOutboxRelay(dbPool.SQL, rabbitConn, "transaction-events", 500*time.Millisecond)
		if err != nil {
			slog.Error("Failed to initialize Transactional Outbox Relay", "err", err)
		} else {
			outboxRelay = relay
			outboxRelay.Start(ctx)
		}
	}

	// Seed demo data in non-production environments
	if db.ShouldSeed() {
		slog.Info("🌱 Seeding demo transactions (APP_ENV != production)...")
		db.SeedTransactions(repo)
	}

	handler := db.NewTransactionHandler(repo, rabbitConn)

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

	api := e.Group("/api/v1")
	api.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:    db.GetJWTKeyFunc(),
		SigningKey: db.JWTSecret,
		ContextKey: "user",
	}))
	api.POST("/transactions", handler.CreateTransaction)
	api.GET("/transactions", handler.GetTransactions)
	api.GET("/transactions/:id", handler.GetTransaction)
	api.GET("/transactions/:id/escrow", handler.GetEscrow)
	api.PUT("/transactions/:id/status", handler.UpdateStatus)
	api.POST("/transactions/:id/escrow/fund", handler.FundEscrow)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      e,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	slog.Info("🔒 Security hardening: timeouts, headers, 1MB body limit enabled")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down server gracefully...")

	if outboxRelay != nil {
		outboxRelay.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("Server stopped")
}
