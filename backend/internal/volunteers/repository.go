package volunteers

import (
	"database/sql"

	"PamojaBuild/internal/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(v *models.Volunteer) error {
	return r.DB.QueryRow(
		`INSERT INTO volunteers (task_id, user_id, status) VALUES ($1,$2,$3) RETURNING id, created_at`,
		v.TaskID, v.UserID, v.Status,
	).Scan(&v.ID, &v.CreatedAt)
}

func (r *Repository) GetByID(id int64) (*models.Volunteer, error) {
	v := &models.Volunteer{}
	err := r.DB.QueryRow(
		`SELECT id, task_id, user_id, status, created_at FROM volunteers WHERE id=$1`, id,
	).Scan(&v.ID, &v.TaskID, &v.UserID, &v.Status, &v.CreatedAt)
	return v, err
}

func (r *Repository) CountForTask(taskID int64) (int64, error) {
	var count int64
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM volunteers WHERE task_id=$1 AND status IN ('pending','approved')`, taskID).Scan(&count)
	return count, err
}

func (r *Repository) AlreadyApplied(taskID, userID int64) (bool, error) {
	var count int64
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM volunteers WHERE task_id=$1 AND user_id=$2`, taskID, userID).Scan(&count)
	return count > 0, err
}

func (r *Repository) UpdateStatus(id int64, status string) error {
	_, err := r.DB.Exec(`UPDATE volunteers SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *Repository) GetApprovedForTask(taskID int64) ([]models.Volunteer, error) {
	rows, err := r.DB.Query(`SELECT id, task_id, user_id, status, created_at FROM volunteers WHERE task_id=$1 AND status='approved'`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []models.Volunteer
	for rows.Next() {
		var v models.Volunteer
		rows.Scan(&v.ID, &v.TaskID, &v.UserID, &v.Status, &v.CreatedAt)
		vs = append(vs, v)
	}
	return vs, nil
}

func (r *Repository) ListForTask(taskID int64) ([]models.Volunteer, error) {
	rows, err := r.DB.Query(`SELECT id, task_id, user_id, status, created_at FROM volunteers WHERE task_id=$1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []models.Volunteer
	for rows.Next() {
		var v models.Volunteer
		rows.Scan(&v.ID, &v.TaskID, &v.UserID, &v.Status, &v.CreatedAt)
		vs = append(vs, v)
	}
	return vs, nil
}
