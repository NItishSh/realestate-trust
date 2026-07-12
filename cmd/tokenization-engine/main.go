package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Fractional Tokenization Engine API on :8083...")

	repo := db.NewInMemoryTokenizationRepository()
	handler := db.NewTokenizationHandler(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/pools", handler.CreatePool)
	mux.HandleFunc("POST /api/v1/pools/{id}/buy", handler.BuyShares)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8083", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
