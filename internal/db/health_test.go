package db

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v5"
)

func TestHealthEndpoints(t *testing.T) {
	e := echo.New()
	RegisterHealthEndpoints(e, nil, nil)

	// 1. Test /api/v1/health
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /health, got %d", rec.Code)
	}

	var legacyResp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &legacyResp); err != nil || legacyResp["status"] != "UP" {
		t.Errorf("unexpected legacy health response: %v", rec.Body.String())
	}

	// 2. Test /api/v1/health/live
	reqLive := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	recLive := httptest.NewRecorder()
	e.ServeHTTP(recLive, reqLive)

	if recLive.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /health/live, got %d", recLive.Code)
	}

	var liveResp map[string]string
	if err := json.Unmarshal(recLive.Body.Bytes(), &liveResp); err != nil || liveResp["status"] != "ALIVE" {
		t.Errorf("unexpected liveness response: %v", recLive.Body.String())
	}

	// 3. Test /api/v1/health/ready (mock / in-memory mode)
	reqReady := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	recReady := httptest.NewRecorder()
	e.ServeHTTP(recReady, reqReady)

	if recReady.Code != http.StatusOK {
		t.Fatalf("expected status 200 on /health/ready in mock mode, got %d", recReady.Code)
	}

	var readyResp map[string]interface{}
	if err := json.Unmarshal(recReady.Body.Bytes(), &readyResp); err != nil || readyResp["status"] != "READY" {
		t.Errorf("unexpected readiness response: %v", recReady.Body.String())
	}
}
