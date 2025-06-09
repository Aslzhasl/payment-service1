package repository

import (
	"context"
	"time"
)

// PaymentMethod описывает запись из таблицы payment_methods
type PaymentMethod struct {
	UserID     string    `db:"user_id"` // ← добавьте это поле
	StripePMID string    `db:"stripe_pm_id"`
	Brand      string    `db:"card_brand"`
	Last4      string    `db:"card_last4"`
	ExpMonth   int       `db:"exp_month"`
	ExpYear    int       `db:"exp_year"`
	CreatedAt  time.Time `db:"created_at"`
}

// PaymentMethodRepo описывает операции над saved cards

type PaymentMethodRepo interface {
	SavePaymentMethod(ctx context.Context, pm PaymentMethod) error
	ListPaymentMethods(ctx context.Context, userID string) ([]PaymentMethod, error)
}
