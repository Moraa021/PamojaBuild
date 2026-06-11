package router

import (
	"encoding/json"
	"net/http"

	"pamoja-build/backend/internal/auth"
	"pamoja-build/backend/internal/config"
	"pamoja-build/backend/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupRouter(jwtSecret string) *mux.Router {
	// Initialize middleware
	middleware.InitAuthMiddleware(jwtSecret)

	// Initialize auth handlers
	authService := auth.NewService(jwtSecret)
	authHandler := auth.NewHandler(authService)

	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// Public auth routes
	r.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Public config routes
	r.HandleFunc("/config/counties", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"counties": config.KenyanCounties,
		})
	}).Methods("GET")

	// Public task list and single task (for Teammate 2)
	r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Teammate 2 implements this
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet"})
	}).Methods("GET")

	r.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Teammate 2 implements this
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet"})
	}).Methods("GET")

	// Public donation total (for Teammate 3)
	r.HandleFunc("/donations/{task_id}/total", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Teammate 3 implements this
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet"})
	}).Methods("GET")

	// Protected routes (require authentication)
	// Using middleware directly on each route instead of subrouter
	r.HandleFunc("/tasks", middleware.AuthMiddleware(createTaskPlaceholder)).Methods("POST")
	r.HandleFunc("/tasks/{id}/volunteer", middleware.AuthMiddleware(applyVolunteerPlaceholder)).Methods("POST")
	r.HandleFunc("/tasks/{id}/volunteers/{vid}/approve", middleware.AuthMiddleware(approveVolunteerPlaceholder)).Methods("POST")
	r.HandleFunc("/tasks/{id}/complete", middleware.AuthMiddleware(completeTaskPlaceholder)).Methods("POST")
	r.HandleFunc("/donations/{task_id}", middleware.AuthMiddleware(createDonationPlaceholder)).Methods("POST")
	r.HandleFunc("/wallet/payout/{id}/sign", middleware.AuthMiddleware(signPayoutPlaceholder)).Methods("POST")
	r.HandleFunc("/wallet/payout/{id}/reject", middleware.AuthMiddleware(rejectPayoutPlaceholder)).Methods("POST")

	return r
}

// Placeholder handlers for Teammate 2 & 3 to implement
func createTaskPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 2"})
}

func applyVolunteerPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 2"})
}

func approveVolunteerPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 2"})
}

func completeTaskPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 2"})
}

func createDonationPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 3"})
}

func signPayoutPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 3"})
}

func rejectPayoutPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet - Teammate 3"})
}
