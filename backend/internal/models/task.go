package models

import (
	"database/sql"
	"time"
)

// type Task struct {
// 	ID             int64          `json:"id"`
// 	CreatorID      int64          `json:"creator_id"`
// 	Title          string         `json:"title"`
// 	Description    string         `json:"description"`
// 	Category       string         `json:"category"`
// 	Region         string         `json:"region"`
// 	LocationDetail sql.NullString `json:"location_detail,omitempty"`
// 	Status         string         `json:"status"`
// 	GoalSats       sql.NullInt64  `json:"goal_sats,omitempty"`
// 	MaxVolunteers  int64          `json:"max_volunteers"`
// 	VolunteerMode  string         `json:"volunteer_mode"`
// 	CreatedAt      string         `json:"created_at"`
// 	ImagePath      sql.NullString `json:"image_path,omitempty"`
// }

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

// TaskState represents the strict financial lifecycle of the campaign's money.
type TaskState string

const (
	TaskStateActive         TaskState = "ACTIVE"
	TaskStateLiquidating    TaskState = "LIQUIDATING"
	TaskStateReadyForPayout TaskState = "READY_FOR_PAYOUT"
	TaskStateSystemLockdown TaskState = "SYSTEM_LOCKDOWN"
	TaskStateArchived       TaskState = "ARCHIVED"
)

// Task represents the unified business metadata and financial state of a donation campaign.
type Task struct {
	// --- Existing Volunteer & Social Metadata ---
	ID             int64          `json:"id" db:"id"`
	CreatorID      int64          `json:"creator_id" db:"creator_id"`
	Title          string         `json:"title" db:"title"`
	Description    string         `json:"description" db:"description"`
	Category       string         `json:"category" db:"category"`
	Region         string         `json:"region" db:"region"`
	LocationDetail sql.NullString `json:"location_detail,omitempty" db:"location_detail"`
	Status         TaskStatus     `json:"status" db:"status"` // Business/Volunteer state
	GoalSats       sql.NullInt64  `json:"goal_sats,omitempty" db:"goal_sats"`
	MaxVolunteers  int64          `json:"max_volunteers" db:"max_volunteers"`
	VolunteerMode  VolunteerMode  `json:"volunteer_mode" db:"volunteer_mode"`
	ImagePath      sql.NullString `json:"image_path,omitempty" db:"image_path"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"` // Standardized to time.Time

	// --- New PamojaBuild Financial & Cryptographic Core ---
	Slug           string         `json:"slug" db:"slug"`             // e.g., "kibera-clean-water-004" (Cryptographic Tenant ID)
	FinancialState TaskState      `json:"financial_state" db:"financial_state"` // Escrow lifecycle state
	L2BalanceSats  int64          `json:"l2_balance_sats" db:"l2_balance_sats"` // Active Lightning channels wallet balance
	L1BalanceSats  int64          `json:"l1_balance_sats" db:"l1_balance_sats"` // On-chain multi-sig vault balance
	CurrentIndex   int32          `json:"current_index" db:"current_index"`     // HD derivation index for generating 3-of-5 addresses
}
