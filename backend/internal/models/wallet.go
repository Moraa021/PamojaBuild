package models

type PayoutRequest struct {
	ID        int64  `json:"id"`
	TaskID    int64  `json:"task_id"`
	TotalSats int64  `json:"total_sats"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type PayoutStatus string

const (
	PayoutStatusPending   PayoutStatus = "pending"
	PayoutStatusApproved  PayoutStatus = "approved"
	PayoutStatusReleased  PayoutStatus = "released"
	PayoutStatusRejected  PayoutStatus = "rejected"
)

type PayoutSignature struct {
	ID              int64  `json:"id"`
	PayoutRequestID int64  `json:"payout_request_id"`
	KeyholderID     int64  `json:"keyholder_id"`
	Action          string `json:"action"`
	CreatedAt       string `json:"created_at"`
}

type Keyholder struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
}