package main

import (
	"fmt"
	"net/http"

	"github.com/realestate-trust/monorepo/internal/db"
)

func main() {
	fmt.Println("Starting Property Registry Service API on :8085...")

	repo := db.NewInMemoryPropertyRepository()

	if db.ShouldSeed() {
		fmt.Println("🌱 Seeding demo properties (APP_ENV != production)...")
		db.SeedProperties(repo)
	}

	handler := db.NewPropertyHandler(repo)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/properties", handler.ListProperties)
	mux.HandleFunc("GET /api/v1/properties/{id}", handler.GetProperty)
	mux.HandleFunc("POST /api/v1/properties/{id}/documents/unlock", handler.UnlockDocuments)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})

	if err := http.ListenAndServe(":8085", db.EnableCORS(mux)); err != nil {
		panic(err)
	}
}
