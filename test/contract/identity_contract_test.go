package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	echo "github.com/labstack/echo/v5"
	"github.com/realestate-trust/monorepo/internal/core"
	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityServiceContract(t *testing.T) {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("../../.loki/specs/user-identity-openapi.yaml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	// Setup application handler
	e := echo.New()
	repo := db.NewInMemoryUserRepository()
	handler := db.NewUserHandler(repo)

	// We only need the routes defined in the spec
	e.POST("/api/v1/users", handler.RegisterUser)
	e.GET("/api/v1/users/:id", handler.GetUser)

	// Helper to validate request/response against OpenAPI spec
	validateInteraction := func(valReq *http.Request, rec *httptest.ResponseRecorder) {
		route, pathParams, err := router.FindRoute(valReq)
		require.NoError(t, err, "Route not found in OpenAPI spec")

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    valReq,
			PathParams: pathParams,
			Route:      route,
		}

		// Validate Request
		err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
		require.NoError(t, err, "Request validation failed")

		// Validate Response
		// Note: The response body is consumed by ValidateResponse, so we need to pass a copy or just bytes.
		// httptest.ResponseRecorder's Body is a *bytes.Buffer, so we can just pass it.
		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 rec.Code,
			Header:                 rec.Header(),
			Body:                   io.NopCloser(rec.Body),
		}

		err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
		require.NoError(t, err, "Response validation failed")
	}

	t.Run("POST /users", func(t *testing.T) {
		reqBody, _ := json.Marshal(core.RegisterUserRequest{
			Email:    "contract@example.com",
			FullName: "Contract Test",
			Password: "password123",
			Role:     core.Buyer,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		valReq := httptest.NewRequest(http.MethodPost, "http://localhost:3001/api/v1/users", bytes.NewBuffer(reqBody))
		valReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		validateInteraction(valReq, rec)
	})
}
