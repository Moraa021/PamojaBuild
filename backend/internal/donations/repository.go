package donations

import (
	"context"
	"database/sql"
	"errors"

	"PamojaBuild/internal/models"
)

type Repository interface {
	TaskExists(ctx context.Context, taskID int64) (bool, error)
	CreateDonation(ctx context.Context, donation *models.Donation) error
	GetPendingDonations(ctx context.Context, taskID int64) ([]*models.Donation, error)
	UpdateDonationStatus(ctx context.Context, donationID int64, status string) error
	GetConfirmedTotal(ctx context.Context, taskID int64) (int64, error)
}

type sqliteRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) TaskExists(ctx context.Context, taskID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)"
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *sqliteRepository) CreateDonation(ctx context.Context, d *models.Donation) error {
	query := `
		INSERT INTO donations (task_id, donor_id, amount_sats, payment_hash, payment_request, is_anonymous, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	isAnon := 0
	if d.IsAnonymous {
		isAnon = 1
	}

	result, err := r.db.ExecContext(ctx, query, d.TaskID, d.DonorID, d.AmountSats, d.PaymentHash, d.PaymentRequest, isAnon, d.Status)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	d.ID = id

	var createdAt string
	err = r.db.QueryRowContext(ctx, "SELECT created_at FROM donations WHERE id = ?", id).Scan(&createdAt)
	if err != nil {
		return err
	}
	d.CreatedAt = createdAt
	return nil
}

func (r *sqliteRepository) GetPendingDonations(ctx context.Context, taskID int64) ([]*models.Donation, error) {
	query := `
		SELECT id, task_id, donor_id, amount_sats, payment_hash, payment_request, is_anonymous, status, created_at
		FROM donations
		WHERE task_id = ? AND status = 'pending'
	`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations []*models.Donation
	for rows.Next() {
		var d models.Donation
		var isAnon int
		err := rows.Scan(&d.ID, &d.TaskID, &d.DonorID, &d.AmountSats, &d.PaymentHash, &d.PaymentRequest, &isAnon, &d.Status, &d.CreatedAt)
		if err != nil {
			return nil, err
		}
		d.IsAnonymous = (isAnon == 1)
		donations = append(donations, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *sqliteRepository) UpdateDonationStatus(ctx context.Context, donationID int64, status string) error {
	query := "UPDATE donations SET status = ? WHERE id = ?"
	res, err := r.db.ExecContext(ctx, query, status, donationID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("donation not found")
	}
	return nil
}

func (r *sqliteRepository) GetConfirmedTotal(ctx context.Context, taskID int64) (int64, error) {
	query := "SELECT COALESCE(SUM(amount_sats), 0) FROM donations WHERE task_id = ? AND status = 'confirmed'"
	var total int64
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
