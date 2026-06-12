package tasks

import (
	"database/sql"
	"fmt"

	"PamojaBuild/internal/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(t *models.Task) error {
	query := `INSERT INTO tasks (creator_id, title, description, category, region, location_detail, status, goal_sats, max_volunteers, volunteer_mode, image_path) VALUES ($1,$2,$3,$4,$5,$6,'open',$7,$8,$9,$10) RETURNING id, created_at`
	err := r.DB.QueryRow(query, t.CreatorID, t.Title, t.Description, t.Category, t.Region, t.LocationDetail, t.GoalSats, t.MaxVolunteers, t.VolunteerMode, t.ImagePath).Scan(&t.ID, &t.CreatedAt)
	if err == nil {
		t.Status = "open"
	}
	return err
}

func (r *Repository) List(region, status, category string) ([]models.Task, error) {
	query := `SELECT id, creator_id, title, description, category, region, location_detail, status, goal_sats, max_volunteers, volunteer_mode, image_path, created_at FROM tasks WHERE 1=1`
	args := []interface{}{}
	i := 1
	if region != "" {
		query += fmt.Sprintf(" AND region=$%d", i)
		args = append(args, region)
		i++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, status)
		i++
	}
	if category != "" {
		query += fmt.Sprintf(" AND category=$%d", i)
		args = append(args, category)
		i++
	}
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.CreatorID, &t.Title, &t.Description, &t.Category, &t.Region, &t.LocationDetail, &t.Status, &t.GoalSats, &t.MaxVolunteers, &t.VolunteerMode, &t.ImagePath, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *Repository) GetByID(id int64) (*models.Task, error) {
	t := &models.Task{}
	err := r.DB.QueryRow(`SELECT id, creator_id, title, description, category, region, location_detail, status, goal_sats, max_volunteers, volunteer_mode, image_path, created_at FROM tasks WHERE id=$1`, id).Scan(&t.ID, &t.CreatorID, &t.Title, &t.Description, &t.Category, &t.Region, &t.LocationDetail, &t.Status, &t.GoalSats, &t.MaxVolunteers, &t.VolunteerMode, &t.ImagePath, &t.CreatedAt)
	return t, err
}

func (r *Repository) UpdateStatus(taskID int64, status string) error {
	_, err := r.DB.Exec(`UPDATE tasks SET status=$1 WHERE id=$2`, status, taskID)
	return err
}

func (r *Repository) RaiseCap(taskID int64, newCap int64) error {
	_, err := r.DB.Exec(`UPDATE tasks SET max_volunteers=$1 WHERE id=$2`, newCap, taskID)
	return err
}

func (r *Repository) CountApprovedVolunteers(taskID int64) (int64, error) {
	var count int64
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM volunteers WHERE task_id=$1 AND status='approved'`, taskID).Scan(&count)
	return count, err
}
