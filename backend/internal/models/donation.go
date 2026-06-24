package models

type Donation struct {
	ID            int64          `json:"id"`
	TaskID        int64          `json:"task_id"`
	DonorID       int64          `json:"donor_id"`
	AmountSats    int64          `json:"amount_sats"`
	PaymentHash   string         `json:"payment_hash"`
	PaymentRequest string        `json:"payment_request"`
	IsAnonymous   bool           `json:"is_anonymous"`
	Status        string         `json:"status"`
	CreatedAt     string         `json:"created_at"`
}

type DonationStatus string

const (
	DonationStatusPending   DonationStatus = "pending"
	DonationStatusConfirmed DonationStatus = "confirmed"
)