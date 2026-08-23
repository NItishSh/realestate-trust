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
	"github.com/labstack/echo/v5"

	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionManagerContract(t *testing.T) {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("../../.loki/specs/escrow-manager-openapi.yaml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	// Setup application handler
	e := echo.New()
	repo := db.NewInMemoryTransactionRepository()
	// Mock RabbitMQ publisher with a nil channel (handlers should handle it safely or we provide a mock)
	handler := db.NewTransactionHandler(repo, nil)

	// We only need the routes defined in the spec
	e.POST("/api/v1/transactions", handler.CreateTransaction)

	// Helper to validate request/response against OpenAPI spec
	validateInteraction := func(req *http.Request, valReq *http.Request, rec *httptest.ResponseRecorder) {
		route, pathParams, err := router.FindRoute(valReq)
		require.NoError(t, err, "Route not found in OpenAPI spec")

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    valReq,
			PathParams: pathParams,
			Route:      route,
		}

		err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
		require.NoError(t, err, "Request validation failed")

		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 rec.Code,
			Header:                 rec.Header(),
			Body:                   io.NopCloser(rec.Body),
		}

		err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
		require.NoError(t, err, "Response validation failed")
	}

	t.Run("POST /transactions", func(t *testing.T) {
		reqBody, _ := json.Marshal(db.CreateTxRequest{
			PropertyID:  "prop-1234-abcd",
			BuyerID:     "usr-buyer",
			SellerID:    "usr-seller",
			TotalAmount: 250000.0,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		valReq := httptest.NewRequest(http.MethodPost, "http://localhost:3000/api/v1/transactions", bytes.NewBuffer(reqBody))
		valReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		validateInteraction(req, valReq, rec)
	})
}
