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
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	maxVol, err := strconv.ParseInt(r.FormValue("max_volunteers"), 10, 64)
	if err != nil || maxVol <= 0 {
		writeError(w, http.StatusBadRequest, "max_volunteers is required and must be positive")
		return
	}

	var goalSats sql.NullInt64
	if gs := r.FormValue("goal_sats"); gs != "" {
		v, err := strconv.ParseInt(gs, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid goal_sats")
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
		writeError(w, http.StatusBadRequest, "error reading image")
		return
	}
	if file != nil {
		defer file.Close()
	}

	if err := h.Service.CreateTask(task, file, header); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Service.ListTasks(
		r.URL.Query().Get("region"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("category"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	task, err := h.Service.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeSuccess(w, http.StatusOK, task)
}

func (h *Handler) RaiseCap(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
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

	var body struct {
		MaxVolunteers int64 `json:"max_volunteers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.Service.RaiseCap(id, body.MaxVolunteers, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, map[string]string{"message": "cap updated"})
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
