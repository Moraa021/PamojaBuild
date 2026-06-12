package router

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"PamojaBuild/internal/auth"
	"PamojaBuild/internal/config"
	"PamojaBuild/internal/donations"
	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/middleware"
	"PamojaBuild/internal/tasks"
	"PamojaBuild/internal/volunteers"
	"PamojaBuild/internal/wallet"

	"github.com/gorilla/mux"
)

func SetupRouter(jwtSecret string, db *sql.DB, lndClient *lightning.Client) *mux.Router {
	// Initialize auth middleware
	middleware.InitAuthMiddleware(jwtSecret)

	// Initialize Auth dependencies
	authService := auth.NewService(jwtSecret)
	authHandler := auth.NewHandler(authService)

	// Initialize Tasks dependencies
	tasksRepo := tasks.NewRepository(db)
	tasksService := tasks.NewService(tasksRepo)
	tasksHandler := tasks.NewHandler(tasksService)

	// Initialize Volunteers dependencies
	volunteersRepo := volunteers.NewRepository(db)
	volunteersService := volunteers.NewService(volunteersRepo, tasksRepo)
	volunteersHandler := volunteers.NewHandler(volunteersService)

	// Initialize Donations dependencies
	donationsRepo := donations.NewRepository(db)
	donationsService := donations.NewService(donationsRepo, lndClient)
	donationsHandler := donations.NewHandler(donationsService)

	// Initialize Wallet dependencies
	walletRepo := wallet.NewRepository(db)
	walletService := wallet.NewService(walletRepo, lndClient)
	walletHandler := wallet.NewHandler(walletService)

	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// Public auth routes
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Public config routes
	r.HandleFunc("/config/counties", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"counties": config.KenyanCounties,
		})
	}).Methods("GET")

	// Public task routes
	r.HandleFunc("/tasks", tasksHandler.ListTasks).Methods("GET")
	r.HandleFunc("/tasks/{id}", tasksHandler.GetTask).Methods("GET")

	// Public donation total
	r.HandleFunc("/donations/{task_id}/total", donationsHandler.GetTotal).Methods("GET")

	// Protected routes (require authentication)
	r.Handle("/tasks", middleware.AuthMiddleware(http.HandlerFunc(tasksHandler.CreateTask))).Methods("POST")
	r.Handle("/tasks/{id}/raise-cap", middleware.AuthMiddleware(http.HandlerFunc(tasksHandler.RaiseCap))).Methods("POST")
	r.Handle("/tasks/{id}/volunteer", middleware.AuthMiddleware(http.HandlerFunc(volunteersHandler.Apply))).Methods("POST")
	r.Handle("/tasks/{id}/volunteers", middleware.AuthMiddleware(http.HandlerFunc(volunteersHandler.ListVolunteers))).Methods("GET")
	r.Handle("/tasks/{id}/volunteers/{vid}/approve", middleware.AuthMiddleware(http.HandlerFunc(volunteersHandler.Approve))).Methods("POST")
	r.Handle("/tasks/{id}/volunteers/{vid}/reject", middleware.AuthMiddleware(http.HandlerFunc(volunteersHandler.Reject))).Methods("POST")
	r.Handle("/tasks/{id}/complete", middleware.AuthMiddleware(http.HandlerFunc(walletHandler.Complete))).Methods("POST")
	r.Handle("/donations/{task_id}", middleware.AuthMiddleware(http.HandlerFunc(donationsHandler.Donate))).Methods("POST")

	// Keyholder protected routes (require authentication + keyholder role)
	r.Handle("/wallet/payout/{id}/sign", middleware.AuthMiddleware(middleware.RequireKeyholder(http.HandlerFunc(walletHandler.Sign)))).Methods("POST")
	r.Handle("/wallet/payout/{id}/reject", middleware.AuthMiddleware(middleware.RequireKeyholder(http.HandlerFunc(walletHandler.Reject)))).Methods("POST")

	// Serve uploaded task images
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	return r
}
