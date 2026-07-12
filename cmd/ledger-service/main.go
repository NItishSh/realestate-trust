package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Immutable Audit Ledger API on :8084...")

	repo := db.NewInMemoryLedgerRepository()
	handler := db.NewLedgerHandler(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/logs", handler.WriteLog)
	mux.HandleFunc("GET /api/v1/logs/{index}", handler.GetLog)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8084", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
