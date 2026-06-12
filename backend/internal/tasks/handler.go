package tasks

import (
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"

	"PamojaBuild/internal/models"

	"github.com/gorilla/mux"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(5 << 20)
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	maxVol, err := strconv.ParseInt(r.FormValue("max_volunteers"), 10, 64)
	if err != nil || maxVol <= 0 {
		http.Error(w, "max_volunteers is required and must be positive", http.StatusBadRequest)
		return
	}

	var goalSats sql.NullInt64
	if gs := r.FormValue("goal_sats"); gs != "" {
		v, err := strconv.ParseInt(gs, 10, 64)
		if err != nil {
			http.Error(w, "invalid goal_sats", http.StatusBadRequest)
			return
		}
		goalSats = sql.NullInt64{Int64: v, Valid: true}
	}

	var locDetail sql.NullString
	if ld := r.FormValue("location_detail"); ld != "" {
		locDetail = sql.NullString{String: ld, Valid: true}
	}

	task := &models.Task{
		CreatorID:      userID,
		Title:          r.FormValue("title"),
		Description:    r.FormValue("description"),
		Category:       r.FormValue("category"),
		Region:         r.FormValue("region"),
		LocationDetail: locDetail,
		GoalSats:       goalSats,
		MaxVolunteers:  maxVol,
		VolunteerMode:  r.FormValue("volunteer_mode"),
	}

	var file multipart.File
	var header *multipart.FileHeader
	file, header, err = r.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "error reading image", http.StatusBadRequest)
		return
	}
	if file != nil {
		defer file.Close()
	}

	if err := h.Service.CreateTask(task, file, header); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Service.ListTasks(
		r.URL.Query().Get("region"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("category"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	task, err := h.Service.GetTask(id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) RaiseCap(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var body struct {
		MaxVolunteers int64 `json:"max_volunteers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.Service.RaiseCap(id, body.MaxVolunteers, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(`{"message":"cap updated"}`))
}
