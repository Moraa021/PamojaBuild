package donations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type donateRequest struct {
	AmountSats  int64 `json:"amount_sats"`
	IsAnonymous bool  `json:"is_anonymous"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Donate(w http.ResponseWriter, r *http.Request) {
	// Parse task_id from path wildcard or gorilla mux
	taskIDStr := r.PathValue("task_id")
	if taskIDStr == "" {
		taskIDStr = mux.Vars(r)["task_id"]
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Parse user_id from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Parse request body
	var req donateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service
	donation, err := h.service.Donate(r.Context(), taskID, userID, req.AmountSats, req.IsAnonymous)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		if errors.Is(err, ErrInvalidAmount) {
			writeError(w, http.StatusBadRequest, "Invalid donation amount")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create donation: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, donation)
}

func (h *Handler) GetTotal(w http.ResponseWriter, r *http.Request) {
	// Parse task_id from path wildcard or gorilla mux
	taskIDStr := r.PathValue("task_id")
	if taskIDStr == "" {
		taskIDStr = mux.Vars(r)["task_id"]
	}
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Call service
	total, err := h.service.GetTotalConfirmed(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, "Task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get donation total: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"task_id":    taskID,
		"total_sats": total,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error": message,
	})
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, map[string]interface{}{
		"data": data,
	})
}
