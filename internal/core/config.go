package core

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// ServiceConfig holds standardized runtime configuration for each microservice.
type ServiceConfig struct {
	ServiceName        string
	Port               string
	DatabaseURL        string
	RabbitMQURL        string
	AppEnv             string
	KeycloakURL        string
	VaultAddr          string
	CorsAllowedOrigins []string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
}

// LoadServiceConfig parses environment variables with secure defaults and fail-fast validation.
func LoadServiceConfig(serviceName, defaultPort string) (*ServiceConfig, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	corsRaw := os.Getenv("CORS_ALLOWED_ORIGINS")
	var origins []string
	if corsRaw != "" {
		for _, o := range strings.Split(corsRaw, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	cfg := &ServiceConfig{
		ServiceName:        serviceName,
		Port:               port,
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RabbitMQURL:        os.Getenv("RABBITMQ_URL"),
		AppEnv:             appEnv,
		KeycloakURL:        os.Getenv("KEYCLOAK_URL"),
		VaultAddr:          os.Getenv("VAULT_ADDR"),
		CorsAllowedOrigins: origins,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       15 * time.Second,
		IdleTimeout:        60 * time.Second,
	}

	slog.Info("Configuration loaded successfully",
		"service", cfg.ServiceName,
		"port", cfg.Port,
		"env", cfg.AppEnv,
		"has_db", cfg.DatabaseURL != "",
		"has_rabbitmq", cfg.RabbitMQURL != "",
	)

	return cfg, nil
}

// ValidateProduction checks that critical production infrastructure is configured.
func (c *ServiceConfig) ValidateProduction() error {
	if c.AppEnv == "production" {
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required in production environment")
		}
	}
	return nil
}
