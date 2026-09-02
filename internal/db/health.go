package db

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RegisterHealthEndpoints mounts standard liveness, readiness, and legacy health endpoints.
func RegisterHealthEndpoints(e *echo.Echo, sqlDB *sql.DB, amqpConn *amqp.Connection) {
	// 1. Legacy / Generic Health Check (Always returns 200 UP)
	e.GET("/api/v1/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	// 2. Kubernetes Liveness Probe: Checks if the HTTP process itself is responsive
	e.GET("/api/v1/health/live", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ALIVE",
		})
	})

	// 3. Kubernetes Readiness Probe: Checks connectivity to DB and RabbitMQ dependencies
	e.GET("/api/v1/health/ready", func(c *echo.Context) error {
		type ComponentStatus struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}

		response := map[string]interface{}{
			"status":     "READY",
			"components": map[string]ComponentStatus{},
		}
		components := response["components"].(map[string]ComponentStatus)

		isReady := true

		// Check PostgreSQL connection pool
		if sqlDB != nil {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(ctx); err != nil {
				isReady = false
				components["database"] = ComponentStatus{
					Status:  "DOWN",
					Message: err.Error(),
				}
			} else {
				components["database"] = ComponentStatus{
					Status: "UP",
				}
			}
		} else {
			components["database"] = ComponentStatus{
				Status:  "IN_MEMORY",
				Message: "Running in mock/in-memory mode",
			}
		}

		// Check RabbitMQ connection
		if amqpConn != nil {
			if amqpConn.IsClosed() {
				isReady = false
				components["rabbitmq"] = ComponentStatus{
					Status:  "DOWN",
					Message: "RabbitMQ connection is closed",
				}
			} else {
				components["rabbitmq"] = ComponentStatus{
					Status: "UP",
				}
			}
		}

		if !isReady {
			response["status"] = "UNAVAILABLE"
			return c.JSON(http.StatusServiceUnavailable, response)
		}

		return c.JSON(http.StatusOK, response)
	})
}
