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
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/realestate-trust/monorepo/internal/events"
)

func main() {
	slog.Info("Starting Immutable Audit Ledger API on :8084...")

	var repo db.LedgerRepository
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		slog.Info("Connecting to database...", "url", dbURL)
		dbPool, err := db.Connect()
		if err != nil {
			slog.Error("Database connection failed", "err", err)
			os.Exit(1)
		}
		defer dbPool.Close()
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
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL != "" {
		slog.Info("Connecting to RabbitMQ...", "url", rabbitURL)
		conn, err := events.Connect(rabbitURL)
		if err != nil {
			slog.Error("Failed to connect to RabbitMQ", "err", err)
		} else {
			slog.Info("Connected to RabbitMQ successfully!")
			rabbitConn = conn
			defer rabbitConn.Close()

			// Start RabbitMQ background consumer
			err = events.Consume(rabbitConn, "transaction-events", func(event events.TransactionEvent) error {
				slog.Info("Writing consumed event to immutable ledger", "payload", event.Payload)
				_, err := repo.WriteLog(event.Payload)
				return err
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
	e.Use(middleware.BodyLimit(1 << 20))

	e.GET("/api/v1/health", func(c *echo.Context) error {
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
