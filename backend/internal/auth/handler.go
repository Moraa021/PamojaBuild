package auth

import (
	"encoding/json"
	"net/http"

	"PamojaBuild/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, token, err := h.service.Register(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user already exists" {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	user, token, err := h.service.Login(&req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
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