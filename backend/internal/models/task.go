package models

import "database/sql"

type Task struct {
	ID             int64          `json:"id"`
	CreatorID      int64          `json:"creator_id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Category       string         `json:"category"`
	Region         string         `json:"region"`
	LocationDetail sql.NullString `json:"location_detail,omitempty"`
	Status         string         `json:"status"`
	GoalSats       sql.NullInt64  `json:"goal_sats,omitempty"`
	MaxVolunteers  int64          `json:"max_volunteers"`
	VolunteerMode  string         `json:"volunteer_mode"`
	CreatedAt      string         `json:"created_at"`
	ImagePath      sql.NullString `json:"image_path,omitempty"`
}

type TaskStatus string

const (
	StatusOpen               TaskStatus = "open"
	StatusInProgress         TaskStatus = "in_progress"
	StatusPendingVerification TaskStatus = "pending_verification"
	StatusCompleted          TaskStatus = "completed"
)

type VolunteerMode string

const (
	VolunteerModeOpen             VolunteerMode = "open"
	VolunteerModeApprovalRequired VolunteerMode = "approval_required"
)
