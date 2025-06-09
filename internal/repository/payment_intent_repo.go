package repository

import "context"

type PaymentIntent struct {
	StripePIID string
	BookingID  string
	UserID     string
	Amount     int64
	Currency   string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

// PaymentIntentRepo описывает операции над payment_intents

type PaymentIntentRepo interface {
	CreatePaymentIntent(ctx context.Context, pi PaymentIntent) error
	UpdatePaymentIntentStatus(ctx context.Context, stripePIID, status string) error
	GetPaymentIntentByID(ctx context.Context, stripePIID string) (PaymentIntent, error)
}
