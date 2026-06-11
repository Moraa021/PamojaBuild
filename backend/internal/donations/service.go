package donations

import (
	"context"
	"errors"
	"fmt"

	"PamojaBuild/internal/lightning"
	"PamojaBuild/internal/models"
)

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidAmount = errors.New("invalid donation amount")
)

type Service struct {
	repo            Repository
	lightningClient *lightning.Client
}

func NewService(repo Repository, lightningClient *lightning.Client) *Service {
	return &Service{
		repo:            repo,
		lightningClient: lightningClient,
	}
}

func (s *Service) Donate(ctx context.Context, taskID int64, donorID int64, amountSats int64, isAnonymous bool) (*models.Donation, error) {
	if amountSats <= 0 {
		return nil, ErrInvalidAmount
	}

	// 1. Verify task exists
	exists, err := s.repo.TaskExists(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to check task existence: %w", err)
	}
	if !exists {
		return nil, ErrTaskNotFound
	}

	// 2. Generate Lightning invoice
	memo := fmt.Sprintf("Donation to task #%d", taskID)
	paymentRequest, paymentHash, err := s.lightningClient.CreateInvoice(amountSats, memo)
	if err != nil {
		return nil, fmt.Errorf("failed to create lightning invoice: %w", err)
	}

	// 3. Create donation record
	d := &models.Donation{
		TaskID:         taskID,
		DonorID:        donorID,
		AmountSats:     amountSats,
		PaymentHash:    paymentHash,
		PaymentRequest: paymentRequest,
		IsAnonymous:    isAnonymous,
		Status:         string(models.DonationStatusPending),
	}

	if err := s.repo.CreateDonation(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to save donation: %w", err)
	}

	return d, nil
}

func (s *Service) GetTotalConfirmed(ctx context.Context, taskID int64) (int64, error) {
	// 1. Verify task exists
	exists, err := s.repo.TaskExists(ctx, taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to check task existence: %w", err)
	}
	if !exists {
		return 0, ErrTaskNotFound
	}

	// 2. Find pending donations and poll status
	pending, err := s.repo.GetPendingDonations(ctx, taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending donations: %w", err)
	}

	for _, d := range pending {
		settled, err := s.lightningClient.CheckPaymentStatus(d.PaymentHash)
		if err != nil {
			// Log error but continue checking other donations
			// We don't want a single node/invoice failure to block everything
			fmt.Printf("Error checking payment status for hash %s: %v\n", d.PaymentHash, err)
			continue
		}
		if settled {
			err = s.repo.UpdateDonationStatus(ctx, d.ID, string(models.DonationStatusConfirmed))
			if err != nil {
				fmt.Printf("Error updating donation status to confirmed for id %d: %v\n", d.ID, err)
			}
		}
	}

	// 3. Calculate sum of confirmed donations
	total, err := s.repo.GetConfirmedTotal(ctx, taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate confirmed total: %w", err)
	}

	return total, nil
}
