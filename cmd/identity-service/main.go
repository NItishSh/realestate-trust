package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting User & Identity Service API on :8081...")

	// Initialize In-Memory Repository (default for local validation / dev)
	repo := db.NewInMemoryUserRepository()
	handler := db.NewUserHandler(repo)

	mux := http.NewServeMux()

	// Endpoints configured with native Go 1.22+ routing features
	mux.HandleFunc("POST /api/v1/users", handler.RegisterUser)
	mux.HandleFunc("GET /api/v1/users/{id}", handler.GetUser)
	mux.HandleFunc("POST /api/v1/users/{id}/kyc", handler.SubmitKYC)
	mux.HandleFunc("GET /api/v1/users/{id}/kyc/status", handler.GetKYCStatus)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8081", mux); err != nil {
		panic(err)
	}
}
