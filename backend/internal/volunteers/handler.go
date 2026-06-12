package volunteers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
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

	v, err := h.Service.Apply(taskID, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, v)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	vid, err := strconv.ParseInt(mux.Vars(r)["vid"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid volunteer id")
		return
	}
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

	if err := h.Service.Approve(vid, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "volunteer approved"})
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	vid, err := strconv.ParseInt(mux.Vars(r)["vid"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid volunteer id")
		return
	}
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

	if err := h.Service.Reject(vid, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "volunteer rejected"})
}

func (h *Handler) ListVolunteers(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	vs, err := h.Service.ListForTask(taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, vs)
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
