package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Transaction & Escrow Manager API on :8080...")

	repo := db.NewInMemoryTransactionRepository()
	handler := db.NewTransactionHandler(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/transactions", handler.CreateTransaction)
	mux.HandleFunc("GET /api/v1/transactions/{id}", handler.GetTransaction)
	mux.HandleFunc("PUT /api/v1/transactions/{id}/status", handler.UpdateStatus)
	mux.HandleFunc("POST /api/v1/transactions/{id}/escrow/fund", handler.FundEscrow)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8080", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
