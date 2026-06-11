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
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	userID := r.Context().Value("user_id").(int)
	v, err := h.Service.Apply(taskID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	vid, err := strconv.Atoi(mux.Vars(r)["vid"])
	if err != nil {
		http.Error(w, "invalid volunteer id", http.StatusBadRequest)
		return
	}
	userID := r.Context().Value("user_id").(int)
	if err := h.Service.Approve(vid, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"volunteer approved"}`))
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	vid, err := strconv.Atoi(mux.Vars(r)["vid"])
	if err != nil {
		http.Error(w, "invalid volunteer id", http.StatusBadRequest)
		return
	}
	userID := r.Context().Value("user_id").(int)
	if err := h.Service.Reject(vid, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"volunteer rejected"}`))
}

func (h *Handler) ListVolunteers(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	vs, err := h.Service.ListForTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vs)
}
