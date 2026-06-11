package models

type Volunteer struct {
	ID             int64  `json:"id"`
	TaskID         int64  `json:"task_id"`
	UserID         int64  `json:"user_id"`
	Status         string `json:"status"`
	PaymentRequest string `json:"payment_request"`
	CreatedAt      string `json:"created_at"`
}

type VolunteerStatus string

const (
	VolunteerStatusPending   VolunteerStatus = "pending"
	VolunteerStatusApproved  VolunteerStatus = "approved"
	VolunteerStatusPaid    VolunteerStatus = "paid"
	VolunteerStatusPayFailed VolunteerStatus = "pay_failed"
)