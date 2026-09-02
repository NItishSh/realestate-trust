package core

import (
	"os"
	"testing"
)

func TestLoadServiceConfig_Defaults(t *testing.T) {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("CORS_ALLOWED_ORIGINS")

	cfg, err := LoadServiceConfig("test-service", ":8080")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.ServiceName != "test-service" {
		t.Errorf("expected ServiceName test-service, got %s", cfg.ServiceName)
	}
	if cfg.Port != ":8080" {
		t.Errorf("expected Port :8080, got %s", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected AppEnv development, got %s", cfg.AppEnv)
	}
	if len(cfg.CorsAllowedOrigins) != 1 || cfg.CorsAllowedOrigins[0] != "*" {
		t.Errorf("expected default wildcard CORS origin, got %v", cfg.CorsAllowedOrigins)
	}
}

func TestLoadServiceConfig_CustomEnv(t *testing.T) {
	_ = os.Setenv("PORT", "9090")
	_ = os.Setenv("APP_ENV", "production")
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "https://app.trust.realestate.com, https://admin.trust.realestate.com")
	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("APP_ENV")
		_ = os.Unsetenv("CORS_ALLOWED_ORIGINS")
	}()

	cfg, err := LoadServiceConfig("custom-service", ":8080")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Port != ":9090" {
		t.Errorf("expected Port :9090, got %s", cfg.Port)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected AppEnv production, got %s", cfg.AppEnv)
	}
	if len(cfg.CorsAllowedOrigins) != 2 || cfg.CorsAllowedOrigins[0] != "https://app.trust.realestate.com" {
		t.Errorf("expected parsed CORS origins, got %v", cfg.CorsAllowedOrigins)
	}
}
