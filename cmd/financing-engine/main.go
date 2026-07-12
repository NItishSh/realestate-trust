package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Embedded Financing Engine API on :8082...")

	repo := db.NewInMemoryFinancingRepository()
	handler := db.NewFinancingHandler(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/loans", handler.ApplyLoan)
	mux.HandleFunc("GET /api/v1/loans/{id}", handler.GetLoan)
	mux.HandleFunc("POST /api/v1/loans/{id}/disburse", handler.DisburseLoan)
	mux.HandleFunc("POST /api/v1/loans/webhooks/bank", handler.BankWebhook)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8082", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
