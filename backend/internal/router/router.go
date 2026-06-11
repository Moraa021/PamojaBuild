package router

import (
	"database/sql"
	"net/http"

	"PamojaBuild/internal/donations"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/middleware"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB, lndClient *lightning.Client) {
	// Initialize Donations dependencies
	donationsRepo := donations.NewRepository(db)
	donationsService := donations.NewService(donationsRepo, lndClient)
	donationsHandler := donations.NewHandler(donationsService)

	// Register Donations routes
	mux.Handle("POST /donations/{task_id}", middleware.AuthMiddleware(http.HandlerFunc(donationsHandler.Donate)))
	mux.HandleFunc("GET /donations/{task_id}/total", donationsHandler.GetTotal)
}
