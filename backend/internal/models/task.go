package models

import "time"

type Task struct {
	ID             int       `json:"id"`
	CreatorID      int       `json:"creator_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	Region         string    `json:"region"`
	LocationDetail string    `json:"location_detail"`
	Status         string    `json:"status"`
	GoalSats       *int64    `json:"goal_sats"`
	MaxVolunteers  int       `json:"max_volunteers"`
	VolunteerMode  string    `json:"volunteer_mode"`
	ImagePath      *string   `json:"image_path"`
	CreatedAt      time.Time `json:"created_at"`
}
