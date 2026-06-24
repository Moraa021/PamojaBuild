package wallet

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"PamojaBuild/internal/models"

	"github.com/btcsuite/btcd/btcec/v2"
)

type Repository interface {
	GetKeyholders(ctx context.Context) ([]*models.Keyholder, error)
	GetTaskCreatorAndStatus(ctx context.Context, taskID int64) (int64, string, error)
	GetApprovedVolunteers(ctx context.Context, taskID int64) ([]*models.Volunteer, error)
	GetConfirmedDonationsTotal(ctx context.Context, taskID int64) (int64, error)
	CreatePayoutRequest(ctx context.Context, pr *models.PayoutRequest) error
	GetPayoutRequest(ctx context.Context, id int64) (*models.PayoutRequest, error)
	UpdatePayoutRequestStatus(ctx context.Context, id int64, status string) error
	UpdatePayoutRequestStatusAndTxID(ctx context.Context, id int64, status string, txID string) error
	GetKeyholderByUserID(ctx context.Context, userID int64) (*models.Keyholder, error)
	HasKeyholderSigned(ctx context.Context, payoutRequestID, keyholderID int64) (bool, error)
	SavePayoutSignature(ctx context.Context, sig *models.PayoutSignature) error
	GetPayoutSignatures(ctx context.Context, payoutRequestID int64) ([]*models.PayoutSignature, error)
	UpdateTaskStatus(ctx context.Context, taskID int64, status string) error
	UpdateVolunteersStatus(ctx context.Context, taskID int64, oldStatus, newStatus string) error
	UpdateVolunteerStatus(ctx context.Context, id int64, status string) error
	GetStalePayoutRequests(ctx context.Context, beforeTime string) ([]*models.PayoutRequest, error)
}

type sqliteRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) GetKeyholders(ctx context.Context) ([]*models.Keyholder, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, COALESCE(public_key, '') FROM keyholders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var khs []*models.Keyholder
	for rows.Next() {
		var kh models.Keyholder
		err := rows.Scan(&kh.ID, &kh.UserID, &kh.PublicKey)
		if err != nil {
			return nil, err
		}
		if kh.PublicKey == "" {
			// Derive public key deterministically for keyholder 1-5 fallback
			seed := fmt.Sprintf("keyholder-seed-%d-pamojabuild", kh.ID)
			hash := sha256.Sum256([]byte(seed))
			privKey, _ := btcec.PrivKeyFromBytes(hash[:])
			kh.PublicKey = hex.EncodeToString(privKey.PubKey().SerializeCompressed())
			_, _ = r.db.ExecContext(ctx, "UPDATE keyholders SET public_key = ? WHERE id = ?", kh.PublicKey, kh.ID)
		}
		khs = append(khs, &kh)
	}
	return khs, nil
}

func (r *sqliteRepository) GetTaskCreatorAndStatus(ctx context.Context, taskID int64) (int64, string, error) {
	var creatorID int64
	var status string
	err := r.db.QueryRowContext(ctx, "SELECT creator_id, status FROM tasks WHERE id = ?", taskID).Scan(&creatorID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", fmt.Errorf("task not found")
		}
		return 0, "", err
	}
	return creatorID, status, nil
}

func (r *sqliteRepository) GetApprovedVolunteers(ctx context.Context, taskID int64) ([]*models.Volunteer, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, task_id, user_id, status, COALESCE(payment_request, '') FROM volunteers WHERE task_id = ? AND status = 'approved'", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vols []*models.Volunteer
	for rows.Next() {
		var v models.Volunteer
		err := rows.Scan(&v.ID, &v.TaskID, &v.UserID, &v.Status, &v.PaymentRequest)
		if err != nil {
			return nil, err
		}
		vols = append(vols, &v)
	}
	return vols, nil
}

func (r *sqliteRepository) GetConfirmedDonationsTotal(ctx context.Context, taskID int64) (int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount_sats), 0) FROM donations WHERE task_id = ? AND status = 'confirmed'", taskID).Scan(&total)
	return total, err
}

func (r *sqliteRepository) CreatePayoutRequest(ctx context.Context, pr *models.PayoutRequest) error {
	query := "INSERT INTO payout_requests (task_id, total_sats, status, psbt, tx_id, created_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)"
	res, err := r.db.ExecContext(ctx, query, pr.TaskID, pr.TotalSats, pr.Status, pr.PSBT, pr.TxID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	pr.ID = id

	var createdAt string
	_ = r.db.QueryRowContext(ctx, "SELECT created_at FROM payout_requests WHERE id = ?", id).Scan(&createdAt)
	pr.CreatedAt = createdAt
	return nil
}

func (r *sqliteRepository) GetPayoutRequest(ctx context.Context, id int64) (*models.PayoutRequest, error) {
	var pr models.PayoutRequest
	var psbtHex sql.NullString
	var txId sql.NullString
	query := "SELECT id, task_id, total_sats, status, psbt, tx_id, created_at FROM payout_requests WHERE id = ?"
	err := r.db.QueryRowContext(ctx, query, id).Scan(&pr.ID, &pr.TaskID, &pr.TotalSats, &pr.Status, &psbtHex, &txId, &pr.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("payout request not found")
		}
		return nil, err
	}
	if psbtHex.Valid {
		pr.PSBT = psbtHex.String
	}
	if txId.Valid {
		pr.TxID = txId.String
	}
	return &pr, nil
}

func (r *sqliteRepository) UpdatePayoutRequestStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE payout_requests SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *sqliteRepository) UpdatePayoutRequestStatusAndTxID(ctx context.Context, id int64, status string, txID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE payout_requests SET status = ?, tx_id = ? WHERE id = ?", status, txID, id)
	return err
}

func (r *sqliteRepository) GetKeyholderByUserID(ctx context.Context, userID int64) (*models.Keyholder, error) {
	var kh models.Keyholder
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, COALESCE(public_key, '') FROM keyholders WHERE user_id = ?", userID).Scan(&kh.ID, &kh.UserID, &kh.PublicKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user is not a keyholder")
		}
		return nil, err
	}
	if kh.PublicKey == "" {
		// Derive public key deterministically
		seed := fmt.Sprintf("keyholder-seed-%d-pamojabuild", kh.ID)
		hash := sha256.Sum256([]byte(seed))
		privKey, _ := btcec.PrivKeyFromBytes(hash[:])
		kh.PublicKey = hex.EncodeToString(privKey.PubKey().SerializeCompressed())
		_, _ = r.db.ExecContext(ctx, "UPDATE keyholders SET public_key = ? WHERE id = ?", kh.PublicKey, kh.ID)
	}
	return &kh, nil
}

func (r *sqliteRepository) HasKeyholderSigned(ctx context.Context, payoutRequestID, keyholderID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payout_signatures WHERE payout_request_id = ? AND keyholder_id = ?", payoutRequestID, keyholderID).Scan(&count)
	return count > 0, err
}

func (r *sqliteRepository) SavePayoutSignature(ctx context.Context, sig *models.PayoutSignature) error {
	var psbtVal interface{}
	if sig.PSBTPartial != "" {
		psbtVal = sig.PSBTPartial
	} else {
		psbtVal = nil
	}

	query := "INSERT INTO payout_signatures (payout_request_id, keyholder_id, action, psbt_partial, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)"
	res, err := r.db.ExecContext(ctx, query, sig.PayoutRequestID, sig.KeyholderID, sig.Action, psbtVal)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	sig.ID = id
	return nil
}

func (r *sqliteRepository) GetPayoutSignatures(ctx context.Context, payoutRequestID int64) ([]*models.PayoutSignature, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, payout_request_id, keyholder_id, action, COALESCE(psbt_partial, '') FROM payout_signatures WHERE payout_request_id = ?", payoutRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sigs []*models.PayoutSignature
	for rows.Next() {
		var sig models.PayoutSignature
		err := rows.Scan(&sig.ID, &sig.PayoutRequestID, &sig.KeyholderID, &sig.Action, &sig.PSBTPartial)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, &sig)
	}
	return sigs, nil
}

func (r *sqliteRepository) UpdateTaskStatus(ctx context.Context, taskID int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE tasks SET status = ? WHERE id = ?", status, taskID)
	return err
}

func (r *sqliteRepository) UpdateVolunteersStatus(ctx context.Context, taskID int64, oldStatus, newStatus string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE volunteers SET status = ? WHERE task_id = ? AND status = ?", newStatus, taskID, oldStatus)
	return err
}

func (r *sqliteRepository) UpdateVolunteerStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE volunteers SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *sqliteRepository) GetStalePayoutRequests(ctx context.Context, beforeTime string) ([]*models.PayoutRequest, error) {
	query := "SELECT id, task_id, total_sats, status, psbt, tx_id, created_at FROM payout_requests WHERE status = 'pending' AND created_at < ?"
	rows, err := r.db.QueryContext(ctx, query, beforeTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*models.PayoutRequest
	for rows.Next() {
		var pr models.PayoutRequest
		var psbtHex sql.NullString
		var txId sql.NullString
		err := rows.Scan(&pr.ID, &pr.TaskID, &pr.TotalSats, &pr.Status, &psbtHex, &txId, &pr.CreatedAt)
		if err != nil {
			return nil, err
		}
		if psbtHex.Valid {
			pr.PSBT = psbtHex.String
		}
		if txId.Valid {
			pr.TxID = txId.String
		}
		prs = append(prs, &pr)
	}
	return prs, nil
}
