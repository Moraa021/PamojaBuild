package wallet

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	// Parse task ID from wildcard
	taskIDStr := r.PathValue("id")
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

	// Call service
	pr, err := h.service.CompleteTask(r.Context(), taskID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not the creator") {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fetch multisig address to return to caller
	addr, err := h.service.GetMultisigAddress(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to derive multisig address: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]interface{}{
		"payout_request_id": pr.ID,
		"multisig_address":  addr,
	})
}

func (h *Handler) Sign(w http.ResponseWriter, r *http.Request) {
	// Parse payout request ID
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payout request ID")
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

	// Call service
	pr, err := h.service.SignPayoutRequest(r.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not a keyholder") {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, pr)
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	// Parse payout request ID
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payout request ID")
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

	// Call service
	pr, err := h.service.RejectPayoutRequest(r.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not a keyholder") {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, pr)
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
