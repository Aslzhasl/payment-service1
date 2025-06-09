package service

import (
	"context"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"

	"Payment-service/internal/repository"
	"Payment-service/internal/stripeadapter"
)

// PaymentService defines logic for PaymentIntents: authorize, capture, cancel.
type PaymentService interface {
	Authorize(ctx context.Context, customerID, userID, bookingID, currency string, amount int64, paymentMethod string) (clientSecret string, paymentIntentID string, err error)
	Capture(ctx context.Context, paymentIntentID string) error
	Cancel(ctx context.Context, paymentIntentID string) error
}

// paymentService is a concrete implementation of PaymentService.
type paymentService struct {
	repo   repository.PaymentIntentRepo
	stripe *stripeadapter.Client
}

func NewPaymentService(repo repository.PaymentIntentRepo, client *stripeadapter.Client) PaymentService {
	return &paymentService{repo: repo, stripe: client}
}

// CreatePaymentIntentRequest holds all input fields for creating an intent.
type CreatePaymentIntentRequest struct {
	UserID        string
	CustomerID    string
	BookingID     string
	Amount        int64
	Currency      string
	PaymentMethod string // optional
}

// CreatePaymentIntent returns a Stripe PaymentIntent, confirmed if off_session
func (s *paymentService) CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		Customer: stripe.String(req.CustomerID),
	}
	params.AddMetadata("user_id", req.UserID)
	params.AddMetadata("booking_id", req.BookingID)

	if req.PaymentMethod != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethod)
		params.Confirm = stripe.Bool(true)
		params.OffSession = stripe.Bool(true)
	}

	return paymentintent.New(params)
}

// Authorize creates a PaymentIntent (with or without saved card) and stores it.
func (s *paymentService) Authorize(ctx context.Context, customerID, userID, bookingID, currency string, amount int64, paymentMethod string) (string, string, error) {
	pi, err := s.CreatePaymentIntent(ctx, CreatePaymentIntentRequest{
		UserID:        userID,
		CustomerID:    customerID,
		BookingID:     bookingID,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		return "", "", err
	}

	intent := repository.PaymentIntent{
		StripePIID: pi.ID,
		BookingID:  bookingID,
		UserID:     userID,
		Amount:     amount,
		Currency:   currency,
		Status:     string(pi.Status),
	}

	if err := s.repo.CreatePaymentIntent(ctx, intent); err != nil {
		return "", "", err
	}

	return pi.ClientSecret, pi.ID, nil
}

// Capture charges a previously authorized PaymentIntent.
func (s *paymentService) Capture(ctx context.Context, paymentIntentID string) error {
	pi, err := s.stripe.CapturePaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	return s.repo.UpdatePaymentIntentStatus(ctx, pi.ID, string(pi.Status))
}

// Cancel releases a hold without charging.
func (s *paymentService) Cancel(ctx context.Context, paymentIntentID string) error {
	pi, err := s.stripe.CancelPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return err
	}
	return s.repo.UpdatePaymentIntentStatus(ctx, pi.ID, string(pi.Status))
}
