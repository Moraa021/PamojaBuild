package router

import (
	"database/sql"
	"net/http"

	"PamojaBuild/internal/donations"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/middleware"
	"PamojaBuild/internal/wallet"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB, lndClient *lightning.Client) {
	// Initialize Donations dependencies
	donationsRepo := donations.NewRepository(db)
	donationsService := donations.NewService(donationsRepo, lndClient)
	donationsHandler := donations.NewHandler(donationsService)

	// Register Donations routes
	mux.Handle("POST /donations/{task_id}", middleware.AuthMiddleware(http.HandlerFunc(donationsHandler.Donate)))
	mux.HandleFunc("GET /donations/{task_id}/total", donationsHandler.GetTotal)

	// Initialize Wallet dependencies
	walletRepo := wallet.NewRepository(db)
	walletService := wallet.NewService(walletRepo, lndClient)
	walletHandler := wallet.NewHandler(walletService)

	// Register Wallet/Multisig routes
	mux.Handle("POST /tasks/{id}/complete", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.Complete)))
	mux.Handle("POST /wallet/payout/{id}/sign", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.Sign)))
	mux.Handle("POST /wallet/payout/{id}/reject", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.Reject)))
}
