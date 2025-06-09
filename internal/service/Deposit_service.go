package service

import (
	"Payment-service/internal/repository"
	"Payment-service/internal/stripeadapter"
	"context"
)

type DepositService interface {
	// AuthorizeDeposit ставит hold и сохраняет в deposits
	AuthorizeDeposit(ctx context.Context, customerID, userID, bookingID, listingID, currency string, amount int64) (clientSecret, depositID string, err error)
	// CaptureDeposit захватывает hold (списание) и обновляет статус
	CaptureDeposit(ctx context.Context, depositID string) error
	// RefundDeposit отменяет hold и обновляет статус
	RefundDeposit(ctx context.Context, depositID string) error
}

type depositService struct {
	repo   repository.DepositRepo
	stripe *stripeadapter.Client
}

func NewDepositService(repo repository.DepositRepo, stripe *stripeadapter.Client) DepositService {
	return &depositService{repo: repo, stripe: stripe}
}

func (s *depositService) AuthorizeDeposit(ctx context.Context, customerID, userID, bookingID, listingID, currency string, amount int64) (string, string, error) {
	pi, err := s.stripe.CreatePaymentIntent(ctx, customerID, amount, currency, bookingID, userID, listingID)
	if err != nil {
		return "", "", err
	}
	d := repository.Deposit{
		StripePIID: pi.ID,
		BookingID:  bookingID,
		ListingID:  listingID,
		UserID:     userID,
		Amount:     amount,
		Currency:   currency,
		Status:     string(pi.Status),
	}
	if err := s.repo.CreateDeposit(ctx, d); err != nil {
		return "", "", err
	}
	return pi.ClientSecret, pi.ID, nil
}

func (s *depositService) CaptureDeposit(ctx context.Context, depositID string) error {
	pi, err := s.stripe.CapturePaymentIntent(ctx, depositID)
	if err != nil {
		return err
	}
	return s.repo.UpdateDepositStatus(ctx, pi.ID, string(pi.Status))
}

func (s *depositService) RefundDeposit(ctx context.Context, depositID string) error {
	pi, err := s.stripe.CancelPaymentIntent(ctx, depositID)
	if err != nil {
		return err
	}
	return s.repo.UpdateDepositStatus(ctx, pi.ID, string(pi.Status))
}
