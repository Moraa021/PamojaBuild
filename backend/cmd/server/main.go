package main

import (
	"fmt"
	"log"
	"net/http"

	"PamojaBuild/internal/config"
	"PamojaBuild/internal/db"
	"PamojaBuild/internal/lightning"
)

func main() {
	cfg := config.Load()

	database, err := db.New("pamojabuild.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	lightningClient, err := lightning.NewClient(cfg)
	if err != nil {
		log.Fatal("Failed to create lightning client:", err)
	}
	_ = lightningClient // Used by donations and wallet services

	router := http.NewServeMux()

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}